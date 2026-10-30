// /home/krylon/go/src/github.com/blicero/chili/probe/probe.go
// -*- mode: go; coding: utf-8; -*-
// Created on 09. 07. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-10-30 11:47:47 krylon>

// Package probe implements the detailed interrogration of Devices
// the Scanner has discovered.
package probe

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/blicero/chili/common"
	"github.com/blicero/chili/control"
	"github.com/blicero/chili/database"
	"github.com/blicero/chili/logdomain"
	"github.com/blicero/chili/model"
	"github.com/blicero/chili/model/attribute"
	"github.com/blicero/chili/model/device"
	"github.com/blicero/krylib"
	"golang.org/x/crypto/ssh"
)

// Schedule defines how often we query each attribute on each Device.
//
// TODO Adjust intervals, make them configurable.
var Schedule = map[attribute.ID]time.Duration{
	attribute.Updates:   time.Second * 3600,
	attribute.DiskSpace: time.Second * 900,
	attribute.Uptime:    time.Second * 60,
	attribute.Packages:  time.Second * 86400,
	attribute.SNMP:      time.Minute * 5,
	attribute.Services:  time.Minute * 10,
	attribute.DMI:       time.Hour * 168, // DMI data is unlikely to change
}

func portFor(a attribute.ID) int {
	switch a {
	case attribute.SNMP:
		return portSNMP
	default:
		return portSSH
	}
} // func portFor(a attribute.ID) int

type probeFn func(id int, dev *model.Device, port int) error

// Probe queries Devices about their OS, installed software, hardware,
// pending updates, etc...
type Probe struct {
	log         *log.Logger
	lock        sync.RWMutex
	CmdQ        chan control.Message
	interval    time.Duration
	pool        *database.Pool
	parCnt      int
	active      atomic.Bool
	scanRunning atomic.Bool
	pending     atomic.Int32
	cfg         *ssh.ClientConfig
	clients     map[int64]*ssh.Client
	privCmd     map[int64]string
	probes      map[attribute.ID]probeFn
	pmap        *krylib.SyncMap[int, string]
}

// Create returns a fresh Probe.
func Create(cnt int, userName string, keyPath ...string) (*Probe, error) {
	var (
		err error
		p   = &Probe{
			parCnt:   cnt,
			clients:  make(map[int64]*ssh.Client),
			interval: common.DefaultProbeInterval,
			privCmd:  make(map[int64]string),
			pmap:     krylib.NewSyncMap[int, string](),
		}
	)

	if cnt < 1 {
		fmt.Fprintf(
			os.Stderr,
			"Trying to create a Probe with %d workers? That's not how that works.\n",
			cnt)

		return nil, nil
	} else if p.log, err = common.GetLogger(logdomain.Probe); err != nil {
		return nil, err
	} else if err = p.initConfig(userName, keyPath...); err != nil {
		return nil, err
	} else if p.pool, err = database.NewPool(cnt + 2); err != nil {
		p.log.Printf("[CRITICAL] Cannot open database connection pool: %s\n",
			err.Error())
		return nil, err
	}

	p.probes = map[attribute.ID]probeFn{
		attribute.Updates:   p.qUpdate,
		attribute.DiskSpace: p.qDiskSpace,
		attribute.Uptime:    p.qUptime,
		attribute.Packages:  p.queryInstalled,
		attribute.SNMP:      p.qSNMP,
		attribute.Services:  p.qSvc,
		attribute.DMI:       p.qDMI,
	}

	p.CmdQ = make(chan control.Message, 1)
	return p, nil
} // func Create() (*Probe, error)

