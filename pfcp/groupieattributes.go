// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package pfcp

import (
	"fmt"

	"golang.org/x/exp/maps"
)

type IeID uint32 // actual ID values vary from 8 bits to 32 bits
const IEInvalid IeID = 0xffffffff

func (ieId *IeID) String() string {
	if ieId == nil {
		return "nil"
	} else {
		return fmt.Sprintf("%d", uint32(*ieId))
	}
}

type groupIeAttributes struct {
	required bool
	multiple bool
	isID     bool
	isDelete bool
	isUpdate bool
	baseIe   IeTypeCode
}
type groupIeAttributeSet map[IeTypeCode]groupIeAttributes

// these are compile time known values, for real use an init function should probably cache them in a var
func (set groupIeAttributeSet) properties() (countRequired uint8, idRequired bool) {
	for _, attributes := range set {
		if attributes.required {
			countRequired += 1
		}
		if attributes.isID {
			idRequired = true
		}
	}
	return
}

// required is the actual map not just the count of it
// If there were no perf concerns (compile time, not run time, eval)
// this would be returned in properties.
// This value is only used to report the missing elt when a check fails based on the count.
// TDOD - use init() for at least some of this....
func (set groupIeAttributeSet) required() map[IeTypeCode]struct{} {
	requiredMap := make(map[IeTypeCode]struct{})
	for ieCode, attributes := range set {
		if attributes.required {
			requiredMap[ieCode] = struct{}{}
		}
	}
	return requiredMap
}

func setRemove(a, b map[IeTypeCode]struct{}) []IeTypeCode {
	for k := range b {
		delete(a, k)
	}
	return maps.Keys(a)
}

var groupIeAttributeSets = map[IeTypeCode]groupIeAttributeSet{
	Create_PDR: {
		PDR_ID:               groupIeAttributes{required: true, isID: true},
		PDI:                  groupIeAttributes{required: true},
		Precedence:           groupIeAttributes{},
		Outer_Header_Removal: groupIeAttributes{},
		FAR_ID:               groupIeAttributes{},
		URR_ID:               groupIeAttributes{},
		QER_ID:               groupIeAttributes{multiple: true},
	},
	Create_FAR: {
		FAR_ID:                groupIeAttributes{required: true, isID: true},
		Forwarding_Parameters: groupIeAttributes{},
		BAR_ID:                groupIeAttributes{},
		Apply_Action:          groupIeAttributes{},
	},
	Create_URR: {
		URR_ID:             groupIeAttributes{required: true, isID: true},
		Volume_Threshold:   groupIeAttributes{},
		Monitoring_Time:    groupIeAttributes{},
		Reporting_Triggers: groupIeAttributes{},
		Measurement_Method: groupIeAttributes{},
		Measurement_Period: groupIeAttributes{},
	},
	Create_BAR: {
		BAR_ID: groupIeAttributes{required: true, isID: true},
	},
	Create_QER: {
		QER_ID:      groupIeAttributes{required: true, isID: true},
		Gate_Status: groupIeAttributes{},
		MBR:         groupIeAttributes{},
		GBR:         groupIeAttributes{},
		QFI:         groupIeAttributes{},
	},
	PDI: {
		Source_Interface: groupIeAttributes{required: true},
		F_TEID:           groupIeAttributes{},
		Network_Instance: groupIeAttributes{},
		UE_IP_Address:    groupIeAttributes{},
		QFI:              groupIeAttributes{},
		SDF_Filter:       groupIeAttributes{},
	},
	Forwarding_Parameters: {
		Destination_Interface: groupIeAttributes{},
		Network_Instance:      groupIeAttributes{},
		Outer_Header_Creation: groupIeAttributes{},
	},
	Created_PDR: {
		PDR_ID:               groupIeAttributes{required: true, isID: true},
		Precedence:           groupIeAttributes{},
		PDI:                  groupIeAttributes{},
		Outer_Header_Removal: groupIeAttributes{},
		FAR_ID:               groupIeAttributes{},
		QER_ID:               groupIeAttributes{multiple: true},
		F_TEID:               groupIeAttributes{},
	},
	Update_FAR: {
		Forwarding_Parameters:        groupIeAttributes{},
		Update_Forwarding_Parameters: groupIeAttributes{isUpdate: true, baseIe: Forwarding_Parameters},
		Apply_Action:                 groupIeAttributes{},
		FAR_ID:                       groupIeAttributes{required: true, isID: true},
		BAR_ID:                       groupIeAttributes{},
	},
	Remove_FAR: {
		FAR_ID: groupIeAttributes{required: true, isID: true},
	},
	Remove_PDR: {
		PDR_ID: groupIeAttributes{required: true, isID: true},
	},
	Update_Forwarding_Parameters: {
		Destination_Interface: groupIeAttributes{},
		Network_Instance:      groupIeAttributes{},
		Outer_Header_Creation: groupIeAttributes{},
		PfcpsmreqFlags:        groupIeAttributes{},
	},
}

