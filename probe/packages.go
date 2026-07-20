// /home/krylon/go/src/github.com/blicero/chili/probe/packages.go
// -*- mode: go; coding: utf-8; -*-
// Created on 20. 07. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-07-20 13:18:24 krylon>

package probe

import (
	"regexp"
	"strings"

	"github.com/blicero/chili/model"
)

var patPkgDebian = regexp.MustCompile(`^([^/]+)/(\S+)\s+(\S+)\s+(\S+)`)

// QueryUpdatesDebian asks a Debian-ish system for a list of available updates.
func (p *Probe) QueryPackagesDebian(d *model.Device, port int) ([]string, error) {
	const cmd = "/usr/bin/apt list --installed"
	var (
		err     error
		output  []string
		match   []string
		updates []string
	)

	if output, err = p.executeCommand(d, port, cmd); err != nil {
		if err == ErrPingOffline {
			return nil, err
		}
		_ = p.disconnect(d)
		p.log.Printf("[ERROR] Failed to execute command %q on %s: %s\n",
			cmd,
			d.Name,
			err.Error())
		return nil, err
	}

	updates = make([]string, 0)

	for _, l := range output {
		if match = patPkgDebian.FindStringSubmatch(l); len(match) > 0 {
			var upd = strings.Join(match[1:], pkgSep)
			updates = append(updates, upd)
		}
	}

	return updates, nil
} // func (p *Probe) QueryPackagesDebian(d *model.Device, port int) ([]string, error)

// QueryPackagesFreeBSD asks a Device running FreeBSD for the list of
// installed packages.
func (p *Probe) QueryPackagesFreeBSD(d *model.Device, port int) ([]string, error) {
	const cmd = "/usr/sbin/pkg info"
	var (
		err     error
		output  []string
		match   []string
		updates []string
	)

	if output, err = p.executeCommand(d, port, cmd); err != nil {
		if err == ErrPingOffline {
			return nil, err
		}
		_ = p.disconnect(d)
		p.log.Printf("[ERROR] Failed to execute command %q on %s: %s\n",
			cmd,
			d.Name,
			err.Error())
		return nil, err
	}

	updates = make([]string, 0)

	for _, l := range output {
		if match = patPkgDebian.FindStringSubmatch(l); len(match) > 0 {
			var upd = strings.Join(match[1:], pkgSep)
			updates = append(updates, upd)
		}
	}

	return updates, nil
} // func (p *Probe) QueryPackagesFreeBSD(d *model.Device, port int) ([]string, error)
