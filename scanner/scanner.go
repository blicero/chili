// /home/krylon/go/src/github.com/blicero/chili/scanner/scanner.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 01. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-01-10 16:15:34 krylon>

// Package scanner implements traversing a range of IP addresses
// and probing which of those correspond to live devices.
package scanner

// NB: IPv6 is NOT supported currently.
import (
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/blicero/chili/common"
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
	lock      *sync.RWMutex
	active    atomic.Bool
	workerCnt int
	dbPool    *database.Pool
	scanQ     chan *scanTarget
	devQ      chan *model.Device
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

	sc.scanQ = make(chan *scanTarget, wcnt)
	sc.devQ = make(chan *model.Device, wcnt)

	return sc, nil
} // func New(wcnt int) (*Scanner, error)

// IsActive returns the state of the Scanner's active flag.
func (sc *Scanner) IsActive() bool {
	return sc.active.Load()
} // func (sc *Scanner) IsActive() bool

// Start starts the Scanners worker goroutines
func (sc *Scanner) Start() error {
	sc.active.Store(true)
	go sc.gatherDevices()

	for i := range sc.workerCnt {
		go sc.scanWorker(i + 1)
	}

	return nil
} // func (sc *Scanner) Start() error

func (sc *Scanner) scanWorker(id int) {
	sc.log.Printf("[TRACE] Scanner worker #%d starting up...\n", id)
	defer sc.log.Printf("[TRACE] Scanner worker #%d quitting...\n", id)

	var ticker = time.NewTicker(common.ActiveTimeout)
	defer ticker.Stop()

	for sc.IsActive() {
		select {
		case <-ticker.C:
			continue
		case addr := <-sc.scanQ:
			sc.scanAddr(addr)
		}
	}
} // func (sc *Scanner) worker(id int)

func (sc *Scanner) scanAddr(target *scanTarget) {
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

	var names []string

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

	sc.log.Printf("[DEBUG] Discovered one device: %s / %s\n",
		names[0],
		target.addr)

	var dev = &model.Device{
		NetID:  target.net.ID,
		Name:   names[0],
		Addr:   target.addr,
		Added:  time.Now(),
		Active: true,
	}

	sc.devQ <- dev
} // func (sc *Scanner) scanAddr(addr net.IP)

func (sc *Scanner) gatherDevices() {
	var (
		err    error
		ticker *time.Ticker
		db     *database.Database
	)

	db = sc.dbPool.Get()
	defer sc.dbPool.Put(db)

	sc.log.Println("[TRACE] Scanner gather worker coming up...")
	defer sc.log.Println("[TRACE] Scanner gather worker quitting...")

	ticker = time.NewTicker(common.ActiveTimeout)
	defer ticker.Stop()

	for sc.IsActive() {
		select {
		case <-ticker.C:
			continue
		case dev := <-sc.devQ:
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
} // func (sc *Scanner) gatherDevices()
