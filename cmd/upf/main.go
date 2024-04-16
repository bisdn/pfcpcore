// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package main

import (
	"net/netip"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"

	"pfcpcore/endpoint"
	"pfcpcore/loginit"
	"pfcpcore/udpserver"
)

var mu sync.Mutex

func main() {
	loginit.Init("trace")
	mu.Lock()
	log.Info("upf started")

	if len(os.Args) < 2 {
		log.Fatal(("listen address required"))
	} else if upfAddrPort, err := netip.ParseAddrPort(os.Args[1]); err != nil {
		log.Fatalf("valid listen address required, ip:port expected, got %s", os.Args[1])
	} else {
		go Start(upfAddrPort)
		mu.Lock()
	}
}

func Start(localAddrPort netip.AddrPort) {
	if localEndpoint, err := endpoint.NewPfcpEndpoint(localAddrPort); err != nil {
		log.Fatalf("newConnection error %s", err.Error())
	} else {

		log.Debug("UPF endpoint starts")

		var (
			peerAddrPort netip.AddrPort
			peerEndpoint *endpoint.PfcpPeer
		)

		for m := range localEndpoint.EventChannel {
			switch event := m.(type) {

			case udpserver.UdpEventNewPeer:
				log.Debugf("UPF got new peer event from %s", event.PeerAddr)
				peerAddrPort = event.PeerAddr
				peerEndpoint = localEndpoint.Peer(peerAddrPort)
				peerEndpoint.Recirculate(event.Payload, event.PeerAddr.Port())
				upfFsm := endpoint.NewPfcpAssociationState(endpoint.PfcpAssociationConfig{
					PeerName:               "",
					NodeName:               localAddrPort.Addr().String(),
					LocalSignallingAddress: localAddrPort.Addr(),
					Application:            &UpfApplication{},
					PeerEndpoint:           peerEndpoint,
				})

				upfFsm.Wait()
				upfFsm.Drop()

			case udpserver.UdpEventNetworkError:
				log.Errorf(event.Err.Error())
			}
		}
	}
}
