// /home/krylon/go/src/hertz/web/tmpl_data.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 05. 2020 by Benjamin Walkenhorst
// (c) 2020 Benjamin Walkenhorst
// Time-stamp: <2026-07-24 15:38:19 krylon>
//
// This file contains data structures to be passed to HTML templates.

package web

import "github.com/blicero/chili/model"

// nolint: unused
type tmplDataBase struct {
	Title    string
	Debug    bool
	URL      string
	Messages []string
}

// nolint: unused
type tmplDataIndex struct {
	tmplDataBase
	Networks []*model.Network
	Devices  []*model.Device
}

// nolint: unused
type tmplDataHosts struct {
	tmplDataBase
}

// nolint: unused
type tmplDataSingleHost struct {
	tmplDataBase
	Host *model.Device
}
