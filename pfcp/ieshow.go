// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package pfcp

import (
	"encoding/binary"
	"fmt"
)

type ieType uint8

const (
	ieTgroup    ieType = iota
	ieTid       ieType = iota
	ieTenum     ieType = iota
	ieTintegral ieType = iota
	ieTstring   ieType = iota
	ieTbytes    ieType = iota
	ieTspecial  ieType = iota
	ieTbits     ieType = iota

	ieTenumInterface     ieType = iota
	ieTueIpAddress       ieType = iota
	ieTbitRates          ieType = iota
	ieTapn               ieType = iota
	ieTOuterHeaderCreate ieType = iota
	ieTApplyAction       ieType = iota
	ieTfteid             ieType = iota
	ieTfseid             ieType = iota
	ieTnodeid            ieType = iota
	ieTUPIpResInfo       ieType = iota
)

var ieTypes = map[IeTypeCode]ieType{
	Create_PDR:                   ieTgroup,
	PDI:                          ieTgroup,
	Create_FAR:                   ieTgroup,
	Forwarding_Parameters:        ieTgroup,
	Create_URR:                   ieTgroup,
	Create_QER:                   ieTgroup,
	Created_PDR:                  ieTgroup,
	Update_FAR:                   ieTgroup,
	Update_Forwarding_Parameters: ieTgroup,
	// all IEs below 19 are group Ies (and a few above)
	Cause:            ieTenum,
	Source_Interface: ieTenumInterface, // only 4 bits used
	F_TEID:           ieTfteid,         // IPV4/6+TEID, TEID 32 bits mandatory
	Network_Instance: ieTapn,
	SDF_Filter:       ieTbytes,
	Gate_Status:      ieTbits,     // only LSB 1,3of 8 used
	MBR:              ieTbitRates, // two 32 bit numbers
	GBR:              ieTbitRates, // two 32 bit numbers
	Precedence:       ieTintegral, // 32 bits
	Volume_Threshold: ieTspecial,  // up to 3 64 bit numbers
	// Monitoring_Time:                    "Monitoring Time",
	// Reporting_Triggers:                 "Reporting Triggers",
	// Report_Type:                        "Report Type",
	Destination_Interface: ieTenumInterface, // only 4 bits used - see Source_Interface
	// UP_Function_Features:               "UP Function Features",
	Apply_Action: ieTApplyAction, // 11bits used
	PDR_ID:       ieTid,          // 16 bits
	F_SEID:       ieTfseid,       // IPV4/6+SEID, SEID 64 bits mandatory
	Node_ID:      ieTnodeid,      // one of string or IPv4/6, IPs not strings...
	// Measurement_Method:                 "Measurement Method",
	// Measurement_Period:                 "Measurement Period",
	// Usage_Report_SDR:                   "Usage Report (Session Deletion Response)",
	// Usage_Report_SRR:                   "Usage Report (Session Report Response)",
	URR_ID:                ieTid,                // 32 bits
	Outer_Header_Creation: ieTOuterHeaderCreate, // many forms, GTPu TEID is our main interest, 32 bits
	Create_BAR:            ieTgroup,
	BAR_ID:                ieTid, // 8 bits
	// CP_Function_Features:               "CP Function Features",
	Recovery_Time_Stamp:                ieTintegral,    // 32 bits, seconds since 01/01/1900 00:00:00
	UE_IP_Address:                      ieTueIpAddress, // can be ipv4 or 6, or empty, requesting them...
	Outer_Header_Removal:               ieTenum,        // strictly not an enum, because another bit can be set...
	FAR_ID:                             ieTid,          // 32 bits
	QER_ID:                             ieTid,          // 32 bits
	PDN_Type:                           ieTenum,        // 3 bits, ip4 ip6 ip4/6 eth other
	User_Plane_IP_Resource_Information: ieTUPIpResInfo, // "User Plane IP Resource Information (rel 15 only)",
	// QFI:                                "QFI",
	// User_ID:                            "User ID",
	APN_DNN:           ieTapn,
	Source_IP_Address: ieTspecial, // can be ipv4 or 6,optionally with a mask
}

func (node IeNode) deserialiseIntegral() uint64 {
	switch ieTypes[node.IeTypeCode] {
	case ieTid, ieTintegral, ieTenumInterface, ieTenum:
		return readIntegral(node.bytes)
	default:
		panic(fmt.Sprintf("deserialiseIntegral: invalid IE type %s", node.IeTypeCode))
	}
}

func (node IeNode) DeserialiseEnumInterface() EnumInterface {
	if ieTypes[node.IeTypeCode] != ieTenumInterface {
		panic(fmt.Sprintf("deserialiseEnumInterface: invalid IE type %s", node.IeTypeCode))
	} else if enumInterface, err := readEnumInterface(node.bytes); err != nil {
		panic(fmt.Sprintf("deserialiseEnumInterface: invalid IE type %s", node.IeTypeCode))
	} else {
		return enumInterface
	}
}

