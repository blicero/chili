// /home/krylon/go/src/hertz/web/helpers_web.go
// -*- mode: go; coding: utf-8; -*-
// Created on 04. 09. 2019 by Benjamin Walkenhorst
// (c) 2019 Benjamin Walkenhorst
// Time-stamp: <2026-08-06 10:23:06 krylon>
//
// Helper functions for use by the HTTP request handlers

package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/blicero/chili/common"
	"github.com/blicero/chili/database"
	"github.com/blicero/chili/model"
	"github.com/blicero/chili/model/attribute"
)

func errJSON(msg string) []byte {
	var res = fmt.Sprintf(`{ "Status": false, "Message": %q }`,
		jsonEscape(msg))

	return []byte(res)
} // func errJSON(msg string) []byte

func jsonEscape(i string) string { // nolint: unused
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	// Trim the beginning and trailing " character
	return string(b[1 : len(b)-1])
}

func (srv *Server) baseData(title string, r *http.Request) tmplDataBase { // nolint: unused
	return tmplDataBase{
		Title: title,
		Debug: common.Debug,
		URL:   r.URL.String(),
	}
} // func (srv *Server) baseData(title string, r *http.Request) tmplDataBase

func (srv *Server) getNotifications(dev *model.Device) ([]notification, error) {
	var (
		err        error
		db         *database.Database
		attributes []*model.Attribute
	)

	db = srv.pool.Get()
	defer srv.pool.Put(db)

	if attributes, err = db.AttributeGetMostRecent(dev); err != nil {
		srv.log.Printf("[ERROR] Failed to get Attributes for %s: %s\n",
			dev.Name,
			err.Error())
		return nil, err
	}

	var notifications = make([]notification, 0)

	for _, att := range attributes {
		switch att.Type {
		case attribute.Updates:
			if updates, ok := att.Value.(model.Updates); ok && len(updates) > 0 {
				var n = notification{
					Attribute: att,
					Text:      fmt.Sprintf("%d pending Updates", len(updates)),
					Icon:      "status_updates.png",
				}

				notifications = append(notifications, n)
			}
		case attribute.Services:
			if svc, ok := att.Value.(*model.Services); ok && len(svc.Failed) > 0 {
				var n = notification{
					Attribute: att,
					Text: fmt.Sprintf("%d failed Services",
						len(svc.Failed)),
					Icon: "status_fail.png",
				}

				notifications = append(notifications, n)
			}
		}
	}

	return notifications, nil
} // func (srv *Server) getNotifications(dev *model.Device) ([]notification, error)

// func generateChartTicks(n int) []float64 {
// 	var (
// 		ticks = make([]float64, n)
// 	)

// 	for i := range n {
// 		ticks[i] = float64(i)
// 	}

// 	return ticks
// } // func generateChartTicks(n int) []float64
