// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package main

import "pfcpcore/pfcp"

type mode interface {
	sessionCreateWithModify(sessionRequest *UeSessionRequest) []*pfcp.PfcpMessage
	sessionCreate(sessionRequest *UeSessionRequest) *pfcp.PfcpMessage
}

type sgwModeT struct{}

func (sgwMode sgwModeT) sessionCreate(sessionRequest *UeSessionRequest) *pfcp.PfcpMessage {
	return sessionRequest.sgwCompleteSER()
}

func (sgwMode sgwModeT) sessionCreateWithModify(sessionRequest *UeSessionRequest) []*pfcp.PfcpMessage {
	return []*pfcp.PfcpMessage{
		sessionRequest.sgwPartialSER(),
		sessionRequest.sgwSMR1(),
		sessionRequest.sgwSMR2(),
	}
}

var sgwMode = sgwModeT{}

type pgwModeT struct{}

func (pgwMode pgwModeT) sessionCreate(sessionRequest *UeSessionRequest) *pfcp.PfcpMessage {
	return sessionRequest.pgwCompleteSER()
}

func (pgwMode pgwModeT) sessionCreateWithModify(sessionRequest *UeSessionRequest) []*pfcp.PfcpMessage {
	panic("unimplemented")
	// return []*pfcp.PfcpMessage{
	// 	sessionRequest.pgwPartialSER(),
	// 	sessionRequest.pgwSMR1(),
	// 	sessionRequest.pgwSMR2(),
	// }
}

var pgwMode = pgwModeT{}

type upfModeT struct{}

func (upfMode upfModeT) sessionCreate(sessionRequest *UeSessionRequest) *pfcp.PfcpMessage {
	return sessionRequest.completeUpfSER()
}

func (upfMode upfModeT) sessionCreateWithModify(sessionRequest *UeSessionRequest) []*pfcp.PfcpMessage {
	return []*pfcp.PfcpMessage{
		sessionRequest.partialUpfSER(),
		sessionRequest.upfSMR(),
	}
}

var upfMode = upfModeT{}
