// /home/krylon/go/src/github.com/blicero/chili/model/model.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 07. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-07-20 10:28:01 krylon>

package model

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/blicero/chili/model/attribute"
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

// Payload is the interface for different types of data we can query
// from Devices.
type Payload interface {
	fmt.Stringer
	Type() attribute.ID
}

// Updates is a list of software packages that updates are available for.
type Updates []string

func (u Updates) Type() attribute.ID { return attribute.Updates }

func (u Updates) String() string {
	var (
		err error
		buf []byte
	)

	if buf, err = json.Marshal(u); err != nil {
		panic(err)
	}

	return string(buf)
} // func (u *Updates) String() string

// DiskSpace is the number of free bytes on a Device's root filesystem.
type DiskSpace int64

func (d DiskSpace) Type() attribute.ID { return attribute.DiskSpace }

func (d DiskSpace) String() string {
	return strconv.FormatInt(int64(d), 10)
} // func (d *DiskSpace) String() string

// Uptime is uptime of Device, and its system load.
type Uptime struct {
	Uptime time.Duration
	Load   [3]float64
}

func (u *Uptime) Type() attribute.ID { return attribute.Uptime }
func (u *Uptime) String() string {
	return fmt.Sprintf("{ Uptime: %s, Load: %.1f/%.1f/%.1f }",
		u.Uptime,
		u.Load[0],
		u.Load[1],
		u.Load[2])
}

// Attribute is a property of a Device that we can query/measure.
type Attribute struct {
	ID        int64
	DevID     int64
	Timestamp time.Time
	Type      attribute.ID
	Value     Payload
}
