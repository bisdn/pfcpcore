// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package endpoint

import (
	"net/netip"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"pfcpcore/pfcp"
	"pfcpcore/session"
)

type IeNode = pfcp.IeNode

var CauseUnknown = []IeNode{pfcp.IE_Cause(pfcp.CauseUnspecified)}

type PfcpAssociationConfig struct {
	PeerName               string
	NodeName               string
	LocalGtpAddress        *netip.Addr
	LocalSignallingAddress netip.Addr
	Application            session.Application
	PeerEndpoint           *PfcpPeer
}

type PfcpAssociationState struct {
	PfcpAssociationConfig
	peerRecoveryTime, recoveryTime    uint32
	requestStats                      map[pfcp.MessageTypeCode]uint32
	peerStartTime, lastRequestMessage time.Time
	*session.SessionStateStore
	exit sync.Mutex
}

func (pfcpAssociationState *PfcpAssociationState) recoveryTimeIe() IeNode {
	return pfcp.IE_RecoveryTimeStamp(pfcpAssociationState.recoveryTime)
}

func (pfcpAssociationState *PfcpAssociationState) nodeIdIe() IeNode {
	return pfcp.IE_NodeIdFqdn(pfcpAssociationState.NodeName)
}

func (state *PfcpAssociationState) serviceAssociationSetupRequest(ser *IeNode) []IeNode {
	root := ser.Getter()
	if nodeId, err := root.GetByTc(pfcp.Node_ID).DeserialiseNodeIdString(); err != nil {
		return CauseUnknown
	} else if recoveryTimestamp, err := root.GetByTc(pfcp.Recovery_Time_Stamp).DeserialiseU32(); err != nil {
		return CauseUnknown
	} else {
		state.peerRecoveryTime = recoveryTimestamp
		state.PeerName = nodeId
		state.peerStartTime = time.Now()
		if state.LocalGtpAddress == nil {
			return []IeNode{state.nodeIdIe(), pfcp.IE_Cause(pfcp.CauseAccepted), state.recoveryTimeIe()}
		} else {
			return []IeNode{state.nodeIdIe(), pfcp.IE_Cause(pfcp.CauseAccepted), state.recoveryTimeIe(), pfcp.IE_UpIpRsrcInfo(*state.LocalGtpAddress)}
		}
	}
}

func (*PfcpAssociationState) serviceAssociationReleaseRequest(*IeNode) []IeNode {
	return nil
}

func (state *PfcpAssociationState) serviceHeartbeatRequest(hbReq *IeNode) []IeNode {
	root := hbReq.Getter()

	if recoveryTimestamp, err := root.GetByTc(pfcp.Recovery_Time_Stamp).DeserialiseU32(); err != nil {
		return CauseUnknown
	} else {
		if state.peerRecoveryTime != recoveryTimestamp {
			log.Warn("peer recovery timestamp has changed")
		}
		return []IeNode{pfcp.IE_RecoveryTimeStamp(state.recoveryTime)}
	}
}

func (state *PfcpAssociationState) serviceSessionEstablishmentRequest(ser *IeNode) (pfcp.SEID, []IeNode) {
	if smfFSeid, err := session.ParseSERSeid(ser); err != nil {
		// should not happen since the prior validation guarantees that the request is valid
		log.Errorf("ParseSERSeid() failed %s", err.Error())
		return 0, []IeNode{pfcp.IE_Cause(pfcp.CauseUnspecified)}
	} else {
		// insert the request before calling FP, in order to acquire the local Seid which is used for other requests to FP
		upfSeid := state.SessionStateStore.Insert(pfcp.SEID(smfFSeid.Seid), ser)
		if cause, err := state.Application.CallbackSessionEstablishmentRequest(upfSeid, ser); err != nil {
			// note, a reject IE from callbackSessionEstablishmentRequest() is used, but if the requests succeeds we must build the reply here
			log.Errorf("callbackSessionEstablishmentRequest() failed")
			state.SessionStateStore.Remove(upfSeid)
			return pfcp.SEID(smfFSeid.Seid), []IeNode{pfcp.IE_Cause(cause)}
		} else {
			// NB- the foreign SEID must be used here, it is needed in future for response to requests for this session.  See also session/statestore.go

			return pfcp.SEID(smfFSeid.Seid), []IeNode{state.nodeIdIe(), pfcp.IE_Cause(pfcp.CauseAccepted), pfcp.IE_FSeid(uint64(upfSeid), state.LocalSignallingAddress)}
		}
	}
}

func (state *PfcpAssociationState) serviceSessionModificationRequest(smr *IeNode, upfSeid pfcp.SEID) (pfcp.SEID, []IeNode) {
	if peerSeid, ser, err := state.SessionStateStore.Retrieve(upfSeid); err != nil {
		return 0, []IeNode{pfcp.IE_Cause(pfcp.SessionContextNotFound)}
	} else {
		sessionModificationAttributeSet := pfcp.MessageIeAttributeSets[pfcp.PFCP_Session_Modification_Request]

		// should we be protecting these type casts with error checks?
		// in theory they can never fail.....
		target := ser.(*IeNode)
		targetIes := target.Ies()
		smrIes := smr.Ies()
		if err := pfcp.MergeIes(targetIes, smrIes, sessionModificationAttributeSet); err != nil {
			log.Errorf(("problem merging session state, %s"), err.Error())
			state.SessionStateStore.Remove(upfSeid)
			return peerSeid, []IeNode{pfcp.IE_Cause(pfcp.CauseUnspecified)}
		} else if cause, err := state.Application.CallbackSessionEstablishmentRequest(upfSeid, target); err != nil {
			// note, a reject IE from callbackSessionEstablishmentRequest() is used, but if the requests succeeds we must build the reply here
			state.SessionStateStore.Remove(upfSeid)
			return peerSeid, []IeNode{pfcp.IE_Cause(cause)}
		} else {
			state.SessionStateStore.Modify(upfSeid, target)

			return peerSeid, []IeNode{pfcp.IE_Cause(pfcp.CauseAccepted)}
		}
	}
}

