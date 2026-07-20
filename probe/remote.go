// /home/krylon/go/src/github.com/blicero/chili/probe/remote.go
// -*- mode: go; coding: utf-8; -*-
// Created on 10. 07. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-07-20 11:40:51 krylon>

package probe

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/blicero/chili/common"
	"github.com/blicero/chili/model"
	probing "github.com/prometheus-community/pro-bing"
	"golang.org/x/crypto/ssh"
)

var (
	// ErrPingOffline indicates a Device did not respond to a ping.
	ErrPingOffline = errors.New("device did not respond to ping")
	// ErrNoData indicates that required data is missing or incomplete.
	ErrNoData = errors.New("missing data")
)

const (
	osReleaseCmd = "/bin/cat /etc/os-release"
	unameCmd     = "/usr/bin/uname -s"
	exit100      = "Process exited with status 100"
	exit2        = "Process exited with status 2"
	portSSH      = 22
	pkgSep       = "\t"
)

func (p *Probe) pingDevice(dev *model.Device) bool {
	var (
		err    error
		pinger *probing.Pinger
	)
	pinger = probing.New(dev.Addr.String())
	pinger.Count = common.PingCount
	pinger.Interval = common.PingInterval
	pinger.Timeout = common.PingCount * common.PingInterval * 2

	if err = pinger.Run(); err != nil {
		p.log.Printf("[ERROR] Error pinging %s: %s\n",
			dev.Addr,
			err.Error())
		return false
	}

	var stats = pinger.Statistics()

	return stats.PacketLoss < 95
} // func (p *Probe) pingDevice(dev *model.Device) bool

// nolint: unused
func (p *Probe) executeCommand(d *model.Device, port int, cmd string) ([]string, error) {
	var (
		err     error
		session *ssh.Session
	)

	// 05. 08. 2025
	// I get a panic originating in NewSession when connecting to a Device that is offline.
	if session, err = p.getSession(d, port); err != nil {
		if err == ErrPingOffline {
			return nil, err
		}
		var ex = fmt.Errorf("failed to create SSH session for %s: %w",
			d.Name,
			err)
		p.log.Printf("[ERROR] %s\n", ex.Error())
		return nil, ex
	}

	defer session.Close() // nolint: errcheck

	var rawOutput []byte

	if rawOutput, err = session.CombinedOutput(cmd); err != nil {
		if strings.Contains(cmd, "dnf") &&
			strings.HasPrefix(err.Error(), exit100) {
			// dnf check-upgrade exits with status 100 if there
			// are updates available.
		} else if strings.Contains(cmd, "checkupdates") &&
			strings.HasPrefix(err.Error(), exit2) {
			// checkupdates on Arch exits with status 2 if no
			// updates are available.
			return nil, nil
		} else if d.OS == "FreeBSD" && strings.HasPrefix(err.Error(), "Process exited with status 2") {
			// On FreeBSD, "freebsd-update updatesready" exits
			// with status 2 if no updates are available.
			return nil, nil
		} else {
			var ex = fmt.Errorf("failed to execute command on %s: %w\n>>> Command: %s",
				d.Name,
				err,
				cmd)
			p.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		}
	}

	var lines = strings.Split(string(rawOutput), "\n")

	return lines, nil
} // func (p *Probe) executeCommand(d *model.Device, port int, cmd string) ([]string, error)

func (p *Probe) getClient(d *model.Device, port int) (*ssh.Client, error) {
	var (
		err error
		ok  bool
		c   *ssh.Client
	)

	p.lock.Lock()
	defer p.lock.Unlock()

	if c, ok = p.clients[d.ID]; ok {
		return c, nil
	} else if c, err = p.connect(d, port); err != nil {
		return nil, err
	} else if c == nil {
		var ex = fmt.Errorf("probe.connect did not return an error, but connection to %s is nil",
			d.Name)
		p.log.Printf("[ERROR] %s\n", ex.Error())
		return nil, ex
	}

	p.clients[d.ID] = c
	return c, nil
} // func (p *Probe) getClient(d *model.Device, port int) (*ssh.Client, error)

