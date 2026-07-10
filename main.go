// /home/krylon/go/src/github.com/blicero/chili/main.go
// -*- mode: go; coding: utf-8; -*-
// Created on 07. 01. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-07-10 12:09:26 krylon>

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"syscall"
	"time"

	"github.com/blicero/chili/common"
	"github.com/blicero/chili/control"
	"github.com/blicero/chili/probe"
	"github.com/blicero/chili/scanner"
)

func main() {
	fmt.Printf("%s %s, built on %s\n",
		common.AppName,
		common.Version,
		common.BuildStamp.Format(common.TimestampFormat))

	var (
		err                           error
		sc                            *scanner.Scanner
		p                             *probe.Probe
		scanWorkerCnt, probeWorkerCnt int
		profOut, username             string
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

	if pubKeys, err = findKeyFiles(); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Failed to find SSH keys: %s\n",
			err.Error())
		os.Exit(1)
	} else if sc, err = scanner.New(scanWorkerCnt); err != nil {
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
	}

	ticker = time.NewTicker(common.ActiveTimeout)
	defer ticker.Stop()

	sigQ = make(chan os.Signal, 1)
	signal.Notify(sigQ, os.Interrupt, syscall.SIGTERM)

	sc.Start()
	p.Start()

	for {
		select {
		case <-ticker.C:
			continue
		case s := <-sigQ:
			fmt.Fprintf(
				os.Stderr,
				"Caught signal: %s\n",
				s)
			sc.CmdQ <- control.Stop
			p.Stop()
			return
		}
	}
} // func main()

func findKeyFiles() ([]string, error) {
	var (
		err   error
		dh    *os.File
		path  string
		names []string
		files = make([]string, 0, 8)
	)

	path = filepath.Join(
		os.Getenv("HOME"),
		".ssh")

	if dh, err = os.Open(path); err != nil {
		return nil, err
	}

	defer dh.Close()

	if names, err = dh.Readdirnames(-1); err != nil {
		return nil, err
	}

	for _, file := range names {
		if strings.HasPrefix(file, "id_") && !strings.HasSuffix(file, ".pub") {
			files = append(
				files,
				filepath.Join(path, file))
		}
	}

	return files, nil
} // func findKeyFiles() ([]string, error)
