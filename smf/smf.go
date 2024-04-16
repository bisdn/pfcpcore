// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package smf

import (
	"fmt"
	"net/netip"

	log "github.com/sirupsen/logrus"

	"pfcpcore/endpoint"
	"pfcpcore/pfcp"
)

type Association struct {
	*endpoint.PfcpEndpoint
	*endpoint.PfcpPeer
	nextSeid uint64
	nodeId   netip.Addr
}

func (association *Association) baseSER(ies ...pfcp.IeNode) (pfcp.SEID, *pfcp.PfcpMessage) {
	association.nextSeid += 1
	baseIes := []pfcp.IeNode{
		pfcp.IE_NodeIdIpV4(association.nodeId),
		pfcp.IE_FSeid(association.nextSeid, association.nodeId),
	}
	return pfcp.SEID(association.nextSeid), pfcp.NewSessionMessage(
		pfcp.PFCP_Session_Establishment_Request,
		0,
		append(baseIes, ies...)...,
	)
}

type Session struct {
	local, peer pfcp.SEID
	*Association
}

func (Session *Session) Modify(ies ...pfcp.IeNode) error {
	_, err := doRequest(
		Session.PfcpPeer,
		pfcp.NewSessionMessage(
			pfcp.PFCP_Session_Modification_Request,
			pfcp.SEID(Session.peer),
			ies...,
		),
	)
	return err
}

func (Session *Session) Delete() error {
	_, err := doRequest(
		Session.PfcpPeer,
		pfcp.NewSessionMessage(
			pfcp.PFCP_Session_Deletion_Request,
			pfcp.SEID(Session.peer),
		),
	)
	return err
}

func (association *Association) CreateSession(ies ...pfcp.IeNode) (*Session, error) {
	localSeid, ser := association.baseSER(ies...)
	if peerSeid, err := doSessionEstablishmentRequest(association.PfcpPeer, ser); err != nil {
		return nil, err
	} else {
		return &Session{
			local:       localSeid,
			peer:        peerSeid,
			Association: association,
		}, nil
	}
}

// (association *Association) Clone() allows an endpoint to reuse the local endpoint for a second peer
func (association *Association) Clone(peerAddr netip.AddrPort) (*Association, error) {
	nodeIp := association.nodeId
	upfPeer := association.PfcpEndpoint.Peer(peerAddr)
	recoveryTime := pfcp.GetRecoveryTime()

	if _, err := doRequest(upfPeer, associationRequest(nodeIp, recoveryTime)); err != nil {
		return nil, err
	} else {
		go heartbeatResponder(upfPeer, pfcp.GetRecoveryTime())
		return &Association{
			PfcpPeer: upfPeer,
			nextSeid: 42,
			nodeId:   nodeIp,
		}, nil
	}
}

func CreateAssociation(localAddr, peerAddr netip.AddrPort) (*Association, error) {
	if local, err := endpoint.NewPfcpEndpoint(localAddr); err != nil {
		return nil, fmt.Errorf("failed to create endpoint for  %s (%s)", localAddr, err)
	} else {
		nodeIp := localAddr.Addr()
		upfPeer := local.Peer(peerAddr)
		recoveryTime := pfcp.GetRecoveryTime()

		log.Printf("using local:%s peer:%s for peer upf\n", localAddr, peerAddr)

		if _, err := doRequest(upfPeer, associationRequest(nodeIp, recoveryTime)); err != nil {
			return nil, err
		} else {

			go heartbeatResponder(upfPeer, pfcp.GetRecoveryTime())
			return &Association{
				PfcpEndpoint: local,
				PfcpPeer:     upfPeer,
				nextSeid:     42,
				nodeId:       nodeIp,
			}, nil
		}
	}
}

func heartbeatResponder(peer *endpoint.PfcpPeer, recoveryTime uint32) {
	log.Trace(("start heartbeatResponder"))

	for m := range peer.RequestChan {
		switch m.Message.MessageTypeCode {

		case pfcp.PFCP_Heartbeat_Request:
			reply := pfcp.NewNodeMessage(pfcp.PFCP_Heartbeat_Response, pfcp.IE_RecoveryTimeStamp(recoveryTime))
			peer.EnterResponse(reply, m)
			log.Trace("process heartbeat request")

		default:
			log.Errorf("got unexpected PFCP request (%s)", m.Message.MessageTypeCode)
		}
	}

	log.Trace("exit heartbeatResponder")
}

// func (sessionRequest *SessionDeleteRequest) sessionDelete() *pfcp.PfcpMessage {
// 	return pfcp.NewSessionMessage(
// 		pfcp.PFCP_Session_Deletion_Request,
// 		pfcp.SEID(sessionRequest.SEID),
// 	)
// }

// func (sessionRequest *SessionCreateRequest) completeSER(nodeId netip.Addr) *pfcp.PfcpMessage {

// 	return pfcp.NewSessionMessage(pfcp.PFCP_Session_Establishment_Request,
// 		0,
// 		pfcp.IE_NodeIdIpV4(nodeId),
// 		pfcp.IE_FSeid(uint64(sessionRequest.SEID), nodeId),
// 		b.DecapPdr(b.ModeCreate, sessionRequest.Request.accessPdr),
// 		b.DecapFar(b.ModeCreate),
// 		b.EncapPdr(b.ModeCreate, sessionRequest.Request.corePdr),
// 		b.EncapFar(b.ModeCreate, sessionRequest.Request.coreFar),
// 	)
// }

func associationRequest(nodeId netip.Addr, recoveryTime uint32) *pfcp.PfcpMessage {
	return pfcp.NewNodeMessage(pfcp.PFCP_Association_Setup_Request, pfcp.IE_NodeIdIpV4(nodeId), pfcp.IE_RecoveryTimeStamp(recoveryTime))
}

func doRequest(peer *endpoint.PfcpPeer, request *pfcp.PfcpMessage) (*pfcp.PfcpMessage, error) {
	if response, err := peer.BlockingRequest(request); err != nil {
		return nil, err
	} else if cause, err := response.Node().ReadCauseCode(); err != nil {
		return nil, fmt.Errorf(("%s request failed with missing cause"), request.MessageTypeCode)
	} else if cause != pfcp.CauseAccepted {
		return nil, fmt.Errorf(("%s request failed with cause %d"), request.MessageTypeCode, cause)
	} else {
		log.Tracef(("%s request success"), request.MessageTypeCode)
		return response, nil
	}
}

func doSessionEstablishmentRequest(peer *endpoint.PfcpPeer, request *pfcp.PfcpMessage) (pfcp.SEID, error) {
	if response, err := doRequest(peer, request); err != nil {
		return 0, err
	} else if fseid, err := response.Node().Getter().GetByTc(pfcp.F_SEID).DeserialiseFSeid(); err != nil {
		return 0, fmt.Errorf("sessionRequest request failed - missing SEID in reply")
	} else {
		seid := pfcp.SEID(fseid.Seid)
		log.Debugf(("sessionRequest success, seid: %s"), seid)
		return seid, nil
	}
}
