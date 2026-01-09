// /home/krylon/go/src/github.com/blicero/chili/scanner/scanner.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 01. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-01-09 21:26:40 krylon>

// Package scanner implements traversing a range of IP addresses
// and probing which of those correspond to live devices.
package scanner

// NB: IPv6 is NOT supported currently.
import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/blicero/chili/common"
	"github.com/blicero/chili/database"
	"github.com/blicero/chili/logdomain"
	"github.com/blicero/krylib"
)

// http://play.golang.org/p/m8TNTtygK0
// nolint: unused
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

type Scanner struct {
	log       *log.Logger
	lock      *sync.RWMutex
	active    bool
	workerCnt int
	dbPool    *database.Pool
	scanQ     chan net.IP
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

	sc.scanQ = make(chan net.IP, wcnt)

	return sc, nil
} // func New(wcnt int) (*Scanner, error)

// Start starts the Scanners worker goroutines
func (sc *Scanner) Start() error {
	return krylib.ErrNotImplemented
} // func (sc *Scanner) Start() error
