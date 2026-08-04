// /home/krylon/go/src/github.com/blicero/chili/database/attribute.go
// -*- mode: go; coding: utf-8; -*-
// Created on 11. 07. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-08-04 10:03:02 krylon>

package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/blicero/chili/database/query"
	"github.com/blicero/chili/model"
	"github.com/blicero/chili/model/attribute"
)

func (db *Database) unpackPayload(attr *model.Attribute, val string) (bool, error) {
	var (
		err error
	)
	switch attr.Type {
	case attribute.DiskSpace:
		var x int64
		if err = json.Unmarshal([]byte(val), &x); err != nil {
			db.log.Printf("[ERROR] Cannot parse JSON %q: %s\n",
				val,
				err.Error())
			return false, err
		}
		attr.Value = model.DiskSpace(x)
	case attribute.Uptime:
		var u = new(model.Uptime)
		if err = json.Unmarshal([]byte(val), u); err != nil {
			db.log.Printf("[ERROR] Cannot parse JSON %q: %s\n",
				val,
				err.Error())
			return false, err
		}
		attr.Value = u
	case attribute.Updates:
		var updates = make([]string, 0, 8)
		if err = json.Unmarshal([]byte(val), &updates); err != nil {
			db.log.Printf("[ERROR] Cannot parse JSON %q: %s\n",
				val,
				err.Error())
			return false, err
		}
		attr.Value = model.Updates(updates)
	case attribute.Packages:
		var pkg = make([]string, 0, 8)
		if err = json.Unmarshal([]byte(val), &pkg); err != nil {
			db.log.Printf("[ERROR] Cannot parse JSON %q: %s\n",
				val,
				err.Error())
			return false, err
		}
		attr.Value = model.Updates(pkg)
	case attribute.SNMP:
		var info = make(map[string]string)
		if err = json.Unmarshal([]byte(val), &info); err != nil {
			db.log.Printf("[ERROR] Cannot parse JSON %q: %s\n",
				val,
				err.Error())
			return false, err
		}
		attr.Value = model.SNMPInfo(info)
	case attribute.Services:
		var svc = new(model.Services)
		if err = json.Unmarshal([]byte(val), svc); err != nil {
			db.log.Printf("[ERROR] Cannot parse JSON %q: %s\n",
				val,
				err.Error())
			return false, err
		}
		attr.Value = svc
	default:
		err = fmt.Errorf("don't know how to decode %s",
			attr.Type)
		db.log.Printf("[ERROR] %s\n", err.Error())
		return false, err

	}

	return true, nil
} // func (db *Database) unpackPayload(dev *model.Device, attr *model.Attribute, val string) (bool, error)

// AttributeAdd adds an Attribute to the Database.
func (db *Database) AttributeAdd(a *model.Attribute) error {
	const qid query.ID = query.AttributeAdd
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

	var res sql.Result

EXEC_QUERY:
	if res, err = stmt.Exec(
		a.DevID,
		a.Timestamp.Unix(),
		a.Type,
		a.Value.String()); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		} else {
			err = fmt.Errorf("cannot add Attribute %s of Device %d to Database: %w (%q)",
				a.Type,
				a.DevID,
				err,
				a.Value)
			db.log.Printf("[ERROR] %s\n", err.Error())
			return err
		}
	}

	var id int64

	if id, err = res.LastInsertId(); err != nil {
		db.log.Printf("[ERROR] Failed to get ID for added Attribute %s on Device %d: %s\n",
			a.Type,
			a.DevID,
			err.Error())
		return err
	}

	// if !rows.Next() {
	// 	msg := fmt.Sprintf("Query %s did not return an ID", qid)
	// 	db.log.Printf("[ERROR] %s\n",
	// 		msg)
	// 	return errors.New(msg)
	// } else if err = rows.Scan(&id); err != nil {
	// 	var ex = fmt.Errorf("failed to get ID  from query %s: %w",
	// 		qid,
	// 		err)
	// 	db.log.Printf("[ERROR] %s\n", ex.Error())
	// 	return ex
	// }

	a.ID = id

	return nil
} // func (db *Database) AttributeAdd(a *model.Attribute) error

