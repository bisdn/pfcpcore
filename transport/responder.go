// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package transport

/*

The reliable transport layer responder.
It has questions and demands answers.
It has many questions, all at the same time.
So, it sends the questions with a token, and expects answers with the same tokens as questions.
The underlying message type is the same in both question and answer.

NB sync.Lock() strategy is very sensitive for handleReesponse / enterRequest;
  variations can cause runtime panics (send on closed channel), or deadlocks.
*/

import (
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
	"pfcpcore/pfcp"
	"pfcpcore/udpserver"
)

type PeerRequest struct {
	Message        *pfcp.PfcpMessage
	sequenceNumber pfcp.PfcpSequenceNumber
	port           uint16
}

func (response PeerRequest) String() string {
	return fmt.Sprintf("Response: port:%d %s", response.port, response.Message)
}

type peerRequestState struct {
	reply *udpserver.UdpMessage
}

type Responder struct {
	requestChannel chan PeerRequest
	inFlight       map[pfcp.PfcpSequenceNumber]*peerRequestState
	mutex          sync.Mutex
}

func (responder *Responder) Drop() {
	responder.mutex.Lock()
	responder.inFlight = nil
	responder.mutex.Unlock()
}

// *** TODO ! *** set an expiry timer to eventually discard the state

func (r *Responder) enterResponse(reply *pfcp.PfcpMessage, response PeerRequest, udpSendChannel chan *udpserver.UdpMessage) {
	r.mutex.Lock()
	state, found := r.inFlight[response.sequenceNumber]

	if found {
		if state.reply != nil {
			r.mutex.Unlock()
			log.Errorf("Responder.enterResponse - error, reply already sent! %s\n", response.sequenceNumber)
		} else if udpSendChannel == nil {
			r.mutex.Unlock()
			log.Debug("Responder.enterResponse - error, udpserver channel closed\n")
		} else {
			reply.SetPfcpSequenceNumber(response.sequenceNumber)
			udpReply := udpserver.ToUdpMessage(reply.Serialise())
			state.reply = udpReply
			r.inFlight[response.sequenceNumber] = state
			r.mutex.Unlock()
			udpSendChannel <- udpReply
		}
	} else {
		r.mutex.Unlock()
		log.Errorf("Responder.enterResponse - error, reply for unknown request! %s\n", response.sequenceNumber)
	}
}

// func (r Requestor) handleRequest(message *pfcp.Message, replyChannel chan RequestReturn, udpSendChannel chan *UdpMessage)
// Note,this is an incoming request from the peer, not the local client.
func (r *Responder) handleRequest(requestMessage *pfcp.PfcpMessage, udpSendChannel chan *udpserver.UdpMessage) {
	if r.inFlight == nil {
		log.Debug("responder - drop inbound request for closed endpoint")
	} else {

		sequenceNumber := requestMessage.PfcpSequenceNumber()

		r.mutex.Lock()
		state, found := r.inFlight[sequenceNumber]

		if found {
			r.mutex.Unlock()
			if state.reply == nil {
				log.Errorf("Responder.handleRequest - error, retranmission request with no pending reply %s\n", sequenceNumber)
			} else {
				log.Infof("Responder.handleRequest - warning, retranmission requested %s\n", sequenceNumber)
				udpSendChannel <- state.reply
			}

		} else if r.requestChannel == nil {
			r.mutex.Unlock()
			log.Debug("responder - drop inbound request for closed endpoint")
		} else {
			r.inFlight[sequenceNumber] = &peerRequestState{}
			r.mutex.Unlock()
			requestMessage.SetPfcpSequenceNumber(0) // the client must not know anything of sequence numbers!
			r.requestChannel <- PeerRequest{Message: requestMessage, sequenceNumber: sequenceNumber, port: 0}
		}
	}
}
