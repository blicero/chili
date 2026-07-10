// /home/krylon/go/src/github.com/blicero/chili/nexus/nexus.go
// -*- mode: go; coding: utf-8; -*-
// Created on 10. 07. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-07-10 13:38:57 krylon>

// Package nexus provides a very primitive kind of message bus for the other
// pieces of the appplication to talk to each other.
package nexus

import (
	"log"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/blicero/chili/common"
	"github.com/blicero/chili/database"
	"github.com/blicero/chili/logdomain"
	"github.com/blicero/chili/nexus/event"
)

// Message is a piece of data sent over the ether, to inform other subsystems
// that something noteworthy happened.
type Message struct {
	Timestamp time.Time
	Event     event.ID
	Payload   any
}

// Nexus is the communications hub of the application.
type Nexus struct {
	log    *log.Logger
	active atomic.Bool
	pool   *database.Pool
	MsgQ   chan Message
}

// New creates and returns a new Nexus
func New() (*Nexus, error) {
	var (
		err error
		nx  = new(Nexus)
	)

	if nx.log, err = common.GetLogger(logdomain.Nexus); err != nil {
		return nil, err
	} else if nx.pool, err = database.NewPool(max(runtime.NumCPU()-2, 2)); err != nil {
		nx.log.Printf("[CRITICAL] Failed to open database connection: %s\n",
			err.Error())
		return nil, err
	}

	nx.MsgQ = make(chan Message)

	return nx, nil
} // func New() (*Nexus, error)

func (nx *Nexus) Start() {
	nx.active.Store(true)

} // func (nx *Nexus) Start()
