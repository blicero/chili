// /home/krylon/go/src/github.com/blicero/chili/model/net.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 01. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-01-06 16:07:59 krylon>

package model

import "time"

// Network defines a range of IP addresses where our devices live.
type Network struct {
	ID       int64
	Name     string
	Addr     string
	Added    time.Time
	LastScan time.Time
}
