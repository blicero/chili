// /home/krylon/go/src/github.com/blicero/chili/probe/attribute.go
// -*- mode: go; coding: utf-8; -*-
// Created on 26. 10. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-10-26 19:11:42 krylon>

package probe

import (
	"time"

	"github.com/blicero/chili/database"
	"github.com/blicero/chili/model"
	"github.com/blicero/chili/model/attribute"
)

func (p *Probe) attributeAdd(db *database.Database, dev *model.Device, atype attribute.ID, val model.Payload) error {
	var (
		err  error
		attr *model.Attribute
	)

	attr = &model.Attribute{
		DevID:     dev.ID,
		Timestamp: time.Now().Truncate(time.Second),
		Type:      atype,
		Value:     val,
	}

	if err = db.AttributeAdd(attr); err != nil {
		p.log.Printf("[ERROR] Cannot add Attribute %s for %s to database: %s\n",
			atype,
			dev.Name,
			err.Error())
		return err
	}

	return nil
} // func (p *Probe) attributeAdd(db *database.Database, dev *model.Device, atype attribute.ID, val *model.Payload) error
