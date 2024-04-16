// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package pfcp

import "fmt"

type MessageTypeCode uint16

const (
	PFCP_Heartbeat_Request              MessageTypeCode = 1
	PFCP_Heartbeat_Response             MessageTypeCode = 2
	PFCP_Association_Setup_Request      MessageTypeCode = 5
	PFCP_Association_Setup_Response     MessageTypeCode = 6
	PFCP_Association_Release_Request    MessageTypeCode = 9
	PFCP_Association_Release_Response   MessageTypeCode = 10
	PFCP_Session_Establishment_Request  MessageTypeCode = 50
	PFCP_Session_Establishment_Response MessageTypeCode = 51
	PFCP_Session_Modification_Request   MessageTypeCode = 52
	PFCP_Session_Modification_Response  MessageTypeCode = 53
	PFCP_Session_Deletion_Request       MessageTypeCode = 54
	PFCP_Session_Deletion_Response      MessageTypeCode = 55
	PFCP_Session_Report_Request         MessageTypeCode = 56
	PFCP_Session_Report_Response        MessageTypeCode = 57
)

func (typeCode MessageTypeCode) String() string {
	if s, present := messageNames[typeCode]; present {
		return s
	} else {
		return fmt.Sprintf("unknown message type (%d)", typeCode)
	}
}

var messageNames = map[MessageTypeCode]string{
	PFCP_Heartbeat_Request:              "Heartbeat Request",
	PFCP_Heartbeat_Response:             "Heartbeat Response",
	PFCP_Association_Setup_Request:      "Association Setup Request",
	PFCP_Association_Setup_Response:     "Association Setup Response",
	PFCP_Session_Establishment_Request:  "Session Establishment Request",
	PFCP_Session_Establishment_Response: "Session Establishment Response",
	PFCP_Session_Modification_Request:   "Session Modification Request",
	PFCP_Session_Modification_Response:  "Session Modification Response",
	PFCP_Session_Deletion_Request:       "Session Deletion Request",
	PFCP_Session_Deletion_Response:      "Session Deletion Response",
	PFCP_Session_Report_Request:         "Session Report Request",
	PFCP_Session_Report_Response:        "Session Report Response",
}

var requestMessageTypeCodes = []MessageTypeCode{
	PFCP_Heartbeat_Request,
	PFCP_Association_Setup_Request,
	PFCP_Session_Establishment_Request,
	PFCP_Session_Modification_Request,
	PFCP_Session_Deletion_Request,
	PFCP_Session_Report_Request,
}

func (typeCode MessageTypeCode) IsRequest() bool {
	for i := range requestMessageTypeCodes {
		if typeCode == requestMessageTypeCodes[i] {
			return true
		}
	}
	return false
}

var responseMessageTypeCodes = []MessageTypeCode{
	PFCP_Heartbeat_Response,
	PFCP_Association_Setup_Response,
	PFCP_Session_Establishment_Response,
	PFCP_Session_Modification_Response,
	PFCP_Session_Deletion_Response,
	PFCP_Session_Report_Response,
}

func (typeCode MessageTypeCode) IsResponse() bool {
	for i := range responseMessageTypeCodes {
		if typeCode == responseMessageTypeCodes[i] {
			return true
		}
	}
	return false
}

func (typeCode MessageTypeCode) IsValid() bool {
	return typeCode.IsRequest() || typeCode.IsResponse()
}
