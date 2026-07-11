// /home/krylon/go/src/github.com/blicero/chili/database/qinit.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 01. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-07-11 11:17:02 krylon>
//
// This file contains the SQL queries to initialize an empty database.

package database

// nolint: unused
var qInit = []string{
	`
CREATE TABLE network (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    addr TEXT UNIQUE NOT NULL,
    added INTEGER NOT NULL,
    last_scan INTEGER NOT NULL DEFAULT 0,
    CHECK (last_scan >= 0)
) STRICT
`,
	"CREATE INDEX net_scan_idx ON network (last_scan)",
	`
CREATE TABLE device (
    id INTEGER PRIMARY KEY,
    net_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    addr TEXT UNIQUE NOT NULL,
    added INTEGER NOT NULL,
    last_contact INTEGER NOT NULL DEFAULT 0,
    os TEXT,
    class INTEGER NOT NULL,
    active INTEGER NOT NULL DEFAULT 1,
    FOREIGN KEY (net_id) REFERENCES network (id)
        ON UPDATE RESTRICT
        ON DELETE CASCADE,
    CHECK (last_contact >= 0)
) STRICT
`,
	"CREATE INDEX dev_net_idx ON device (net_id)",
	"CREATE INDEX dev_contact_idx ON device (last_contact)",
	`
CREATE TABLE attribute (
    id INTEGER PRIMARY KEY,
    dev_id INTEGER NOT NULL,
    timestamp INTEGER NOT NULL,
    atype INTEGER NOT NULL,
    value TEXT NOT NULL,
    FOREIGN KEY (dev_id) REFERENCES device (id)
       ON UPDATE RESTRICT
       ON DELETE CASCADE,
    UNIQUE (dev_id, timestamp, atype),
    CHECK (json_valid(value))
) STRICT
`,
	"CREATE INDEX att_dev_idx ON attribute (dev_id)",
	"CREATE INDEX att_time_idx ON attribute (timestamp)",
	"CREATE INDEX att_type_idx ON attribute (atype)",
}
