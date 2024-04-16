// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package main

import (
	"pfcpcore/pfcp"
	b "pfcpcore/pfcp/builder"
)

// type sessionAction func(sessionRequest *UeSessionRequest) *pfcp.PfcpMessage

func (sessionRequest *UeSessionRequest) genericSER(mode b.RequestMode, iex ...pfcp.IeNode) *pfcp.PfcpMessage {
	baseIes := []pfcp.IeNode{
		pfcp.IE_NodeIdFqdn(sessionRequest.NodeId),
		pfcp.IE_FSeid(sessionRequest.Seid, nodeIp),
	}
	switch mode {
	case b.ModeCreate:
		return pfcp.NewSessionMessage(pfcp.PFCP_Session_Establishment_Request, 0, append(baseIes, iex...)...)
	case b.ModeUpdate:
		return pfcp.NewSessionMessage(pfcp.PFCP_Session_Modification_Request, pfcp.SEID(sessionRequest.Seid), iex...)
	default:
		panic("")
	}
}

func (sessionRequest *UeSessionRequest) completeUpfSER() *pfcp.PfcpMessage {
	return sessionRequest.genericSER(b.ModeCreate,
		b.DecapPdr(b.ModeCreate, sessionRequest.Request.upfFteid()),
		b.DecapFar(b.ModeCreate),
		b.EncapPdr(b.ModeCreate, sessionRequest.Request.ueIp()),
		b.EncapFar(b.ModeCreate, sessionRequest.Request.eNbFteid()),
	)
}

func (sessionRequest *UeSessionRequest) partialUpfSER() *pfcp.PfcpMessage {
	return sessionRequest.genericSER(b.ModeCreate,
		b.DecapPdr(b.ModeCreate, sessionRequest.Request.upfFteid()),
		b.DecapFar(b.ModeCreate),
		b.EncapPdr(b.ModeCreate, sessionRequest.Request.ueIp()),
		b.BufferFar(b.ModeCreate),
	)
}

func (sessionRequest *UeSessionRequest) upfSMR() *pfcp.PfcpMessage {
	return sessionRequest.genericSER(b.ModeUpdate,
		b.EncapFar(b.ModeUpdate, sessionRequest.Request.eNbFteid()),
	)
}

func (sessionRequest *UeSessionRequest) sgwCompleteSER() *pfcp.PfcpMessage {
	return sessionRequest.genericSER(b.ModeCreate,
		b.TunnelPdr(b.ModeCreate, b.RoleCore, sessionRequest.Request.sgwCoreFteid()),
		b.TunnelFar(b.ModeCreate, b.RoleCore, sessionRequest.Request.eNbFteid()),
		b.TunnelPdr(b.ModeCreate, b.RoleAccess, sessionRequest.Request.upfFteid()),
		b.TunnelFar(b.ModeCreate, b.RoleAccess, sessionRequest.Request.pgwCoreFteid()),
	)
}

func (sessionRequest *UeSessionRequest) sgwPartialSER() *pfcp.PfcpMessage {
	return sessionRequest.genericSER(b.ModeCreate,
		b.TunnelPdr(b.ModeCreate, b.RoleCore, sessionRequest.Request.sgwCoreFteid()),
		b.BufferFar(b.ModeCreate),
		b.TunnelPdr(b.ModeCreate, b.RoleAccess, sessionRequest.Request.upfFteid()),
		b.BufferFar(b.ModeCreate),
	)
}

func (sessionRequest *UeSessionRequest) sgwSMR1() *pfcp.PfcpMessage {
	return sessionRequest.genericSER(b.ModeUpdate,
		b.TunnelFar(b.ModeUpdate, b.RoleCore, sessionRequest.Request.eNbFteid()),
	)
}

func (sessionRequest *UeSessionRequest) sgwSMR2() *pfcp.PfcpMessage {
	return sessionRequest.genericSER(b.ModeUpdate,
		b.TunnelFar(b.ModeUpdate, b.RoleAccess, sessionRequest.Request.pgwCoreFteid()),
	)
}

func (sessionRequest *UeSessionRequest) pgwCompleteSER() *pfcp.PfcpMessage {
	return sessionRequest.genericSER(b.ModeCreate,
		b.DecapPdr(b.ModeCreate, sessionRequest.Request.pgwCoreFteid()),
		b.DecapFar(b.ModeCreate),
		b.EncapPdr(b.ModeCreate, sessionRequest.Request.ueIp()),
		b.EncapFar(b.ModeCreate, sessionRequest.Request.sgwCoreFteid()),
	)
}
