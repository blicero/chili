// /home/krylon/go/src/github.com/blicero/chili/main.go
// -*- mode: go; coding: utf-8; -*-
// Created on 07. 01. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-08-13 11:02:01 krylon>

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/blicero/chili/common"
	"github.com/blicero/chili/config"
	"github.com/blicero/chili/logdomain"
	"github.com/blicero/chili/probe"
	"github.com/blicero/chili/scanner"
	"github.com/blicero/chili/web"
	"github.com/hashicorp/logutils"
)

func main() {
	fmt.Printf("%s %s, built on %s\n",
		common.AppName,
		common.Version,
		common.BuildStamp.Format(common.TimestampFormat))

	time.Sleep(time.Second * 2)

	var (
		err                           error
		sc                            *scanner.Scanner
		p                             *probe.Probe
		srv                           *web.Server
		scanWorkerCnt, probeWorkerCnt int
		profOut, username, webAddr    string
		cfgPath                       string
		cfg                           *config.Config
		ticker                        *time.Ticker
		sigQ                          chan os.Signal
		pubKeys                       []string
	)

	flag.IntVar(
		&scanWorkerCnt,
		"scan",
		runtime.NumCPU(),
		"number of workers to run network scans in parallel")
	flag.IntVar(
		&probeWorkerCnt,
		"probe",
		runtime.NumCPU(),
		"number of workers to probe devices in parallel")
	flag.StringVar(
		&profOut,
		"profile",
		"",
		"file to write profiling data to (ignored if empty)")
	flag.StringVar(
		&username,
		"user",
		os.Getenv("USER"),
		"the user name to use for SSH connections")
	flag.StringVar(
		&webAddr,
		"addr",
		"",
		"the addr the web server listens on")
	flag.StringVar(
		&cfgPath,
		"cfg",
		common.CfgPath,
		"the path to the configuration file",
	)

	flag.Parse()

	if profOut != "" {
		var profH *os.File

		fmt.Printf("Writing profiling data to %s\n",
			profOut)

		if profH, err = os.Create(profOut); err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Failed to open %s: %s\n",
				profOut,
				err.Error())
			os.Exit(1)
		}

		defer profH.Close() // nolint: errcheck

		if err = pprof.StartCPUProfile(profH); err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Error starting CPU profile: %s\n",
				err.Error())
			os.Exit(1)
		}

		defer pprof.StopCPUProfile()
	}

	if cfg, err = config.Read(cfgPath); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Failed to read configuration file %s: %s\n",
			cfgPath,
			err.Error(),
		)
		os.Exit(1)
	} else if webAddr == "" {
		webAddr = cfg.Web.Address
	}

	common.PingCount = int(cfg.Ping.Count)
	common.PingTimeout = time.Second * time.Duration(cfg.Ping.Timeout)
	common.PingInterval = time.Millisecond * time.Duration(cfg.Ping.Interval*100)

	scanner.ScanInterval = time.Second * time.Duration(cfg.Scan.Interval)

	common.PackageLevels[logdomain.Common] = logutils.LogLevel(cfg.Loglevel.Common)
	common.PackageLevels[logdomain.Database] = logutils.LogLevel(cfg.Loglevel.Database)
	common.PackageLevels[logdomain.DBPool] = logutils.LogLevel(cfg.Loglevel.DBPool)
	common.PackageLevels[logdomain.Ping] = logutils.LogLevel(cfg.Loglevel.Ping)
	common.PackageLevels[logdomain.Probe] = logutils.LogLevel(cfg.Loglevel.Probe)
	common.PackageLevels[logdomain.Scanner] = logutils.LogLevel(cfg.Loglevel.Scanner)
	common.PackageLevels[logdomain.Scheduler] = logutils.LogLevel(cfg.Loglevel.Scheduler)
	common.PackageLevels[logdomain.Nexus] = logutils.LogLevel(cfg.Loglevel.Nexus)
	common.PackageLevels[logdomain.Web] = logutils.LogLevel(cfg.Loglevel.Web)

	pubKeys = []string{filepath.Join(os.Getenv("HOME"), ".ssh")}

	common.PackageLevels[logdomain.Scanner] = "INFO"

	if sc, err = scanner.New(scanWorkerCnt); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Failed to create Scanner: %s\n",
			err.Error())
		os.Exit(1)
	} else if p, err = probe.Create(probeWorkerCnt, username, pubKeys...); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Failed to create Probe: %s\n",
			err.Error())
		os.Exit(1)
	} else if srv, err = web.Create(webAddr, p.CmdQ); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Failed to create web server on %s: %s\n",
			webAddr,
			err.Error())
		os.Exit(1)
	}

	ticker = time.NewTicker(common.ActiveTimeout)
	defer ticker.Stop()

	sigQ = make(chan os.Signal, 1)
	signal.Notify(sigQ, os.Interrupt, syscall.SIGTERM)

	sc.Start()
	p.Start()
	go srv.Run()

	for {
		select {
		case <-ticker.C:
			continue
		case s := <-sigQ:
			fmt.Fprintf(
				os.Stderr,
				"Caught signal: %s\n",
				s)
			sc.Stop()
			p.Stop()
			return
		}
	}
} // func main()

// func findKeyFiles() ([]string, error) {
// 	var (
// 		err   error
// 		dh    *os.File
// 		path  string
// 		names []string
// 		files = make([]string, 0, 8)
// 	)

// 	path = filepath.Join(
// 		os.Getenv("HOME"),
// 		".ssh")

// 	if dh, err = os.Open(path); err != nil {
// 		return nil, err
// 	}

// 	defer dh.Close()

// 	if names, err = dh.Readdirnames(-1); err != nil {
// 		return nil, err
// 	}

// 	for _, file := range names {
// 		if strings.HasPrefix(file, "id_") && !strings.HasSuffix(file, ".pub") {
// 			files = append(
// 				files,
// 				filepath.Join(path, file))
// 		}
// 	}

// 	return files, nil
// } // func findKeyFiles() ([]string, error)
