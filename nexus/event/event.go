// /home/krylon/go/src/github.com/blicero/chili/nexus/event/event.go
// -*- mode: go; coding: utf-8; -*-
// Created on 10. 07. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-07-10 12:57:42 krylon>

// Package event provides symbolic constants to describe things happening.
package event

//go:generate stringer -type=ID

// ID describes an event.
type ID uint8

const (
	NetAdded ID = iota
	DeviceAdded
	ScanDue
	ScanFinish
	ProbeDue
	ProbeFinish
	Shutdown
)
