// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package pfcp

import (
	"net/netip"
	"testing"
)

func TestTWHeartBeatRequest(t *testing.T) {
	TTW(t, HeartBeatRequest)
}

func TTW(t *testing.T, subject PfcpMessage) {
	bytes := subject.Serialise()
	if msg, err := ParseValidate(bytes); err != nil {
		t.Logf("message parse failed:%s", err.Error())
	} else {
		t.Logf("original message %s\n", subject.Dumper())
		t.Logf("recreated message %s\n", msg.Dumper())

		// newbytes := msg.Serialise()
		// t.Logf("original message  %x\n", bytes)
		// t.Logf("recreated message %x\n", newbytes)
	}
}

func TestTWHSERFail(t *testing.T) {
	var SER PfcpMessage = *NewSessionMessage(PFCP_Session_Establishment_Request, 1234, IE_RecoveryTimeStamp(2020202))
	TTW(t, SER)
}

var ser = NewSessionMessage(
	PFCP_Session_Establishment_Request,
	1234,
	IE_NodeIdFqdn("id@node"),
	IE_FSeid(77, netip.MustParseAddr("169.254.169.254")),
	IE_CreatePdr(
		IE_PdrId(66),
		IE_Pdi(
			IE_SourceInterface(EnumAccess))),
	IE_CreateFar(IE_FarId(55)),
)

func TestTWHSERok(t *testing.T) {
	// t.Logf("SER %+v\n", ser)
	TTW(t, *ser)
}

func TestTWHSER2(t *testing.T) {
	// t.Logf("SER %+v\n", ser)
	TTW(t, *ser2)
}

func TestTWHSet1(t *testing.T) {
	for i := range TestSet1 {
		TTW(t, *TestSet1[i])
	}
}