// nolint: unused
func (p *Probe) getSession(d *model.Device, port int) (s *ssh.Session, e error) {
	var (
		err    error
		client *ssh.Client
		sess   *ssh.Session
	)

	defer func() {
		if ex := recover(); ex != nil {
			p.log.Printf("[ERROR] Panic trying to get SSH session for %s: %s\n",
				d.Name,
				ex)
			s = nil
			e = ex.(error)
		}
	}()

	if client, err = p.getClient(d, port); err != nil {
		if err == ErrPingOffline {
			return nil, err
		}
		var ex = fmt.Errorf("failed to get SSH client for %s: %w",
			d.Name,
			err)
		p.log.Printf("[ERROR] %s\n", ex.Error())
		return nil, ex
	} else if sess, err = client.NewSession(); err != nil {
		if errors.Is(err, io.EOF) {
			_ = p.disconnect(d)
		}
		var ex = fmt.Errorf("failed to create new SSH session for %s: %w",
			d.Name,
			err)
		p.log.Printf("[ERROR] %s\n", ex.Error())
		return nil, ex
	}

	return sess, nil
} // func (p *Probe) getSession(d *model.Device, port int) (*ssh.Session, error)

// nolint: unused
func (p *Probe) disconnect(d *model.Device) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	var (
		err error
		ok  bool
		c   *ssh.Client
	)

	if c, ok = p.clients[d.ID]; !ok {
		return nil
	}

	delete(p.clients, d.ID)

	if c == nil {
		goto END
	}

	if err = c.Close(); err != nil {
		p.log.Printf("[ERROR] Failed to close SSH connection to %s: %s\n",
			d.Name,
			err.Error())
	}

END:
	return err
} // func (p *Probe) disconnect(d *model.Device) error

func (p *Probe) connect(d *model.Device, port int) (*ssh.Client, error) {
	var (
		err    error
		client *ssh.Client
	)

	var addr = fmt.Sprintf("%s:%d",
		d.Addr,
		port)

	if client, err = ssh.Dial("tcp", addr, p.cfg); err != nil {
		p.log.Printf("[ERROR] Failed to connect to %s at %s: %s\n",
			d.Name,
			addr,
			err.Error())
	} else {
		return client, nil
	}

	return nil, ErrPingOffline
} // func (p *Probe) connect(d *model.Device, port int) (*ssh.Client, error)

// QueryOS attempts to find out what operating system the device runs.
func (p *Probe) QueryOS(d *model.Device, port int) (string, error) {
	var (
		err     error
		client  *ssh.Client
		session *ssh.Session
	)

	if client, err = p.getClient(d, port); err != nil {
		return "", err
	} else if client == nil {
		p.log.Printf("[ERROR] Could not connect to %s on any address.\n",
			d.Name)
		return "", err
	}

	// defer client.Close()

	if session, err = client.NewSession(); err != nil {
		var ex = fmt.Errorf("failed to create SSH session with %s: %w",
			d.Name,
			err)
		p.log.Printf("[ERROR] %s\n", ex.Error())
		return "", ex
	}

	defer session.Close() // nolint: errcheck

	var rawOutput []byte

	if rawOutput, err = session.CombinedOutput(unameCmd); err != nil {
		var ex = fmt.Errorf("failed to run %q on %s: %w",
			unameCmd,
			d.Name,
			err)
		p.log.Printf("[ERROR] %s\n", ex.Error())
		return "", ex
	}

	rawOutput = bytes.Trim(rawOutput, "\n\t ")

	var kernel = string(rawOutput)

	// If the kernel isn't Linux, it almost certainly is some kind of BSD, in which
	// case we have the information we want.
	//
	// If it is, we try to read /etc/os-release to determine what distro we
	// are dealing with.
	if kernel != "Linux" {
		return kernel, nil
	} else if session, err = client.NewSession(); err != nil {
		var ex = fmt.Errorf("failed to create SSH session on %s: %w",
			d.Name,
			err)
		p.log.Printf("[ERROR] %s\n", ex.Error())
	}

	defer session.Close() // nolint: errcheck

	if rawOutput, err = session.CombinedOutput(osReleaseCmd); err != nil {
		var ex = fmt.Errorf("failed to cat(1) /etc/os-release on %s: %w",
			d.Name,
			err)
		p.log.Printf("[ERROR] %s\n", ex.Error())
		return "", ex
	}

	var (
		osname      string
		releaseInfo = string(rawOutput)
	)

	for l := range strings.Lines(releaseInfo) {
		if strings.HasPrefix(l, "NAME=") {
			osname = strings.Trim(l[5:], "\"\n\t ")
		}
	}

	return osname, nil
} // func (p *Probe) QueryOS(d *model.Device) (string, error)

