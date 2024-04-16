// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package transport

/*
the reliable transport layer consumes an instance of the UDP service, UdpServerPeer
the reliable transport layer has two half systems - a requester and a responder.
They function separately, and a message routing function use a table of types to route incoming messages,
depending on whether they are requests or responses, to the correct half system.

the reliable transport layer requester provides a service interface which accepts valid request messages and an accompanying reply channel.
the reply channel carries one of error or message, where message is symmetric with the request

the multiplexer consumes the UdpServerPeer, and calls the appropriate state machine depending on the request/response nature

so, the reliable transport layer requester (RTL requester) provides an upstack method call as well as downstack.
In a null implementation, the upstack call simply delivers the message blindly to a client, as long as it can locate it.
How does it locate it? - by message sequence number


the reliable transport layer responder.
*/

import (
	"time"

	log "github.com/sirupsen/logrus"
	"pfcpcore/pfcp"
	"pfcpcore/udpserver"
)

// N1 = 10 is a useful value for testing, where the UPF may take rather long to start initially
// TODO - Perhaps it is too high for real use, and should be configurable
const (
	N1 = 100
	T1 = time.Second
)

type Transport struct {
	Requestor
	Responder
	*udpserver.UdpServerPeer
}

func NewTransport(udpServerPeer *udpserver.UdpServerPeer, requestChannel chan PeerRequest) *Transport {
	transport := &Transport{
		Requestor:     Requestor{nextSequenceNumber: getSeqStart(), inFlight: make(map[pfcp.PfcpSequenceNumber]*requestState)},
		Responder:     Responder{requestChannel: requestChannel, inFlight: make(map[pfcp.PfcpSequenceNumber]*peerRequestState)},
		UdpServerPeer: udpServerPeer,
	}
	go transport.runLower()
	return transport
}

func (r *Transport) Drop() {
	// on inspection it appears that any transient go routines in this instance will safely terminate
	//  But, it needs to closely checked once working
	r.UdpServerPeer.Drop()
	r.Requestor.Drop()
	r.Responder.Drop()
}

func (r *Transport) runLower() {
	for m := range r.UdpServerPeer.Receive {
		if pfcpMessage, err := pfcp.ParseValidate(m.Payload); err != nil {
			log.Warnf("error in PFCP message format %s\n", err.Error())
		} else if pfcpMessage.IsRequest() {
			r.Responder.handleRequest(pfcpMessage, r.UdpServerPeer.Send)
		} else if pfcpMessage.IsResponse() {
			r.Requestor.handleResponse(pfcpMessage)
		} else {
			// TDOD introduce some way to notify that an invalid message was rejected
			log.Warnf("error in PFCP type code %d\n", pfcpMessage.TypeCode())
		}
	}
}

func (r *Transport) BlockingRequest(message *pfcp.PfcpMessage) (*pfcp.PfcpMessage, error) {
	replyChannel := make(chan RequestReturn)
	r.EnterRequest(message, replyChannel)
	rval := <-replyChannel
	return rval.Value()
}

func (r *Transport) EnterRequest(message *pfcp.PfcpMessage, replyChannel chan RequestReturn) {
	r.Requestor.enterRequest(message, replyChannel, r.UdpServerPeer.Send)
}

func (r *Transport) EnterResponse(message *pfcp.PfcpMessage, response PeerRequest) {
	r.Responder.enterResponse(message, response, r.UdpServerPeer.Send)
}

type RequestReturn struct {
	err     error
	message *pfcp.PfcpMessage
}

func (requestReturn *RequestReturn) Value() (message *pfcp.PfcpMessage, err error) {
	return requestReturn.message, requestReturn.err
}
