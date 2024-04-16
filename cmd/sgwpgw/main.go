// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package main

import (
	"flag"
	"net/netip"
	"sync"

	log "github.com/sirupsen/logrus"

	"pfcpcore/endpoint"
	"pfcpcore/loginit"
	"pfcpcore/smf"
	"pfcpcore/udpserver"
)

func main() {
	var (
		mu          sync.Mutex
		sgwPgwState *SgwPgwState

		sgwParameter,
		pgwParameter,
		upfParameter,
		sgwUpParameter,
		pgwUpParameter,
		logLevel string

		sgwAddrPort,
		pgwAddrPort netip.AddrPort

		sgwUpAddr,
		pgwUpAddr netip.Addr

		err error
	)
	flag.StringVar(&sgwParameter, "sgw", "", "local signalling address for sgw function (IP:port)")
	flag.StringVar(&pgwParameter, "pgw", "", "local signalling address for pgw function (IP:port)")
	flag.StringVar(&upfParameter, "upf", "", "local and remote peer signalling addresses for UPF function (<local IP>:port,<peer IP>:port)")
	flag.StringVar(&sgwUpParameter, "sgwup", "", "local userplane (GTPu) IP address for sgw function")
	flag.StringVar(&pgwUpParameter, "pgwup", "169.254.169.253", "local userplane (GTPu) IP address for pgw function (not required,defaults to 169.254.169.253)")

	flag.StringVar(&logLevel, "loglevel", "debug", "logging level")
	flag.Parse()
	loginit.Init(logLevel)
	mu.Lock()
	log.Info("sgwpgw started")

	if sgwUpAddr, err = netip.ParseAddr(sgwUpParameter); err != nil {
		log.Fatalf("sgw user plane address required (--sgwup=ip) (got %s)", sgwUpParameter)
	} else if pgwUpAddr, err = netip.ParseAddr(pgwUpParameter); err != nil {
		log.Fatal("if specified, pgw user plane address must be valid (--pgwup=ip)")
	} else if sgwAddrPort, err = netip.ParseAddrPort(sgwParameter); err != nil {
		log.Fatal("sgw listen address required (--sgw=ip:port)")
	} else if pgwAddrPort, err = netip.ParseAddrPort(pgwParameter); err != nil {
		log.Fatal("pgw listen address required (--pgw=ip:port)")
	} else {

		var association *smf.Association = nil
		if upfParameter == "" {
			log.Warn("no UPF configured, running in test mode only")
		} else if localAddr, peerAddr, err := parseUpfParameter(upfParameter); err != nil {
			log.Fatalf("invalid UPF configuration (%s, expected (<local IP>:port,<peer IP>:port))", upfParameter)
		} else if association, err = smf.CreateAssociation(localAddr, peerAddr); err != nil {
		} else {
			sgwPgwState = newSgwPgwState(association)

			go StartXgwU(sgwPgwState, sgwAddrPort, roleSgw, sgwUpAddr)
			go StartXgwU(sgwPgwState, pgwAddrPort, rolePgw, pgwUpAddr)

			mu.Lock()
		}
	}
}

func StartXgwU(sgwPgwState *SgwPgwState, localAddrPort netip.AddrPort, role pfcpRole, gtpuAddress netip.Addr) {
	var (
		peerAddrPort netip.AddrPort
		peerEndpoint *endpoint.PfcpPeer
	)

	if passivePeerEndpoint, err := endpoint.NewPfcpEndpoint(localAddrPort); err != nil {
		log.Fatalf("sgwpgw: newConnection error %s", err.Error())
	} else {

		log.Debug("endpoint starts")

		for m := range passivePeerEndpoint.EventChannel {
			switch event := m.(type) {

			case udpserver.UdpEventNewPeer:
				log.Debugf("got new peer event from %s", event.PeerAddr)
				peerAddrPort = event.PeerAddr
				peerEndpoint = passivePeerEndpoint.Peer(peerAddrPort)
				peerEndpoint.Recirculate(event.Payload, event.PeerAddr.Port())
				config := endpoint.PfcpAssociationConfig{
					NodeName:               localAddrPort.Addr().String(),
					LocalGtpAddress:        &gtpuAddress,
					LocalSignallingAddress: localAddrPort.Addr(),
					Application:            newSgwPgwApplication(role, sgwPgwState),
					PeerEndpoint:           peerEndpoint,
				}
				upfFsm := endpoint.NewPfcpAssociationState(config)
				upfFsm.Wait()
				upfFsm.Drop()

			case udpserver.UdpEventNetworkError:
				log.Errorf(event.Err.Error())
			}
		}
	}
}
