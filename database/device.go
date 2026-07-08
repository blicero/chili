// /home/krylon/go/src/github.com/blicero/chili/database/device.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 07. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-07-08 12:25:03 krylon>

package database

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/blicero/chili/database/query"
	"github.com/blicero/chili/model"
	"github.com/blicero/chili/model/device"
)

////////////////////////////////////////////////////////////////////////
// Device //////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////

// DeviceAdd adds a Device to the Database.
func (db *Database) DeviceAdd(d *model.Device) error {
	const qid query.ID = query.DeviceAdd
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
	if rows, err = stmt.Query(d.NetID, d.Name, d.Addr.String(), now.Unix(), d.Class); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		} else {
			err = fmt.Errorf("cannot add Network %s to Database: %w",
				d.Addr,
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
			d.Addr,
			err)
		db.log.Printf("[ERROR] %s\n", ex.Error())
		return ex
	}

	d.ID = id
	d.Added = now

	return nil
} // func (db *Database) DeviceAdd(d *model.Device) error

// DeviceUpdateLastContact updates a Device's LastContact timestamp.
func (db *Database) DeviceUpdateLastContact(d *model.Device, contact time.Time) error {
	const qid query.ID = query.DeviceUpdateLastContact
	var (
		err  error
		stmt *sql.Stmt
	)

	if d == nil {
		db.log.Println("[ERROR] Cannot update Device, pointer is nil")
		return ErrInvalidValue
	} else if d.ID <= 0 {
		db.log.Printf("[ERROR] Cannot update Device %s (%s) - missing ID!\n",
			d.Name,
			d.Addr)
		return ErrInvalidValue
	} else if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Failed to prepare query %s: %s\n",
			qid,
			err.Error())
		panic(err)
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

EXEC_QUERY:
	if _, err = stmt.Exec(contact.Unix(), d.ID); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		} else {
			err = fmt.Errorf("cannot update contact timestamp of Device %s: %w",
				d.Addr,
				err)
			db.log.Printf("[ERROR] %s\n", err.Error())
			return err
		}
	}

	d.LastContact = contact
	return nil
} // func (db *Database) DeviceUpdateLastContact(d *model.Device, contact time.Time) error

// DeviceUpdateOS updates a Device's OS
func (db *Database) DeviceUpdateOS(d *model.Device, osName string) error {
	const qid query.ID = query.DeviceUpdateOS
	var (
		err  error
		stmt *sql.Stmt
	)

	if d == nil {
		db.log.Println("[ERROR] Cannot update Device, pointer is nil")
		return ErrInvalidValue
	} else if d.ID <= 0 {
		db.log.Printf("[ERROR] Cannot update Device %s (%s) - missing ID!\n",
			d.Name,
			d.Addr)
		return ErrInvalidValue
	} else if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Failed to prepare query %s: %s\n",
			qid,
			err.Error())
		panic(err)
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

EXEC_QUERY:
	if _, err = stmt.Exec(osName, d.ID); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		} else {
			err = fmt.Errorf("cannot update OS of Device %s: %w",
				d.Addr,
				err)
			db.log.Printf("[ERROR] %s\n", err.Error())
			return err
		}
	}

	d.OS = osName
	return nil
} // func (db *Database) DeviceUpdateOS(d *model.Device, osName string) error

// DeviceUpdateClass updates a Device's Class.
func (db *Database) DeviceUpdateClass(d *model.Device, class device.Class) error {
	const qid query.ID = query.DeviceUpdateClass
	var (
		err  error
		stmt *sql.Stmt
	)

	if d == nil {
		db.log.Println("[ERROR] Cannot update Device, pointer is nil")
		return ErrInvalidValue
	} else if d.ID <= 0 {
		db.log.Printf("[ERROR] Cannot update Class of Device %s (%s) - missing ID!\n",
			d.Name,
			d.Addr)
		return ErrInvalidValue
	} else if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Failed to prepare query %s: %s\n",
			qid,
			err.Error())
		panic(err)
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

EXEC_QUERY:
	if _, err = stmt.Exec(class, d.ID); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		} else {
			err = fmt.Errorf("cannot update contact timestamp of Device %s: %w",
				d.Addr,
				err)
			db.log.Printf("[ERROR] %s\n", err.Error())
			return err
		}
	}

	d.Class = class
	return nil
} // func (db *Database) DeviceUpdateClass(d *model.Device, class device.Class) error

