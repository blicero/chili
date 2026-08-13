// /home/krylon/go/src/github.com/blicero/hertz/config/config_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 15. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-08-13 10:33:40 krylon>

package config

import (
	"testing"

	"github.com/blicero/chili/common"
	"github.com/davecgh/go-spew/spew"
)

func TestReadNoConfig(t *testing.T) {
	var (
		err    error
		cfg    *Config
		expect = Config{
			Global: Global{
				Debug: true,
			},
			Web: Web{
				Address: "[::1]:4001",
			},
			Ping: Ping{
				Count:    4,
				Timeout:  30,
				Interval: 0.25,
			},
			Scan: Scan{
				Interval: 14400,
			},
			Loglevel: Loglevel{
				Common:    "DEBUG",
				Database:  "DEBUG",
				DBPool:    "DEBUG",
				Ping:      "DEBUG",
				Probe:     "DEBUG",
				Scanner:   "DEBUG",
				Scheduler: "DEBUG",
				Nexus:     "DEBUG",
				Web:       "DEBUG",
			},
		}
	)

	if cfg, err = Read(common.CfgPath); err != nil {
		t.Fatalf("Failed to read %s: %s\n",
			common.CfgPath,
			err.Error())
	} else if cfg == nil {
		t.Fatal("Read() did not return a Config object")
	} else if !cfg.Equal(&expect) {
		t.Fatalf("Read() returned unexpected fresh config:\nExpected: %s\nGot: %s\n",
			spew.Sdump(&expect),
			spew.Sdump(cfg))
	}
} // func TestReadNoConfig(t *testing.T)

func TestReadExampleConfig(t *testing.T) {
	const cfgPath = "testdata/test01.toml"
	var (
		err    error
		cfg    *Config
		expect = Config{
			Global: Global{
				Debug: true,
			},
			Web: Web{
				Address: "[::1]:4242",
			},
			Ping: Ping{
				Count:    13,
				Timeout:  23,
				Interval: 3.141592,
			},
			Scan: Scan{
				Interval: 14400,
			},
			Loglevel: Loglevel{
				Common:    "DEBUG",
				Database:  "DEBUG",
				DBPool:    "DEBUG",
				Ping:      "DEBUG",
				Probe:     "DEBUG",
				Scanner:   "DEBUG",
				Scheduler: "DEBUG",
				Nexus:     "DEBUG",
				Web:       "DEBUG",
			},
		}
	)

	if cfg, err = Read(cfgPath); err != nil {
		t.Fatalf("Failed to read %s: %s\n",
			common.CfgPath,
			err.Error())
	} else if cfg == nil {
		t.Fatal("Read() did not return a Config object")
	} else if !cfg.Equal(&expect) {
		t.Fatalf("Read() returned unexpected config:\nExpected: %s\nGot: %s\n",
			spew.Sdump(&expect),
			spew.Sdump(cfg))
	}
} // func TestReadExampleConfig(t *testing.T)
