// /home/krylon/go/src/github.com/blicero/chili/probe/probe.go
// -*- mode: go; coding: utf-8; -*-
// Created on 09. 07. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-07-10 14:02:40 krylon>

// Package probe implements the detailed interrogration of Devices
// the Scanner has discovered.
package probe

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/blicero/chili/common"
	"github.com/blicero/chili/control"
	"github.com/blicero/chili/database"
	"github.com/blicero/chili/logdomain"
	"github.com/blicero/chili/model"
	"golang.org/x/crypto/ssh"
)

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
	cfg         *ssh.ClientConfig
	clients     map[int64]*ssh.Client
}

// Create returns a fresh Probe.
func Create(cnt int, userName string, keyPath ...string) (*Probe, error) {
	var (
		err error
		p   = &Probe{
			parCnt:   cnt,
			clients:  make(map[int64]*ssh.Client),
			interval: common.DefaultProbeInterval,
		}
	)

	if p.log, err = common.GetLogger(logdomain.Probe); err != nil {
		return nil, err
	} else if err = p.initConfig(userName, keyPath...); err != nil {
		return nil, err
	} else if p.pool, err = database.NewPool(max(cnt-2, 1)); err != nil {
		p.log.Printf("[CRITICAL] Cannot open database connection pool: %s\n",
			err.Error())
		return nil, err
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
	//     InsecureIgnoreHostKey in production code, which makes sense for obvious reasons.
	//     But I intend to only run this application on my local network, where I own and
	//     administer all the devices.
	//     But if anyone ever intends to use this code (or parts of it) for any other purpose,
	//     please, PLEASE rectify this!!! You have been warned.
	p.cfg = &ssh.ClientConfig{
		User: userName,
		// Auth: make([]ssh.AuthMethod, 0, len(keys)),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(keys...),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
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
	p.log.Println("[TRACE] Probe main loop starting up")
	defer p.log.Println("[TRACE] Probe main loop quitting")

	var heartbeat = time.NewTicker(common.ActiveTimeout)
	defer heartbeat.Stop()

	var probeTicker = time.NewTicker(p.interval)
	defer probeTicker.Stop()

	for p.IsActive() {
		select {
		case <-heartbeat.C:
			continue
		case <-probeTicker.C:
			go p.probeDevices()
		case cmd := <-p.CmdQ:
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
	p.log.Println("[TRACE] Probing devices")
	defer p.log.Println("[TRACE] Done probing")

	if swapped := p.scanRunning.CompareAndSwap(false, true); !swapped {
		p.log.Println("[ERROR] It appears a probe is already running. Bye.")
		return
	}

	defer p.scanRunning.Store(false)

	var (
		err     error
		db      *database.Database
		devices []*model.Device
		wg      sync.WaitGroup
		devQ    chan *model.Device
		ticker  = time.NewTicker(common.ActiveTimeout)
	)

	defer ticker.Stop()

	devQ = make(chan *model.Device)
	defer close(devQ) // nolint: errcheck

	for i := range p.parCnt {
		wg.Go(func() { p.probeWorker(i+1, devQ) })
	}

	db = p.pool.Get()
	defer p.pool.Put(db)

	if devices, err = db.DeviceGetAll(); err != nil {
		p.log.Printf("[ERROR] Failed to load all Devices: %s\n",
			err.Error())
		return
	}

	for _, dev := range devices {
	SELECT:
		select {
		case <-ticker.C:
			if p.active.Load() && p.scanRunning.Load() {
				goto SELECT
			}
		case devQ <- dev:
			continue
		}
	}
} // func (p *Probe) probeDevices()

func (p *Probe) probeWorker(id int, devQ <-chan *model.Device) {
	p.log.Printf("[TRACE] Probe worker %02d starting up", id)
	defer p.log.Printf("[TRACE] Probe worker %02d quitting", id)

	for dev := range devQ {
		p.probeOneDevice(dev)
	}
} // func (p *Probe) probeWorker(id int, devQ <-chan *model.Device)

func (p *Probe) probeOneDevice(dev *model.Device) {
	p.log.Printf("[TRACE] Probing %s (%s)\n",
		dev.Name,
		dev.Addr)

	var (
		err    error
		online bool
		db     *database.Database
	)

	if online = p.pingDevice(dev); !online {
		p.log.Printf("[TRACE] Device %s is not online\n",
			dev.Name)
		return
	}

	db = p.pool.Get()
	defer p.pool.Put(db)

	if dev.OS == "" {
		var sysname string
		if sysname, err = p.QueryOS(dev, portSSH); err != nil {
			p.log.Printf("[ERROR] Failed to query device operating system on %s: %s\n",
				dev.Name,
				err.Error())
			return
		} else if err = db.DeviceSetOS(dev, sysname); err != nil {
			p.log.Printf("[ERROR] Failed to set OS for %s to %s: %s\n",
				dev.Name,
				sysname,
				err.Error())
			return
		}
	}
} // func (p *Probe) probeDevice(dev *model.Device)
