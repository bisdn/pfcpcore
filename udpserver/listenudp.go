// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package udpserver

import (
	"net"
	"net/netip"
)

// netip style wrapper for the net.ListenUDP() function
func ListenUDPAddrPort(addr netip.AddrPort) (*net.UDPConn, error) {
	return net.ListenUDP("udp", net.UDPAddrFromAddrPort(addr))
}
