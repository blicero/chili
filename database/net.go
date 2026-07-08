// /home/krylon/go/src/github.com/blicero/chili/database/net.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 07. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-07-08 12:24:44 krylon>

package database

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/blicero/chili/database/query"
	"github.com/blicero/chili/model"
)

////////////////////////////////////////////////////////////////////////
// Network /////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////

// NetAdd adds a Network to the Database.
func (db *Database) NetAdd(n *model.Network) error {
	const qid query.ID = query.NetAdd
	var (
		err  error
		stmt *sql.Stmt
		now  = time.Now()
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Failed to prepare query %s: %s\n",
			qid,
			err.Error())
		panic(err)
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

	var rows *sql.Rows

EXEC_QUERY:
	if rows, err = stmt.Query(n.Name, n.Addr.String(), now.Unix()); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		} else {
			err = fmt.Errorf("cannot add Network %s to Database: %w",
				n.Addr,
				err)
			db.log.Printf("[ERROR] %s\n", err.Error())
			return err
		}
	}

	var id int64

	defer rows.Close() // nolint: errcheck

	if !rows.Next() {
		// CANTHAPPEN
		msg := fmt.Sprintf("Query %s did not return an ID", qid)
		db.log.Printf("[ERROR] %s\n",
			msg)
		return errors.New(msg)
	} else if err = rows.Scan(&id); err != nil {
		var ex = fmt.Errorf("failed to get ID for newly added Network %s: %w",
			n.Addr,
			err)
		db.log.Printf("[ERROR] %s\n", ex.Error())
		return ex
	}

	n.ID = id
	n.Added = now

	return nil
} // func (db *Database) NetAdd(n *model.Network) error

// NetUpdateLastScan updates a Network's scan timestamp.
func (db *Database) NetUpdateLastScan(n *model.Network, t time.Time) error {
	const qid query.ID = query.NetUpdateLastScan
	var (
		err  error
		stmt *sql.Stmt
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Failed to prepare query %s: %s\n",
			qid,
			err.Error())
		panic(err)
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

EXEC_QUERY:
	if _, err = stmt.Exec(t.Unix(), n.ID); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		} else {
			err = fmt.Errorf("cannot update scan timestamp of Network %s: %w",
				n.Addr,
				err)
			db.log.Printf("[ERROR] %s\n", err.Error())
			return err
		}
	}

	n.LastScan = t
	return nil
} // func (db *Database) NetUpdateLastScan(n *model.Network, t time.Time) error

// NetUpdateName updates a Network's name.
func (db *Database) NetUpdateName(n *model.Network, name string) error {
	const qid query.ID = query.NetUpdateName
	var (
		err  error
		stmt *sql.Stmt
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Failed to prepare query %s: %s\n",
			qid,
			err.Error())
		panic(err)
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

EXEC_QUERY:
	if _, err = stmt.Exec(name, n.ID); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		} else {
			err = fmt.Errorf("cannot update scan timestamp of Network %s: %w",
				n.Addr,
				err)
			db.log.Printf("[ERROR] %s\n", err.Error())
			return err
		}
	}

	n.Name = name
	return nil
} // func (db *Database) NetUpdateName(n *model.Network, name string) error

func (db *Database) NetGetByID(id int64) (*model.Network, error) {
	const qid query.ID = query.NetGetByID
	var (
		err  error
		stmt *sql.Stmt
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid,
			err.Error())
		return nil, err
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

	var rows *sql.Rows

EXEC_QUERY:
	if rows, err = stmt.Query(id); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return nil, err
	}

	defer rows.Close() // nolint: errcheck,gosec

	if rows.Next() {
		var (
			addstamp, contact int64
			astr              string
			n                 = &model.Network{ID: id}
		)

		if err = rows.Scan(&n.Name, &astr, &addstamp, &contact); err != nil {
			var ex = fmt.Errorf("failed to scan row: %w", err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		} else if _, n.Addr, err = net.ParseCIDR(astr); err != nil {
			var ex = fmt.Errorf("cannot parse Network address %q: %w",
				astr,
				err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		}

		n.LastScan = time.Unix(contact, 0)
		n.Added = time.Unix(addstamp, 0)

		return n, nil
	}

	return nil, nil
} // func (db *Database) NetGetByID(id int64) (*model.Network, error)

// NetGetAll loads all Networks from the Database, in no particular order.
func (db *Database) NetGetAll() ([]*model.Network, error) {
	const qid query.ID = query.NetGetAll
	var (
		err  error
		stmt *sql.Stmt
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid,
			err.Error())
		return nil, err
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

	var rows *sql.Rows

EXEC_QUERY:
	if rows, err = stmt.Query(); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return nil, err
	}

	defer rows.Close() // nolint: errcheck,gosec
	var networks = make([]*model.Network, 0, 8)

	for rows.Next() {
		var (
			addstamp, contact int64
			astr              string
			n                 = new(model.Network)
		)

		if err = rows.Scan(&n.ID, &n.Name, &astr, &addstamp, &contact); err != nil {
			var ex = fmt.Errorf("failed to scan row: %w", err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		} else if _, n.Addr, err = net.ParseCIDR(astr); err != nil {
			var ex = fmt.Errorf("cannot parse Network address %q: %w",
				astr,
				err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		}

		n.LastScan = time.Unix(contact, 0)
		n.Added = time.Unix(addstamp, 0)

		networks = append(networks, n)
	}

	return networks, nil
} // func (db *Database) NetGetAll() ([]*model.Network, error)
