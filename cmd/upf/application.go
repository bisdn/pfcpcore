// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package main

import (
	log "github.com/sirupsen/logrus"
	"pfcpcore/pfcp"
)

type UpfApplication struct{}

func (st *UpfApplication) CallbackSessionEstablishmentRequest(seid pfcp.SEID, ser *pfcp.IeNode) (uint8, error) {
	log.Info("using upf application")
	if state, err := ParseUpfSER(ser); err != nil {
		log.Printf("UPF parse SER fail: %s\n", err.Error())
	} else {
		log.Printf("UPF parse SER ok: %s\n", state)
	}
	return 0, nil
}

func (UpfApplication) CallbackSessionDeletionRequest(seid pfcp.SEID) (uint8, error) {
	log.Error("UpfApplication: CallbackSessionDeletionRequest() null action")
	return 0, nil
}