func (p *Probe) initConfig(userName string, keyPath ...string) error {
	var (
		err    error
		keyRaw []byte
		signer ssh.Signer
		keys   = make([]ssh.Signer, 0, len(keyPath))
	)

	for _, path := range keyPath {
		var (
			fh       *os.File
			keyFiles []string
		)

		p.log.Printf("[DEBUG] Trying to import %s\n", path)
		if fh, err = os.Open(path); err != nil {
			p.log.Printf("[ERROR] Cannot open %s: %s\n",
				path,
				err.Error())
			continue
		}

		defer fh.Close() // nolint: errcheck

		if keyFiles, err = fh.Readdirnames(-1); err != nil {
			p.log.Printf("[ERROR] Cannot read files in directory %s: %s\n",
				path,
				err.Error())
			continue
		}

		for _, file := range keyFiles {
			if !strings.HasPrefix(file, "id_") || strings.HasSuffix(file, ".pub") {
				continue
			}

			var fullPath = filepath.Join(path, file)

			p.log.Printf("[DEBUG] Import SSH key %s\n", fullPath)

			if keyRaw, err = os.ReadFile(fullPath); err != nil {
				var ex = fmt.Errorf("failed to read SSH key from %s: %w",
					fullPath,
					err)
				p.log.Printf("[ERROR] %s\n", ex.Error())
				return ex
			} else if signer, err = ssh.ParsePrivateKey(keyRaw); err != nil {
				var ex = fmt.Errorf("failed to parse SSH key: %w",
					err)
				p.log.Printf("[ERROR] %s\n", ex.Error())
				return ex
			} else if signer == nil {
				var ex = fmt.Errorf("ParsePrivateKey did not return an error, but signer is nil!\nKey File: %s\nKey: %s",
					fullPath,
					keyRaw)
				p.log.Printf("[ERROR] %s\n",
					ex.Error())
				return ex
			}
			keys = append(keys, signer)
		}
	}

	// XXX The documentation for the ssh package says very explicitly to NOT use
	//     InsecureIgnoreHostKey in production code, which makes sense
	//     for obvious reasons.
	//     But I intend to only run this application on my local network,
	//     where I own and administer all the devices.
	//     But if anyone ever intends to use this code
	//     (or parts of it) for any other purpose, please, PLEASE rectify
	//     this!!! You have been warned.
	p.cfg = &ssh.ClientConfig{
		User: userName,
		// Auth: make([]ssh.AuthMethod, 0, len(keys)),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(keys...),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 300,
	}

	return nil
} // func (p *Probe) initConfig(keyPath string) error

// IsActive returns the Probe's active flag.
func (p *Probe) IsActive() bool {
	return p.active.Load()
} // func (p *Probe) IsActive() bool

// Start starts the Probe's main loop.
func (p *Probe) Start() {
	var swapped bool

	if swapped = p.active.CompareAndSwap(false, true); swapped {
		go p.mainLoop()
	} else {
		p.log.Println("[INFO] Probe main loop is already active.")
	}
} // func (p *Probe) Start()

// Stop tells the Probe's main loop and workers to quit.
func (p *Probe) Stop() {
	p.scanRunning.Store(false)
	if swapped := p.active.CompareAndSwap(true, false); !swapped {
		p.log.Println("[INFO] Probe main loop was not active.")
	}
} // func (p *Probe) Stop()

func (p *Probe) mainLoop() {
	var cnt uint64

	p.log.Println("[TRACE] Probe main loop starting up")
	defer p.log.Println("[TRACE] Probe main loop quitting")

	var heartbeat = time.NewTicker(common.ActiveTimeout)
	defer heartbeat.Stop()

	var probeTicker = time.NewTicker(p.interval)
	defer probeTicker.Stop()

	go p.probeDevices()

	for p.IsActive() {
		select {
		case <-heartbeat.C:
			if cnt++; cnt%5 == 0 && p.scanRunning.Load() {
				// FIXME This debugging code is not pretty! I should
				//       remove it as soon as possible.
				var (
					sb      strings.Builder
					keys    []int
					pending = p.pmap.Map()
				)

				keys = make([]int, 0, len(pending))

				for k := range pending {
					keys = append(keys, k)
				}

				slices.Sort(keys)

				for _, key := range keys {
					var w = key
					var d = pending[w]
					fmt.Fprintf(
						&sb,
						"\t%02d => %s\n",
						w,
						d)
				}

				p.log.Printf("[DEBUG] Pending probes: %d\n%s\n",
					p.pending.Load(),
					sb.String())
			}
			continue
		case <-probeTicker.C:
			if !p.scanRunning.Load() {
				go p.probeDevices()
			}
		case cmd := <-p.CmdQ:
			p.log.Printf("[DEBUG] Received command %s\n", cmd)
			switch cmd {
			case control.Stop:
				p.Stop()
			case control.Scan:
				go p.probeDevices()
			}
		}
	}
} // func (p *Probe) mainLoop()