func (node IeNode) show() (s string) {
	switch ieTypes[node.IeTypeCode] {
	case ieTgroup,
		ieTstring,
		ieTbytes,
		ieTspecial,
		ieTbits:
		return showBytes(node.bytes)
	case ieTid, ieTintegral:
		return showIntegral(node.bytes)
	case ieTenum:
		return showEnum(node.bytes)
	case ieTenumInterface:
		return showEnumInterface(node.bytes)
	case ieTueIpAddress:
		return showUeIpAddress(node.bytes)
	case ieTbitRates:
		return showBitRates(node.bytes)
	case ieTapn:
		return showAPNString(node.bytes)
	case ieTfseid:
		return showFSeid(node.bytes)
	case ieTfteid:
		return showFTeid(node.bytes)
	case ieTnodeid:
		return showNodeId(node.bytes)
	case ieTOuterHeaderCreate:
		return showOuterHeaderCreate(node.bytes)
	case ieTApplyAction:
		return showApplyAction(node.bytes)
	case ieTUPIpResInfo:
		return showUpIpResInfo(node.bytes)
	}
	return ""
}

func showIntegral(bytes []byte) (s string) {
	if len(bytes) > 8 {
		return showBytes(bytes)
	} else {
		u64 := readIntegral(bytes)
		switch {
		case len(bytes) > 8:
			s = fmt.Sprintf("%0x", u64)
		case len(bytes) > 4:
			s = fmt.Sprintf("%020d", u64)
		case len(bytes) > 2:
			s = fmt.Sprintf("%010d", u64)
		case len(bytes) > 1:
			s = fmt.Sprintf("%5d", u64)
		default:
			s = fmt.Sprintf("%3d", u64)
		}
	}
	return
}

func readIntegral(bytes []byte) uint64 {
	var acc uint64
	for i := range bytes {
		acc = acc<<8 + uint64(bytes[i])
	}
	return acc
}

func showEnum(bytes []byte) string {
	var acc uint64
	for i := range bytes {
		acc = acc<<8 + uint64(bytes[i])
	}
	return fmt.Sprintf("%d", acc)
}

func showNodeId(bytes []byte) string {
	if len(bytes) == 0 {
		return "invalid Node-id <nil>"
	} else {
		switch bytes[0] {
		case 0:
			return ReadIpV4(bytes[1:5]).String()
		case 1:
			return showBytes(bytes[1:])
		case 2:
			return showFQNString(bytes[1:])
		default:
			return showBytesAsError(bytes[1:], "invalid Node-id format ID")
		}
	}
}

func showFQNString(bytes []byte) string {
	if int(bytes[0]) == len(bytes)-1 {
		return string(bytes[1:])
	} else if int(bytes[0]) < len(bytes)-1 {
		s := string(bytes[1 : len(bytes)-1])
		return s + string(bytes[len(bytes)-1:])
	} else {
		return fmt.Sprintf("apn bad name %0X", bytes)
	}
}

func showAPNString(bytes []byte) string {
	return string(bytes[:])
}

func showBytes(bytes []byte) string {
	return fmt.Sprintf("%0X", bytes)
}

func showBytesAsError(bytes []byte, msg string) string {
	return fmt.Sprintf("invalid format(%s) - %0X", msg, bytes)
}

func showBitRates(bytes []byte) string {
	if len(bytes) == 10 {
		ulBitRate := readIntegral(bytes[0:5])
		dlBitRate := readIntegral(bytes[5:10])
		return fmt.Sprintf("Up:%d Dn:%d", ulBitRate, dlBitRate)
	}
	return showBytesAsError(bytes, "invalid IE payload length(!=10)")
}

func showUeIpAddress(bytes []byte) string {
	if len(bytes) == 5 && testBit2(bytes[0]) {
		direction := "src"
		if testBit3(bytes[0]) {
			direction = "dst"
		}
		return fmt.Sprintf("%s %s", direction, ReadIpV4(bytes[1:5]))
	} else {
		return showBytesAsError(bytes, "only IPv4 is decoded")
	}
}

func getUeIpAddress(bytes []byte) IpV4 {
	if len(bytes) == 5 && testBit2(bytes[0]) && testBit3(bytes[0]) {
		return ReadIpV4(bytes[1:5])
	} else {
		panic("unsupported UE IP form")
	}
}

func showUpIpResInfo(bytes []byte) string {
	return getUpIpResInfo(bytes).String()
}