// func (state *PfcpAssociationState) serviceSessionEstablishmentRequest(ser *IeNode) (pfcp.SEID, []IeNode) {
// 	if seid, err := session.ParseSER(ser); err != nil {
// 		log.Infof("session reject fail to parse SEID %s", err.Error())
// 		return 0, CauseUnknown
// 	} else {
// 		localSeid := state.SessionStateStore.Insert(pfcp.SEID(seid), ser)
// 		log.Infof("session accept local SEID: %s peer SEID: %s", pfcp.SEID(localSeid), pfcp.SEID(seid))
// 		return pfcp.SEID(localSeid), []IeNode{state.nodeIdIe(), pfcp.IE_Cause(pfcp.CauseAccepted), pfcp.IE_FSeid(uint64(localSeid), state.localGtpAddress)}
// 	}
// }

// func (state *PfcpAssociationState) serviceSessionModificationRequest(_ *IeNode, seid pfcp.SEID) (pfcp.SEID, []IeNode) {
// 	// wrong action here...
// 	if peerSeid, err := state.SessionStateStore.Remove(seid); err != nil {
// 		return 0, CauseUnknown // should be "Session context not found"
// 	} else {
// 		return peerSeid, []IeNode{pfcp.IE_Cause(pfcp.CauseAccepted)}
// 	}
// }

func (state *PfcpAssociationState) serviceSessionDeletionRequest(_ *IeNode, seid pfcp.SEID) (pfcp.SEID, []IeNode) {
	if peerSeid, err := state.SessionStateStore.Remove(seid); err != nil {
		return 0, CauseUnknown // should be "Session context not found"
	} else if cause, err := state.Application.CallbackSessionDeletionRequest(seid); err != nil {
		return peerSeid, []IeNode{pfcp.IE_Cause(cause)}
	} else {
		return peerSeid, []IeNode{pfcp.IE_Cause(pfcp.CauseAccepted)}
	}
}

func (state *PfcpAssociationState) Drop() {
	// probably this action should be done automatically when the association exits
	state.PeerEndpoint.Drop()
}

func (state *PfcpAssociationState) Wait() {
	state.exit.Lock()
}

func NewPfcpAssociationState(config PfcpAssociationConfig) *PfcpAssociationState {
	state := &PfcpAssociationState{
		PfcpAssociationConfig: config,
		recoveryTime:          pfcp.GetRecoveryTime(),
		requestStats:          map[pfcp.MessageTypeCode]uint32{},
		SessionStateStore:     session.NewSessionStateStore(),
	}

	updateStats := func(tc pfcp.MessageTypeCode) {
		n := state.requestStats[tc]
		state.requestStats[tc] = n + 1
		state.lastRequestMessage = time.Now()
	}

	runner := func() {
		for m := range config.PeerEndpoint.RequestChan {
			switch m.Message.MessageTypeCode {

			case pfcp.PFCP_Association_Setup_Request:
				response := state.serviceAssociationSetupRequest(m.Message.Node())
				reply := pfcp.NewNodeMessage(pfcp.PFCP_Association_Setup_Response, response...)
				config.PeerEndpoint.EnterResponse(reply, m)

			case pfcp.PFCP_Association_Release_Request:
				response := state.serviceAssociationReleaseRequest(m.Message.Node())
				reply := pfcp.NewNodeMessage(pfcp.PFCP_Association_Release_Response, response...)
				config.PeerEndpoint.EnterResponse(reply, m)

			case pfcp.PFCP_Heartbeat_Request:
				response := state.serviceHeartbeatRequest(m.Message.Node())
				reply := pfcp.NewNodeMessage(pfcp.PFCP_Heartbeat_Response, response...)
				config.PeerEndpoint.EnterResponse(reply, m)

			case pfcp.PFCP_Session_Establishment_Request:
				seid, response := state.serviceSessionEstablishmentRequest(m.Message.Node())
				reply := pfcp.NewSessionMessage(pfcp.PFCP_Session_Establishment_Response, seid, response...)
				config.PeerEndpoint.EnterResponse(reply, m)

			case pfcp.PFCP_Session_Modification_Request:
				seid, response := state.serviceSessionModificationRequest(m.Message.Node(), *m.Message.SEID)
				reply := pfcp.NewSessionMessage(pfcp.PFCP_Session_Modification_Response, seid, response...)
				config.PeerEndpoint.EnterResponse(reply, m)

			case pfcp.PFCP_Session_Deletion_Request:
				log.Tracef("PFCP_Session_Deletion_Request: seid: %s", *m.Message.SEID)
				seid, response := state.serviceSessionDeletionRequest(m.Message.Node(), *m.Message.SEID)
				reply := pfcp.NewSessionMessage(pfcp.PFCP_Session_Deletion_Response, seid, response...)
				config.PeerEndpoint.EnterResponse(reply, m)
			}
			updateStats(m.Message.MessageTypeCode)
		}
		state.exit.Unlock()
	}
	go runner()
	state.exit.Lock()
	return state
}
