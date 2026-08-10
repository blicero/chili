// /home/krylon/go/src/github.com/blicero/chili/model/model.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 07. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-08-10 13:46:41 krylon>

package model

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/blicero/chili/model/attribute"
	"github.com/blicero/chili/model/device"
)

// // http://play.golang.org/p/m8TNTtygK0
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func clone(ip net.IP) net.IP {
	var b = make(net.IP, len(ip))
	copy(b, ip)
	return b
} // func clone(ip net.IP) net.IP

// Network defines a range of IP addresses where our devices live.
type Network struct {
	ID       int64
	Name     string
	Addr     *net.IPNet
	Added    time.Time
	LastScan time.Time
}

// NewNet creates a new Network with the given name and IP address.
func NewNet(name, addr string) (*Network, error) {
	var err error
	var n = &Network{
		Name: name,
	}

	if _, n.Addr, err = net.ParseCIDR(addr); err != nil {
		return nil, err
	}

	return n, nil
} // func NewNet(name, addr string) (*Network, error)

// Enumerate generates all IP addresses for the Network and sends them through
// the channel passed in as its argument.
func (n *Network) Enumerate(q chan<- net.IP) error {
	go func() {
		var ip net.IP
		defer close(q)
		ip = make(net.IP, len(n.Addr.IP))
		copy(ip, n.Addr.IP)
		inc(ip)
		for n.Addr.Contains(ip) {
			var b = clone(ip)
			inc(ip)
			q <- b
		}
	}()

	return nil
} // func (n *Network) Enumerate(q chan<- net.IP)

// Device is a networked computer.
type Device struct {
	ID          int64
	NetID       int64
	Name        string
	Addr        net.IP
	Added       time.Time
	LastContact time.Time
	OS          string
	Class       device.Class
	Active      bool
}

// NewDevice creates a new Device.
func NewDevice(netID int64, name, addr string) (*Device, error) {
	var d = &Device{
		NetID:  netID,
		Name:   name,
		Active: true,
	}

	if d.Addr = net.ParseIP(addr); d.Addr == nil {
		return nil, fmt.Errorf("cannot parse address '%s'", addr)
	}

	return d, nil
} // func NewDevice(name, addr string) (*Device, error)

// Payload is the interface for different types of data we can query
// from Devices.
type Payload interface {
	fmt.Stringer
	Type() attribute.ID
	HTML() string
}

// Updates is a list of software packages that updates are available for.
type Updates []string

// Type returns the receiver's attribute type.
func (u Updates) Type() attribute.ID { return attribute.Updates }

func (u Updates) String() string {
	var (
		err error
		buf []byte
	)

	if buf, err = json.Marshal(u); err != nil {
		panic(err)
	}

	return string(buf)
} // func (u *Updates) String() string

// HTML renders the value to HTML. Hence the name.
func (u Updates) HTML() string {
	var sb strings.Builder

	sb.WriteString("<ol>\n")
	for _, pkg := range u {
		sb.WriteString("<li>")
		sb.WriteString(pkg)
		sb.WriteString("</li>\n")
	}
	sb.WriteString("</ol>\n")

	return sb.String()
} // func (u *Updates) HTML() string

// Packages is a list of software packages that are installed on a Device.
type Packages []string

// Type returns the receiver's attribute type.
func (p Packages) Type() attribute.ID { return attribute.Packages }

func (p Packages) String() string {
	var (
		err error
		buf []byte
	)

	if buf, err = json.Marshal(p); err != nil {
		panic(err)
	}

	return string(buf)
} // func (p Packages) String() string

// HTML renders the value to HTML. Hence the name.
func (p Packages) HTML() string {
	var sb strings.Builder

	sb.WriteString("<ol>\n")
	for _, pkg := range p {
		sb.WriteString("<li>")
		sb.WriteString(pkg)
		sb.WriteString("</li>\n")
	}
	sb.WriteString("</ol>\n")

	return sb.String()
} // func (u *Updates) HTML() string

// DiskSpace is the number of free bytes on a Device's root filesystem.
type DiskSpace int64

// Type returns the receiver's attribute type.
func (d DiskSpace) Type() attribute.ID { return attribute.DiskSpace }

func (d DiskSpace) String() string {
	return strconv.FormatInt(int64(d), 10)
} // func (d *DiskSpace) String() string

// HTML renders the value to HTML. Hence the name.
func (d DiskSpace) HTML() string {
	return fmt.Sprintf("%3d %%", d)
} // func (d DiskSpace) HTML() string

// Uptime is uptime of Device, and its system load.
type Uptime struct {
	Uptime time.Duration
	Load   [3]float64
}

// Type returns the receiver's attribute type.
func (u *Uptime) Type() attribute.ID { return attribute.Uptime }
func (u *Uptime) String() string {
	return fmt.Sprintf(`{ "Uptime": %d, "Load": [ %.1f, %.1f, %.1f ] }`,
		int64(u.Uptime),
		u.Load[0],
		u.Load[1],
		u.Load[2])
}

