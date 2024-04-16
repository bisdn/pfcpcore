// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package main

import (
	"fmt"
	"strconv"
	"strings"

	"pfcpcore/pfcp"
	b "pfcpcore/pfcp/builder"
)

/*
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
*/
type Session struct {
	sgwCoreFteid pfcp.FTeid
	eNbFteid     pfcp.FTeid
	upfFteid     pfcp.FTeid
	pgwCoreFteid pfcp.FTeid
	ueIp         pfcp.IpV4
}

var session1 = Session{
	sgwCoreFteid: mustParseFTeid("172.19.0.3:30001"),
	eNbFteid:     mustParseFTeid("172.19.0.2:20001"),
	upfFteid:     mustParseFTeid("172.19.0.1:10001"),
	pgwCoreFteid: mustParseFTeid("172.19.0.4:40001"),
	ueIp:         pfcp.MustParseAddr("10.0.100.1"),
}

func mustParseFTeid(s string) pfcp.FTeid {
	if fetid, err := parseFTeid(s); err != nil {
		panic(fmt.Sprintf("cannot parse %s as pfcp.FTeid (%s)", s, err.Error()))
	} else {
		return fetid
	}
}

func parseFTeid(s string) (fteid pfcp.FTeid, err error) {
	if sx := strings.Split(s, ":"); len(sx) != 2 {
		err = fmt.Errorf("failed to parse FTeid=\"%s\" as a pair of IPv4:uint32", s)
	} else if ip, e := pfcp.ParseAddr(sx[0]); e != nil {
		err = fmt.Errorf("failed to parse \"%s\" as an IPv4 ()%s", sx[0], e.Error())
	} else if teidInt, e := strconv.Atoi(sx[1]); err != nil {
		err = fmt.Errorf("failed to parse \"%s\" as an int ()%s", sx[1], e.Error())
	} else {
		teid := pfcp.TEID(teidInt)
		fteid.IpV4 = &ip
		fteid.Teid = &teid
	}
	return
}

func (Session *Session) sgwSessionCreate() []pfcp.IeNode {
	return []pfcp.IeNode{
		b.TunnelPdr(b.ModeCreate, b.RoleCore, Session.sgwCoreFteid),
		b.TunnelFar(b.ModeCreate, b.RoleCore, Session.eNbFteid),
		b.TunnelPdr(b.ModeCreate, b.RoleAccess, Session.upfFteid),
		b.TunnelFar(b.ModeCreate, b.RoleAccess, Session.pgwCoreFteid),
	}
}

func (Session *Session) xgwSessionModifyIdle() []pfcp.IeNode {
	return []pfcp.IeNode{
		b.BufferFar(b.ModeUpdate),
	}
}

func (Session *Session) sgwSessionModifyActive() []pfcp.IeNode {
	Session.eNbFteid.NextTeid()
	return []pfcp.IeNode{
		b.TunnelFar(b.ModeUpdate, b.RoleCore, Session.eNbFteid),
	}
}

func (Session *Session) pgwSessionCreate() []pfcp.IeNode {
	return []pfcp.IeNode{
		b.DecapPdr(b.ModeCreate, Session.pgwCoreFteid),
		b.DecapFar(b.ModeCreate),
		b.EncapPdr(b.ModeCreate, Session.ueIp),
		b.EncapFar(b.ModeCreate, Session.sgwCoreFteid),
	}
}

func (Session *Session) smfSessionCreate() []pfcp.IeNode {
	return []pfcp.IeNode{
		b.DecapPdr(b.ModeCreate, Session.upfFteid),
		b.DecapFar(b.ModeCreate),
		b.EncapPdr(b.ModeCreate, Session.ueIp),
		b.EncapFar(b.ModeCreate, Session.eNbFteid),
	}
}

func (Session *Session) smfSessionModifyActive() []pfcp.IeNode {
	Session.eNbFteid.NextTeid()
	return []pfcp.IeNode{
		b.EncapFar(b.ModeUpdate, Session.eNbFteid),
	}
}

var (
	eNb2Fteid          = mustParseFTeid("172.19.0.101:50001")
	eNb2FteidTemporary = mustParseFTeid("172.19.0.101:50010")
	upfFteidTemporary  = mustParseFTeid("172.19.0.2:20010")
)

func (Session *Session) sgwSessionMobility() []pfcp.IeNode {
	return []pfcp.IeNode{
		b.TunnelFar(b.ModeUpdate, b.RoleCore, eNb2Fteid),
	}
}

func (Session *Session) sgwSessionMobilityNewTunnel() []pfcp.IeNode {
	return []pfcp.IeNode{
		b.CustomOhcPdr(3, 3, b.ModeCreate, pfcp.EnumAccess, upfFteidTemporary),
		b.CustomOhcFar(3, b.ModeCreate, pfcp.EnumAccess, eNb2FteidTemporary),
	}
}

func (Session *Session) sgwSessionMobilityRemoveTunnel() []pfcp.IeNode {
	return []pfcp.IeNode{
		pfcp.IE_RemovePdr(pfcp.IE_PdrId(3)),
		pfcp.IE_RemoveFar(pfcp.IE_FarId(3)),
	}
}
