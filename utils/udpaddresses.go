// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package utils

import (
	"net"
	"net/netip"

	log "github.com/sirupsen/logrus"
)

// the below code is ugly but it seems that forcing an IPv4 net.Addr into an explicitly IPv4 netip.Addrport really is this horrible
// the panics are only in places where it is reasonable to expect the panic is unreachable
// copied from r1upf/cmd/ppcs/udpaddresses.go

func netIpToAddr(netAddr net.IP) netip.Addr {
	// knowingly panic if the slice is len(4) rather than len(16)
	if ip4, ok := netip.AddrFromSlice(netAddr[12:]); !ok {
		panic("netip.AddrFromSlice() failed")
	} else {
		return ip4
	}
}

func udpAddrToAddrPort(netAddr net.Addr) netip.AddrPort {
	if addr, err := net.ResolveUDPAddr(netAddr.Network(), netAddr.String()); err != nil {
		panic("net.ResolveUDPAddr() failed")
	} else {
		ip := netIpToAddr(addr.IP)
		log.Infof("got net.Addr %s, netip.AddrPort %s, from %s:%s", addr, netip.AddrPortFrom(ip, uint16(addr.Port)), netAddr.Network(), netAddr.String())
		return netip.AddrPortFrom(ip, uint16(addr.Port))
	}
}

func GetEndpointAddresses(serverName string, localPort uint16) (netip.AddrPort, netip.AddrPort, error) {
	var zero netip.AddrPort

	if serverAddress, err := net.ResolveUDPAddr("udp4", serverName); err != nil {
		return zero, zero, err
	} else if conn, err := net.DialUDP("udp4", nil, serverAddress); err != nil {
		return zero, zero, err
	} else {
		localAddrPort := udpAddrToAddrPort(conn.LocalAddr())
		if localPort != 0 {
			localAddr := localAddrPort.Addr()
			localAddrPort = netip.AddrPortFrom(localAddr, localPort)
		}
		peerAddrPort := udpAddrToAddrPort(conn.RemoteAddr())
		conn.Close()
		return localAddrPort, peerAddrPort, nil
	}
}