// DeviceUpdateActive updates a Device's active flag.
func (db *Database) DeviceUpdateActive(d *model.Device, active bool) error {
	const qid query.ID = query.DeviceUpdateActive
	var (
		err  error
		stmt *sql.Stmt
	)

	if d == nil {
		db.log.Println("[ERROR] Cannot update Device, pointer is nil")
		return ErrInvalidValue
	} else if d.ID <= 0 {
		db.log.Printf("[ERROR] Cannot update Device %s (%s) - missing ID!\n",
			d.Name,
			d.Addr)
		return ErrInvalidValue
	} else if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Failed to prepare query %s: %s\n",
			qid,
			err.Error())
		panic(err)
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

EXEC_QUERY:
	if _, err = stmt.Exec(active, d.ID); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		} else {
			err = fmt.Errorf("cannot update active flag of Device %s: %w",
				d.Addr,
				err)
			db.log.Printf("[ERROR] %s\n", err.Error())
			return err
		}
	}

	d.Active = active
	return nil
} // func (db *Database) DeviceUpdateActive(d *model.Device, class device.Class) error

// DeviceGetByID loads a Device by its ID, if such a record exists in the Datbase.
func (db *Database) DeviceGetByID(id int64) (*model.Device, error) {
	const qid query.ID = query.DeviceGetByID
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
			d                 = &model.Device{ID: id}
		)

		if err = rows.Scan(&d.NetID, &d.Name, &astr, &addstamp, &contact, &d.OS, &d.Class, &d.Active); err != nil {
			var ex = fmt.Errorf("failed to scan row: %w", err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		} else if d.Addr = net.ParseIP(astr); d.Addr == nil {
			var ex = fmt.Errorf("cannot parse IP address %q: %w",
				astr,
				err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		}

		d.LastContact = time.Unix(contact, 0)
		d.Added = time.Unix(addstamp, 0)

		return d, nil
	}

	return nil, nil
} // func (db *Database) DeviceGetByID(id int64) (*model.Device, error)

// DeviceGetByNet loads all Devices belonging to a particular Network.
func (db *Database) DeviceGetByNet(netID int64) ([]*model.Device, error) {
	const qid query.ID = query.DeviceGetByNet
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
	if rows, err = stmt.Query(netID); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return nil, err
	}

	defer rows.Close() // nolint: errcheck,gosec
	var devices = make([]*model.Device, 0, 16)

	for rows.Next() {
		var (
			addstamp, contact int64
			astr              string
			d                 = new(model.Device)
		)

		if err = rows.Scan(&d.ID, &d.Name, &astr, &addstamp, &contact, &d.OS, &d.Class, &d.Active); err != nil {
			var ex = fmt.Errorf("failed to scan row: %w", err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		} else if d.Addr = net.ParseIP(astr); d.Addr == nil {
			var ex = fmt.Errorf("cannot parse IP address %q: %w",
				astr,
				err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		}

		d.LastContact = time.Unix(contact, 0)
		d.Added = time.Unix(addstamp, 0)

		devices = append(devices, d)
	}

	return devices, nil
} // func (db *Database) DeviceGetByNet(netID int64) ([]*model.Device, error)

func (db *Database) DeviceGetAll() ([]*model.Device, error) {
	const qid query.ID = query.DeviceGetAll
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
	var devices = make([]*model.Device, 0, 16)

	for rows.Next() {
		var (
			addstamp, contact int64
			astr              string
			d                 = new(model.Device)
		)

		if err = rows.Scan(&d.ID, &d.NetID, &d.Name, &astr, &addstamp, &contact, &d.OS, &d.Class, &d.Active); err != nil {
			var ex = fmt.Errorf("failed to scan row: %w", err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		} else if d.Addr = net.ParseIP(astr); d.Addr == nil {
			var ex = fmt.Errorf("cannot parse IP address %q: %w",
				astr,
				err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		}

		d.LastContact = time.Unix(contact, 0)
		d.Added = time.Unix(addstamp, 0)

		devices = append(devices, d)
	}

	return devices, nil
} // func (db *Database) DeviceGetAll() ([]*model.Device, error)
