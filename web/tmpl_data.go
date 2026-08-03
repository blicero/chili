// /home/krylon/go/src/hertz/web/tmpl_data.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 05. 2020 by Benjamin Walkenhorst
// (c) 2020 Benjamin Walkenhorst
// Time-stamp: <2026-08-03 12:46:38 krylon>
//
// This file contains data structures to be passed to HTML templates.

package web

import "github.com/blicero/chili/model"

type tmplDataBase struct {
	Title    string
	Debug    bool
	URL      string
	Messages []string
	Devices  []*model.Device
}

type tmplDataIndex struct {
	tmplDataBase
	Networks []*model.Network
}

// nolint: unused
type tmplDataDeviceDetails struct {
	tmplDataBase
	Device     *model.Device
	Attributes []*model.Attribute
}
