// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package testcases

import (
	"testing"

	"pfcpcore/pfcp"
	"pfcpcore/transport"
	"pfcpcore/utils"
)

func testMergeCommon(t *testing.T, ser, smr *pfcp.PfcpMessage) {
	if ser.TypeCode() != pfcp.PFCP_Session_Establishment_Request {
		t.Fatalf("message #1 is not SER: %s (%d)\n", ser.TypeCode(), ser.TypeCode())
	} else if smr.TypeCode() != pfcp.PFCP_Session_Modification_Request {
		t.Fatalf("message #2 is not SMR: %s (%d)\n", smr.TypeCode(), smr.TypeCode())
	} else {
		t.Logf("success, now attempting to merge messages\n")
		if combinedMessage, err := ser.Merge(smr); err != nil {
			t.Fatalf("failed to merge messages: %s\n", err.Error())
		} else {
			newMessageBin := combinedMessage.Serialise()
			utils.WriteFileMessage("merged.bin", newMessageBin)
			t.Logf("all done (result written on 'merged.bin')\n")
		}
	}
}

func TestMergeStatic(t *testing.T) {
	if messages := utils.GetFileMessages(sersmr); len(messages) != 2 {
		t.Fatalf("exactly two messages expected\n")
	} else if ser, err := pfcp.ParseValidate(messages[0]); err != nil {
		t.Fatalf("failed to parse message #1 (SER) - %s\n", err.Error())
	} else if smr, err := pfcp.ParseValidate(messages[1]); err != nil {
		t.Fatalf("failed to parse message #2 (SMR) - %s\n", err.Error())
	} else {
		testMergeCommon(t, ser, smr)
	}
}

func TestMerge(t *testing.T) {
	if peer1, peer2, err := MakeTestPeers(); err != nil {
		t.Errorf(err.Error())
	} else {

		smf := func() {
			defer Close()

			wantRsp := func(tc pfcp.MessageTypeCode) {
				rr := <-peer1.ResponseChan
				if m, err := rr.Value(); err != nil {
					t.Errorf("request failed %s", err.Error())
				} else if m.MessageTypeCode != tc {
					t.Errorf("got bad response in smf: %s", m.MessageTypeCode)
				} else {
					t.Log("reply ok")
				}
			}

			peer1.EnterRequest(pfcp.SessionEstablishmentRequest)
			wantRsp(pfcp.PFCP_Session_Establishment_Response)
			peer1.EnterRequest(pfcp.SessionModificationRequest)
			wantRsp(pfcp.PFCP_Session_Modification_Response)
		}

		upf := func() {
			defer Close()

			wantReq := func(tc pfcp.MessageTypeCode) transport.PeerRequest {
				rq := <-peer2.RequestChan
				if rq.Message.MessageTypeCode != tc {
					t.Errorf("got bad response in smf: %s", rq.Message.MessageTypeCode)
				} else {
					t.Log("reply ok")
				}
				return rq
			}

			ser := wantReq(pfcp.PFCP_Session_Establishment_Request)
			peer2.EnterResponse(pfcp.SessionEstablishmentResponse, ser)

			smr := wantReq(pfcp.PFCP_Session_Modification_Request)
			peer2.EnterResponse(pfcp.SessionModificationResponse, smr)

			testMergeCommon(t, ser.Message, smr.Message)
		}

		go smf()
		go upf()

		CloseWait()
		CloseWait()
	}
}
