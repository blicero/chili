// /home/krylon/go/src/github.com/blicero/chili/probe/snmp.go
// -*- mode: go; coding: utf-8; -*-
// Created on 01. 08. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-08-03 12:26:46 krylon>

package probe

import (
	"time"

	"github.com/alouca/gosnmp"
	"github.com/blicero/chili/model"
)

const (
	snmpTimeout = 5
	portSNMP    = 161
)

// oids - quite stupidly - contains the OID values we are interested in.
var oids = map[string]string{
	"sysname": ".1.3.6.1.2.1.1.1.0",
	"uptime":  ".1.3.6.1.2.1.1.3.0",
}

// QuerySNMP attempts to query information about a device via SNMP.
func (p *Probe) QuerySNMP(dev *model.Device, port int) (model.SNMPInfo, error) {
	var (
		err    error
		snmp   *gosnmp.GoSNMP
		resp   *gosnmp.SnmpPacket
		result = make(map[string]string)
	)

	if snmp, err = gosnmp.NewGoSNMP(dev.Addr.String(), "public", gosnmp.Version2c, snmpTimeout); err != nil {
		p.log.Printf("[ERROR] Failed to talk SNMP to %s: %s\n",
			dev.Name,
			err.Error())
		return nil, err
	}

	for name, oid := range oids {
		if resp, err = snmp.Get(oid); err == nil {
		VARLOOP:
			for _, v := range resp.Variables {
				switch v.Type {
				case gosnmp.OctetString:
					result[name] = v.Value.(string)
					break VARLOOP
				case gosnmp.TimeTicks:
					var (
						tick int
						ok   bool
					)

					if tick, ok = v.Value.(int); ok {
						var uptime = time.Millisecond * time.Duration(tick*10)
						result[name] = uptime.String()
					}
				}
			}
		}
	}

	return model.SNMPInfo(result), nil
} // func (p *Probe) QuerySNMP(dev *model.Device, port int) (*model.SNMPInfo, error)
