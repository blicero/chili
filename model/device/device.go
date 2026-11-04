// /home/krylon/go/src/github.com/blicero/chili/model/device/device.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 01. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-11-03 13:00:13 krylon>

package device

// Class identifies different kinds of devices
type Class uint8

//go:generate stringer -type=Class

const (
	Unknown Class = iota
	PC
	Server
	VM
	Jail
	Mobile
	Entertainment
	Router
)

// Int64 returns the Class value as an int64.
func (c Class) Int64() int64 {
	return int64(c)
} // func (c Class) Int64() int64

// All returns a slice of all valid Class values.
func All() []Class {
	return []Class{
		Unknown,
		PC,
		Server,
		VM,
		Jail,
		Mobile,
		Entertainment,
		Router,
	}
} // func All() []Class
