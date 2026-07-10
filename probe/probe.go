// /home/krylon/go/src/github.com/blicero/chili/probe/probe.go
// -*- mode: go; coding: utf-8; -*-
// Created on 09. 07. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-07-09 11:49:07 krylon>

// Package probe implements the detailed interrogration of Devices
// the Scanner has discovered.
package probe

import (
	"log"

	"github.com/blicero/chili/common"
	"github.com/blicero/chili/control"
	"github.com/blicero/chili/database"
	"github.com/blicero/chili/logdomain"
)

// Probe queries Devices about their OS, installed software, hardware,
// pending updates, etc...
type Probe struct {
	log    *log.Logger
	CmdQ   chan control.Message
	pool   *database.Pool
	parCnt int
}

// Create returns a fresh Probe.
func Create(cnt int) (*Probe, error) {
	var (
		err error
		p   = &Probe{parCnt: cnt}
	)

	if p.log, err = common.GetLogger(logdomain.Probe); err != nil {
		return nil, err
	} else if p.pool, err = database.NewPool(max(cnt-2, 1)); err != nil {
		p.log.Printf("[CRITICAL] Cannot open database connection pool: %s\n",
			err.Error())
		return nil, err
	}

	p.CmdQ = make(chan control.Message, 2)
	return p, nil
} // func Create() (*Probe, error)
