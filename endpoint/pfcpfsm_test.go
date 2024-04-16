// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package endpoint_test

import (
	"net/netip"
	"testing"

	"pfcpcore/endpoint"
	"pfcpcore/pfcp"
	"pfcpcore/session"
	"pfcpcore/testcases"
)

func TestPfcpFsm(t *testing.T) {
	if peer1, peer2, err := testcases.MakeTestPeers(); err != nil {
		t.Errorf(err.Error())
	} else {
		active := func(peer *endpoint.PfcpPeer) {
			peer.EnterRequest(&pfcp.HeartBeatRequest)

			m := <-peer.ResponseChan
			if m, err := m.Value(); err != nil {
				t.Errorf("request failed %s", err.Error())
			} else if m.MessageTypeCode != pfcp.PFCP_Heartbeat_Response {
				t.Errorf("request failed, wrong message type %d", m.MessageTypeCode)
			} else {
				t.Logf("got response in active: %s", m)
			}
		}

		upfFsm := endpoint.NewPfcpAssociationState(endpoint.PfcpAssociationConfig{
			NodeName:               "upf",
			LocalSignallingAddress: netip.MustParseAddr("169.254.169.252"),
			Application:            session.DefaultApplication{},
			PeerEndpoint:           peer2,
		})

		active(peer1)
		upfFsm.Drop()
	}
}
