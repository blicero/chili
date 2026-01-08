// /home/krylon/go/src/github.com/blicero/chili/model/net.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 01. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-01-08 15:12:10 krylon>

package model

import (
	"net"
	"time"

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
