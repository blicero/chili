// /home/krylon/go/src/github.com/blicero/chili/scanner/scanner.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 01. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-07-08 13:34:53 krylon>

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
	pingCount    = 8
	scanInterval = time.Minute // XXX Set to more reasonable value after testing/debugging.
)

type scanTarget struct {
	net  *model.Network
	addr net.IP
}

type Scanner struct {
	log       *log.Logger
	scanLock  sync.RWMutex
	active    atomic.Bool
	workerCnt int
	dbPool    *database.Pool
	CmdQ      chan control.Message
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
	sc.scanLock.Lock()
	defer sc.scanLock.Unlock()

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

	wg.Wait()
} // func (sc Scanner) runScan()

func (sc *Scanner) feeder(scanQ chan<- *scanTarget) {
	sc.log.Println("[TRACE] Scanner feeder coming up...")
	defer sc.log.Println("[TRACE] Scanner feeder quitting...")

	var (
		err      error
		db       *database.Database
		networks []*model.Network
	)

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
		var ipq = make(chan net.IP)

		if err = n.Enumerate(ipq); err != nil {
			sc.log.Printf("[ERROR] Failed to enumerate network %d (%s): %s\n",
				n.ID,
				n.Addr,
				err.Error())
			return
		}

		for ip := range ipq {
			var target = &scanTarget{
				net:  n,
				addr: ip,
			}
			scanQ <- target
		}
	}
} // func (sc *Scanner) feeder(scanQ chan<- *scanTarget)

func (sc *Scanner) scanWorker(id int, scanQ <-chan *scanTarget, devQ chan<- *model.Device) {
	sc.log.Printf("[TRACE] Scanner worker #%d starting up...\n", id)
	defer sc.log.Printf("[TRACE] Scanner worker #%d quitting...\n", id)

	var ticker = time.NewTicker(common.ActiveTimeout)
	defer ticker.Stop()

	for sc.IsActive() {
		select {
		case <-ticker.C:
			continue
		case addr := <-scanQ:
			sc.scanAddr(addr, devQ)
		}
	}
} // func (sc *Scanner) scanWorker(id int, scanQ <-chan *scanTarget, devQ chan<- *model.Device)

func (sc *Scanner) scanAddr(target *scanTarget, devQ chan<- *model.Device) {
	var (
		err    error
		pinger *probing.Pinger
	)

	pinger = probing.New(target.addr.String())
	pinger.Count = pingCount

	if err = pinger.Run(); err != nil {
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
		sc.log.Printf("[ERROR] Could not resolve address %s to name: %s\n",
			target.addr,
			err.Error())
		return
	} else if len(names) == 0 {
		sc.log.Printf("[INFO] No name(s) were found for %s\n",
			target.addr)
		return
	}

	name, _ = strings.CutSuffix(names[0], ".")

	sc.log.Printf("[DEBUG] Discovered one device: %s / %s\n",
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

	for sc.IsActive() {
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
