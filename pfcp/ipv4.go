// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package pfcp

import (
	"encoding/binary"
	"fmt"
	"net/netip"
)

type IpV4 uint32

func (IpV4 IpV4) String() string {
	var bytes [4]byte
	binary.BigEndian.PutUint32(bytes[:], uint32(IpV4))
	return fmt.Sprintf("%d.%d.%d.%d", bytes[0], bytes[1], bytes[2], bytes[3])
}

func MustParseAddr(s string) IpV4 {
	if ip, err := ParseAddr(s); err != nil {
		panic(fmt.Sprintf("could not parse %s as IPv4, %s", s, err.Error()))
	} else {
		return ip
	}
}

func ParseAddr(s string) (IpV4, error) {
	if netip, err := netip.ParseAddr(s); err != nil {
		return IpV4(0), err
	} else {
		return NetIpAddrToIpV4(netip), nil
	}
}

func (IpV4 IpV4) Addr() netip.Addr {
	return NetipUint32(uint32(IpV4))
}

func ReadIpV4(bytes []byte) IpV4 {
	if len(bytes) != 4 {
		panic("IpV4 must have length 4")
	} else {
		return IpV4(binary.BigEndian.Uint32(bytes[:]))
	}
}

// TODO both below functions will be unified to handle IpV4 not u32

func NetipUint32(u32 uint32) netip.Addr {
	var bytes [4]byte
	binary.BigEndian.PutUint32(bytes[:], u32)
	if ip4, ok := netip.AddrFromSlice(bytes[:]); !ok {
		panic("netip.AddrFromSlice() failed")
	} else {
		return ip4
	}
}

// untested, as yet unused...

func NetIpAddrToUInt32(netAddr netip.Addr) uint32 {
	sl := netAddr.AsSlice()
	return uint32(sl[3]) | uint32(sl[2])<<8 | uint32(sl[1])<<16 | uint32(sl[0])<<24
}

func NetIpAddrToIpV4(netAddr netip.Addr) IpV4 {
	return IpV4(NetIpAddrToUInt32(netAddr))
}
