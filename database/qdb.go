// /home/krylon/go/src/github.com/blicero/chili/database/qdb.go
// -*- mode: go; coding: utf-8; -*-
// Created on 07. 01. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-07-11 13:35:33 krylon>
//
// This files contains the SQL queries we intend to run on the database.

package database

import "github.com/blicero/chili/database/query"

var qdb = map[query.ID]string{
	query.NetAdd:         "INSERT INTO network (name, addr, added) VALUES (?, ?, ?) RETURNING id",
	query.NetSetLastScan: "UPDATE network SET last_scan = ? WHERE id = ?",
	query.NetSetName:     "UPDATE network SET name = ? WHERE id = ?",
	query.NetGetByID: `
SELECT
    name,
    addr,
    added,
    last_scan
FROM network
WHERE id = ?
`,
	query.NetGetAll: `
SELECT
    id,
    name,
    addr,
    added,
    last_scan
FROM network
`,
	query.DeviceAdd: `
INSERT INTO device (net_id, name, addr, added, class)
            VALUES (     ?,    ?,    ?,     ?,     ?)
ON CONFLICT (addr) DO UPDATE SET last_contact = unixepoch()
RETURNING id
`,
	query.DeviceSetLastContact: "UPDATE device SET last_contact = ? WHERE id = ?",
	query.DeviceSetOS:          "UPDATE device SET os = ? WHERE id = ?",
	query.DeviceSetClass:       "UPDATE device SET class = ? WHERE id = ?",
	query.DeviceSetName:        "UPDATE device SET name = ? WHERE id = ?",
	query.DeviceSetActive:      "UPDATE device SET active = ? WHERE id = ?",
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
	query.AttributeAdd: `
INSERT INTO attribute (dev_id, timestamp, atype, value)
VALUES                (     ?,         ?,     ?,     ?)
RETURNING id
`,
	query.AttributeGetByDevice: `
WITH dattribute AS (
SELECT
    id,
    row_number() OVER (PARTITION BY atype ORDER BY timestamp DESC) AS rid,
    atype,
    timestamp,
    value
FROM attribute
WHERE dev_id = ?
ORDER BY timestamp DESC
)

SELECT id, atype, timestamp, value
FROM dattribute
WHERE rid = 1
`,
	query.AttributeGetByDeviceType: `
SELECT
    id,
    timestamp,
    value
FROM attribute
WHERE dev_id = ? AND atype = ?
ORDER BY timestamp DESC
LIMIT ?
`,
	query.AttributeGetByType: `
WITH dattribute AS (
SELECT
    id,
    row_number() OVER (PARTITION BY dev_id ORDER BY timestamp DESC) AS rid,
    dev_id,
    timestamp,
    value
FROM attribute
WHERE atype = ?
ORDER BY timestamp DESC
)

SELECT id, dev_id, timestamp, value
FROM dattribute
WHERE rid = 1
`,
}
