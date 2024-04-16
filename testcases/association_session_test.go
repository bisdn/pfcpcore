// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package testcases

import (
	"testing"

	"pfcpcore/pfcp"
)

func TestEndpoint(t *testing.T) {
	if peer1, peer2, err := MakeTestPeers(); err != nil {
		t.Errorf(err.Error())
	} else {

		active := func() {
			defer Close()

			for _, request := range pfcp.TestSet1Requests {
				peer1.EnterRequest(request)
				m := <-peer1.ResponseChan
				if m, err := m.Value(); err != nil {
					t.Errorf("request failed %s", err.Error())
				} else {
					t.Logf("got response in active: %s", m.MessageTypeCode)
				}
			}
		}

		passive := func() {
			defer Close()

			for _, response := range pfcp.TestSet1Responses {
				m := <-peer2.RequestChan
				t.Logf("got request in passive: %s", m)
				peer2.EnterResponse(response, m)
			}
		}
		go active()
		go passive()
		CloseWait()
		CloseWait()
	}
}
