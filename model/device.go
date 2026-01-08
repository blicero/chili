// /home/krylon/go/src/github.com/blicero/chili/model/host.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 01. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-01-08 15:30:29 krylon>

package model

import (
	"fmt"
	"net"
	"time"

	"github.com/blicero/chili/model/device"
)

// Device is a networked computer.
type Device struct {
	ID          int64
	NetID       int64
	Name        string
	Addr        net.IP
	Added       time.Time
	LastContact time.Time
	OS          string
	Class       device.Class
	Active      bool
}

// NewDevice creates a new Device.
func NewDevice(netID int64, name, addr string) (*Device, error) {
	var d = &Device{
		NetID:  netID,
		Name:   name,
		Active: true,
	}

	if d.Addr = net.ParseIP(addr); d.Addr == nil {
		return nil, fmt.Errorf("cannot parse address '%s'", addr)
	}

	return d, nil
} // func NewDevice(name, addr string) (*Device, error)
