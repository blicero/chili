// /home/krylon/go/src/github.com/blicero/hertz/web/ajax_types.go
// -*- mode: go; coding: utf-8; -*-
// Created on 03. 11. 2022 by Benjamin Walkenhorst
// (c) 2022 Benjamin Walkenhorst
// Time-stamp: <2026-08-11 16:37:53 krylon>

package web

import (
	"time"

	"github.com/blicero/chili/model"
)

type ajaxData struct {
	Status    bool      `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Payload   string    `json:"payload"`
}

type ajaxBeaconData struct {
	ajaxData
	Hostname string `json:"hostname"`
}

type ajaxResponseNetAdd struct {
	ajaxData
	Network *model.Network `json:"net"`
}
