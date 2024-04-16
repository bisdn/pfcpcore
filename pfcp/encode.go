// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package pfcp

import (
	"net/netip"
)

// ===================================================
// plain IEs
// ===================================================

func IE_Cause(code uint8) IeNode {
	// TODO make cause codes an enum type
	return *NewIeNode(Cause, Encode_Uint8(code))
}

func IE_RecoveryTimeStamp(u32 uint32) IeNode {
	return *NewIeNode(Recovery_Time_Stamp, Encode_Uint32(u32))
}

func IE_NodeIdFqdn(sx ...string) IeNode {
	return *NewIeNode(Node_ID, Encode_NodeIdFqdn(sx...))
}

func IE_NodeIdIpV4(ip netip.Addr) IeNode {
	return *NewIeNode(Node_ID, Encode_NodeIdIpV4(ip))
}

// TS29.244: either or both ipv4 and ipv6 are required
// in addition to the 64 bit SEID.
// TDOD - for now only implement ipv4
func IE_FSeid(seid uint64, ip netip.Addr) IeNode {
	return *NewIeNode(F_SEID, Encode_FSeid(seid, ip))
}

func IE_PdrId(u16 uint16) IeNode {
	return *NewIeNode(PDR_ID, Encode_Uint16(u16))
}

func IE_FarId(u32 uint32) IeNode {
	return *NewIeNode(FAR_ID, Encode_Uint32(u32))
}

func IE_QerId(u32 uint32) IeNode {
	return *NewIeNode(QER_ID, Encode_Uint32(u32))
}

func IE_Precedence(u32 uint32) IeNode {
	return *NewIeNode(Precedence, Encode_Uint32(u32))
}

func IE_SourceInterface(t EnumInterface) IeNode {
	return *NewIeNode(Source_Interface, Encode_InterfaceType(t))
}

func IE_FTeid_Choose_IpV4() IeNode {
	return *NewIeNode(F_TEID, Encode_FTeid_Choose_IpV4())
}

func IE_FTeid_IpV4(teid TEID, ip netip.Addr) IeNode {
	return *NewIeNode(F_TEID, Encode_FTeid_IpV4(teid, ip))
}

func IE_NetworkInstance(s string) IeNode {
	return *NewIeNode(Network_Instance, Encode_APN(s))
}

func IE_Qfi(qfi uint8) IeNode {
	return *NewIeNode(QFI, Encode_Qfi(qfi))
}

func IE_GateStatus(gs GateStatus) IeNode {
	return *NewIeNode(Gate_Status, Encode_GateStatus(gs))
}

func IE_UeIpAddress(ip netip.Addr) IeNode {
	return *NewIeNode(UE_IP_Address, Encode_UeIpAddress_IpV4_Dst(ip))
}

func IE_UpIpRsrcInfo(ip netip.Addr) IeNode {
	return *NewIeNode(User_Plane_IP_Resource_Information, Encode_UserPlaneIpResourceInformation(ip))
}

func IE_ApplyAction(action EnumAction) IeNode {
	return *NewIeNode(Apply_Action, Encode_ApplyAction(action))
}

func IE_DestinationInterface(t EnumInterface) IeNode {
	return *NewIeNode(Destination_Interface, Encode_InterfaceType(t))
}

func IE_OuterHeaderCreation(teid TEID, ip netip.Addr) IeNode {
	return *NewIeNode(Outer_Header_Creation, Encode_OuterHeaderCreation(teid, ip))
}

func IE_OuterHeaderRemoval() IeNode {
	return *NewIeNode(Outer_Header_Removal, Encode_OuterHeaderRemoval())
}

func IE_Mbr(br BitRate) IeNode {
	return *NewIeNode(MBR, Encode_BitRates(br))
}

func IE_Gbr(br BitRate) IeNode {
	return *NewIeNode(GBR, Encode_BitRates(br))
}

// ===================================================
// grouped IEs
// ===================================================

func IE_CreatePdr(nodes ...IeNode) IeNode {
	return *NewGroupNode(Create_PDR, nodes...)
}

func IE_UpdatePdr(nodes ...IeNode) IeNode {
	return *NewGroupNode(Update_PDR, nodes...)
}

func IE_CreateFar(nodes ...IeNode) IeNode {
	return *NewGroupNode(Create_FAR, nodes...)
}

func IE_UpdateFar(nodes ...IeNode) IeNode {
	return *NewGroupNode(Update_FAR, nodes...)
}

// TODO this need only take the PDR ID parameter
func IE_RemovePdr(nodes ...IeNode) IeNode {
	return *NewGroupNode(Remove_PDR, nodes...)
}

// TODO this need only take the FAR ID parameter
func IE_RemoveFar(nodes ...IeNode) IeNode {
	return *NewGroupNode(Remove_FAR, nodes...)
}

func IE_CreateQer(nodes ...IeNode) IeNode {
	return *NewGroupNode(Create_QER, nodes...)
}

func IE_Pdi(nodes ...IeNode) IeNode {
	return *NewGroupNode(PDI, nodes...)
}

func IE_ForwardingParameters(nodes ...IeNode) IeNode {
	return *NewGroupNode(Forwarding_Parameters, nodes...)
}

func IE_UpateFar(nodes ...IeNode) IeNode {
	return *NewGroupNode(Update_FAR, nodes...)
}

func IE_UpdateForwardingParameters(nodes ...IeNode) IeNode {
	return *NewGroupNode(Update_Forwarding_Parameters, nodes...)
}
