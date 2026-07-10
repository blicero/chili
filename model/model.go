// /home/krylon/go/src/github.com/blicero/chili/model/model.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 07. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-07-09 13:17:05 krylon>

package model

import (
	"fmt"
	"net"
	"time"

	"github.com/blicero/chili/model/device"
)

// // http://play.golang.org/p/m8TNTtygK0
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func clone(ip net.IP) net.IP {
	var b = make(net.IP, len(ip))
	copy(b, ip)
	return b
} // func clone(ip net.IP) net.IP

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

// Enumerate generates all IP addresses for the Network and sends them through
// the channel passed in as its argument.
func (n *Network) Enumerate(q chan<- net.IP) error {
	go func() {
		var ip net.IP
		defer close(q)
		ip = make(net.IP, len(n.Addr.IP))
		copy(ip, n.Addr.IP)
		inc(ip)
		for n.Addr.Contains(ip) {
			var b = clone(ip)
			inc(ip)
			q <- b
		}
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
