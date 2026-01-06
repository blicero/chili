// /home/krylon/go/src/github.com/blicero/chili/model/device/device.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 01. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-01-06 15:59:28 krylon>

package device

// Class identifies different kinds of devices
type Class uint8

//go:generate stringer -type=Class

const (
	PC Class = iota
	Laptop
	Server
	SBC
	VM
	Jail
	VPS
)
