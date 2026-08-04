// /home/krylon/go/src/github.com/blicero/chili/model/attribute/attribute.go
// -*- mode: go; coding: utf-8; -*-
// Created on 11. 07. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-08-04 09:48:55 krylon>

package attribute

//go:generate stringer -type=ID

// ID identifies an Attribute of a Device
type ID uint8

const (
	Updates ID = iota
	DiskSpace
	Uptime
	Packages
	SNMP
	Services
)
