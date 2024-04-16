// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package main

/*

concrete implementation of the interface types SessionObject and SessionService,
based on the smf library package and the encoders defined locally.

*/

import (
	log "github.com/sirupsen/logrus"
	"pfcpcore/smf"
)

type XgwSessionObject struct {
	*smf.Session
}

type XgwSessionService struct {
	association *smf.Association
}

func newXgwSessionService(association *smf.Association) XgwSessionService {
	return XgwSessionService{
		association: association,
	}
}

func (xss XgwSessionService) create(csv *CommonStateValue) SessionObject {
	log.Trace("(XgwSessionService) create()")
	ies := encodeSer(csv)
	if session, err := xss.association.CreateSession(ies...); err != nil {
		log.Error("failed to create session")
		return XgwSessionObject{Session: nil}
	} else {
		return XgwSessionObject{Session: session}
	}
}

func (xso XgwSessionObject) modify(old, new *CommonStateValue) {
	ies := encodeSmr(old, new)
	if len(ies) == 0 {
		log.Warn("(XgwSessionObject) modify() skip, empty IE list")
	} else if xso.Session == nil {
		log.Warn("(XgwSessionObject) modify() failed, no active session")
	} else if err := xso.Session.Modify(ies...); err != nil {
		log.Warnf("(XgwSessionObject) modify() failed, %s", err.Error())
	} else {
		log.Trace("(XgwSessionObject) modify()")
	}
}

func (xso XgwSessionObject) delete() {
	if xso.Session == nil {
		log.Warn("(XgwSessionObject) delete() failed, no active session")
	} else if err := xso.Session.Delete(); err != nil {
		log.Warnf("(XgwSessionObject) delete() failed, %s", err.Error())
	} else {
		log.Trace("(XgwSessionObject) delete()")
	}
}
