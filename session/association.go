// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package session

import "pfcpcore/pfcp"

type AssociationSetupBody struct {
	NodeId            string // probably not rich enough, can also be an IP address i think
	RecoveryTimestamp uint32
}

func ParseAssociationSetupRequest(ies *pfcp.IeNode) (*AssociationSetupBody, error) {
	root := ies.Getter()
	if nodeId, err := root.GetByTc(pfcp.Node_ID).DeserialiseNodeIdString(); err != nil { // need a deserialiser and type for node ID
		return nil, err
	} else if recoveryTimestamp, err := root.GetByTc(pfcp.Recovery_Time_Stamp).DeserialiseU32(); err != nil {
		return nil, err
	} else {
		return &AssociationSetupBody{
			NodeId:            nodeId,
			RecoveryTimestamp: recoveryTimestamp,
		}, nil
	}
}

func (AssociationSetupBody *AssociationSetupBody) EncodeResponse() []pfcp.IeNode {
	return []pfcp.IeNode{
		pfcp.IE_NodeIdFqdn(AssociationSetupBody.NodeId),
		pfcp.IE_RecoveryTimeStamp(AssociationSetupBody.RecoveryTimestamp),
		pfcp.IE_Cause(pfcp.CauseAccepted),
	}
}
