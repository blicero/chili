// /home/krylon/go/src/github.com/blicero/chili/probe/dmi.go
// -*- mode: go; coding: utf-8; -*-
// Created on 07. 08. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-08-10 13:57:54 krylon>

package probe

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/blicero/chili/model"
)

// There is a tool called dmidecode(1) that can query a lot of information
// about a Device's private parts. Up to things like Hardware manufacturer,
// model number, serial number.
//
// Now THIS is very interesting. VERY interesting.
// The output of dmidecode(1) is a bit difficult to parse, but I shall try.
// NB On OpenBD, one has to add "kern.allowkmem=1" to /etc/sysctl.conf

// QueryDMI attempts to extract information about a Device via DMI
func (p *Probe) QueryDMI(dev *model.Device, port int) (*model.DMI, error) {
	const cmd = "dmidecode"
	var (
		err    error
		output []string
	)

	var command = fmt.Sprintf("%s %s",
		p.getPrivCmd(dev),
		cmd)

	if output, err = p.executeCommand(dev, port, command); err != nil {
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

	return p.parseDMI(output)
} // func (p *Probe) QueryDMI(dev *model.Device, port int) (*model.DMI, error)

// To make testing and debugging easier, I delegate the actual processing
// of dmidecode(1) output to a separate function.

var (
	emptyLine = regexp.MustCompile(`^\s*$`)
	dmiField  = regexp.MustCompile(`^\s+([^:]+): (.*)`)
)

func (p *Probe) parseDMI(output []string) (*model.DMI, error) {
	var (
		idx, begin, end int
		dmi             = new(model.DMI)
	)

LINE:
	for idx < len(output) {
		var line = output[idx]
		// The Handle ID appears to be 2 byte, i.e. 4 hex digits
		// It has no intrinsic meaning, so the same ID can mean different
		// things on different Devices.
		if strings.HasPrefix(line, "Handle 0x") {
			begin = idx
			for !emptyLine.MatchString(output[idx]) && idx < len(output) {
				idx++
			}
			end = idx
		}

		switch output[begin+1] {
		case "Processor Information":
			// Apparently, this information can occur more than once
			if dmi.CPU != "" {
				idx++
				continue LINE
			}
		CPU:
			for lno := begin + 2; lno < end; lno++ {
				var match = dmiField.FindStringSubmatch(output[lno])
				if match == nil {
					continue CPU
				} else if match[1] == "Version" {
					dmi.CPU = match[2]
					idx++
					continue LINE
				}
			}
		case "System Information":
			if dmi.Vendor != "" || dmi.Model != "" {
				idx++
				continue LINE
			}
		SYSINFO:
			for lno := begin + 2; lno < end; lno++ {
				var match = dmiField.FindStringSubmatch(output[lno])
				if match == nil {
					continue SYSINFO
				}

				switch match[1] {
				case "Manufacturer":
					dmi.Vendor = match[2]
				case "Product Name":
					dmi.Model = match[2]
				case "Serial Number":
					dmi.Serial = match[2]
				}
			}
		}

		idx++
	}

	dmi.CPU = trim(dmi.CPU)
	dmi.Vendor = trim(dmi.Vendor)
	dmi.Serial = trim(dmi.Serial)
	dmi.Model = trim(dmi.Model)

	return dmi, nil
} // func (p *Probe) parseDMI(output []string) (*model.DMI, error)
