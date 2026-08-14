// /home/krylon/go/src/github.com/blicero/chili/config/config.go
// -*- mode: go; coding: utf-8; -*-
// Created on 12. 08. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-08-14 10:42:59 krylon>

// Package config deals with the configuration file
package config

import (
	"io"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/blicero/krylib"
)

const defaultConf = `# Time-stamp: <>

[Global]
Debug = true

[Web]
Address = "[::1]:4001"

[Ping]
Count = 4
Timeout = 30
Interval = 0.25

[Scan]
Interval = 14400

[Probe]
Updates = 3600
DiskSpace = 900
Uptime = 60
Packages = 86400
SNMP = 600
Services = 600
DMI = 604800

[Loglevel]
Common = "DEBUG"
Database = "DEBUG"
DBPool = "DEBUG"
Ping = "DEBUG"
Probe = "DEBUG"
Scanner = "DEBUG"
Scheduler = "DEBUG"
Nexus = "DEBUG"
Web = "DEBUG"

`

// Global is the global section of the config file.
type Global struct {
	Debug bool
}

// Equal returns true if other is equal to itself.
func (g *Global) Equal(other any) bool {
	var (
		g2 *Global
		ok bool
	)

	if g2, ok = other.(*Global); !ok {
		return false
	}

	return g.Debug == g2.Debug
} // func (g *Global) Equal(other any) bool

// Web configures the Web server.
type Web struct {
	Address string
}

// Equal returns true if other is equal to itself.
func (w *Web) Equal(other any) bool {
	var (
		x  *Web
		ok bool
	)

	if x, ok = other.(*Web); !ok {
		return false
	}

	return w.Address == x.Address
} // func (w *Web) Equal(other any) bool

// Ping defines values for pinging Devices
type Ping struct {
	Count    int64
	Timeout  int64
	Interval float64
}

// Equal returns true if the argument is of the same type as the receiver
// and if all of its fields have the same values.
func (p *Ping) Equal(other any) bool {
	var (
		p2 *Ping
		ok bool
	)

	if p2, ok = other.(*Ping); !ok {
		return false
	}

	return (p.Count == p2.Count) &&
		(p.Timeout == p2.Timeout) &&
		(p.Interval == p2.Interval)
} // func (p *Ping) Equal(other any) bool

// Scan defines settings for the network scanner
type Scan struct {
	Interval int64
}

// Equal returns true if the argument is of the same type as the receiver
// and if all of its fields have the same values.
func (s *Scan) Equal(other any) bool {
	var (
		t  *Scan
		ok bool
	)

	if t, ok = other.(*Scan); !ok {
		return false
	}

	return s.Interval == t.Interval
} // func (s *Scan) Equal(other any) bool

// Probe defines the settings for the Device Probe
type Probe struct {
	Updates   int64
	DiskSpace int64
	Uptime    int64
	Packages  int64
	SNMP      int64
	Services  int64
	DMI       int64
}

// Equal returns true if the argument is of the same type as the receiver
// and if all of its fields have the same values.
func (p *Probe) Equal(other any) bool {
	var (
		p2 *Probe
		ok bool
	)

	if p2, ok = other.(*Probe); !ok {
		return false
	}

	return (p.Updates == p2.Updates) &&
		(p.DiskSpace == p2.DiskSpace) &&
		(p.Uptime == p2.Uptime) &&
		(p.Packages == p2.Packages) &&
		(p.SNMP == p2.SNMP) &&
		(p.Services == p2.Services) &&
		(p.DMI == p2.DMI)
} // func (p *Probe) Equal(other any) bool

// Loglevel configures the minimum log level for the components of
// the application.
type Loglevel struct {
	Common    string
	Database  string
	DBPool    string
	Ping      string
	Probe     string
	Scanner   string
	Scheduler string
	Nexus     string
	Web       string
}

