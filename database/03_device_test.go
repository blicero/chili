// /home/krylon/go/src/github.com/blicero/chili/database/03_device_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 07. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-07-08 13:10:15 krylon>

package database

import (
	"fmt"
	"testing"

	"github.com/blicero/chili/model"
	"github.com/korylprince/ipnetgen"
)

const (
	dCount = 16
)

var tdevs = make([]*model.Device, 0, dCount)

func TestDeviceAdd(t *testing.T) {
	if tdb == nil {
		t.SkipNow()
	} else if tnet == nil || tnet.ID == 0 {
		t.SkipNow()
	}

	var (
		err   error
		ipgen *ipnetgen.IPNetGenerator
		dev   *model.Device
	)

	ipgen, _ = ipnetgen.New(tNetAddr)

	for i := range dCount {
		if dev, err = model.NewDevice(
			tnet.ID,
			fmt.Sprintf("host%02d", i+1),
			ipgen.Next().String()); err != nil {
			t.Fatalf("Failed to create Device: %s", err.Error())
		} else if err = tdb.DeviceAdd(dev); err != nil {
			t.Fatalf("Failed to add Device %s (%s): %s",
				dev.Name,
				dev.Addr,
				err.Error())
		}

		tdevs = append(tdevs, dev)
	}
} // func TestDeviceAdd(t *testing.T)

func TestDeviceAddTwice(t *testing.T) {
	if tdb == nil {
		t.SkipNow()
	} else if tnet == nil || tnet.ID == 0 {
		t.SkipNow()
	}

	var err error

	for _, dev := range tdevs {
		if err = tdb.DeviceAdd(dev); err != nil {
			t.Fatalf("Failed to upsert Device %s (%s): %s",
				dev.Name,
				dev.Addr,
				err.Error())
		}
	}
} // func TestDeviceAddTwice(t *testing.T)