// AttributeGetByDeviceType gets the most recent value for the given Device
// and Attribute.
func (db *Database) AttributeGetByDeviceType(dev *model.Device, aid attribute.ID) (*model.Attribute, error) {
	const qid query.ID = query.AttributeGetByDeviceType
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
	if rows, err = stmt.Query(dev.ID, aid); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return nil, err
	}

	defer rows.Close() // nolint: errcheck,gosec

	if rows.Next() {
		var (
			addstamp int64
			vstr     string
			attr     = &model.Attribute{
				DevID: dev.ID,
				Type:  aid,
			}
		)

		if err = rows.Scan(&attr.ID, &addstamp, &vstr, 1); err != nil {
			var ex = fmt.Errorf("failed to scan row: %w", err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		} else if _, err = db.unpackPayload(attr, vstr); err != nil {
			var ex = fmt.Errorf("cannot parse JSON value %q: %w",
				vstr,
				err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		}

		attr.Timestamp = time.Unix(addstamp, 0)

		return attr, nil
	}

	return nil, nil
} // func (db *Database) AttributeGetByDeviceType(dev *model.Device, aid attribute.ID) (*model.Attribute, error)

// AttributeGetByType returns values of a given type across all devices
func (db *Database) AttributeGetByType(aid attribute.ID) ([]*model.Attribute, error) {
	const qid query.ID = query.AttributeGetByType
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
	if rows, err = stmt.Query(aid); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return nil, err
	}

	defer rows.Close() // nolint: errcheck,gosec

	var results = make([]*model.Attribute, 0, 8)

	for rows.Next() {
		var (
			addstamp int64
			vstr     string
			attr     = &model.Attribute{
				Type: aid,
			}
		)

		if err = rows.Scan(&attr.ID, &attr.DevID, &addstamp, &vstr); err != nil {
			var ex = fmt.Errorf("failed to scan row: %w", err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		} else if _, err = db.unpackPayload(attr, vstr); err != nil {
			var ex = fmt.Errorf("cannot parse JSON value %q: %w",
				vstr,
				err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		}

		attr.Timestamp = time.Unix(addstamp, 0)

		results = append(results, attr)
	}

	return results, nil
} // func (db *Database) AttributeGetByType(aid attribute.ID) ([]*model.Attribute, error)

// AttributeGetByDevice returns the most recent values of each type for the
// given Device.
func (db *Database) AttributeGetByDevice(dev *model.Device) ([]*model.Attribute, error) {
	const qid query.ID = query.AttributeGetByDevice
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
	if rows, err = stmt.Query(dev.ID); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return nil, err
	}

	defer rows.Close() // nolint: errcheck,gosec

	var results = make([]*model.Attribute, 0, 8)

	for rows.Next() {
		var (
			addstamp int64
			vstr     string
			ok       bool
			attr     = &model.Attribute{
				DevID: dev.ID,
			}
		)

		if err = rows.Scan(&attr.ID, &attr.Type, &addstamp, &vstr); err != nil {
			var ex = fmt.Errorf("failed to scan row: %w", err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		} /*else if err = json.Unmarshal([]byte(vstr), &attr.Value); err != nil {
			var ex = fmt.Errorf("cannot parse JSON value %q: %w",
				vstr,
				err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		}*/
		// switch attr.Type {
		// case attribute.DiskSpace:
		// 	var x int64
		// 	if err = json.Unmarshal([]byte(vstr), &x); err != nil {
		// 		db.log.Printf("[ERROR] Cannot parse JSON %q: %s\n",
		// 			vstr,
		// 			err.Error())
		// 		return nil, err
		// 	}
		// 	attr.Value = model.DiskSpace(x)
		// case attribute.Uptime:
		// 	var u = new(model.Uptime)
		// 	if err = json.Unmarshal([]byte(vstr), u); err != nil {
		// 		db.log.Printf("[ERROR] Cannot parse JSON %q: %s\n",
		// 			vstr,
		// 			err.Error())
		// 		return nil, err
		// 	}
		// 	attr.Value = u
		// case attribute.Updates:
		// 	var updates = make([]string, 0, 8)
		// 	if err = json.Unmarshal([]byte(vstr), &updates); err != nil {
		// 		db.log.Printf("[ERROR] Cannot parse JSON %q: %s\n",
		// 			vstr,
		// 			err.Error())
		// 		return nil, err
		// 	}
		// 	attr.Value = model.Updates(updates)
		// case attribute.Packages:
		// 	var pkg = make([]string, 0, 8)
		// 	if err = json.Unmarshal([]byte(vstr), &pkg); err != nil {
		// 		db.log.Printf("[ERROR] Cannot parse JSON %q: %s\n",
		// 			vstr,
		// 			err.Error())
		// 		return nil, err
		// 	}
		// 	attr.Value = model.Updates(pkg)
		// default:
		// 	err = fmt.Errorf("don't know how to decode %s",
		// 		attr.Type)
		// 	db.log.Printf("[ERROR] %s\n", err.Error())
		// 	return nil, err

		// }
		if ok, err = db.unpackPayload(attr, vstr); err != nil {
			db.log.Printf("[ERROR] Failed to parse JSON: %s\n",
				err.Error())
			return nil, nil
		} else if !ok {
			db.log.Printf("[INFO] Did not find attributes for %s\n",
				dev.Name)
			return nil, nil
		}

		attr.Timestamp = time.Unix(addstamp, 0)

		results = append(results, attr)
	}

	return results, nil
} // func (db *Database) AttributeGetByDevice(dev *model.Device) ([]*model.Attribute, error)

// AttributeGetMostRecent returns the most recent Attribute of each type for
// the given Device.
func (db *Database) AttributeGetMostRecent(dev *model.Device) ([]*model.Attribute, error) {
	const qid query.ID = query.AttributeGetMostRecent
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
	if rows, err = stmt.Query(dev.ID); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return nil, err
	}

	defer rows.Close() // nolint: errcheck,gosec

	var results = make([]*model.Attribute, 0, 8)

	for rows.Next() {
		var (
			addstamp int64
			vstr     string
			ok       bool
			attr     = &model.Attribute{
				DevID: dev.ID,
			}
		)

		if err = rows.Scan(&attr.ID, &attr.Type, &addstamp, &vstr); err != nil {
			var ex = fmt.Errorf("failed to scan row: %w", err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		} /*else if err = json.Unmarshal([]byte(vstr), &attr.Value); err != nil {
			var ex = fmt.Errorf("cannot parse JSON value %q: %w",
				vstr,
				err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		}*/
		// switch attr.Type {
		// case attribute.DiskSpace:
		// 	var x int64
		// 	if err = json.Unmarshal([]byte(vstr), &x); err != nil {
		// 		db.log.Printf("[ERROR] Cannot parse JSON %q: %s\n",
		// 			vstr,
		// 			err.Error())
		// 		return nil, err
		// 	}
		// 	attr.Value = model.DiskSpace(x)
		// case attribute.Uptime:
		// 	var u = new(model.Uptime)
		// 	if err = json.Unmarshal([]byte(vstr), u); err != nil {
		// 		db.log.Printf("[ERROR] Cannot parse JSON %q: %s\n",
		// 			vstr,
		// 			err.Error())
		// 		return nil, err
		// 	}
		// 	attr.Value = u
		// case attribute.Updates:
		// 	var updates = make([]string, 0, 8)
		// 	if err = json.Unmarshal([]byte(vstr), &updates); err != nil {
		// 		db.log.Printf("[ERROR] Cannot parse JSON %q: %s\n",
		// 			vstr,
		// 			err.Error())
		// 		return nil, err
		// 	}
		// 	attr.Value = model.Updates(updates)
		// case attribute.Packages:
		// 	var pkg = make([]string, 0, 8)
		// 	if err = json.Unmarshal([]byte(vstr), &pkg); err != nil {
		// 		db.log.Printf("[ERROR] Cannot parse JSON %q: %s\n",
		// 			vstr,
		// 			err.Error())
		// 		return nil, err
		// 	}
		// 	attr.Value = model.Updates(pkg)
		// default:
		// 	err = fmt.Errorf("don't know how to decode %s",
		// 		attr.Type)
		// 	db.log.Printf("[ERROR] %s\n", err.Error())
		// 	return nil, err

		// }
		if ok, err = db.unpackPayload(attr, vstr); err != nil {
			db.log.Printf("[ERROR] Failed to parse JSON: %s\n",
				err.Error())
			return nil, nil
		} else if !ok {
			db.log.Printf("[INFO] Did not find attributes for %s\n",
				dev.Name)
			return nil, nil
		}

		attr.Timestamp = time.Unix(addstamp, 0)

		results = append(results, attr)
	}

	return results, nil
} // func (db *Database) AttributeGetMostRecent(dev *model.Device) ([]*model.Attribute, error)