// Sample output:
// 18:01:18  2 Tage  0:22 an,  2 Benutzer,  Durchschnittslast: 1,08, 0,98, 0,94
// 6:02PM  up 56 days,  5:16, 4 users, load averages: 0.00, 0.01, 0.00

var uptimePat = regexp.MustCompile(
	`:\s+(\d+[,.]\d+),?\s+(\d+[,.]\d+),?\s+(\d+[,.]\d+)$`)

// QueryUptime attempts to extract the system load average from the given Device.
func (p *Probe) QueryUptime(d *model.Device, port int) (*model.Uptime, error) {
	const cmd = "/usr/bin/uptime"
	var (
		err   error
		res   []string
		match []string
		up    = new(model.Uptime)
	)

	if res, err = p.executeCommand(d, port, cmd); err != nil {
		if err == ErrPingOffline {
			return nil, err
		}
		var ex = fmt.Errorf("failed to query uptime/loadavg on %s: %w",
			d.Name,
			err)
		p.log.Printf("[ERROR] %s\n", ex.Error())
		return nil, ex
	} else if res == nil {
		var ex = fmt.Errorf("querying uptime on %s did not return a value",
			d.Name)
		p.log.Printf("[CANTHAPPEN] %s\n", ex.Error())
		return nil, ex
	} else if match = uptimePat.FindStringSubmatch(res[0]); match == nil {
		var ex = fmt.Errorf("cannot parse the output of uptime(1) from %s: %q",
			d.Name,
			res[0])
		p.log.Printf("[ERROR] %s\n", ex.Error())
		return nil, ex
	} else if len(match) > 0 {
		for idx, val := range match[1:] {
			var (
				s = strings.ReplaceAll(val, ",", ".")
				f float64
			)

			if f, err = strconv.ParseFloat(s, 64); err != nil {
				var ex = fmt.Errorf("cannot parse load avg %q: %w",
					s,
					err)
				p.log.Printf("[ERROR] %s\n", ex.Error())
				return nil, ex
			}

			up.Load[idx] = f
		}
	}

	// ...

	return up, nil
} // func (p *Probe) QueryUptime(d *model.Device, port int) (*model.Uptime, error)

var dfPat = regexp.MustCompile(`(\d+)%`)

// QueryDiskFree queries a Device for the free disk space on its root filesystem.
func (p *Probe) QueryDiskFree(d *model.Device, port int) (int64, error) {
	const cmd = "env LC_ALL=en_EN.UTF-8 df -h /"
	var (
		err        error
		res, match []string
		used, free int64
	)

	if res, err = p.executeCommand(d, port, cmd); err != nil {
		if err == ErrPingOffline {
			return 0, err
		}
		var ex = fmt.Errorf("failed to query free disk space on %s: %w",
			d.Name,
			err)
		p.log.Printf("[ERROR] %s\n", ex.Error())
		return 0, ex
	} else if len(res) < 2 {
		var ex = fmt.Errorf("cannot parse output of \"df -h\" on %s: %s",
			d.Name,
			strings.Join(res, "\n"))
		p.log.Printf("[ERROR] %s\n", ex.Error())
		return 0, ex
	} else if match = dfPat.FindStringSubmatch(res[1]); match == nil {
		var ex = fmt.Errorf("cannot parse output of \"df -h\" on %s: %s",
			d.Name,
			strings.Join(res, "\n"))
		p.log.Printf("[ERROR] %s\n", ex.Error())
		return 0, ex
	} else if used, err = strconv.ParseInt(match[1], 10, 64); err != nil {
		var ex = fmt.Errorf("cannot parse free disk space on %s: %q - %w",
			d.Name,
			match[1],
			err)
		p.log.Printf("[ERROR] %s\n", ex.Error())
		return 0, ex
	}

	free = 100 - used

	return free, nil
} // func (p *Probe) QueryDiskFree(d *model.Device, port int) (int64, error)