func (p *Probe) probeDevices() {
	p.log.Println("[INFO] Probing devices")
	defer p.log.Println("[INFO] Done probing")

	if swapped := p.scanRunning.CompareAndSwap(false, true); !swapped {
		p.log.Println("[ERROR] It appears a probe is already running. Bye.")
		return
	}

	defer p.scanRunning.Store(false)

	var (
		err     error
		devices []*model.Device
		wg      sync.WaitGroup
		devQ    chan *model.Device
		ticker  = time.NewTicker(common.ActiveTimeout)
	)

	defer ticker.Stop()

	if devices, err = p.getAllDevices(); err != nil {
		p.log.Printf("[ERROR] Failed to load all Devices: %s\n",
			err.Error())
		return
	} else if len(devices) == 0 {
		p.log.Printf("[INFO] No Devices were found in database. Bye.\n")
		return
	}

	p.log.Printf("[INFO] About to probe %d Devices using %d workers\n",
		len(devices),
		p.parCnt)

	devQ = make(chan *model.Device)

	for i := range p.parCnt {
		wg.Go(func() { p.probeWorker(i+1, devQ) })
	}

	var indices = rand.Perm(len(devices))

	for _, i := range indices {
		dev := devices[i]

		switch dev.Class {
		case device.Mobile, device.Unknown, device.Entertainment, device.Router:
			p.log.Printf("[INFO] Won't probe %s, it is a %s\n",
				dev.Name,
				dev.Class)
			continue
		}

	DEVICE:
		select {
		case <-ticker.C:
			if !p.active.Load() || !p.scanRunning.Load() {
				p.log.Printf("[INFO] Probe Feeder quitting prematurely\n")
				close(devQ)
				return
			}
			goto DEVICE
		case devQ <- dev:
			p.log.Printf("[TRACE] Enqueueing %s for a Probing\n", dev.Name)
			p.pending.Add(1)
		}
	}

	p.log.Printf("[TRACE] That's enough probing for now!\n")
	close(devQ)
	p.log.Printf("[TRACE] Closed device queue, waiting for workers to finish.\n")
	wg.Wait()
} // func (p *Probe) probeDevices()

func (p *Probe) probeWorker(id int, devQ <-chan *model.Device) {
	p.log.Printf("[TRACE] Probe worker %02d starting up", id)
	defer p.log.Printf("[TRACE] Probe worker %02d quitting", id)

	for dev := range devQ {
		// p.pending.Add(-1)
		p.probeOneDevice(id, dev)
	}
} // func (p *Probe) probeWorker(id int, devQ <-chan *model.Device)

