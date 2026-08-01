// /home/krylon/go/src/github.com/blicero/chili/probe/packages.go
// -*- mode: go; coding: utf-8; -*-
// Created on 20. 07. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-08-01 06:05:57 krylon>

package probe

import (
	"regexp"
	"strings"

	"github.com/blicero/chili/model"
)

var (
	patPkgDebian = regexp.MustCompile(`^([^/]+)/(\S+)\s+(\S+)\s+(\S+)`)
	patPkgFedora = regexp.MustCompile(`^(\S+)\s+(\S+)\s+\S+$`)
)

// QueryPackagesDebian asks a Debian-ish system for a list of available updates.
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

// QueryPackagesOpenBSD attempts to get the installed packages from an OpenBSD device.
func (p *Probe) QueryPackagesOpenBSD(d *model.Device, port int) ([]string, error) {
	const cmd = "/usr/sbin/pkg_info"
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
} // func (p *Probe) QueryPackagesOpenBSD(d *model.Device, port int) ([]string, error)

// QueryPackagesNetBSD attempts to get the installed packages from an NetBSD device.
func (p *Probe) QueryPackagesNetBSD(d *model.Device, port int) ([]string, error) {
	const cmd = "/usr/pkg/bin/pkgin list"
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
} // func (p *Probe) QueryPackagesNetBSD(d *model.Device, port int) ([]string, error)

// QueryPackagesSuse attempts to get the installed packages from an openSuse device.
func (p *Probe) QueryPackagesSuse(d *model.Device, port int) ([]string, error) {
	const cmd = "/usr/bin/zypper pa -i"
	var (
		err      error
		output   []string
		packages []string
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

	packages = make([]string, 0)

	for _, l := range output[4:] {
		l = strings.Trim(l, " \t\n")
		var pieces = patUpdateSuse.Split(l, -1)
		if len(pieces) > 0 {
			var upd = strings.Join(pieces[1:], pkgSep)
			packages = append(packages, upd)
		}
	}

	return packages, nil
} // func (p *Probe) QueryPackagesSuse(d *model.Device, port int) ([]string, error)

// QueryPackagesArch attempts to get the installed packages from an Arch Linux device.
func (p *Probe) QueryPackagesArch(d *model.Device, port int) ([]string, error) {
	const cmd = "/sbin/pacman -Q"
	var (
		err      error
		output   []string
		packages []string
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

	packages = make([]string, 0)

	for _, l := range output {
		l = strings.Trim(l, " \t\n")
		packages = append(packages, l)
	}

	return packages, nil
} // func (p *Probe) QueryPackagesArch(d *model.Device, port int) ([]string, error)

// QueryPackagesFedora attempts to get the installed packages from a Fedora device.
func (p *Probe) QueryPackagesFedora(d *model.Device, port int) ([]string, error) {
	const cmd = "/usr/sbin/dnf list -q"
	var (
		err      error
		output   []string
		packages []string
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

	packages = make([]string, 0)

	for _, l := range output[1:] {
		var pkg string
		l = strings.Trim(l, " \t\n")
		if match := patPkgFedora.FindStringSubmatch(l); len(match) == 3 {
			pkg = match[1] + " " + match[2]
			packages = append(packages, pkg)
		} else {
			p.log.Printf("[ERROR] Don't understand output of dnf(8) on %s: %s\n",
				d.Name,
				l)
		}
	}

	return packages, nil
} // func (p *Probe) QueryPackagesArch(d *model.Device, port int) ([]string, error)

// QueryPackages attempts to query a list of installed packages from a Device.
func (p *Probe) QueryPackages(d *model.Device, port int) ([]string, error) {
	switch d.OS {
	case "Debian GNU/Linux", "Raspbian GNU/Linux":
		return p.QueryPackagesDebian(d, port)
	case "openSUSE Leap", "openSUSE Tumbleweed":
		return p.QueryPackagesSuse(d, port)
	case "FreeBSD":
		return p.QueryPackagesFreeBSD(d, port)
	case "OpenBSD":
		return p.QueryPackagesOpenBSD(d, port)
	case "NetBSD":
		return p.QueryPackagesNetBSD(d, port)
	case "Arch Linux":
		return p.QueryPackagesArch(d, port)
	default:
		p.log.Printf("[INFO] Don't know how to query installed packages from %s (%s)\n",
			d.Name,
			d.OS)
		return nil, nil
	}
} // func (p *Probe) QueryPackages(d *model.Device, port int) ([]string, error)
