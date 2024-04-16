// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package transport_test

import (
	"net/netip"
	"testing"

	"pfcpcore/pfcp"
	"pfcpcore/transport"
	"pfcpcore/udpserver"
)

var (
	localAddr = netip.MustParseAddrPort("127.0.0.2:8805")
	peerAddr  = netip.MustParseAddrPort("127.0.0.3:8805")
)

type t struct{}

var exitChan chan t = make(chan t)

func close() { exitChan <- t{} }

func TestTransport(t *testing.T) {
	if local, err := udpserver.NewUDPServer(localAddr); err != nil {
		t.Error(err)
	} else if remote, err := udpserver.NewUDPServer(peerAddr); err != nil {
		t.Error(err)
	} else {
		localUdp := local.Register(peerAddr)
		peerUdp := remote.Register(localAddr)
		go chatter(t, localUdp)
		go chattee(t, peerUdp)
		<-exitChan
		<-exitChan
		// time.Sleep(time.Second * 5) // only use this to check there is no bad stuff happening in the background
	}
}

func chatter(t *testing.T, udp *udpserver.UdpServerPeer) {
	requestChan := make(chan transport.PeerRequest)
	endpoint := transport.NewTransport(udp, requestChan)
	responseChan := make(chan transport.RequestReturn)
	endpoint.EnterRequest(&pfcp.HeartBeatRequest, responseChan)
	m := <-responseChan
	if m, err := m.Value(); err != nil {
		t.Errorf("request failed %s", err.Error())
	} else if m.MessageTypeCode != pfcp.PFCP_Heartbeat_Response {
		t.Errorf("request failed, wrong message type %d", m.MessageTypeCode)
	} else {
		t.Logf("got response in chatter: %s", m)
	}

	close()
}

func chattee(t *testing.T, udp *udpserver.UdpServerPeer) {
	requestChan := make(chan transport.PeerRequest)
	conn := transport.NewTransport(udp, requestChan)
	m := <-requestChan
	t.Logf("got request in chattee: %s", m)
	conn.EnterResponse(&pfcp.HeartBeatResponse, m)
	close()
}
