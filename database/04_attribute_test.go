// /home/krylon/go/src/github.com/blicero/chili/database/04_attribute_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 11. 07. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-07-21 11:17:35 krylon>

package database

import (
	"math/rand/v2"
	"testing"
	"time"

	"github.com/blicero/chili/model"
	"github.com/blicero/chili/model/attribute"
)

const atCnt = 4

var tPackages = []string{
	"emacs",
	"vim",
	"git",
	"zsh",
	"rsync",
	"vlc",
	"firefox",
	"libreoffice",
	"perl",
	"python",
	"fortune",
	"stellarium",
	"strawberry",
	"gimp",
	"inkscape",
	"sqlite",
	"ghostscript",
}

var tattr []*model.Attribute

func rndPack() model.Updates {
	var perm = rand.Perm(len(tPackages))
	var cnt = rand.IntN(len(tPackages)) + 1

	var res = make([]string, cnt)

	for i := range cnt {
		res[i] = tPackages[perm[i]]
	}

	return res
}

func TestAttributeAdd(t *testing.T) {
	if tdb == nil ||
		tnet == nil ||
		tnet.ID == 0 ||
		len(tdevs) == 0 {
		t.SkipNow()
	}

	var (
		err    error
		tstamp = time.Now().Add(time.Hour * -atCnt)
	)

	tattr = make([]*model.Attribute, 0, len(tdevs)*atCnt)

	for h := range atCnt {
		var tstamp = tstamp.Add(time.Minute * 15 * time.Duration(h))

		for _, dev := range tdevs {
			var bit = &model.Attribute{
				DevID:     dev.ID,
				Type:      attribute.DiskSpace,
				Timestamp: tstamp,
				Value:     model.DiskSpace(1024 * rand.Int64N(1<<16)),
			}

			if err = tdb.AttributeAdd(bit); err != nil {
				t.Fatalf("Failed to add Attribute %s for device %s: %s",
					bit.Type,
					dev.Name,
					err.Error())
			} else {
				tattr = append(tattr, bit)
			}
		}

		for _, dev := range tdevs {
			var bit = &model.Attribute{
				DevID:     dev.ID,
				Type:      attribute.Updates,
				Timestamp: tstamp,
				Value:     rndPack(),
			}

			if err = tdb.AttributeAdd(bit); err != nil {
				t.Fatalf("Failed to add Attribute %s for device %s: %s",
					bit.Type,
					dev.Name,
					err.Error())
			} else {
				tattr = append(tattr, bit)
			}
		}
	}
} // func TestAttributeAdd(t *testing.T)

func TestAttributeGetByDevice(t *testing.T) {
	if tdb == nil ||
		tnet == nil ||
		tnet.ID == 0 ||
		len(tdevs) == 0 ||
		len(tattr) == 0 {
		t.SkipNow()
	}

	var (
		err  error
		attr []*model.Attribute
	)

	for _, dev := range tdevs {
		if attr, err = tdb.AttributeGetByDevice(dev); err != nil {
			t.Fatalf("Failed to get attributes per Device %s: %s",
				dev.Name,
				err.Error())
		} else if len(attr) != 2 {
			t.Fatalf("Unexpected number of attributes for %s: %d (expected 2)",
				dev.Name,
				len(attr))
		}
	}
} // func TestAttributeGetByDevice(t *testing.T)

func TestAttributeGetByType(t *testing.T) {
	if tdb == nil ||
		tnet == nil ||
		tnet.ID == 0 ||
		len(tdevs) == 0 ||
		len(tattr) == 0 {
		t.SkipNow()
	}

	var (
		err  error
		attr []*model.Attribute
	)

	if attr, err = tdb.AttributeGetByType(attribute.DiskSpace); err != nil {
		t.Fatalf("Failed to load attributes DiskSpace: %s",
			err.Error())
	} else if len(attr) == 0 {
		t.Fatalf("Unexpected nummber of attributes %d (expected %d)",
			len(attr),
			len(tdevs)*atCnt)
	}
} // func TestAttributeGetByType(t *testing.T)
