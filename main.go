// /home/krylon/go/src/github.com/blicero/chili/main.go
// -*- mode: go; coding: utf-8; -*-
// Created on 07. 01. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-01-07 15:37:07 krylon>

package main

import (
	"fmt"

	"github.com/blicero/chili/common"
)

func main() {
	fmt.Printf("%s %s, built on %s\n",
		common.AppName,
		common.Version,
		common.BuildStamp.Format(common.TimestampFormat))

	fmt.Println("Nothing to see here, yet. Have a nice day.")
} // func main()