// Equal returns true if other is equal to itself.
func (l *Loglevel) Equal(other any) bool {
	var (
		m  *Loglevel
		ok bool
	)

	if m, ok = other.(*Loglevel); !ok {
		return false
	}

	return (l.Common == m.Common) &&
		(l.Database == m.Database) &&
		(l.DBPool == m.DBPool) &&
		(l.Ping == m.Ping) &&
		(l.Probe == m.Probe) &&
		(l.Scanner == m.Scanner) &&
		(l.Scheduler == m.Scheduler) &&
		(l.Nexus == m.Nexus) &&
		(l.Web == m.Web)
} // func (l *Loglevel) Equal(other any) bool

// Config defines the configurable settings of Chili
type Config struct {
	Global   Global
	Web      Web
	Ping     Ping
	Scan     Scan
	Probe    Probe
	Loglevel Loglevel
}

// Equal returns true if the argument is of the same type as the receiver
// and if all of its fields have the same values.
func (cfg *Config) Equal(other any) bool {
	switch c2 := other.(type) {
	case *Config:
		return cfg.Global.Equal(&c2.Global) &&
			cfg.Web.Equal(&c2.Web) &&
			cfg.Ping.Equal(&c2.Ping) &&
			cfg.Scan.Equal(&c2.Scan) &&
			cfg.Probe.Equal(&c2.Probe) &&
			cfg.Loglevel.Equal(&c2.Loglevel)
	case Config:
		return cfg.Global.Equal(&c2.Global) &&
			cfg.Web.Equal(&c2.Web) &&
			cfg.Ping.Equal(&c2.Ping) &&
			cfg.Scan.Equal(&c2.Scan) &&
			cfg.Probe.Equal(&c2.Probe) &&
			cfg.Loglevel.Equal(&c2.Loglevel)
	default:
		return false
	}
} // func (cfg *Config) Equal(other any) bool

// Read attempts to read the configuration from the given file.
func Read(path string) (*Config, error) {
	var (
		err    error
		exists bool
		fh     *os.File
		buf    []byte
		cfg    = new(Config)
	)

	if exists, err = krylib.Fexists(path); err != nil {
		return nil, err
	} else if !exists {
		if err = writeDefaultCfg(path); err != nil {
			return nil, err
		}
	}

	if fh, err = os.Open(path); err != nil {
		return nil, err
	}

	defer fh.Close() // nolint: errcheck

	if buf, err = io.ReadAll(fh); err != nil {
		return nil, err
	} else if err = toml.Unmarshal(buf, cfg); err != nil {
		return nil, err
	}

	if cfg.Loglevel.Common == "" {
		cfg.Loglevel.Common = "DEBUG"
	}

	if cfg.Loglevel.Database == "" {
		cfg.Loglevel.Database = "DEBUG"
	}

	if cfg.Loglevel.DBPool == "" {
		cfg.Loglevel.DBPool = "DEBUG"
	}

	if cfg.Loglevel.Ping == "" {
		cfg.Loglevel.Ping = "DEBUG"
	}

	if cfg.Loglevel.Web == "" {
		cfg.Loglevel.Web = "DEBUG"
	}

	if cfg.Loglevel.Probe == "" {
		cfg.Loglevel.Probe = "DEBUG"
	}

	if cfg.Loglevel.Scanner == "" {
		cfg.Loglevel.Scanner = "DEBUG"
	}

	if cfg.Loglevel.Scheduler == "" {
		cfg.Loglevel.Scheduler = "DEBUG"
	}

	if cfg.Loglevel.Nexus == "" {
		cfg.Loglevel.Nexus = "DEBUG"
	}

	return cfg, nil
} // func Read(path string) (*Config, error)

func writeDefaultCfg(path string) error {
	var (
		err error
		fh  *os.File
	)

	if fh, err = os.Create(path); err != nil {
		return err
	}

	defer fh.Close() // nolint: errcheck

	if _, err = fh.Write([]byte(defaultConf)); err != nil {
		return err
	}

	return nil
} // func writeDefaultCfg(path string) error
