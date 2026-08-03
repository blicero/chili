// /home/krylon/go/src/github.com/blicero/chili/probe/updates.go
// -*- mode: go; coding: utf-8; -*-
// Created on 20. 07. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-08-03 11:07:37 krylon>

package probe

import (
	"regexp"
	"strings"

	"github.com/blicero/chili/model"
)

var patUpdateDebian = regexp.MustCompile(`^([^/]+)/(\S+)\s+(\S+)\s+(\S+)`)

// QueryUpdatesDebian asks a Debian-ish system for a list of available updates.
func (p *Probe) QueryUpdatesDebian(d *model.Device, port int) ([]string, error) {
	const cmd = "/usr/bin/apt list --upgradable"
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
		if match = patUpdateDebian.FindStringSubmatch(l); len(match) > 0 {
			var upd = strings.Join(match[1:], pkgSep)
			updates = append(updates, upd)
		}
	}

	return updates, nil
} // func (p *Probe) QueryUpdatesDebian(d *model.Device, port int) ([]string, error)

var patUpdateSuse = regexp.MustCompile(`\s+\|\s+`)

// QueryUpdatesSuse asks an openSuse system for a list of available updates.
func (p *Probe) QueryUpdatesSuse(d *model.Device, port int) ([]string, error) {
	const cmd = "zypper lu"
	var (
		err     error
		output  []string
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
	} else if output == nil {
		p.log.Printf("[ERROR] Querying updates on %s did not return an error, but no output either\n",
			d.Name)
		return nil, ErrNoData
	}

	updates = make([]string, 0)

	for _, l := range output[4:] {
		l = strings.Trim(l, " \t\n")
		var pieces = patUpdateSuse.Split(l, -1)
		if len(pieces) > 0 {
			var upd = strings.Join(pieces[1:], pkgSep)
			updates = append(updates, upd)
		}
	}

	return updates, nil
} // func (p *Probe) QueryUpdatesSuse(d *model.Device, port int) ([]string, error)

var patUpdateDNF = regexp.MustCompile(`\s+`)

// QueryUpdatesFedora asks a Fedora system for a list of available updates.
// Or any other system based the dnf package manager.
func (p *Probe) QueryUpdatesFedora(d *model.Device, port int) ([]string, error) {
	const cmd = "env DNF5_FORCE_INTERACTIVE=0 dnf check-upgrade"
	var (
		err             error
		output, updates []string
	)

	if output, err = p.executeCommand(d, port, cmd); err != nil {
		if err == ErrPingOffline {
			return nil, err
		}
		_ = p.disconnect(d)
		return nil, err
	}

	updates = make([]string, 0)

	for _, l := range output {
		var pieces = patUpdateDNF.Split(l, -1)
		if len(pieces) == 3 {
			var upd = strings.Join(pieces, pkgSep)
			updates = append(updates, upd)
		}
	}

	return updates, nil
} // func (p *Probe) QueryUpdatesFedora(d *model.Device, port int) ([]string, error)

var patUpdateArch = regexp.MustCompile(`^(\S+)\s+(\S+)\s+->\s+(\S+)$`)

// QueryUpdatesArch asks an Arch Linux system for a list of pending updates.
func (p *Probe) QueryUpdatesArch(d *model.Device, port int) ([]string, error) {
	const cmd = "checkupdates"
	var (
		err             error
		output, updates []string
	)

	if output, err = p.executeCommand(d, port, cmd); err != nil {
		if err == ErrPingOffline {
			return nil, err
		}
		_ = p.disconnect(d)
		return nil, err
	}

	updates = make([]string, 0)

	for _, l := range output {
		var match []string
		if match = patUpdateArch.FindStringSubmatch(l); len(match) > 0 {
			var upd = strings.Join(match[1:], pkgSep)
			updates = append(updates, upd)
		}
	}

	return updates, nil
} // func (p *Probe) QueryUpdatesArch(d *model.Device, port int) ([]string, error)

var patUpdateOpenBSD = regexp.MustCompile(`\w+`)

// QueryUpdatesOpenBSD checks for available updates on OpenBSD.
func (p *Probe) QueryUpdatesOpenBSD(d *model.Device, port int) ([]string, error) {
	const cmd = "/usr/bin/doas /usr/sbin/syspatch -c"
	var (
		err             error
		output, updates []string
	)

	if output, err = p.executeCommand(d, port, cmd); err != nil {
		if err == ErrPingOffline {
			return nil, err
		}
		_ = p.disconnect(d)
		return nil, err
	}

	updates = make([]string, 0, len(output))

	for _, l := range output {
		if patUpdateOpenBSD.MatchString(l) {
			updates = append(updates, l)
		}
	}

	if len(updates) == 0 {
		return nil, nil
	}

	return updates, nil
} // func (p *Probe) QueryUpdatesOpenBSD(d *model.Device, port int) ([]string, error)

// QueryUpdatesFreeBSD checks for available updates on FreeBSD.
func (p *Probe) QueryUpdatesFreeBSD(d *model.Device, port int) ([]string, error) {
	const cmd = "doas freebsd-update updatesready"
	var (
		err             error
		output, updates []string
	)

	if output, err = p.executeCommand(d, port, cmd); err != nil {
		if err == ErrPingOffline {
			return nil, err
		}
		_ = p.disconnect(d)
		return nil, err
	}

	updates = make([]string, 0, len(output))

	for _, l := range output {
		if patUpdateOpenBSD.MatchString(l) {
			updates = append(updates, l)
		}
	}

	if len(updates) == 0 {
		return nil, nil
	}

	return updates, nil
} // func (p *Probe) QueryUpdatesFreeBSD(d *model.Device, port int) ([]string, error)

// QueryUpdates attempts to query the given Device for available updates.
func (p *Probe) QueryUpdates(d *model.Device, port int) ([]string, error) {
	switch d.OS {
	case "Debian GNU/Linux", "Raspbian GNU/Linux":
		return p.QueryUpdatesDebian(d, port)
	case "openSUSE Tumbleweed", "openSUSE Leap":
		return p.QueryUpdatesSuse(d, port)
	case "Fedora Linux":
		return p.QueryUpdatesFedora(d, port)
	case "Arch Linux":
		return p.QueryUpdatesArch(d, port)
	case "OpenBSD":
		return p.QueryUpdatesOpenBSD(d, port)
	case "FreeBSD":
		return p.QueryUpdatesFreeBSD(d, port)
	default:
		p.log.Printf("[TRACE] Don't know how to query %s (running %s) for updates\n",
			d.Name,
			d.OS)
		return nil, nil
	}
} // func (p *Probe) QueryUpdates(d *model.Device, port int) ([]string, error)
