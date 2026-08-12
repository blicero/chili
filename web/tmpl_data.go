// /home/krylon/go/src/hertz/web/tmpl_data.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 05. 2020 by Benjamin Walkenhorst
// (c) 2020 Benjamin Walkenhorst
// Time-stamp: <2026-08-12 10:08:05 krylon>
//
// This file contains data structures to be passed to HTML templates.

package web

import (
	"github.com/blicero/chili/model"
	"github.com/blicero/chili/model/attribute"
)

type notification struct {
	Attribute *model.Attribute
	Icon      string
	Text      string
}

type tmplDataBase struct {
	Title    string
	Debug    bool
	URL      string
	Messages []string
	Devices  []*model.Device
}

type tmplDataIndex struct {
	tmplDataBase
	Networks      []*model.Network
	Notifications map[int64][]notification
}

type tmplDataDeviceDetails struct {
	tmplDataBase
	Device        *model.Device
	Attributes    map[attribute.ID]*model.Attribute
	Notifications []notification
}

type tmplDataNetworkAll struct {
	tmplDataBase
	Networks []*model.Network
}

type tmplDataNetworkDetails struct {
	tmplDataBase
	Network *model.Network
}
