// /home/krylon/go/src/github.com/blicero/chili/probe/services.go
// -*- mode: go; coding: utf-8; -*-
// Created on 04. 08. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-08-08 19:57:06 krylon>
//
// query remote devices for running and failed services

package probe

import (
	"regexp"
	"slices"
	"strings"

	"github.com/Feralthedogg/go-functional/pkg/functional"
	"github.com/blicero/chili/model"
)

var (
	patSystemDUnits    = regexp.MustCompile(`^\s*(\S+)\s+(\w+)\s+(\w+)\s+(\w+)`)
	patFreeBSDServices = regexp.MustCompile("/([^/]+)$")
)

// QueryServicesSystemd attempts to query running and failed services on
// Linux Devices using systemd.
func (p *Probe) QueryServicesSystemd(dev *model.Device, port int) (*model.Services, error) {
	const cmd = "systemctl --no-pager list-units"

	var (
		err    error
		output []string
		svc    *model.Services
	)

	if output, err = p.executeCommand(dev, port, cmd); err != nil {
		if err == ErrPingOffline {
			return nil, err
		}

		_ = p.disconnect(dev)
		p.log.Printf("[ERROR] Failed to execute command %q on %s: %s\n",
			cmd,
			dev.Name,
			err.Error())
		return nil, err
	} else if output == nil {
		p.log.Printf("[TRACE] Output of %q on %s was nil?\n",
			cmd,
			dev.Name)
		return nil, nil
	}

	svc = &model.Services{
		Running: make([]string, 0, 16),
		Failed:  make([]string, 0, 4),
	}

	// p.log.Printf("[TRACE] Query %q on %s returned %d lines of output.\n",
	// 	cmd,
	// 	dev.Name,
	// 	len(output))

	for _, line := range output[1:] {
		var match []string
		if match = patSystemDUnits.FindStringSubmatch(line); len(match) > 0 {
			var name, loaded, active, sub string
			name = match[1]
			loaded = match[2]
			active = match[3]
			sub = match[4]

			if !strings.HasSuffix(name, ".service") || loaded != "loaded" || active != "active" {
				continue
			}

			switch sub {
			case "running":
				svc.Running = append(svc.Running, name)
			case "failed":
				svc.Failed = append(svc.Failed, name)
			case "waiting", "exited":
				// don't do anything
				continue
			default:
				p.log.Printf("[DEBUG] Don't know what to do with %s: %s\n",
					sub,
					chomp(line))
			}
		}
	}

	return svc, nil
} // func (p *Probe) QueryServicesSystemd(dev *model.Device, port int) (*model.Services, error)

func chomp(s string) string {
	return strings.Trim(s, "\n\t ")
} // func chomp(s string) string

// QueryServicesOpenBSD attempts to query running and failed services on a device running OpenBSD.
func (p *Probe) QueryServicesOpenBSD(dev *model.Device, port int) (*model.Services, error) {
	const (
		cmdRun  = "doas rcctl ls started"
		cmdFail = "doas rcctl ls failed"
	)

	var (
		err    error
		output []string
		svc    *model.Services
	)

	p.log.Printf("[TRACE] Query running services on %s\n", dev.Name)

	if output, err = p.executeCommand(dev, port, cmdRun); err != nil {
		if err == ErrPingOffline {
			return nil, err
		}
		_ = p.disconnect(dev)
		p.log.Printf("[ERROR] Failed to execute command %q on %s: %s\n",
			cmdRun,
			dev.Name,
			err.Error())
		return nil, err
	}

	// p.log.Printf("[TRACE] Query %q on %s returned %d lines of output.\n",
	// 	cmdRun,
	// 	dev.Name,
	// 	len(output))

	svc = &model.Services{Running: functional.Map(chomp, output)}

	// p.log.Printf("[TRACE] Query dysfunctional services on %s\n", dev.Name)

	if output, err = p.executeCommand(dev, port, cmdFail); err != nil {
		if err == ErrPingOffline {
			return nil, err
		}
		_ = p.disconnect(dev)
		p.log.Printf("[ERROR] Failed to execute command %q on %s: %s\n",
			cmdRun,
			dev.Name,
			err.Error())
		return nil, err
	}

	svc.Failed = slices.DeleteFunc(
		functional.Map(chomp, output),
		func(s string) bool { return s == "" })

	return svc, nil
} // func (p *Probe) QueryServicesOpenBSD(dev *model.Device, port int) (*model.Services, error)

// QueryServicesFreeBSD attempts to query the running services on a Device running FreeBSD.
// I currently have no idea how to query *failed* services on FreeBSD, so for the time
// being, we'll pretend everything is fine.
func (p *Probe) QueryServicesFreeBSD(dev *model.Device, port int) (*model.Services, error) {
	const cmd = "doas /usr/sbin/service -e"

	var (
		err    error
		output []string
		svc    *model.Services
	)

	if output, err = p.executeCommand(dev, port, cmd); err != nil {
		if err == ErrPingOffline {
			return nil, err
		}
		_ = p.disconnect(dev)
		p.log.Printf("[ERROR] Failed to execute command %q on %s: %s\n",
			cmd,
			dev.Name,
			err.Error())
		return nil, err
	}

	svc = &model.Services{
		Running: make([]string, 0, len(output)),
		Failed:  make([]string, 0),
	}

	for _, line := range output {
		var match []string

		line = strings.Trim(line, "\n\t ")

		if line == "" {
			continue
		} else if match = patFreeBSDServices.FindStringSubmatch(line); match == nil {
			p.log.Printf("[DEBUG] Cannot parse output of %q on %s: %s\n",
				cmd,
				dev.Name,
				line)
		} else {
			svc.Running = append(svc.Running, match[1])
		}
	}

	// svc = &model.Services{Running: functional.Map(chomp, output)}

	return svc, nil
} // func (p *Probe) QueryServicesFreeBSD(dev *model.Device, port int) (*model.Services, error)

// QueryServices attempts to query a Device for running and failed services.
func (p *Probe) QueryServices(dev *model.Device, port int) (*model.Services, error) {
	switch dev.OS {
	case "Debian GNU/Linux", "Raspbian GNU/Linux":
		fallthrough
	case "openSUSE Tumbleweed", "openSUSE Leap":
		fallthrough
	case "Fedora Linux":
		fallthrough
	case "Arch Linux":
		return p.QueryServicesSystemd(dev, port)
	case "OpenBSD":
		return p.QueryServicesOpenBSD(dev, port)
	case "FreeBSD", "NetBSD":
		return p.QueryServicesFreeBSD(dev, port)
	default:
		p.log.Printf("[DEBUG] Don't know how to query %s (on %s) for service status.\n",
			dev.OS,
			dev.Name)
		return nil, nil
	}
} // func (p *Probe) QueryServices(dev *model.Device, port int) (*model.Services, error)
