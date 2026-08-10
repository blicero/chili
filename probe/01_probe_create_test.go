// /home/krylon/go/src/github.com/blicero/chili/probe/01_probe_create.go
// -*- mode: go; coding: utf-8; -*-
// Created on 10. 08. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-08-10 12:53:21 krylon>

package probe

import "testing"

var tprobe *Probe

func TestProbeCreate(t *testing.T) {
	var err error

	if tprobe, err = Create(1, "user"); err != nil {
		tprobe = nil
		t.Fatalf("Failed to create Probe: %s",
			err.Error())
	}
} // func TestProbeCreate(t *testing.T)