// this is only a subset of the defined User Plane IP Resource Information formats,
// but since it is used for a limited test case it is ok for now.
func getUpIpResInfo(bytes []byte) IpV4 {
	if len(bytes) == 5 && bytes[0] == 1 {
		return ReadIpV4(bytes[1:5])
	} else {
		panic("unsupported UP IP Res Info format")
	}
}

func showOuterHeaderCreate(bytes []byte) string {
	// tdod - use the deserialiser after resolving error handling issues
	if len(bytes) == 10 && bytes[0] == 1 && bytes[1] == 0 {
		teid := readIntegral(bytes[2:6])
		ipv4 := ReadIpV4(bytes[6:10])
		return fmt.Sprintf("teid:%05x ipv4:%s", teid, ipv4)
	}
	return showBytesAsError(bytes, "only GTPu/UDP/IPV4 format decoded")
}

type OuterHeaderCreate struct {
	Teid TEID
	IpV4
}

func (OuterHeaderCreate OuterHeaderCreate) FTeid() FTeid {
	return FTeid{
		Teid: &OuterHeaderCreate.Teid,
		IpV4: &OuterHeaderCreate.IpV4,
	}
}

func (FTeid FTeid) OuterHeaderCreate() (OuterHeaderCreate, error) {
	if FTeid.IpV4 == nil {
		return OuterHeaderCreate{}, fmt.Errorf("missing IpV4 in FTeid -> OuterHeaderCreate")
	} else if FTeid.Teid == nil {
		return OuterHeaderCreate{}, fmt.Errorf("missing Teid in FTeid -> OuterHeaderCreate")
	} else {
		return OuterHeaderCreate{
			Teid: *FTeid.Teid,
			IpV4: *FTeid.IpV4,
		}, nil
	}
}

func getOuterHeaderCreate(bytes []byte) (t OuterHeaderCreate) {
	if len(bytes) == 10 && bytes[0] == 1 && bytes[1] == 0 {
		t.Teid = TEID(binary.BigEndian.Uint32(bytes[2:6]))
		t.IpV4 = ReadIpV4(bytes[6:10])
	} else {
		panic("invalid IE format")
	}
	return
}

// for now FTEID is not capable of expressing 'choose IPv6'
type FTeid struct {
	Teid *TEID
	IpV4 *IpV4
}

func (a FTeid) Eq(b FTeid) bool {
	teidEq := (a.Teid == nil && b.Teid == nil) || ((a.Teid != nil && b.Teid != nil) && *a.Teid == *b.Teid)

	ipv4Eq := (a.IpV4 == nil && b.IpV4 == nil) || ((a.IpV4 != nil && b.IpV4 != nil) && *a.IpV4 == *b.IpV4)

	return teidEq && ipv4Eq
}

func (a *FTeid) NextTeid() {
	if a.Teid == nil {
		panic("cannot apply next to nil TEID")
	} else {
		*a.Teid += 1
	}
}

func NewFTeid(Teid TEID, IpV4 IpV4) *FTeid {
	return &FTeid{
		Teid: &Teid,
		IpV4: &IpV4,
	}
}

func (FTeid FTeid) String() string {
	// TDOD probably needs changing since some options are not valid....
	switch {
	case FTeid.IpV4 == nil && FTeid.Teid == nil:
		return ("choose")
	case FTeid.Teid == nil:
		return FTeid.IpV4.String()
	case FTeid.IpV4 == nil:
		return fmt.Sprintf("%d", *FTeid.Teid)
	default:
		return fmt.Sprintf("%s:%d", FTeid.IpV4, *FTeid.Teid)
	}
}

func showFTeid(bytes []byte) string {
	if fteid, err := getFteid(bytes); err == nil {
		return fteid.String()
	} else {
		return showBytesAsError(bytes, err.Error())
	}
}

func getFteid(bytes []byte) (t *FTeid, err error) {
	if len(bytes) == 0 {
		err = fmt.Errorf("empty IE")
	} else if len(bytes) == 1 && bytes[0] == fteid_flag_V4|fteid_flag_CH {
		return &FTeid{Teid: nil, IpV4: nil}, nil // for now FTEID is not capable of expressing 'choose IPv6'
	} else if len(bytes) == 2 && bytes[0] == fteid_flag_CHID|fteid_flag_V4|fteid_flag_CH {
		return &FTeid{Teid: nil, IpV4: nil}, nil
	} else if len(bytes) == 9 && bytes[0] == fteid_flag_V4 {
		return NewFTeid(TEID(binary.BigEndian.Uint32(bytes[1:5])), ReadIpV4(bytes[5:9])), nil
	} else {
		err = fmt.Errorf("invalid IE format")
	}
	return
}

// ******************************************************

type FSeid struct {
	Seid SEID
	IpV4 *IpV4
}

