// /home/krylon/go/src/github.com/blicero/chili/model/attribute/attribute.go
// -*- mode: go; coding: utf-8; -*-
// Created on 11. 07. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-08-08 11:22:07 krylon>

package attribute

import (
	"fmt"
	"strings"
)

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
	default:
		panic(fmt.Sprintf("Unknown attribute ID %s", s))
	}
} // func Parse(s string) ID
