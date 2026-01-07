// /home/krylon/go/src/github.com/blicero/chili/database/query/query.go
// -*- mode: go; coding: utf-8; -*-
// Created on 07. 01. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-01-07 16:24:00 krylon>

package query

//go:generate stringer -type=ID

// ID signifies a particular SQL query we perform on the database.
type ID uint8

const (
	NetAdd ID = iota
	NetUpdateLastScan
	NetUpdateName
	NetGetByID
	NetGetAll
	DeviceAdd
	DeviceUpdateLastContact
	DeviceUpdateOS
	DeviceUpdateClass
	DeviceUpdateName
	DeviceUpdateActive
	DeviceGetByID
	DeviceGetByNet
	DeviceGetAll
)
