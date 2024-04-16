// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package pfcp

import (
	"fmt"
)

// TODO - return error here, remove the print statements
func ExcerciseParser(rawMsg []byte) {
	rawCopy := make([]byte, len(rawMsg))
	copy(rawCopy, rawMsg)
	if parsedMsg, err := ParseValidate(rawCopy); err != nil {
		fmt.Println("message parse failed:", err.Error())
	} else if err := ReserialiseCheck(parsedMsg, rawMsg); err != nil {
		fmt.Printf("check reserialisation failed: %s\n", err.Error())
		// panic(fmt.Sprintf("check reserialisation failed: %s", err.Error()))
	} else {
		fmt.Println("validation and reserialisation succeeded")
		fmt.Printf("message dump:\n%s\n", parsedMsg.Dumper())
	}
}

func ParseValidate(rawMsg []byte) (*PfcpMessage, error) {
	if msg, err := ParsePFCPHeader(rawMsg); err != nil {
		return nil, err
	} else if parsedMsg, err := msg.ParsePFCPPayload(); err != nil {
		return nil, err
	} else if err := parsedMsg.Validate(); err != nil {
		return nil, err
	} else {
		return parsedMsg, nil
	}
}
