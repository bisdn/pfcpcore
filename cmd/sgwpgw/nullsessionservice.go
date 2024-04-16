// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package main

/*

definition and null implementation for the interface types SessionObject and SessionService

*/

import log "github.com/sirupsen/logrus"

// ******************************
// Interface Definition
// ******************************

type SessionObject interface {
	modify(old, new *CommonStateValue)
	delete()
}

type SessionService interface {
	create(*CommonStateValue) SessionObject
}

// ******************************
// null implementation
// ******************************

type (
	NullSessionObject  struct{}
	NullSessionService struct{}
)

func (NullSessionService) create(*CommonStateValue) SessionObject {
	log.Trace("(NullSessionService) create()")
	return NullSessionObject{}
}

func (NullSessionObject) modify(old, new *CommonStateValue) {
	log.Trace("(NullSessionObject) modify()")
}

func (NullSessionObject) delete() {
	log.Trace("(NullSessionObject) delete()")
}