func (FSeid FSeid) String() string {
	if FSeid.IpV4 == nil {
		return fmt.Sprintf("%010d", FSeid.Seid)
	} else {
		return fmt.Sprintf("%010d : %s", FSeid.Seid, FSeid.IpV4)
	}
}

func showFSeid(bytes []byte) string {
	if FSeid, err := getFSeid(bytes); err == nil {
		return FSeid.String()
	} else {
		return showBytesAsError(bytes, err.Error())
	}
}

func getFSeid(bytes []byte) (*FSeid, error) {
	if len(bytes) < 9 {
		return nil, fmt.Errorf("FSEID too short")
	} else {
		seid := SEID(binary.BigEndian.Uint64(bytes[1:9]))

		if bytes[0] == 0b00000000 {
			return &FSeid{
				Seid: seid,
				IpV4: nil,
			}, nil
		} else if len(bytes) == 13 && bytes[0] == 0b00000010 {
			ipv4 := ReadIpV4(bytes[9:13])
			return &FSeid{
				Seid: seid,
				IpV4: &ipv4,
			}, nil
		} else {
			return nil, fmt.Errorf("invalid IE format")
		}
	}
}

// ******************************************************
type EnumInterface uint8

const (
	EnumAccess         EnumInterface = 0
	EnumCore           EnumInterface = 1
	EnumSGi_LAN        EnumInterface = 2
	EnumCP_function    EnumInterface = 3
	Enum5G_VN_Internal EnumInterface = 4
)

func (enumInterface EnumInterface) String() string {
	return EnumInterfaceNames[enumInterface]
}

var EnumInterfaceNames = map[EnumInterface]string{EnumAccess: "Access", EnumCore: "Core", EnumSGi_LAN: "SGi-LAN", EnumCP_function: "CP-function", Enum5G_VN_Internal: "5G VN Internal"}

func readEnumInterface(bytes []byte) (EnumInterface, error) {
	if len(bytes) == 1 && bytes[0] <= uint8(Enum5G_VN_Internal) {
		return EnumInterface(bytes[0]), nil
	} else if len(bytes) == 0 {
		return EnumInterface(255), fmt.Errorf("invalid empty interface IE")
	} else {
		return EnumInterface(255), fmt.Errorf("invalid format in interface IE, %0x", bytes[0])
	}
}

func showEnumInterface(bytes []byte) string {
	if enumInterface, err := readEnumInterface(bytes); err == nil {
		return EnumInterfaceNames[enumInterface]
	} else {
		return err.Error()
	}
}

type EnumAction uint8

const (
	EnumDrop EnumAction = 1
	EnumForw EnumAction = 2
	EnumBuff EnumAction = 3
	EnumNocp EnumAction = 4
)

func (EnumAction EnumAction) String() string {
	return EnumActionNames[EnumAction]
}

var EnumActionNames = map[EnumAction]string{EnumDrop: "DROP", EnumForw: "FORW", EnumBuff: "BUFF", EnumNocp: "CP-function"}

// NB the actions appear to be exclusive,but in principle can be combined...!
func showApplyAction(bytes []byte) string {
	switch {
	case testBit1(bytes[0]):
		return "DROP"
	case testBit2(bytes[0]):
		return "FORW"
	case testBit3(bytes[0]):
		return "BUFF"
	case testBit4(bytes[0]):
		return "NOCP"
	}
	return "unknown"
}

// in 3gpp the lowest numbered bit is 'one'
func testBit1(b byte) bool {
	const mask = 1
	return b&mask == mask
}

func testBit2(b byte) bool {
	const mask = 1 << 1
	return b&mask == mask
}

func testBit3(b byte) bool {
	const mask = 1 << 2
	return b&mask == mask
}

func testBit4(b byte) bool {
	const mask = 1 << 3
	return b&mask == mask
}

type EnumGateStatus uint8

const (
	EnumOpen   EnumGateStatus = 0
	EnumClosed EnumGateStatus = 1
)

type GateStatus struct{ UlGate, DlGate EnumGateStatus }

func (GateStatus *GateStatus) String() string {
	return fmt.Sprintf("Gate: UL:%s DL:%s", GateStatus.UlGate, GateStatus.UlGate)
}

func (EnumGateStatus EnumGateStatus) String() string {
	return EnumGateStatusnNames[EnumGateStatus]
}

var GateOpen = GateStatus{UlGate: EnumOpen, DlGate: EnumOpen}

var EnumGateStatusnNames = map[EnumGateStatus]string{EnumOpen: "OPEN", EnumClosed: "CLOSED"}

type BitRate struct{ Uplink, Downlink uint64 }

func (BitRate *BitRate) String() string {
	return fmt.Sprintf("UL:%d DL:%d", &BitRate.Uplink, &BitRate.Downlink)
}