func (p *Probe) probeOneDevice(id int, dev *model.Device) {
	p.log.Printf("[INFO] Probe#%d querying %s (%s)\n",
		id,
		dev.Name,
		dev.Addr)
	defer p.log.Printf("[INFO] Probe#%d finished probing %s (%s)\n",
		id,
		dev.Name,
		dev.Addr)
	defer p.pending.Add(-1)

	p.pmap.Put(id, dev.Name)
	defer p.pmap.Delete(id)

	var (
		err    error
		online bool
	)

	if online = p.pingDevice(dev); !online {
		p.log.Printf("[TRACE] Device %s is not online\n",
			dev.Name)
		return
	}

	p.log.Printf("[TRACE] Probe#%d getting Database from pool\n",
		id)

	if dev.OS == "" {
		var sysname string
		if sysname, err = p.QueryOS(dev, portSSH); err != nil {
			p.log.Printf("[ERROR] Failed to query device operating system on %s: %s\n",
				dev.Name,
				err.Error())
			return
		} else if err = p.setOS(id, dev, sysname); err != nil {
			p.log.Printf("[ERROR] Failed to set OS for %s to %s: %s\n",
				dev.Name,
				sysname,
				err.Error())
			return
		}
	}

	var (
		knownAttr  []*model.Attribute
		lastProbed map[attribute.ID]time.Time
		now        = time.Now().Truncate(time.Second)
	)

	if knownAttr, err = p.getMostRecentAttributes(id, dev); err != nil {
		p.log.Printf("[ERROR] Failed to load attributes for %s: %s\n",
			dev.Name,
			err.Error())
		return
	}

	lastProbed = make(map[attribute.ID]time.Time, len(knownAttr))

	for _, a := range knownAttr {
		lastProbed[a.Type] = a.Timestamp
	}

	// FIXME I should really organize these steps more sensibly, but
	//       to get me going, I will just perform them one after another,
	//       each time.

	for _, att := range attribute.All() {
		if stamp, ok := lastProbed[att]; !ok || stamp.Add(Schedule[att]).Before(now) {
			p.log.Printf("[TRACE] Probe#%d probing %s for %s\n",
				id,
				dev.Name,
				att)

			if err = p.probes[att](id, dev, portFor(att)); err != nil {
				p.log.Printf("[ERROR] Probe#%d failed to probe %s for %s: %s\n",
					id,
					dev.Name,
					att,
					err.Error())
			}
		}
	}

	// if stamp, ok := lastProbed[attribute.Updates]; !ok || stamp.Add(Schedule[attribute.Updates]).Before(now) {
	// 	if err = p.qUpdate(id, dev, portSSH); err != nil {
	// 		p.log.Printf("[ERROR] Probe#%d failed to query updates on %s: %s\n",
	// 			id,
	// 			dev.Name,
	// 			err.Error())
	// 	}
	// }

	// // FIXME Querying for installed software runs into some sort of tar pit.
	// //       At first I thought I had gotten too clever with concurrency, but
	// //       it appears the problem is a timeout?
	// if stamp, ok := lastProbed[attribute.Packages]; !ok || stamp.Add(Schedule[attribute.Packages]).Before(now) {
	// 	if err = p.queryInstalled(id, dev, portSSH); err != nil {
	// 		p.log.Printf("[ERROR] Probe#%d failed to query packages on %s: %s\n",
	// 			id,
	// 			dev.Name,
	// 			err.Error())
	// 	}
	// }

	// if stamp, ok := lastProbed[attribute.Uptime]; !ok || stamp.Add(Schedule[attribute.Uptime]).Before(now) {
	// 	if err = p.qUptime(id, dev, portSSH); err != nil {
	// 		p.log.Printf("[ERROR] Failed to query Uptime on %s: %s\n",
	// 			dev.Name,
	// 			err.Error())
	// 	}
	// }

	// if stamp, ok := lastProbed[attribute.DiskSpace]; !ok || stamp.Add(Schedule[attribute.DiskSpace]).Before(now) {
	// 	if err = p.qDiskSpace(id, dev, portSSH); err != nil {
	// 		p.log.Printf("[ERROR] Probe#%d failed to query free disk space on %s: %s\n",
	// 			id,
	// 			dev.Name,
	// 			err.Error())
	// 	}
	// }

	// if stamp, ok := lastProbed[attribute.SNMP]; !ok || stamp.Add(Schedule[attribute.SNMP]).Before(now) {
	// 	if err = p.qSNMP(id, dev, portSNMP); err != nil {
	// 		p.log.Printf("[ERROR] Probe#%d failed to query %s via SNMP: %s\n",
	// 			id,
	// 			dev.Name,
	// 			err.Error())
	// 	}
	// }

	// if stamp, ok := lastProbed[attribute.Services]; !ok || stamp.Add(Schedule[attribute.Services]).Before(now) {
	// 	if err = p.qSvc(id, dev, portSSH); err != nil {
	// 		p.log.Printf("[ERROR] Probe#%d failed to query Services on %s: %s\n",
	// 			id,
	// 			dev.Name,
	// 			err.Error())
	// 	}
	// }

	// if stamp, ok := lastProbed[attribute.DMI]; !ok || stamp.Add(Schedule[attribute.DMI]).Before(now) {
	// 	if err = p.qDMI(id, dev, portSSH); err != nil {
	// 		p.log.Printf("[ERROR] Probe#%d failed to query DMI on %s: %s\n",
	// 			id,
	// 			dev.Name,
	// 			err.Error())
	// 	}
	// }

	p.log.Printf("[TRACE] Probe#%d finished probing %s\n",
		id,
		dev.Name)
} // func (p *Probe) probeOneDevice(id int, dev *model.Device)

func (p *Probe) getAllDevices() ([]*model.Device, error) {
	var (
		err  error
		db   *database.Database
		devs []*model.Device
	)

	db = p.pool.Get()
	defer p.pool.Put(db)

	if devs, err = db.DeviceGetAll(); err != nil {
		p.log.Printf("[ERROR] Failed to load all Devices: %s\n",
			err.Error())
		return nil, err
	}

	return devs, nil
} // func (p *Probe) getAllDevices() ([]*model.Device, error)

