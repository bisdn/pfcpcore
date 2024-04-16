// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package pfcp

import "fmt"

type PfcpSequenceNumber uint32

func (sequenceNumber PfcpSequenceNumber) String() string {
	return fmt.Sprintf("SN:%06x", uint32(sequenceNumber))
}

func (message *PfcpMessage) SetPfcpSequenceNumber(sn PfcpSequenceNumber) {
	message.SequenceNumber = uint32(sn)
}

func (message *PfcpMessage) PfcpSequenceNumber() PfcpSequenceNumber {
	return PfcpSequenceNumber(message.SequenceNumber)
}

func (message *PfcpMessage) IsRequest() bool {
	return message.typeCode().IsRequest()
}

func (message *PfcpMessage) IsResponse() bool {
	return message.typeCode().IsResponse()
}