// HTML renders the value to HTML. Hence the name.
func (u *Uptime) HTML() string {
	var sb strings.Builder

	sb.WriteString("<ul>\n")
	sb.WriteString("<li>")
	sb.WriteString("<b>Uptime:</b> ")
	sb.WriteString(u.Uptime.String())
	sb.WriteString("</li>\n")
	sb.WriteString("<li>")
	sb.WriteString("<b>Load Average:</b> ")
	fmt.Fprintf(
		&sb,
		"%.1f / %.1f / %1.f</li>\n",
		u.Load[0],
		u.Load[1],
		u.Load[2])
	sb.WriteString("</ul>\n")

	return sb.String()
} // func (u *Uptime) HTML() string

// SNMPInfo represents data obtained via SNMP.
type SNMPInfo map[string]string

// Type returns the receiver's attribute type.
func (i SNMPInfo) Type() attribute.ID { return attribute.SNMP }
func (i SNMPInfo) String() string {
	if len(i) == 0 {
		return "{}"
	}

	var (
		sb     strings.Builder
		fields = make([]string, 0, len(i))
	)

	// XXX I naively assume that none of the strings we encounter here
	//     will be dangerous/malformed to be valid JSON.
	//     If/When I find out that was overly optimistic, I will look into
	//     properly handling that. I think html/template has a helper
	//     function to deal with this.

	sb.WriteString("{\n")
	for k, v := range i {
		var s = fmt.Sprintf(
			`"%s": "%s"`,
			k,
			v)
		fields = append(fields, s)
	}

	var body = strings.Join(fields, ",\n")
	sb.WriteString(body)
	sb.WriteString("\n}\n")

	return sb.String()
} // func (i SNMPInfo) String() string

// HTML returns a HTML representation of the receiver.
func (i SNMPInfo) HTML() string {
	var sb strings.Builder

	// ...
	sb.WriteString(`<table class="horizontal table table-striped">`)
	sb.WriteString("\n")

	for k, v := range i {
		fmt.Fprintf(
			&sb,
			"<tr>\n\t<th>%s</th>\n\t<td>%s</td>\n",
			template.JSEscapeString(k),
			template.JSEscapeString(v),
		)
	}

	sb.WriteString("</table>\n")

	return sb.String()
} // func (i SNMPInfo) HTML() string

// Services represents running and failed services (aka daemons) on a Device.
type Services struct {
	Running []string `json:"running"`
	Failed  []string `json:"failed"`
}

func (s *Services) Type() attribute.ID { return attribute.Services }
func (s *Services) String() string {
	var (
		err error
		buf []byte
	)

	if buf, err = json.Marshal(s); err != nil {
		panic(err)
	}

	return string(buf)
} // func (s *Services) String() string

// HTML returns a HTML representation of the receiver.
func (s *Services) HTML() string {
	var sb strings.Builder

	if len(s.Failed) > 0 {
		sb.WriteString(`<div>
  <table class="table table-striped caption-top failed">
    <caption>Failed Services</caption>
    <thead>
      <tr>
        <th>#</th>
        <th>Name</th>
      </tr>
    </thead>
    <tbody>
`)

		for i, svc := range s.Failed {
			fmt.Fprintf(
				&sb,
				`<tr>
  <td>%d</td>
  <td>%s</td>
</tr>
`,
				i+1,
				svc)
		}

		sb.WriteString(`
    </tbody>
  </table>
</div>`)
	}

	sb.WriteString(`<div>
  <table class="table table-striped caption-top">
  <caption>Running Services</caption>
  <thead>
    <tr>
      <th>#</th>
      <th>Name</th>
    </tr>
  </thead>
  <tbody>
`)

	for i, svc := range s.Running {
		fmt.Fprintf(
			&sb,
			`<tr>
  <td>%d</td>
  <td>%s</td>
</tr>
`,
			i+1,
			svc)
	}

	sb.WriteString(`
    </tbody>
  </table>
</div>
`)

	return sb.String()
} // func (s *Services) HTML() string

// DMI contains information extracted from a Device's DMI... yeah.
type DMI struct {
	Vendor string `json:"vendor"`
	Model  string `json:"model"`
	Serial string `json:"serial"`
	CPU    string `json:"cpu"`
}

func (d *DMI) Type() attribute.ID { return attribute.DMI }
func (d *DMI) String() string {
	var buf []byte

	buf, _ = json.Marshal(d)
	return string(buf)
} // func (d *DMI) String() string

// HTML returns a HTML representation of the receiver.
func (d *DMI) HTML() string {
	return d.String()
} // func (d *DMI) HTML() string

// Attribute is a property of a Device that we can query/measure.
type Attribute struct {
	ID        int64
	DevID     int64
	Timestamp time.Time
	Type      attribute.ID
	Value     Payload
}