func (p *Probe) getMostRecentAttributes(id int, dev *model.Device) ([]*model.Attribute, error) {
	var (
		err  error
		db   *database.Database
		attr []*model.Attribute
	)

	db = p.pool.Get()
	defer p.pool.Put(db)

	if attr, err = db.AttributeGetMostRecent(dev); err != nil {
		p.log.Printf("[ERROR] Probe#%d failed to load attributes for %s: %s\n",
			id,
			dev.Name,
			err.Error())
		return nil, err
	}

	return attr, nil
} // func (p *Probe) getMostRecentAttributes(id int, dev *model.Device) ([]*model.Attribute, error)

func (p *Probe) qUptime(id int, d *model.Device, port int) error {
	var (
		err  error
		db   *database.Database
		up   *model.Uptime
		attr *model.Attribute
	)

	if up, err = p.QueryUptime(d, port); err != nil {
		p.log.Printf("[ERROR] Probe#%d failed to query Uptime on %s: %s\n",
			id,
			d.Name,
			err.Error())
		return err
	}

	attr = &model.Attribute{
		DevID:     d.ID,
		Timestamp: time.Now().Truncate(time.Second),
		Type:      attribute.Uptime,
		Value:     up,
	}

	db = p.pool.Get()
	defer p.pool.Put(db)

	if err = db.AttributeAdd(attr); err != nil {
		p.log.Printf("[ERROR] Failed to add Uptime of %s to database: %s\n",
			d.Name,
			err.Error())
		return err
	}

	return nil
} // func (p *Probe) qUptime(d *model.Device, port int) error

func (p *Probe) qDiskSpace(id int, d *model.Device, port int) error {
	var (
		err  error
		db   *database.Database
		ds   int64
		attr *model.Attribute
	)

	if ds, err = p.QueryDiskFree(d, port); err != nil {
		p.log.Printf("[ERROR] Probe#%d failed to query free space on %s: %s\n",
			id,
			d.Name,
			err.Error())
		return err
	}

	attr = &model.Attribute{
		DevID:     d.ID,
		Timestamp: time.Now().Truncate(time.Second),
		Type:      attribute.DiskSpace,
		Value:     model.DiskSpace(ds),
	}

	db = p.pool.Get()
	defer p.pool.Put(db)

	if err = db.AttributeAdd(attr); err != nil {
		p.log.Printf("[ERROR] Failed to add free space for %s to database: %s\n",
			d.Name,
			err.Error())
		return err
	}

	return nil
} // func (p *Probe) qDiskSpace(d *model.Device, port int) error

func (p *Probe) qSNMP(id int, d *model.Device, port int) error {
	var (
		err  error
		db   *database.Database
		val  model.SNMPInfo
		attr *model.Attribute
	)

	p.log.Printf("[TRACE] Probe#%d about to query SNMP data from %s\n",
		id,
		d.Name)

	if val, err = p.QuerySNMP(d, port); err != nil {
		p.log.Printf("[ERROR] Probe#%d failed to query %s via SNMP: %s\n",
			id,
			d.Name,
			err.Error())
	}

	attr = &model.Attribute{
		DevID:     d.ID,
		Timestamp: time.Now().Truncate(time.Second),
		Type:      attribute.SNMP,
		Value:     val,
	}

	db = p.pool.Get()
	defer p.pool.Put(db)

	if err = db.AttributeAdd(attr); err != nil {
		p.log.Printf("[ERROR] Failed to add SNMP data for %s to database: %s\n",
			d.Name,
			err.Error())
		return err
	}

	return nil
} // func (p *Probe) qSNMP(id int, d *model.Device, port int) error

func (p *Probe) qDMI(id int, d *model.Device, port int) error {
	var (
		err  error
		db   *database.Database
		val  *model.DMI
		attr *model.Attribute
	)

	switch d.Class {
	case device.VM, device.Jail, device.Router, device.Entertainment:
		// p.log.Printf("[DEBUG] %s is a %s, so no DMI\n",
		// 	dev.Name,
		// 	dev.Class)
		return nil
	default:
		p.log.Printf("[TRACE] Probe#%d about to query DMI data from %s\n",
			id,
			d.Name)
	}

	if val, err = p.QueryDMI(d, port); err != nil {
		p.log.Printf("[ERROR] Probe#%d failed to query DMI on %s: %s\n",
			id,
			d.Name,
			err.Error())
	}

	attr = &model.Attribute{
		DevID:     d.ID,
		Timestamp: time.Now().Truncate(time.Second),
		Type:      attribute.DMI,
		Value:     val,
	}

	db = p.pool.Get()
	defer p.pool.Put(db)

	if err = db.AttributeAdd(attr); err != nil {
		p.log.Printf("[ERROR] Failed to add DMI data for %s to database: %s\n",
			d.Name,
			err.Error())
		return err
	}

	return nil
} // func (p *Probe) qDMI(id int, d *model.Device, port int) error

