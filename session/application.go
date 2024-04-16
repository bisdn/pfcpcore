// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package session

import (
	log "github.com/sirupsen/logrus"
	"pfcpcore/pfcp"
)

type Application interface {
	CallbackSessionEstablishmentRequest(seid pfcp.SEID, ser *pfcp.IeNode) (uint8, error)
	CallbackSessionDeletionRequest(seid pfcp.SEID) (uint8, error)
}

type DefaultApplication struct{}

func (DefaultApplication) CallbackSessionEstablishmentRequest(seid pfcp.SEID, ser *pfcp.IeNode) (uint8, error) {
	log.Error("using undefined application in CallbackSessionEstablishmentRequest()")
	return 0, nil
}

func (DefaultApplication) CallbackSessionDeletionRequest(seid pfcp.SEID) (uint8, error) {
	log.Error("using undefined application in CallbackSessionDeletionRequest()")
	return 0, nil
}
