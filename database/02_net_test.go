// /home/krylon/go/src/github.com/blicero/chili/database/02_net_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 07. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-07-08 12:57:41 krylon>

package database

import (
	"testing"

	"github.com/blicero/chili/model"
)

const tNetAddr = "172.16.42.0/24"

var tnet *model.Network

func TestNetAdd(t *testing.T) {
	if tdb == nil {
		t.SkipNow()
	}

	var err error

	if tnet, err = model.NewNet("Test Network 01", tNetAddr); err != nil {
		tnet = nil

		t.Fatalf("Failed to create test network %s", tNetAddr)

	} else if err = tdb.NetAdd(tnet); err != nil {
		t.Fatalf("Adding network to database failed: %s", err.Error())
	} else if tnet.ID == 0 {
		t.Fatalf("NetAdd did not set Network ID for %s", tNetAddr)
	}
} // func TestNetAdd(t *testing.T)

func TestNetAddTwice(t *testing.T) {
	if tdb == nil {
		t.SkipNow()
	}

	var (
		err  error
		xnet *model.Network
	)

	if xnet, err = model.NewNet("Test Network 01", tNetAddr); err != nil {
		tnet = nil

		t.Fatalf("Failed to create test network %s", tNetAddr)

	} else if err = tdb.NetAdd(xnet); err == nil {
		t.Fatalf("Adding network to database should have failed: %s", err.Error())
	}
} // func TestNetAddTwice(t *testing.T)
