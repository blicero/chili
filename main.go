// /home/krylon/go/src/github.com/blicero/chili/main.go
// -*- mode: go; coding: utf-8; -*-
// Created on 07. 01. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-07-08 13:38:57 krylon>

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/blicero/chili/common"
	"github.com/blicero/chili/scanner"
)

func main() {
	fmt.Printf("%s %s, built on %s\n",
		common.AppName,
		common.Version,
		common.BuildStamp.Format(common.TimestampFormat))

	var (
		err             error
		sc              *scanner.Scanner
		scanWorkerCount int
		ticker          *time.Ticker
		sigQ            chan os.Signal
	)

	flag.IntVar(
		&scanWorkerCount,
		"parallel",
		runtime.NumCPU(),
		"number of workers to run network scans in parallel")

	flag.Parse()

	if sc, err = scanner.New(scanWorkerCount); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Failed to create Scanner: %s\n",
			err.Error())
		os.Exit(1)
	}

	ticker = time.NewTicker(common.ActiveTimeout)
	defer ticker.Stop()

	sigQ = make(chan os.Signal, 1)
	signal.Notify(sigQ, os.Interrupt, syscall.SIGTERM)

	sc.Start()

	for {
		select {
		case <-ticker.C:
			continue
		case s := <-sigQ:
			fmt.Fprintf(
				os.Stderr,
				"Caught signal: %s\n",
				s)

			return
		}
	}
} // func main()
