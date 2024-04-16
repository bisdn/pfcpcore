// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package testcases

import (
	"net/netip"

	"pfcpcore/endpoint"
)

var sersmr []string = []string{"../samples/SEreq.bin", "../samples/SMreq.bin"}

type t struct{}

var exitChan chan t = make(chan t)

func Close()     { exitChan <- t{} }
func CloseWait() { <-exitChan }

var nextIp = netip.MustParseAddr("127.0.0.100")

func AddrFactory() netip.AddrPort {
	nextIp = nextIp.Next()
	return netip.AddrPortFrom(nextIp, 8805)
}

func MakeTestPeers() (*endpoint.PfcpPeer, *endpoint.PfcpPeer, error) {
	addr1 := AddrFactory()
	addr2 := AddrFactory()
	if ep1, err := endpoint.NewPfcpEndpoint(addr1); err != nil {
		return nil, nil, err
	} else if ep2, err := endpoint.NewPfcpEndpoint(addr2); err != nil {
		return nil, nil, err
	} else {
		peer1 := ep1.Peer(addr2)
		peer2 := ep2.Peer(addr1)
		return peer1, peer2, nil
	}
}
