// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package pfcp

import (
	"encoding/binary"
	"fmt"
	"strings"
)

type (
	SEID uint64
	TEID uint32
)

func (SEID SEID) String() string {
	return fmt.Sprintf("%016x", uint64(SEID))
}

func (TEID TEID) String() string {
	return fmt.Sprintf("%08x", uint32(TEID))
}

type rawPfcpMessageHeader struct {
	MessageTypeCode
	*SEID
	SequenceNumber uint32 // actually only 24 bits
	Priority       *uint8 // actually only 4 bits
}

func (header *rawPfcpMessageHeader) typeCode() MessageTypeCode { return header.MessageTypeCode }

type rawPfcpMessage struct {
	rawPfcpMessageHeader
	bytes []byte
}

type PfcpMessage struct {
	rawPfcpMessageHeader
	iEnodes []IeNode
}

func NewSessionMessage(tc MessageTypeCode, seid SEID, iEnodes ...IeNode) *PfcpMessage {
	msg := &PfcpMessage{
		rawPfcpMessageHeader: rawPfcpMessageHeader{MessageTypeCode: tc, SEID: &seid},
		iEnodes:              iEnodes,
	}
	return msg
}

func NewNodeMessage(tc MessageTypeCode, iEnodes ...IeNode) *PfcpMessage {
	msg := &PfcpMessage{
		rawPfcpMessageHeader: rawPfcpMessageHeader{MessageTypeCode: tc, SEID: nil},
		iEnodes:              iEnodes,
	}
	return msg
}

func (msg *PfcpMessage) typeName() string          { return msg.rawPfcpMessageHeader.typeCode().String() }
func (msg *PfcpMessage) TypeCode() MessageTypeCode { return msg.rawPfcpMessageHeader.typeCode() }

func (msg *PfcpMessage) Node() *IeNode {
	return &IeNode{
		IeTypeCode: 0,
		bytes:      []byte{},
		IeID:       nil,
		ies:        msg.iEnodes,
	}
}

func (msg rawPfcpMessageHeader) String() string {
	var seid string

	if msg.SEID == nil {
		seid = "not present"
	} else {
		seid = fmt.Sprintf("%08d", *msg.SEID)
	}
	return fmt.Sprintf("PFCP Message type=%s,seid=%s, seq=%d", msg.MessageTypeCode, seid, msg.SequenceNumber)
}

func (msg rawPfcpMessage) ParsePFCPPayload() (*PfcpMessage, error) {
	if ieNodes, err := readIes(msg.bytes); err == nil {
		return &PfcpMessage{
			rawPfcpMessageHeader: rawPfcpMessageHeader{
				MessageTypeCode: msg.MessageTypeCode,
				SEID:            msg.SEID,
				SequenceNumber:  msg.SequenceNumber,
				Priority:        msg.Priority,
			},
			iEnodes: ieNodes,
		}, nil
	} else {
		return nil, err
	}
}

func (msg PfcpMessage) Dumper() string {
	sb := new(strings.Builder)
	msg.dump(sb, 0)
	return sb.String()
}

func (msg PfcpMessage) dump(sb *strings.Builder, level int) {
	// TODO: just implement dump for rawPfcpMessageHeader?
	fmt.Fprintf(sb, "%s\n", msg.rawPfcpMessageHeader.String())
	for _, thisIe := range msg.iEnodes {
		thisIe.dump(sb, level)
	}
}

func (msg *PfcpMessage) Validate() error {
	sb := new(strings.Builder)
	if attributeSet, present := MessageIeAttributeSets[msg.MessageTypeCode]; present {
		// first directly validate the top level list, which is not a group IE itself
		validateAttributeSet(sb, &msg.iEnodes, msg, attributeSet)
		// now call recursive validate on the elements
		for i := range msg.iEnodes {
			msg.iEnodes[i].validate(sb)
		}
	} else {
		fmt.Fprintf(sb, "parser exception - unknown message type %d in validator (%s)\n", msg.MessageTypeCode, msg.MessageTypeCode.String())
	}

	if sb.Len() == 0 {
		return nil
	} else {
		return fmt.Errorf(sb.String())
	}
}

func readBE24(b []byte) uint32 {
	return uint32(b[0])<<16 + uint32(b[1])<<8 + uint32(b[2])
}

func ParsePFCPHeader(b []byte) (rawPfcpMessage, error) {
	if len(b) < 8 {
		return rawPfcpMessage{}, fmt.Errorf("buffer length below minimum PFCP header")
	}
	flags := b[0]
	seidFlag := (flags & 0b00000001) == 0b00000001
	mpFlag := (flags & 0b00000010) == 0b00000010
	typeCode := b[1]
	messageLength := binary.BigEndian.Uint16(b[2:4])

	if int(messageLength)+4 != len(b) {
		return rawPfcpMessage{}, fmt.Errorf("buffer length mismatch with header length field %d %d", messageLength, len(b))
	}

	header := rawPfcpMessageHeader{
		MessageTypeCode: MessageTypeCode(typeCode),
	}

	if !seidFlag {
		header.SequenceNumber = readBE24(b[4:7])
		return rawPfcpMessage{
			rawPfcpMessageHeader: header,
			bytes:                b[8:],
		}, nil
	} else if len(b) < 16 {
		return rawPfcpMessage{}, fmt.Errorf("buffer length below minimum PFCP header with SEID")
	} else {
		seid := SEID(binary.BigEndian.Uint64(b[4:12]))
		header.SEID = &seid
		header.SequenceNumber = readBE24(b[12:15])
		if mpFlag {
			header.Priority = new(uint8)
			*header.Priority = b[15] >> 4
		}
		return rawPfcpMessage{
			rawPfcpMessageHeader: header,
			bytes:                b[16:],
		}, nil
	}
}
