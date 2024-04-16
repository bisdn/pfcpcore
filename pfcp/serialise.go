// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package pfcp

import (
	"bytes"
	"encoding/binary"
	"fmt"

	log "github.com/sirupsen/logrus"
)

func (msg *PfcpMessage) Serialise() []byte {
	payload := bytes.NewBuffer(nil)
	serialise(payload, msg.iEnodes)
	return msg.rawPfcpMessageHeader.serialise(payload.Bytes())
}

func ReserialiseCheck(msg *PfcpMessage, raw []byte) error {
	payload := bytes.NewBuffer(nil)
	serialise(payload, msg.iEnodes)
	wireformatCopy := msg.rawPfcpMessageHeader.serialise(payload.Bytes())
	return BytesCompare(wireformatCopy, raw)
}

func BytesCompare(a, b []byte) error {
	if bytes.Equal(a, b) {
		return nil
	} else {
		compare(a, b)
		return fmt.Errorf("recursive parse result IS NOT identical to original packet,\nrecursive parse result length %d, original packet length %d",
			len(a),
			len(b))
	}
}

func compare(a, b []byte) {
	mismatches := 0
	for i := 0; i < len(a) && i < len(b); i++ {
		switch {
		case mismatches == 0 && a[i] != b[i]:
			mismatches += 1
			log.Errorf("first mismatch at offset %d, ", i)
		case a[i] < b[i]:
			mismatches += 1
		}
	}
	if mismatches > 0 {
		log.Errorf("%d total mismatches\n", mismatches)
		log.Errorf("a: % x\n", a)
		log.Errorf("b: % x\n", b)
	}
}

func writeBE24(b []byte, u uint32) {
	b[0] = uint8(u >> 16)
	b[1] = uint8(u >> 8)
	b[2] = uint8(u)
}

func (header rawPfcpMessageHeader) serialise(payload []byte) (resultSlice []byte) {
	if header.SEID == nil {
		resultSlice = make([]byte, 8)
		binary.BigEndian.PutUint16(resultSlice[2:4], uint16(4+len(payload)))
		writeBE24(resultSlice[4:7], header.SequenceNumber)
	} else {
		resultSlice = make([]byte, 16)
		resultSlice[0] = resultSlice[0] | 0b00000001
		binary.BigEndian.PutUint16(resultSlice[2:4], uint16(12+len(payload)))

		if header.Priority != nil {
			resultSlice[0] = resultSlice[0] | 0b00000010
			resultSlice[15] = *header.Priority << 4

		}
		binary.BigEndian.PutUint64(resultSlice[4:12], uint64(*header.SEID))
		writeBE24(resultSlice[12:15], header.SequenceNumber)
	}
	resultSlice[0] = resultSlice[0] | 0b00100000
	// binary.BigEndian.PutUint16(resultSlice[2:4], uint16(len(payload)))
	resultSlice[1] = uint8(header.MessageTypeCode)
	resultSlice = append(resultSlice, payload...)
	return
}
