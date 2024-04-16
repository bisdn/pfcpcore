// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package main

import (
	log "github.com/sirupsen/logrus"
	"pfcpcore/pfcp"
	b "pfcpcore/pfcp/builder"
)

func encodeSer(csv *CommonStateValue) (ies []pfcp.IeNode) {
	// prior logic should dictate that there are no nil pointers in the CSV
	// the only expected variant is a potentially null core FAR

	if csv.PgwSessionState == nil || csv.SgwSessionState == nil {
		panic("Xgw states cannot be nil here")
	}

	encapFar := b.BufferFar(b.ModeCreate)
	if csv.coreFar != spFteidNull {
		encapFar = b.EncapFar(b.ModeCreate, csv.coreFar.PfcpFteid())
	}

	return []pfcp.IeNode{
		b.DecapPdr(b.ModeCreate, csv.accessPdr.PfcpFteid()),
		b.DecapFar(b.ModeCreate),
		b.EncapPdr(b.ModeCreate, csv.corePdr),
		encapFar,
	}
}

func encodeSmr(oldCsv, newCsv *CommonStateValue) (ies []pfcp.IeNode) {
	if oldCsv.PgwSessionState == nil || oldCsv.SgwSessionState == nil || newCsv.PgwSessionState == nil || newCsv.SgwSessionState == nil {
		panic("Xgw states cannot be nil here")
	}

	// the expected changes here are in the core FAR, though the access PDR could also change based on the same logic

	updates := []pfcp.IeNode{}

	if newCsv.coreFar != oldCsv.coreFar {
		if newCsv.coreFar == spFteidZero {
			log.Warn("invalid core FAR")
		} else if newCsv.coreFar == spFteidNull {
			log.Trace("SMR core FAR: set buffer mode")
			updates = append(updates, b.BufferFar(b.ModeUpdate))
		} else {
			log.Trace("SMR core FAR: new")
			updates = append(updates, b.EncapFar(b.ModeUpdate, newCsv.coreFar.PfcpFteid()))
		}
	}

	if newCsv.accessPdr != oldCsv.accessPdr {
		if newCsv.accessPdr == spFteidNull || newCsv.accessPdr == spFteidZero {
			log.Warn("SMR: invalid access PDR")
		} else {
			log.Trace("SMR: new access PDR")
			updates = append(updates, b.DecapPdr(b.ModeUpdate, newCsv.accessPdr.PfcpFteid()))
		}
	}

	if newCsv.corePdr != oldCsv.corePdr {
		log.Warn("SMR: unexpected change in core PDR")
	}

	if len(updates) == 0 {
		log.Warn("SMR payload unexpectedly empty")
	}
	return updates
}
