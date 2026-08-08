// /home/krylon/go/src/github.com/blicero/chili/probe/dmi.go
// -*- mode: go; coding: utf-8; -*-
// Created on 07. 08. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-08-07 16:52:52 krylon>

package probe

// There is a tool called dmidecode(1) that can query a lot of information
// about a Device's private parts. Up to things like Hardware manufacturer,
// model number, serial number.
//
// Now THIS is very interesting. VERY interesting.
// The output of dmidecode(1) is a bit difficult to parse, but I shall try.
// NB On OpenBD, one has to add "kern.allowkmem=1" to /etc/sysctl.conf
