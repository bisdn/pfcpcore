// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package session

import (
	"fmt"
	"math/rand"

	log "github.com/sirupsen/logrus"

	"pfcpcore/pfcp"
)

// SeidCopyMode forces the upf to reuse the session ID provided by the SMF.
// It is required for compatibility with PPCA
const SeidCopyMode = true

/*
Important note concerning SEID usage

In PFCP there exists for every UE session two SEID values, one assigned by each side.
Apart from Session Establishment Request, every session related message carries SEID in its message header,
and the SEID must be the one WHICH WAS PROVIDED BY THE INTENDED RECIPIENT OF A MESSAGE.
This means that, unless a UPF function simply reuses the SEID assigned by the SMF, then the UPF must track incoming message by its locally assigned SEID, but reply with the peer SEID.
Therefore, the session state must in every case save the peer SEID, even if it does not use it as its own key for session state.

This explains the usage in this source file of 'internalSessionStateElement', which is a tuple over the peer SEID and the wanted local session state.
*/

type SessionStateStoreKey = pfcp.SEID

type SessionStateElement interface{}

type internalSessionStateElement struct {
	peerSeid pfcp.SEID
	SessionStateElement
}

type SessionStateStore struct {
	sessions map[SessionStateStoreKey]internalSessionStateElement
}

func NewSessionStateStore() *SessionStateStore {
	return &SessionStateStore{
		sessions: map[SessionStateStoreKey]internalSessionStateElement{},
	}
}

func (SessionStateStore *SessionStateStore) Insert(peerSeid pfcp.SEID, SessionStateElement SessionStateElement) SessionStateStoreKey {
	upfSeid := SessionStateStore.nextSeid()
	if SeidCopyMode {
		// suppress our local seid
		upfSeid = peerSeid
		if _, present := SessionStateStore.sessions[upfSeid]; present {
			log.Errorf("peer SMF is reusing an old SEID")
		}
	}
	SessionStateStore.sessions[upfSeid] = internalSessionStateElement{
		peerSeid:            peerSeid,
		SessionStateElement: SessionStateElement,
	}
	return upfSeid
}

func (SessionStateStore *SessionStateStore) Modify(upfSeid pfcp.SEID, SessionStateElement SessionStateElement) error {
	if internalState, ok := SessionStateStore.sessions[SessionStateStoreKey(upfSeid)]; !ok {
		log.Warn("invalid SEID in session modification request")
		return fmt.Errorf("seid not found %s ", upfSeid)
	} else {
		SessionStateStore.sessions[SessionStateStoreKey(upfSeid)] = internalSessionStateElement{
			peerSeid:            internalState.peerSeid,
			SessionStateElement: SessionStateElement,
		}
		return nil
	}
}

func (SessionStateStore *SessionStateStore) Retrieve(upfSeid pfcp.SEID) (pfcp.SEID, SessionStateElement, error) {
	if internalState, ok := SessionStateStore.sessions[SessionStateStoreKey(upfSeid)]; !ok {
		log.Warn("invalid SEID in session request")
		return 0, nil, fmt.Errorf("seid not found %s ", upfSeid)
	} else {
		return internalState.peerSeid, internalState.SessionStateElement, nil
	}
}

func (SessionStateStore *SessionStateStore) Remove(upfSeid pfcp.SEID) (pfcp.SEID, error) {
	if internalState, ok := SessionStateStore.sessions[SessionStateStoreKey(upfSeid)]; !ok {
		log.Warnf("invalid SEID in session remove request (%s)", upfSeid)
		return 0, fmt.Errorf("seid not found %s ", upfSeid)
	} else {
		delete(SessionStateStore.sessions, SessionStateStoreKey(upfSeid)) // assumes that the retrieved key is safely copied before the delete is requested!
		return internalState.peerSeid, nil
	}
}

func (SessionStateStore *SessionStateStore) nextSeid() SessionStateStoreKey {
	// assign random SEID to distinguish local and peer SEID usage
	// in future, the local SEID could be usefully distinguished from peer assigned SEID
	// And for SMF, it is essential to have a local source of unique SEID
	r := SessionStateStoreKey(rand.Uint64())
	if _, ok := SessionStateStore.sessions[r]; ok {
		return SessionStateStore.nextSeid()
	} else {
		return r
	}
}
