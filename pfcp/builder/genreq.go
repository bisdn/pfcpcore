// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package builder

import (
	"pfcpcore/pfcp"
)

const (
	networkInstance             = "internet"
	defaultPdrPrecedence uint32 = 255
)

type InterfaceRole uint8

const (
	RoleAccess InterfaceRole = iota + 1
	RoleCore
)

// TODO make return value struct
func (role InterfaceRole) interfaces() (destinationInterface, sourceInterface pfcp.EnumInterface) {
	switch role {
	case RoleCore:
		destinationInterface = pfcp.EnumAccess
		sourceInterface = pfcp.EnumCore

	case RoleAccess:
		destinationInterface = pfcp.EnumCore
		sourceInterface = pfcp.EnumAccess
	}
	return
}

type RequestMode uint8

const (
	ModeCreate RequestMode = iota + 1
	ModeUpdate
)

func (mode RequestMode) makeFar() func(...pfcp.IeNode) pfcp.IeNode {
	switch mode {
	case ModeCreate:
		return pfcp.IE_CreateFar
	case ModeUpdate:
		return pfcp.IE_UpdateFar
	default:
		panic("unallowed mode")
	}
}

func (mode RequestMode) makePdr() func(...pfcp.IeNode) pfcp.IeNode {
	switch mode {
	case ModeCreate:
		return pfcp.IE_CreatePdr
	case ModeUpdate:
		return pfcp.IE_UpdatePdr
	default:
		panic("unallowed mode")
	}
}

func DecapPdr(mode RequestMode, fteid pfcp.FTeid) pfcp.IeNode {
	return mode.makePdr()(
		pfcp.IE_PdrId(uint16(RoleAccess)),
		pfcp.IE_Precedence(defaultPdrPrecedence),
		pfcp.IE_Pdi(
			pfcp.IE_SourceInterface(pfcp.EnumAccess),
			pfcp.IE_FTeid_IpV4(*fteid.Teid, fteid.IpV4.Addr()),
			pfcp.IE_NetworkInstance(networkInstance),
		),
		pfcp.IE_OuterHeaderRemoval(),
		pfcp.IE_FarId(uint32(RoleAccess)),
	)
}

func DecapFar(mode RequestMode) pfcp.IeNode {
	return mode.makeFar()(
		pfcp.IE_FarId(uint32(RoleAccess)),
		pfcp.IE_ApplyAction(pfcp.EnumForw),
		pfcp.IE_ForwardingParameters(
			pfcp.IE_DestinationInterface(pfcp.EnumCore),
			pfcp.IE_NetworkInstance(networkInstance),
		),
	)
}

func EncapPdr(mode RequestMode, ueIp pfcp.IpV4) pfcp.IeNode {
	return mode.makePdr()(
		pfcp.IE_PdrId(uint16(RoleCore)),
		pfcp.IE_Precedence(defaultPdrPrecedence),
		pfcp.IE_Pdi(
			pfcp.IE_SourceInterface(pfcp.EnumCore),
			pfcp.IE_NetworkInstance(networkInstance),
			pfcp.IE_UeIpAddress(ueIp.Addr()),
		),
		pfcp.IE_FarId(uint32(RoleCore)),
	)
}

func EncapFar(mode RequestMode, fteid pfcp.FTeid) pfcp.IeNode {
	return mode.makeFar()(
		pfcp.IE_FarId(uint32(RoleCore)),
		pfcp.IE_ApplyAction(pfcp.EnumForw),
		pfcp.IE_ForwardingParameters(
			pfcp.IE_DestinationInterface(pfcp.EnumAccess),
			pfcp.IE_NetworkInstance(networkInstance),
			pfcp.IE_OuterHeaderCreation(*fteid.Teid, fteid.IpV4.Addr()),
		),
	)
}

func TunnelPdr(mode RequestMode, role InterfaceRole, fteid pfcp.FTeid) pfcp.IeNode {
	_, sourceInterface := role.interfaces()

	return mode.makePdr()(
		pfcp.IE_PdrId(uint16(role)),
		pfcp.IE_Precedence(defaultPdrPrecedence),
		pfcp.IE_Pdi(
			pfcp.IE_SourceInterface(sourceInterface),
			pfcp.IE_FTeid_IpV4(*fteid.Teid, fteid.IpV4.Addr()),
			pfcp.IE_NetworkInstance(networkInstance),
		),
		pfcp.IE_FarId(uint32(role)),
	)
}

// TODO the tunnel make pdr/far functions should call these custom ones
func CustomOhcFar(farId uint32, mode RequestMode, destinationInterface pfcp.EnumInterface, fteid pfcp.FTeid) pfcp.IeNode {
	return mode.makeFar()(
		pfcp.IE_FarId(farId),
		pfcp.IE_ApplyAction(pfcp.EnumForw),
		pfcp.IE_ForwardingParameters(
			pfcp.IE_DestinationInterface(destinationInterface),
			pfcp.IE_NetworkInstance(networkInstance),
			pfcp.IE_OuterHeaderCreation(*fteid.Teid, fteid.IpV4.Addr()),
		),
	)
}

func CustomOhcPdr(pdrId uint16, farId uint32, mode RequestMode, sourceInterface pfcp.EnumInterface, fteid pfcp.FTeid) pfcp.IeNode {
	return mode.makePdr()(
		pfcp.IE_PdrId(pdrId),
		pfcp.IE_Precedence(defaultPdrPrecedence),
		pfcp.IE_Pdi(
			pfcp.IE_SourceInterface(sourceInterface),
			pfcp.IE_FTeid_IpV4(*fteid.Teid, fteid.IpV4.Addr()),
			pfcp.IE_NetworkInstance(networkInstance),
		),
		pfcp.IE_FarId(farId),
	)
}

func TunnelFar(mode RequestMode, role InterfaceRole, fteid pfcp.FTeid) pfcp.IeNode {
	destinationInterface, _ := role.interfaces()

	return mode.makeFar()(
		pfcp.IE_FarId(uint32(role)),
		pfcp.IE_ApplyAction(pfcp.EnumForw),
		pfcp.IE_ForwardingParameters(
			pfcp.IE_DestinationInterface(destinationInterface),
			pfcp.IE_NetworkInstance(networkInstance),
			pfcp.IE_OuterHeaderCreation(*fteid.Teid, fteid.IpV4.Addr()),
		),
	)
}

func BufferFar(mode RequestMode) pfcp.IeNode {
	// only used for now on downlinks, so role is fixed as destination = interface

	return mode.makeFar()(
		pfcp.IE_FarId(uint32(RoleCore)),
		pfcp.IE_ApplyAction(pfcp.EnumBuff),
		pfcp.IE_ForwardingParameters(
			pfcp.IE_DestinationInterface(pfcp.EnumAccess),
		),
	)
}
