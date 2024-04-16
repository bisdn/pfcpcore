// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package endpoint

import (
	"net/netip"

	"pfcpcore/pfcp"
	"pfcpcore/transport"
	"pfcpcore/udpserver"
)

// TDOD consider if this wrapper is really needed around UdpServer
type PfcpEndpoint struct {
	*udpserver.UdpServer
}

type PfcpPeer struct {
	*udpserver.UdpServer
	*udpserver.UdpServerPeer
	*transport.Transport
	RequestChan  chan transport.PeerRequest
	ResponseChan chan transport.RequestReturn
}

func NewPfcpEndpointByName(addrPortString string) (pfcpEndpoint *PfcpEndpoint, err error) {
	if addrPort, err := netip.ParseAddrPort(addrPortString); err != nil {
		return nil, err
	} else {
		return NewPfcpEndpoint(addrPort)
	}
}

func NewPfcpEndpoint(addrPort netip.AddrPort) (pfcpEndpoint *PfcpEndpoint, err error) {
	if udpServer, err := udpserver.NewUDPServer(addrPort); err != nil {
		return nil, err
	} else {
		pfcpEndpoint = &PfcpEndpoint{
			UdpServer: udpServer,
		}
	}
	return
}

func (PfcpPeer *PfcpPeer) Drop() {
	PfcpPeer.Transport.Drop()
	close(PfcpPeer.RequestChan)
	close(PfcpPeer.ResponseChan)
}

func (pfcpEndpoint *PfcpEndpoint) Peer(addrPort netip.AddrPort) *PfcpPeer {
	udpPeer := pfcpEndpoint.UdpServer.Register(addrPort)
	requestChan := make(chan transport.PeerRequest)
	responseChan := make(chan transport.RequestReturn)
	transport := transport.NewTransport(udpPeer, requestChan)

	pfcpPeer := &PfcpPeer{
		UdpServer:     pfcpEndpoint.UdpServer,
		UdpServerPeer: udpPeer,
		Transport:     transport,
		RequestChan:   requestChan,
		ResponseChan:  responseChan,
	}
	return pfcpPeer
}

func (pfcpPeer *PfcpPeer) EnterRequest(pfcpMessage *pfcp.PfcpMessage) {
	pfcpPeer.Transport.EnterRequest(pfcpMessage, pfcpPeer.ResponseChan)
}

func (pfcpPeer *PfcpPeer) BlockingRequest(pfcpMessage *pfcp.PfcpMessage) (*pfcp.PfcpMessage, error) {
	return pfcpPeer.Transport.BlockingRequest(pfcpMessage)
}
