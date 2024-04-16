// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package main

import (
	"net/netip"

	"pfcpcore/pfcp"
)

// this value copies the UE session in customer file 'xyz.com' in upfca unit test
var ue_session_1 = UeSessionRequest{
	Seid:   50000,
	NodeId: nodeId,
	Request: SessionData{
		TeidUl:  10001,
		TeidDl:  20001,
		UeIpv4:  netip.MustParseAddr("10.0.100.1"),
		EnbIpv4: netip.MustParseAddr("172.19.0.1"),
		UpfIpv4: netip.MustParseAddr("172.19.0.2"),
		SgwData: &SgwData{
			TeidSgw: 30001,
			TeidPgw: 40001,
			SgwIpv4: netip.MustParseAddr("172.19.0.3"),
			PgwIpv4: netip.MustParseAddr("172.19.0.4"),
		},
	},
}

/*
  teid usage

  Uplink/downlink is used to name the TEID present in GTPu packets in the named direction.
  The TEID is assigned by the destination and signalled in PDR to the destination and FAR to the source.
  As FTEID it is the PDR, as outer header creation the source.
  The IP address in FTEID is the destination IP address - the source address for GTPu packets is unspecified.
  So, for upf operation the uplink session (access) has PDR/FTEID where the FTEID IPv4 is a local GTPu endpoint,
  and the TEID is the value set in GTPu packet to that address.
  This corresponds precisely with the matched outer header creation parameters.

  If the parameters for UPF operation and sgw/pgw combined are 'UPF tunnel address' and 'eNb tunnel address',
  and the inserted SGW has two tunnel addresses - 'SGW access tunnel address' and 'SGW core tunnel address',
  then in sgw/pgw mode the parameters sent to sgw and pgw are:

  pgw: core: pdr: ue ip far: 'SGW core tunnel address'
  pgw: access: pdr: 'UPF tunnel address' far: 'SGW core tunnel address'

  sgw: core: pdr: 'SGW core tunnel address' far: 'eNb tunnel address'
  sgw: access: pdr: 'SGW core tunnel address' far: 'UPF tunnel address'

*/

type SessionData struct {
	TeidUl  pfcp.TEID
	TeidDl  pfcp.TEID
	UeIpv4  netip.Addr
	EnbIpv4 netip.Addr
	UpfIpv4 netip.Addr
	*SgwData
}

type SgwData struct {
	TeidSgw pfcp.TEID
	TeidPgw pfcp.TEID
	SgwIpv4 netip.Addr
	PgwIpv4 netip.Addr
}

func (SessionData *SessionData) upfFteid() pfcp.FTeid {
	return *pfcp.NewFTeid(SessionData.TeidUl, pfcp.NetIpAddrToIpV4(SessionData.UpfIpv4))
}

func (SessionData *SessionData) eNbFteid() pfcp.FTeid {
	return *pfcp.NewFTeid(SessionData.TeidDl, pfcp.NetIpAddrToIpV4(SessionData.EnbIpv4))
}

func (SessionData *SessionData) sgwCoreFteid() pfcp.FTeid {
	if SessionData.SgwData == nil {
		panic("no sgw present")
	} else {
		return *pfcp.NewFTeid(SessionData.SgwData.TeidSgw, pfcp.NetIpAddrToIpV4(SessionData.SgwData.SgwIpv4))
	}
}

func (SessionData *SessionData) pgwCoreFteid() pfcp.FTeid {
	if SessionData.SgwData == nil {
		panic("no sgw present")
	} else {
		return *pfcp.NewFTeid(SessionData.SgwData.TeidPgw, pfcp.NetIpAddrToIpV4(SessionData.SgwData.PgwIpv4))
	}
}

func (SessionData *SessionData) ueIp() pfcp.IpV4 {
	return pfcp.NetIpAddrToIpV4(SessionData.UeIpv4)
}

type UeSessionRequest struct {
	Seid    uint64
	NodeId  string
	Request SessionData
}

func (ueSessionRequest *UeSessionRequest) next() {
	ueSessionRequest.Seid += 1
	ueSessionRequest.Request.TeidDl += 1
	ueSessionRequest.Request.TeidUl += 1
	ueSessionRequest.Request.UeIpv4 = ueSessionRequest.Request.UeIpv4.Next()
}

func getLocalNodeId() pfcp.IeNode {
	if send_nodeID_as_IPv4 {
		return pfcp.IE_NodeIdIpV4(nodeIp)
	} else {
		return pfcp.IE_NodeIdFqdn(nodeId)
	}
}

func heartBeatRequest() *pfcp.PfcpMessage {
	return pfcp.NewNodeMessage(pfcp.PFCP_Heartbeat_Request, pfcp.IE_RecoveryTimeStamp(recoveryTime))
}

func associationRequest() *pfcp.PfcpMessage {
	return pfcp.NewNodeMessage(pfcp.PFCP_Association_Setup_Request, getLocalNodeId(), pfcp.IE_RecoveryTimeStamp(recoveryTime))
}

func associationReleaseRequest() *pfcp.PfcpMessage {
	return pfcp.NewNodeMessage(pfcp.PFCP_Association_Release_Request, getLocalNodeId())
}

func sessionDelete(seid pfcp.SEID) *pfcp.PfcpMessage {
	return pfcp.NewSessionMessage(pfcp.PFCP_Session_Deletion_Request, seid)
}
