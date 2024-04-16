// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package pfcp

import (
	"encoding/binary"
	"net/netip"
	"runtime/debug"

	log "github.com/sirupsen/logrus"
)

func Encode_Uint8(u8 uint8) []byte {
	bytes := make([]byte, 1)
	bytes[0] = u8
	return bytes
}

func Encode_Uint16(u16 uint16) []byte {
	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes[:], u16)
	return bytes
}

func Encode_Uint32(u32 uint32) []byte {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes[:], u32)
	return bytes
}

func Encode_Uint64(u64 uint64) []byte {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes[:], u64)
	return bytes
}

func Encode_IpV4(ip netip.Addr) []byte {
	bytes := make([]byte, 4)
	if ip.Is4() {
		addr := ip.As4()
		copy(bytes[:], addr[:])
	} else {
		debug.PrintStack()
		log.Warn("invalid IP address")
	}
	return bytes
}

func Encode_InterfaceType(t EnumInterface) []byte {
	return Encode_Uint8(byte(t))
}

func Encode_Qfi(qfi uint8) []byte {
	if qfi&0b11000000 != 0 {
		log.Warnf("invalid QFI, %x greater than 0x20", qfi)
	}
	return Encode_Uint8(qfi)
}

// // TS29.244 8.2.37
func Encode_FSeid(seid uint64, ip netip.Addr) []byte {
	bytes := make([]byte, 1+8, 1+8+4)
	bytes[0] = 0b00000010 // hardcodes v4
	binary.BigEndian.PutUint64(bytes[1:9], seid)
	bytes = append(bytes, Encode_IpV4(ip)...)
	return bytes
}

// TS 23.003, TS29.244 8.2.4
// TODO check for variants, 29.244 implies there may be some
func Encode_APN(s string) (bytes []byte) {
	l := uint8(len(s))
	bytes = append(bytes, l)
	bytes = append(bytes, []byte(s)[:l]...)

	return
}

// RFC 1035, TS29.244 8.2.38
func Encode_FQDN(sx ...string) (bytes []byte) {
	for i := range sx {
		l := uint8(len(sx[i]))
		s := []byte(sx[i])[:l]
		bytes = append(bytes, l)
		bytes = append(bytes, s...)
	}
	return
}

// TS29.244 8.2.56
const outer_header_flag_GTPU_UDP_IPv4 = 0b00000001

func Encode_OuterHeaderCreation(teid TEID, ip netip.Addr) (bytes []byte) {
	bytes = make([]byte, 2+4, 2+4+4)
	bytes[0] = outer_header_flag_GTPU_UDP_IPv4
	binary.BigEndian.PutUint32(bytes[2:], uint32(teid))
	bytes = append(bytes, Encode_IpV4(ip)...)
	return
}

// TS29.244 8.2.64

func Encode_OuterHeaderRemoval() (bytes []byte) {
	return
}

// TS29.244 8.2.38
func Encode_NodeIdFqdn(sx ...string) (bytes []byte) {
	bytes = append(bytes, 2) // hardcodes FQDN
	bytes = append(bytes, Encode_FQDN(sx...)...)
	return
}

func Encode_NodeIdIpV4(ip netip.Addr) (bytes []byte) {
	bytes = append(bytes, 0) // hardcodes v4
	bytes = append(bytes, Encode_IpV4(ip)...)
	return
}

// TS29.244 8.2.62
const (
	ueip_flag_SD uint8 = 0b00000100 // set indicates destination
	ueip_flag_V4 uint8 = 0b00000010
	ueip_flag_V6 uint8 = 0b00000001
)

func Encode_UeIpAddress_IpV4_Dst(ip netip.Addr) (bytes []byte) {
	flags := ueip_flag_SD | ueip_flag_V4
	bytes = append(bytes, flags)
	bytes = append(bytes, Encode_IpV4(ip)...)
	return
}

// TS29.244 - 8.2.82- User Plane IP Resource Information
// Release 15 only
func Encode_UserPlaneIpResourceInformation(ip netip.Addr) (bytes []byte) {
	bytes = append(bytes, 1) // V4 flag is bit 1
	bytes = append(bytes, Encode_IpV4(ip)...)
	return
}

// TS29.244 8.2.3
const (
	fteid_flag_CHID = 0b00001000
	fteid_flag_CH   = 0b00000100
	fteid_flag_V6   = 0b00000010
	fteid_flag_V4   = 0b00000001
)

func Encode_FTeid_Choose_IpV4() (bytes []byte) {
	bytes = make([]byte, 1)
	bytes[0] = fteid_flag_CH | fteid_flag_V4
	return
}

func Encode_FTeid_IpV4(teid TEID, ip netip.Addr) (bytes []byte) {
	bytes = make([]byte, 1+4, 1+4+4)
	bytes[0] = fteid_flag_V4
	binary.BigEndian.PutUint32(bytes[1:], uint32(teid))
	bytes = append(bytes, Encode_IpV4(ip)...)
	return
}

const (
	action_flag_drop = 0b00000001
	action_flag_forw = 0b00000010
	action_flag_buff = 0b00000100
	action_flag_nocp = 0b00001000
)

func Encode_ApplyAction(action EnumAction) []byte {
	switch action {
	case EnumDrop:
		return Encode_Uint8(action_flag_drop)
	case EnumForw:
		return Encode_Uint8(action_flag_forw)
	case EnumBuff:
		return Encode_Uint8(action_flag_buff)
	case EnumNocp:
		return Encode_Uint8(action_flag_nocp)
	default:
		log.Warn("invalid action")
		return Encode_Uint8(0)
	}
}

const (
	// NB there are _2_ bits for each, but only 0 and 1 are allowed!
	gate_status_flag_ul byte = 0b00000100
	gate_status_flag_dl byte = 0b00000001
)

func Encode_GateStatus(gs GateStatus) []byte {
	var flags byte
	if gs.UlGate == EnumClosed {
		flags |= gate_status_flag_ul
	}
	if gs.DlGate == EnumClosed {
		flags |= gate_status_flag_dl
	}
	return Encode_Uint8(flags)
}

func encodeBitRate(u40 uint64) []byte {
	// bit rates are 40 bits!
	if u40 >= 2<<39 {
		log.Warn("bit rate exceeds limit of 40 bits")
	}
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes[:], u40)
	return bytes[3:]
}

func Encode_BitRates(br BitRate) (bytes []byte) {
	bytes = make([]byte, 10)
	copy(bytes[:6], encodeBitRate(br.Uplink))
	copy(bytes[5:], encodeBitRate(br.Downlink))
	return
}
