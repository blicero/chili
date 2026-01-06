// /home/krylon/go/src/github.com/blicero/chili/model/host.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 01. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-01-06 16:04:46 krylon>

package model

import (
	"time"

	"github.com/blicero/chili/model/device"
)

type Host struct {
	ID          int64
	Name        string
	Addr        string
	Added       time.Time
	LastContact time.Time
	OS          string
	Type        device.Class
}
