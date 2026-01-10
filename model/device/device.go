// /home/krylon/go/src/github.com/blicero/chili/model/device/device.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 01. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-01-10 16:00:54 krylon>

package device

// Class identifies different kinds of devices
type Class uint8

//go:generate stringer -type=Class

const (
	Unknown Class = iota
	PC
	Server
	SBC
	VM
	Jail
	VPS
	Mobile
	Entertainment
)