func (p *Probe) qSvc(id int, d *model.Device, port int) error {
	var (
		err  error
		db   *database.Database
		val  *model.Services
		attr *model.Attribute
	)

	p.log.Printf("[TRACE] Probe#%d about to query services on %s\n",
		id,
		d.Name)

	if val, err = p.QueryServices(d, port); err != nil {
		p.log.Printf("[ERROR] Probe#%d failed to query services on %s: %s\n",
			id,
			d.Name,
			err.Error())
	}

	attr = &model.Attribute{
		DevID:     d.ID,
		Timestamp: time.Now().Truncate(time.Second),
		Type:      attribute.Services,
		Value:     val,
	}

	db = p.pool.Get()
	defer p.pool.Put(db)

	if err = db.AttributeAdd(attr); err != nil {
		p.log.Printf("[ERROR] Failed to add Uptime of %s to database: %s\n",
			d.Name,
			err.Error())
		return err
	}

	return nil
} // func (p *Probe) qSvc(id int, d *model.Device, port int) error

func (p *Probe) qUpdate(id int, d *model.Device, port int) error {
	var (
		err  error
		db   *database.Database
		upd  []string
		attr *model.Attribute
	)

	p.log.Printf("[TRACE] Probe#%d about to query pending updates on %s\n",
		id,
		d.Name)

	if upd, err = p.QueryUpdates(d, port); err != nil {
		p.log.Printf("[ERROR] Probe#%d failed to query updates on %s: %s\n",
			id,
			d.Name,
			err.Error())
	}

	attr = &model.Attribute{
		DevID:     d.ID,
		Timestamp: time.Now().Truncate(time.Second),
		Type:      attribute.Updates,
		Value:     model.Updates(upd),
	}

	db = p.pool.Get()
	defer p.pool.Put(db)

	if err = db.AttributeAdd(attr); err != nil {
		p.log.Printf("[ERROR] Failed to add Updates on %s to database: %s\n",
			d.Name,
			err.Error())
		return err
	}

	return nil
} // func (p *Probe) qUpdate(id int, d *model.Device, port int) error

func (p *Probe) setOS(id int, d *model.Device, osname string) error {
	var (
		err error
		db  *database.Database
	)

	db = p.pool.Get()
	defer p.pool.Put(db)

	if err = db.DeviceSetOS(d, osname); err != nil {
		p.log.Printf("[ERROR] Probe#%d failed to set OS on %s to %s: %s\n",
			id,
			d.Name,
			osname,
			err.Error())
		return err
	}

	d.OS = osname
	return nil
} // func (p *Probe) setOS(id int, d *model.Device, osname string) error

const (
	cmdSudo = "sudo"
	cmdDoas = "doas"
)

// Some systems use sudo(1) to execute a command with elevated privileges,
// some use doas(1). There are others, too, that I know nothing about.
// We attempt to figure out which command to use for a given Device.
// NB sudo is basically available everywhere as a package, doas is
// part of the base system on OpenBSD, but it is also available on other
// platforms. So some Devices might have BOTH sudo(1) and doas(1), and even
// other similary tools (systemd has its own)
func (p *Probe) getPrivCmd(dev *model.Device) string {
	switch dev.OS {
	case "Debian GNU/Linux", "Raspbian GNU/Linux":
		fallthrough
	case "openSUSE Leap", "openSUSE Tumbleweed":
		return cmdSudo
	case "FreeBSD", "OpenBSD", "NetBSD", "Arch Linux":
		return cmdDoas
	default:
		return cmdSudo
	}
} // func (p *Probe) getPrivCmd(dev *model.Device) string

func trim(s string) string {
	return strings.Trim(s, "\n\t ")
}
