// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package pfcp

import "net/netip"

var (
	TestSet1 []*PfcpMessage = []*PfcpMessage{
		AssociationRequest,
		AssociationResponse,
		SessionEstablishmentRequest,
		SessionEstablishmentResponse,
		SessionModificationRequest,
		SessionModificationResponse,
		SessionDeletionRequest,
		SessionDeletionResponse,
	}

	TestSet1Requests []*PfcpMessage = []*PfcpMessage{
		AssociationRequest,
		SessionEstablishmentRequest,
		SessionModificationRequest,
		SessionDeletionRequest,
	}
	TestSet1Responses []*PfcpMessage = []*PfcpMessage{
		AssociationResponse,
		SessionEstablishmentResponse,
		SessionModificationResponse,
		SessionDeletionResponse,
	}

	/*
		PFCP Message type=Association Setup Request,seid=not present, seq=0
		Node ID: 620b55cf891b93e8d253fcfaaaaaaaac : customer1
		Recovery Time Stamp: 3889157877
	*/
	AssociationRequest = NewNodeMessage(
		PFCP_Association_Setup_Request,
		IE_NodeIdFqdn("620b55cf891b93e8d253fcfaaaaaaaac", "customer1"),
		IE_RecoveryTimeStamp(3889157877),
	)

	/*
	   PFCP Message type=Association Setup Response,seid=not present, seq=0
	   Node ID: marcel@bisdnde
	   Cause: 1
	   Recovery Time Stamp: 3889157880
	*/
	AssociationResponse = NewNodeMessage(
		PFCP_Association_Setup_Response,
		IE_NodeIdFqdn("marcel", "bisdnde"),
		IE_Cause(CauseAccepted),
		IE_RecoveryTimeStamp(3889157880),
	)
	/*validation and reserialisation succeeded
	message dump:
	PFCP Message type=Session Deletion Request,seid=00000001, seq=6
	*/
	SessionDeletionRequest = NewSessionMessage(PFCP_Session_Deletion_Request, 1)

	/*
	   validation and reserialisation succeeded
	   message dump:
	   PFCP Message type=Session Deletion Response,seid=00000001, seq=6
	   Cause: 64
	*/
	SessionDeletionResponse = NewSessionMessage(PFCP_Session_Deletion_Response, 1,
		IE_Cause(CauseUnspecified),
	)

	/*
			PFCP Message type=Session Establishment Request,seid=00000000, seq=3
		Node ID: 620b55cf891b93e8d253fcfaaaaaaaac : customer1
		F-SEID: 0000000001 : 162.118.51.1
		Create PDR (0)
		  PDR ID:     0
		  Precedence: 0000000040
		  PDI
		    Source Interface: Access
		    F-TEID: 824634925736 : 162.117.255.201
		    QFI: 05
		  Outer Header Removal: 0
		  FAR ID: 0000000000
		  QER ID: 1074069508
		Create PDR (32768)
		  PDR ID: 32768
		  Precedence: 0000000020
		  PDI
		    Source Interface: Core
		    UE IP Address: dst 14.0.0.2
		  FAR ID: 1073741824
		  QER ID: 0000327684
		Create FAR (0)
		  FAR ID: 0000000000
		  Apply Action: FORW
		  Forwarding Parameters
		    Destination Interface: Core
		Create FAR (1073741824)
		  FAR ID: 1073741824
		  Apply Action: BUFF
		  Forwarding Parameters
		    Destination Interface: Access
		Create QER (1074069508)
		  QER ID: 1074069508
		  Gate Status: 00
		  MBR: Up:10000000 Dn:10000000
		  GBR: Up:5000000 Dn:5000000
		  QFI: 05
		Create QER (327684)
		  QER ID: 0000327684
		  Gate Status: 00
		  MBR: Up:10000000 Dn:10000000
		  GBR: Up:5000000 Dn:5000000
		  QFI: 05
		PDN Type: 1

	*/

	SessionEstablishmentRequest = NewSessionMessage(

		PFCP_Session_Establishment_Request,
		1234,
		IE_NodeIdFqdn("620b55cf891b93e8d253fcfaaaaaaaac", "customer1"),
		IE_FSeid(1, netip.MustParseAddr("162.118.51.1")),

		IE_CreatePdr(
			IE_PdrId(0),
			IE_Precedence(0x40),
			IE_Pdi(
				IE_SourceInterface(EnumAccess),
				IE_FTeid_IpV4(1234, netip.MustParseAddr("162.118.51.1")),

				IE_NetworkInstance("sgi"),
				IE_Qfi(5),
			),
			IE_OuterHeaderRemoval(),
			IE_FarId(0),
			IE_QerId(1074069508),
		),

		IE_CreatePdr(
			IE_PdrId(32768),
			IE_Precedence(0x20),
			IE_Pdi(
				IE_SourceInterface(EnumCore),
				IE_UeIpAddress(netip.MustParseAddr("14.0.0.2")),
			),
			IE_FarId(1073741824),
			IE_QerId(327684),
		),

		IE_CreateFar(
			IE_FarId(0),
			IE_ApplyAction(EnumForw),
			IE_ForwardingParameters(
				IE_DestinationInterface(EnumCore),
			),
		),

		IE_CreateFar(
			IE_FarId(1073741824),
			IE_ApplyAction(EnumBuff),
			IE_ForwardingParameters(
				IE_DestinationInterface(EnumAccess),
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

	/*
	   PFCP Message type=Session Establishment Response,seid=00000001, seq=3
	   Node ID: marcel@bisdnde
	   Cause: 1
	   F-SEID: 0000000001 : 162.118.255.42
	*/

	SessionEstablishmentResponse = NewSessionMessage(PFCP_Session_Establishment_Response, 1,

		IE_NodeIdFqdn("marcel", "bisdnde"),
		IE_Cause(CauseAccepted),
		IE_FSeid(1, netip.MustParseAddr("162.118.255.42")),
	)

	/*
		PFCP Message type=Session Modification Request,seid=00000001, seq=4
		Update FAR (1073741824)
		  FAR ID: 1073741824
		  Apply Action: FORW
		  Update Forwarding Parameters
		    Outer Header Creation: teid:00001 ipv4:162.117.1.1
	*/

	SessionModificationRequest = NewSessionMessage(
		PFCP_Session_Modification_Request,
		1,
		IE_UpateFar(
			IE_FarId(1073741824),
			IE_ApplyAction(EnumForw),
			IE_UpdateForwardingParameters(
				IE_OuterHeaderCreation(1, netip.MustParseAddr("162.117.1.1")),
			),
		),
	)

	/*
		PFCP Message type=Session Modification Response,seid=00000001, seq=4
		Cause: 1
	*/

	SessionModificationResponse = NewSessionMessage(PFCP_Session_Modification_Response, 1,
		IE_Cause(CauseAccepted),
	)
)
