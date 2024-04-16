// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package udpserver_test

import (
	"net/netip"
	"testing"

	"pfcpcore/pfcp"
	"pfcpcore/udpserver"
)

var (
	local1  netip.AddrPort = netip.MustParseAddrPort("127.0.0.4:8805")
	remote1                = netip.MustParseAddrPort("127.0.0.5:8805")
)

func TestSimple(t *testing.T) {
	if local, err := udpserver.NewUDPServer(local1); err != nil {
		t.Error(err)
	} else if remote, err := udpserver.NewUDPServer(remote1); err != nil {
		t.Error(err)
	} else {
		go echoService(local.Register(remote1))
		echoClient(remote.Register(local1))
	}
}

func echoService(udpServerPeer *udpserver.UdpServerPeer) {
	hbRsp := pfcp.HeartBeatResponse.Serialise()

	// TODO parse/validate the received messages
	<-udpServerPeer.Receive
	udpServerPeer.Enqueue(udpserver.ToUdpMessage(hbRsp))
}

func echoClient(udpServerPeer *udpserver.UdpServerPeer) {
	hbReq := pfcp.HeartBeatRequest.Serialise()
	udpServerPeer.Enqueue(udpserver.ToUdpMessage(hbReq))
	<-udpServerPeer.Receive
}
