// /home/krylon/go/src/github.com/blicero/chili/model/attribute/attribute.go
// -*- mode: go; coding: utf-8; -*-
// Created on 11. 07. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-10-30 13:14:52 krylon>

package attribute

import (
	"fmt"
	"strings"
)

//go:generate stringer -type=ID

// ID identifies an Attribute of a Device
type ID uint8

const (
	None ID = iota
	Updates
	DiskSpace
	Uptime
	Packages
	SNMP
	Services
	DMI
	Finished
)

// All returns a slice of all valid ID values.
func All() []ID {
	return []ID{
		Updates,
		DiskSpace,
		Uptime,
		Packages,
		SNMP,
		Services,
		DMI,
	}
} // func All() []ID

// Parse attempts to convert the given string to an attribute ID.
func Parse(s string) ID {
	switch strings.ToLower(s) {
	case "updates":
		return Updates
	case "diskspace":
		return DiskSpace
	case "uptime":
		return Uptime
	case "packages":
		return Packages
	case "snmp":
		return SNMP
	case "services":
		return Services
	case "dmi":
		return DMI
	default:
		panic(fmt.Sprintf("Unknown attribute ID %s", s))
	}
} // func Parse(s string) ID
