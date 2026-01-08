// /home/krylon/go/src/github.com/blicero/chili/database/qdb.go
// -*- mode: go; coding: utf-8; -*-
// Created on 07. 01. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-01-08 15:46:47 krylon>
//
// This files contains the SQL queries we intend to run on the database.

package database

import "github.com/blicero/chili/database/query"

var qdb = map[query.ID]string{
	query.NetAdd:            "INSERT INTO network (name, addr, added) VALUE (?, ?, ?) RETURNING id",
	query.NetUpdateLastScan: "UPDATE network SET last_scan = ? WHERE id = ?",
	query.NetUpdateName:     "UPDATE network SET name = ? WHERE id = ?",
	query.NetGetByID: `
SELECT
    name,
    addr,
    added,
    last_contact
FROM network
WHERE id = ?
`,
	query.NetGetAll: `
SELECT
    id,
    name,
    addr,
    added,
    last_contact
FROM network
`,
	query.DeviceAdd: `
INSERT INTO device (net_id, name, addr, added, class)
            VALUES (     ?,    ?,    ?,     ?,     ?)
RETURNING id
`,
	query.DeviceUpdateLastContact: "UPDATE device SET last_contact = ? WHERE id = ?",
	query.DeviceUpdateOS:          "UPDATE device SET os = ? WHERE id = ?",
	query.DeviceUpdateClass:       "UPDATE device SET class = ? WHERE id = ?",
	query.DeviceUpdateName:        "UPDATE device SET name = ? WHERE id = ?",
	query.DeviceUpdateActive:      "UPDATE device SET active = ? WHERE id = ?",
	query.DeviceGetByID: `
SELECT
    net_id,
    name,
    addr,
    added,
    last_contact,
    COALESCE(os, ''),
    class,
    active
FROM device
WHERE id = ?
`,
	query.DeviceGetByNet: `
SELECT
    id,
    name,
    addr,
    added,
    last_contact,
    COALESCE(os, ''),
    class,
    active
FROM device
WHERE net_id = ?
`,
	query.DeviceGetAll: `
SELECT
    id,
    net_id,
    name,
    addr,
    added,
    last_contact,
    COALESCE(os, ''),
    class,
    active
FROM device
`,
}
