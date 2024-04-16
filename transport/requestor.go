// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package transport

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"pfcpcore/pfcp"
	"pfcpcore/udpserver"
)

type requestState struct {
	replyChannel chan RequestReturn
}

type Requestor struct {
	nextSequenceNumber pfcp.PfcpSequenceNumber
	inFlight           map[pfcp.PfcpSequenceNumber]*requestState
	mutex              sync.Mutex
}

func (requestor *Requestor) Drop() {
	requestor.mutex.Lock()
	requestor.inFlight = nil
	requestor.mutex.Unlock()
}

// Note - the retry work/state is held in a go routine, so there is very little explicit state
// More state could be held in the map, in order to make it visible....

// How does retry work?
// The message state held by sequence number holds the reply channel.
// When a reply has been received the receive side sets the channel pointer to nil,
// this allows the send side to exit gracefully only when the work is done.

func (r *Requestor) enterRequest(message *pfcp.PfcpMessage, replyChannel chan RequestReturn, udpSendChannel chan *udpserver.UdpMessage) {
	if replyChannel == nil {
		panic("pfcpcore: nil reply channel is fatal")
	}

	r.mutex.Lock()
	r.inFlight[r.nextSequenceNumber] = &requestState{
		replyChannel: replyChannel,
	}
	sequenceNumber := r.nextSequenceNumber
	r.nextSequenceNumber++
	r.mutex.Unlock()
	message.SetPfcpSequenceNumber(sequenceNumber)

	go func() {
		timedOut := true
		for n := N1; n > 0; n -= 1 {
			if n < N1 {
				log.Debug("pfcpcore: resending request")
			}
			udpSendChannel <- &udpserver.UdpMessage{Payload: message.Serialise()}
			time.Sleep(T1)
			r.mutex.Lock()
			state, found := r.inFlight[sequenceNumber]
			r.mutex.Unlock()
			if !found {
				log.Errorf("pfcpcore: impossible condition in enterRequest() poll loop")
			} else {
				if state.replyChannel == nil {
					timedOut = false
					n = 0
				} else {
					log.Trace("pfcpcore: no response yet seen")
				}
			}
		}
		if timedOut {
			log.Tracef("pfcpcore: request failed with timeout %s\n", sequenceNumber)
			replyChannel <- RequestReturn{err: fmt.Errorf("pfcpcore: request failed with timeout %s", sequenceNumber)}
		}
		r.mutex.Lock()
		delete(r.inFlight, sequenceNumber)
		r.mutex.Unlock()
	}()
}

func (r *Requestor) handleResponse(pfcpMessage *pfcp.PfcpMessage) {
	sequenceNumber := pfcpMessage.PfcpSequenceNumber()
	r.mutex.Lock()
	request, found := r.inFlight[sequenceNumber]
	r.mutex.Unlock()

	if !found {
		log.Warnf("pfcpcore: unknown SEID in response from peer - seid: %s\n", sequenceNumber)
	} else if request.replyChannel == nil {
		log.Warnf("pfcpcore: unexpected repeat response from peer - seid: %s\n", sequenceNumber)
	} else {
		pfcpMessage.SetPfcpSequenceNumber(0) // the client must not know anything of sequence numbers!
		request.replyChannel <- RequestReturn{message: pfcpMessage}
		request.replyChannel = nil
		r.mutex.Lock()
		r.inFlight[sequenceNumber] = request
		r.mutex.Unlock()
	}
}
