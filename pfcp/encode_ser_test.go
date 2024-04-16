// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package pfcp

import "net/netip"

var ser2 = NewSessionMessage(

	PFCP_Session_Establishment_Request,
	1234,
	IE_NodeIdFqdn("smf0", "ng4tcom"),
	IE_FSeid(1, netip.MustParseAddr("162.118.51.1")),

	IE_CreatePdr(
		IE_PdrId(0),
		IE_Precedence(0x40),
		IE_Pdi(
			IE_SourceInterface(EnumAccess),
			IE_FTeid_Choose_IpV4(),
			IE_NetworkInstance("sgi"),
			IE_Qfi(5),
		),
		IE_FarId(0),
		IE_QerId(1074069508),
	),

	IE_CreatePdr(
		IE_PdrId(32768),
		IE_Precedence(0x20),
		IE_Pdi(
			IE_SourceInterface(EnumCore),
			IE_UeIpAddress(netip.MustParseAddr("14.0.0.4")),
			IE_NetworkInstance("epc"),
		),
		IE_FarId(1073741824),
		IE_QerId(327684),
	),

	IE_CreateFar(
		IE_FarId(0),
		IE_ApplyAction(EnumForw),
		IE_ForwardingParameters(
			IE_DestinationInterface(EnumCore),
			IE_NetworkInstance("sgi"),
		),
	),

	IE_CreateFar(
		IE_FarId(1073741824),
		IE_ApplyAction(EnumBuff),
		IE_ForwardingParameters(
			IE_DestinationInterface(EnumAccess),
			IE_NetworkInstance("epc"),
		),
	),

	IE_CreateQer(
		IE_QerId(1074069508),
		IE_GateStatus(GateOpen),
		IE_Mbr(BitRate{Uplink: 10000000, Downlink: 2560000000}),
		IE_Gbr(BitRate{Uplink: 5000000, Downlink: 1280000000}),
		IE_Qfi(5),
	),

	IE_CreateQer(
		IE_QerId(327684),
		IE_GateStatus(GateOpen),
		IE_Mbr(BitRate{Uplink: 10000000, Downlink: 2560000000}),
		IE_Gbr(BitRate{Uplink: 5000000, Downlink: 1280000000}),
		IE_Qfi(5),
	),
)
