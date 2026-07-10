// /home/krylon/go/src/github.com/blicero/chili/scanner/scanner.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 01. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-07-10 10:50:18 krylon>

// Package scanner implements traversing a range of IP addresses
// and probing which of those correspond to live devices.
package scanner

// NB: IPv6 is NOT supported currently.
import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/blicero/chili/common"
	"github.com/blicero/chili/control"
	"github.com/blicero/chili/database"
	"github.com/blicero/chili/logdomain"
	"github.com/blicero/chili/model"
	probing "github.com/prometheus-community/pro-bing"
)

// // http://play.golang.org/p/m8TNTtygK0
// // nolint: unused
// func inc(ip net.IP) {
// 	for j := len(ip) - 1; j >= 0; j-- {
// 		ip[j]++
// 		if ip[j] > 0 {
// 			break
// 		}
// 	}
// }

const (
	pingCount    = 4
	pingTimeout  = time.Second * 10
	pingInterval = time.Millisecond * 250
	scanInterval = time.Second * 30 // XXX Set to more reasonable value after testing/debugging.
)

type scanTarget struct {
	net  *model.Network
	addr net.IP
}

type Scanner struct {
	log *log.Logger
	// scanLock    sync.RWMutex
	active      atomic.Bool
	scanRunning atomic.Bool
	workerCnt   int
	dbPool      *database.Pool
	CmdQ        chan control.Message
}

// New creates a new Scanner with the given number of worker goroutines.
func New(wcnt int) (*Scanner, error) {
	if wcnt < 1 {
		return nil, fmt.Errorf("number of workers must be a positive integer, not %d", wcnt)
	}

	var (
		err error
		sc  = &Scanner{workerCnt: wcnt}
	)

	if sc.log, err = common.GetLogger(logdomain.Scanner); err != nil {
		return nil, err
	} else if sc.dbPool, err = database.NewPool(min(wcnt>>1, 2)); err != nil {
		sc.log.Printf("[ERROR] Failed to create database pool: %s\n",
			err.Error())
		return nil, err
	}

	sc.CmdQ = make(chan control.Message, 2)

	return sc, nil
} // func New(wcnt int) (*Scanner, error)

// IsActive returns the state of the Scanner's active flag.
func (sc *Scanner) IsActive() bool {
	return sc.active.Load()
} // func (sc *Scanner) IsActive() bool

// Start starts the Scanners worker goroutines
func (sc *Scanner) Start() error {
	sc.active.Store(true)

	go sc.mainLoop()

	return nil
} // func (sc *Scanner) Start() error

func (sc *Scanner) mainLoop() {
	var ticker = time.NewTicker(scanInterval)
	defer ticker.Stop()

	sc.log.Println("[TRACE] Scanner mainloop initiated")
	defer sc.log.Println("[TRACE] Scanner mainloop finished")

	for sc.active.Load() {
		select {
		case <-ticker.C:
			sc.runScan()
		case cmd := <-sc.CmdQ:
			switch cmd {
			case control.Scan:
				go sc.runScan()
			case control.Stop:
				sc.log.Println("[DEBUG] Someone told me to stop. Bye.")
				sc.active.Store(false)
				return
			}
		}
	}
} // func (sc *Scanner) mainLoop()

func (sc *Scanner) runScan() {
	if swapped := sc.scanRunning.CompareAndSwap(false, true); !swapped {
		sc.log.Println("[INFO] Another scan appears to be in progress. Bye.")
		return
	}
	defer sc.scanRunning.Store(false)

	sc.log.Println("[TRACE] Begin network scan")
	defer sc.log.Println("[TRACE] Finished network scan")

	var (
		wg    sync.WaitGroup
		scanQ = make(chan *scanTarget, sc.workerCnt)
		devQ  = make(chan *model.Device, sc.workerCnt)
	)

	wg.Go(func() { sc.gatherDevices(devQ) })
	wg.Go(func() { sc.feeder(scanQ) })

	for i := range sc.workerCnt {
		wg.Go(func() { sc.scanWorker(i+1, scanQ, devQ) })
	}

	//sc.feeder(scanQ)

	wg.Wait()
} // func (sc Scanner) runScan()

