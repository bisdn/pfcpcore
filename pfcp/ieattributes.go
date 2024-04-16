// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package pfcp

import "fmt"

const (
	CauseAccepted                uint8 = 1
	CauseUnspecified             uint8 = 64
	SessionContextNotFound       uint8 = 65
	MandatoryIeMissing           uint8 = 66
	NoEstablishedPFCPAssociation uint8 = 72
)

func (typeCode IeTypeCode) isGroupIe() bool {
	_, exists := groupIeAttributeSets[typeCode]
	return exists
}

var ieNames = map[IeTypeCode]string{
	Create_PDR:                         "Create PDR",
	PDI:                                "PDI",
	Create_FAR:                         "Create FAR",
	Forwarding_Parameters:              "Forwarding Parameters",
	Create_URR:                         "Create URR",
	Create_QER:                         "Create QER",
	Created_PDR:                        "Created PDR",
	Update_FAR:                         "Update FAR",
	Remove_PDR:                         "Remove PDR",
	Remove_FAR:                         "Remove FAR",
	Update_Forwarding_Parameters:       "Update Forwarding Parameters",
	Cause:                              "Cause",
	Source_Interface:                   "Source Interface",
	F_TEID:                             "F-TEID",
	Network_Instance:                   "Network Instance",
	SDF_Filter:                         "SDF Filter",
	Gate_Status:                        "Gate Status",
	MBR:                                "MBR",
	GBR:                                "GBR",
	Precedence:                         "Precedence",
	Volume_Threshold:                   "Volume Threshold",
	Monitoring_Time:                    "Monitoring Time",
	Reporting_Triggers:                 "Reporting Triggers",
	Report_Type:                        "Report Type",
	Destination_Interface:              "Destination Interface",
	UP_Function_Features:               "UP Function Features",
	Apply_Action:                       "Apply Action",
	PDR_ID:                             "PDR ID",
	F_SEID:                             "F-SEID",
	Node_ID:                            "Node ID",
	Measurement_Method:                 "Measurement Method",
	Measurement_Period:                 "Measurement Period",
	Usage_Report_SDR:                   "Usage Report (Session Deletion Response)",
	Usage_Report_SRR:                   "Usage Report (Session Report Response)",
	URR_ID:                             "URR ID",
	Outer_Header_Creation:              "Outer Header Creation",
	Create_BAR:                         "Create BAR",
	BAR_ID:                             "BAR ID",
	CP_Function_Features:               "CP Function Features",
	Recovery_Time_Stamp:                "Recovery Time Stamp",
	UE_IP_Address:                      "UE IP Address",
	Outer_Header_Removal:               "Outer Header Removal",
	FAR_ID:                             "FAR ID",
	QER_ID:                             "QER ID",
	PDN_Type:                           "PDN Type",
	User_Plane_IP_Resource_Information: "User Plane IP Resource Information (rel 15 only)",
	QFI:                                "QFI",
	User_ID:                            "User ID",
	APN_DNN:                            "APN/DNN",
	Source_IP_Address:                  "Source IP Address",
	PfcpsereqFlags:                     "PFCPSEReq-Flags",
	PfcpsmreqFlags:                     "PFCPSMReq-Flags",
}

func (typeCode IeTypeCode) String() string {
	if name, exists := ieNames[typeCode]; exists {
		return name
	} else {
		return fmt.Sprintf("unknown IE(%d)", typeCode)
	}
}

const (
	Create_PDR                         IeTypeCode = 1
	PDI                                IeTypeCode = 2
	Create_FAR                         IeTypeCode = 3
	Forwarding_Parameters              IeTypeCode = 4
	Create_URR                         IeTypeCode = 6
	Create_QER                         IeTypeCode = 7
	Created_PDR                        IeTypeCode = 8
	Update_PDR                         IeTypeCode = 9
	Update_FAR                         IeTypeCode = 10
	Update_Forwarding_Parameters       IeTypeCode = 11
	Remove_PDR                         IeTypeCode = 15
	Remove_FAR                         IeTypeCode = 16
	Cause                              IeTypeCode = 19
	Source_Interface                   IeTypeCode = 20
	F_TEID                             IeTypeCode = 21
	Network_Instance                   IeTypeCode = 22
	SDF_Filter                         IeTypeCode = 23
	Gate_Status                        IeTypeCode = 25
	MBR                                IeTypeCode = 26
	GBR                                IeTypeCode = 27
	Precedence                         IeTypeCode = 29
	Volume_Threshold                   IeTypeCode = 31
	Monitoring_Time                    IeTypeCode = 33
	Reporting_Triggers                 IeTypeCode = 37
	Report_Type                        IeTypeCode = 39
	Destination_Interface              IeTypeCode = 42
	UP_Function_Features               IeTypeCode = 43
	Apply_Action                       IeTypeCode = 44
	PfcpsmreqFlags                     IeTypeCode = 49
	PDR_ID                             IeTypeCode = 56
	F_SEID                             IeTypeCode = 57
	Node_ID                            IeTypeCode = 60
	Measurement_Method                 IeTypeCode = 62
	Measurement_Period                 IeTypeCode = 64
	Usage_Report_SDR                   IeTypeCode = 79
	Usage_Report_SRR                   IeTypeCode = 80
	URR_ID                             IeTypeCode = 81
	Outer_Header_Creation              IeTypeCode = 84
	Create_BAR                         IeTypeCode = 85
	BAR_ID                             IeTypeCode = 88
	CP_Function_Features               IeTypeCode = 89
	UE_IP_Address                      IeTypeCode = 93
	Outer_Header_Removal               IeTypeCode = 95
	Recovery_Time_Stamp                IeTypeCode = 96
	FAR_ID                             IeTypeCode = 108
	QER_ID                             IeTypeCode = 109
	PDN_Type                           IeTypeCode = 113
	User_Plane_IP_Resource_Information IeTypeCode = 116
	QFI                                IeTypeCode = 124
	User_ID                            IeTypeCode = 141
	APN_DNN                            IeTypeCode = 159
	PfcpsereqFlags                     IeTypeCode = 186
	Source_IP_Address                  IeTypeCode = 192
	S_NSSAI                            IeTypeCode = 257
)
