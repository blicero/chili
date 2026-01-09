// /home/krylon/go/src/github.com/blicero/chili/control/message.go
// -*- mode: go; coding: utf-8; -*-
// Created on 09. 01. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-01-09 21:33:41 krylon>

// Package control provides symbolic constants for controlling the subsystems
// and their worker goroutines.
package control

//go:generate stringer -type=Message

// Message represents a command sent to a subsystem or one of its workers.
type Message uint8

const (
	Nothing Message = iota
	Scan
	Stop
	Pause
)
