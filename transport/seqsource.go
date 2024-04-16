// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package transport

import (
	"crypto/rand"
	"encoding/binary"

	"pfcpcore/pfcp"
)

func getSeqStart() pfcp.PfcpSequenceNumber {
	b := make([]byte, 4)
	rand.Read(b[1:])
	u32 := binary.BigEndian.Uint32(b)
	return pfcp.PfcpSequenceNumber(u32)
}
