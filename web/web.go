// /home/krylon/go/src/github.com/blicero/chili/web/web.go
// -*- mode: go; coding: utf-8; -*-
// Created on 22. 07. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-07-22 12:48:32 krylon>

package web

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"sync"

	"github.com/blicero/chili/database"
	"github.com/gorilla/mux"
)

//go:embed assets
var assets embed.FS

type Web struct {
	addr      string
	log       *log.Logger
	pool      *database.Pool
	lock      sync.RWMutex
	router    *mux.Router
	tmpl      *template.Template
	web       http.Server
	mimeTypes map[string]string
}
