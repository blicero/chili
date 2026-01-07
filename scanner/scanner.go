// /home/krylon/go/src/github.com/blicero/chili/scanner/scanner.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 01. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-01-07 15:43:28 krylon>

// Package scanner implements traversing a range of IP addresses
// and probing which of those correspond to live devices.
package scanner

// NB: IPv6 is NOT supported currently.
import "net"

// http://play.golang.org/p/m8TNTtygK0
// nolint: unused
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