var MessageIeAttributeSets = map[MessageTypeCode]groupIeAttributeSet{
	PFCP_Heartbeat_Request: {
		Recovery_Time_Stamp: groupIeAttributes{required: true},
		Source_IP_Address:   groupIeAttributes{},
	},
	PFCP_Heartbeat_Response: {
		Recovery_Time_Stamp: groupIeAttributes{required: true},
	},

	PFCP_Session_Establishment_Request: {
		Node_ID:        groupIeAttributes{required: true},
		F_SEID:         groupIeAttributes{required: true},
		Create_PDR:     groupIeAttributes{required: true, multiple: true},
		Create_FAR:     groupIeAttributes{required: true, multiple: true},
		Create_URR:     groupIeAttributes{multiple: true},
		Create_QER:     groupIeAttributes{multiple: true},
		Create_BAR:     groupIeAttributes{multiple: true},
		PDN_Type:       groupIeAttributes{},
		User_ID:        groupIeAttributes{},
		APN_DNN:        groupIeAttributes{},
		SDF_Filter:     groupIeAttributes{},
		S_NSSAI:        groupIeAttributes{},
		PfcpsereqFlags: groupIeAttributes{},
	},
	PFCP_Session_Establishment_Response: {
		Node_ID:     groupIeAttributes{required: true},
		F_SEID:      groupIeAttributes{}, // 'required: true' only if cause is success (hard to encode here)
		Created_PDR: groupIeAttributes{multiple: true},
		Cause:       groupIeAttributes{required: true},
	},

	PFCP_Session_Modification_Request: {
		Update_FAR: groupIeAttributes{multiple: true, isUpdate: true, baseIe: Create_FAR},
		Update_PDR: groupIeAttributes{multiple: true, isUpdate: true, baseIe: Create_PDR},
		Remove_FAR: groupIeAttributes{multiple: true, isDelete: true, baseIe: Create_FAR},
		Remove_PDR: groupIeAttributes{multiple: true, isDelete: true, baseIe: Create_PDR},
		Create_FAR: groupIeAttributes{multiple: true},
		Create_PDR: groupIeAttributes{multiple: true},
	},
	PFCP_Session_Modification_Response: {
		Cause: groupIeAttributes{required: true},
	},

	PFCP_Session_Deletion_Request: {},
	PFCP_Session_Deletion_Response: {
		Cause:            groupIeAttributes{required: true},
		Usage_Report_SDR: {multiple: true},
	},

	PFCP_Association_Setup_Request: {
		Node_ID:              groupIeAttributes{required: true},
		Recovery_Time_Stamp:  groupIeAttributes{required: true},
		CP_Function_Features: groupIeAttributes{},
		UP_Function_Features: groupIeAttributes{},
	},
	PFCP_Association_Setup_Response: {
		Cause:                              groupIeAttributes{required: true},
		Node_ID:                            groupIeAttributes{required: true},
		Recovery_Time_Stamp:                groupIeAttributes{required: true},
		CP_Function_Features:               groupIeAttributes{},
		UP_Function_Features:               groupIeAttributes{},
		User_Plane_IP_Resource_Information: groupIeAttributes{}, // an R15 only IE, allowed for now, because no way to manage flexibility
	},

	PFCP_Session_Report_Request: {
		Report_Type:      {},
		Usage_Report_SRR: {multiple: true},
	},
	PFCP_Session_Report_Response: {
		Cause: groupIeAttributes{required: true},
	},
}
