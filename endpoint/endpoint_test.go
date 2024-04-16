// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package endpoint_test

import (
	"net/netip"
	"testing"

	"pfcpcore/endpoint"
	"pfcpcore/pfcp"
	"pfcpcore/testcases"
	"pfcpcore/udpserver"
)

func TestEndpoint(t *testing.T) {
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

			testcases.Close()
		}

		passive := func(peer *endpoint.PfcpPeer) {
			m := <-peer.RequestChan
			t.Logf("got request in passive: %s", m)
			peer.EnterResponse(&pfcp.HeartBeatResponse, m)
			testcases.Close()
		}
		go active(peer1)
		go passive(peer2)
		testcases.CloseWait()
		testcases.CloseWait()
	}
}

/*
	TestDynamicEndpoint

construct one peer as active, register the passive partner, and use it to send message to the passive peer.
The passive peer has no registration for the active until the message arrives.
*/
func TestDynamicEndpoint(t *testing.T) {
	activePeerAddrPort := testcases.AddrFactory()
	passivePeerAddrPort := testcases.AddrFactory()
	activeExitStatusChan := make(chan error)

	go func() {
		t.Logf("active running")

		if activePeerEndpoint, err := endpoint.NewPfcpEndpoint(activePeerAddrPort); err != nil {
			activeExitStatusChan <- err
			return
		} else {
			peer := activePeerEndpoint.Peer(passivePeerAddrPort)

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
		activeExitStatusChan <- nil
	}()

	var (
		passivePeerEndpoint *endpoint.PfcpEndpoint
		err                 error
	)

	if passivePeerEndpoint, err = endpoint.NewPfcpEndpoint(passivePeerAddrPort); err != nil {
		t.Fatalf(err.Error())
	}

	var (
		dynamicPeerAddrPort netip.AddrPort
		dynamicPeerPeer     *endpoint.PfcpPeer
	)
	t.Logf("passive running")

	m := <-passivePeerEndpoint.EventChannel
	t.Logf("passive got event")

	switch event := m.(type) {

	case udpserver.UdpEventNewPeer:
		t.Logf("got new peer event from %s", event.PeerAddr)
		dynamicPeerAddrPort = event.PeerAddr
		dynamicPeerPeer = passivePeerEndpoint.Peer(dynamicPeerAddrPort)
		dynamicPeerPeer.Recirculate(event.Payload, event.PeerAddr.Port())

	case udpserver.UdpEventNetworkError:
		t.Fatalf(event.Err.Error())
	}

	passive := func() {
		m := <-dynamicPeerPeer.RequestChan
		t.Logf("got request in passive: %s", m)
		dynamicPeerPeer.EnterResponse(&pfcp.HeartBeatResponse, m)
	}

	passive()

	if activeExitStatus := <-activeExitStatusChan; activeExitStatus != nil {
		t.Errorf(activeExitStatus.Error())
	}
}
