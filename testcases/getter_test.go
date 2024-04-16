// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package testcases

import (
	"testing"

	"pfcpcore/pfcp"
	"pfcpcore/session"
	"pfcpcore/utils"
)

func GetterTest(t *testing.T) {
	if messages := utils.GetFileMessages(sersmr); len(messages) > 1 {
		t.Fatal("at least one message expected")
	} else if ser, err := pfcp.ParseValidate(messages[0]); err != nil {
		t.Fatalf("failed to parse message #1 (SER) - %s", err.Error())
	} else if ser.TypeCode() != pfcp.PFCP_Session_Establishment_Request {
		t.Fatalf("message #1 is not SER: %s (%d)\n", ser.TypeCode(), ser.TypeCode())
	} else {
		session.AccessSER(ser)
		sessionState := session.ParseUpfSER(ser.Node())
		t.Logf("success reading session state from SER:\n%s\n", sessionState)
	}
}