func (sc *Scanner) feeder(scanQ chan<- *scanTarget) {
	sc.log.Println("[TRACE] Scanner feeder coming up...")
	defer sc.log.Println("[TRACE] Scanner feeder quitting...")

	defer close(scanQ) // nolint: errcheck

	const reportInterval = 1024

	var (
		err      error
		db       *database.Database
		networks []*model.Network
		ticker   = time.NewTicker(common.ActiveTimeout)
	)

	defer ticker.Stop()

	db = sc.dbPool.Get()
	defer sc.dbPool.Put(db)

	if networks, err = db.NetGetAll(); err != nil {
		sc.log.Printf("[ERROR] Failed to get networks from database: %s\n",
			err.Error())
		return
	}

	for _, n := range networks {
		sc.log.Printf("[DEBUG] Scanning network %d (%s)\n",
			n.ID,
			n.Addr)
		var (
			ipq = make(chan net.IP)
			cnt int
		)

		if err = n.Enumerate(ipq); err != nil {
			sc.log.Printf("[ERROR] Failed to enumerate network %d (%s): %s\n",
				n.ID,
				n.Addr,
				err.Error())
			return
		}

	ADDR:
		for ip := range ipq {
			var target = &scanTarget{
				net:  n,
				addr: ip,
			}

		TIMEOUT:
			for sc.active.Load() && sc.scanRunning.Load() {
				select {
				case <-ticker.C:
					continue TIMEOUT
				case scanQ <- target:
					if cnt++; cnt%reportInterval == 0 {
						sc.log.Printf("[TRACE] Feeder dispatched %d IP addresses, most recently %s.\n",
							cnt,
							target.addr)
					}
					continue ADDR
				}
			}
		}
	}
} // func (sc *Scanner) feeder(scanQ chan<- *scanTarget)

func (sc *Scanner) scanWorker(id int, scanQ <-chan *scanTarget, devQ chan<- *model.Device) {
	sc.log.Printf("[TRACE] Scanner worker #%d starting up...\n", id)
	defer sc.log.Printf("[TRACE] Scanner worker #%d quitting...\n", id)

	for target := range scanQ {
		// sc.log.Printf("[TRACE] Scan worker %03d about to scan %s\n",
		// 	id,
		// 	target.addr)
		sc.scanAddr(id, target, devQ)
	}
} // func (sc *Scanner) scanWorker(id int, scanQ <-chan *scanTarget, devQ chan<- *model.Device)

func (sc *Scanner) scanAddr(wid int, target *scanTarget, devQ chan<- *model.Device) {
	var (
		err    error
		pinger *probing.Pinger
	)

	pinger = probing.New(target.addr.String())
	pinger.Count = pingCount
	pinger.Interval = pingInterval
	pinger.Timeout = pingCount * pingInterval * 2

	if err = pinger.Run(); err != nil {
		sc.log.Printf("[ERROR] sc#%03d Error scanning %s: %s\n",
			wid,
			target.addr,
			err.Error())
		return
	}

	var stats = pinger.Statistics()

	if stats.PacketLoss >= 95 {
		return
	}

	sc.log.Printf("[INFO] Discovered one Device at %s\n", target.addr)

	var (
		name  string
		names []string
	)

	if names, err = net.LookupAddr(target.addr.String()); err != nil {
		sc.log.Printf("[ERROR] sc#%03d Could not resolve address %s to name: %s\n",
			wid,
			target.addr,
			err.Error())
		return
	} else if len(names) == 0 {
		sc.log.Printf("[INFO] sc#%03d No name(s) were found for %s\n",
			wid,
			target.addr)
		return
	}

	name, _ = strings.CutSuffix(names[0], ".")

	sc.log.Printf("[DEBUG] sc#%03d Discovered one device: %s / %s\n",
		wid,
		name,
		target.addr)

	var dev = &model.Device{
		NetID:  target.net.ID,
		Name:   name,
		Addr:   target.addr,
		Added:  time.Now(),
		Active: true,
	}

	devQ <- dev
} // func (sc *Scanner) scanAddr(target *scanTarget, devQ chan<- *model.Device)

func (sc *Scanner) gatherDevices(devQ <-chan *model.Device) {
	var (
		err    error
		ticker *time.Ticker
		db     *database.Database
	)

	sc.log.Println("[TRACE] Scanner gather worker coming up...")
	defer sc.log.Println("[TRACE] Scanner gather worker quitting...")

	db = sc.dbPool.Get()
	defer sc.dbPool.Put(db)

	ticker = time.NewTicker(common.ActiveTimeout)
	defer ticker.Stop()

	for sc.scanRunning.Load() {
		select {
		case <-ticker.C:
			continue
		case dev := <-devQ:
			if err = db.DeviceAdd(dev); err != nil {
				sc.log.Printf("[ERROR] Cannot add Device %s/%s to Database: %s\n",
					dev.Name,
					dev.Addr,
					err.Error())
			} else {
				sc.log.Printf("[INFO] Device %s/%s was added to Database.\n",
					dev.Name,
					dev.Addr)
			}
		}
	}
} // func (sc *Scanner) gatherDevices(devQ <-chan *model.Device)
