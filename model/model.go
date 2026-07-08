// /home/krylon/go/src/github.com/blicero/chili/model/model.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 07. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-07-08 12:40:03 krylon>

package model

import (
	"fmt"
	"net"
	"time"

	"github.com/blicero/chili/model/device"
	"github.com/korylprince/ipnetgen"
)

// Network defines a range of IP addresses where our devices live.
type Network struct {
	ID       int64
	Name     string
	Addr     *net.IPNet
	Added    time.Time
	LastScan time.Time
}

func NewNet(name, addr string) (*Network, error) {
	var err error
	var n = &Network{
		Name: name,
	}

	if _, n.Addr, err = net.ParseCIDR(addr); err != nil {
		return nil, err
	}

	return n, nil
} // func NewNet(name, addr string) (*Network, error)

// Enumerate generates all IP addresses for the Network and sends them through the channel
// passed in as its argument.
func (n *Network) Enumerate(q chan<- net.IP) error {
	gen, err := ipnetgen.New(n.Addr.String())

	if err != nil {
		return err
	}

	go func() {
		for ip := gen.Next(); ip != nil; ip = gen.Next() {
			if !ip.IsMulticast() {
				q <- ip
			}
		}
		close(q)
	}()

	return nil
} // func (n *Network) Enumerate(q chan<- net.IP)

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
