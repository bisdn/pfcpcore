// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package main

import (
	"fmt"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"pfcpcore/pfcp"
)

type CommonStateKey struct {
	// these are the PGW - SGW tunnel descriptors
	uplink, downlink spFteid
}

type SgwSessionState struct {
	// these are the eNB - SGW tunnel descriptors
	// the PDR/FAR are from the SGW perspective
	accessPdr, coreFar spFteid
}

type PgwSessionState struct {
	// PGW only defines a core ingress rule
	corePdr pfcp.IpV4
}

type XgwSessionState interface {
	isXgwSessionState()
}

func (SgwSessionState) isXgwSessionState() {}
func (PgwSessionState) isXgwSessionState() {}

func (SgwSessionState SgwSessionState) String() string {
	return fmt.Sprintf("sgw: accessPdr: %s  coreFar: %s ",
		SgwSessionState.accessPdr,
		SgwSessionState.coreFar,
	)
}

func (PgwSessionState PgwSessionState) String() string {
	return fmt.Sprintf("pgw: corePdr: %s",
		PgwSessionState.corePdr,
	)
}

func (CommonStateKey CommonStateKey) String() string {
	return fmt.Sprintf("css: uplink: %s downlink: %s",
		CommonStateKey.uplink,
		CommonStateKey.downlink,
	)
}

type CommonStateValue struct {
	*PgwSessionState
	*SgwSessionState
	pgwSeid, sgwSeid pfcp.SEID
	SessionObject
}

func (csv *CommonStateValue) String() string {
	return fmt.Sprintf("sgw: %s:%s, pgw: %s:%s", csv.sgwSeid, csv.SgwSessionState, csv.pgwSeid, csv.PgwSessionState)
}

// (*CommonStateValue) Differs()
// semi-deep introspection.  Session object is opaque, but would expect the pointer value to be maintained.
// For the other components and field change is significant.

func (a *CommonStateValue) Differs(b *CommonStateValue) string {
	if a.SessionObject != b.SessionObject {
		return "(*CommonStateValue) Differs() - a.SessionObject != b.SessionObject"
	}

	if (a.PgwSessionState == nil) != (b.PgwSessionState == nil) {
		return ("(*CommonStateValue) Differs() - a.PgwSessionState==nil != b.PgwSessionState==nil")
	}

	if (a.SgwSessionState == nil) != (b.SgwSessionState == nil) {
		return ("(*CommonStateValue) Differs() - a.SgwSessionState==nil != b.SgwSessionState==nil")
	}
	if a.PgwSessionState != nil {
		if *a.PgwSessionState != *b.PgwSessionState {
			return ("(*CommonStateValue) Differs() - *a.PgwSessionState != *b.PgwSessionState")
		}

		if *a.SgwSessionState != *b.SgwSessionState {
			return ("(*CommonStateValue) Differs() - *a.SgwSessionState != *b.SgwSessionState")
		}

		if a.pgwSeid != b.pgwSeid {
			return ("(*CommonStateValue) Differs() - a.pgwSeid != b.pgwSeid")
		}

		if a.sgwSeid != b.sgwSeid {
			return ("(*CommonStateValue) Differs() - a.sgwSeid != b.sgwSeid")
		}
	}

	if a.SessionObject != b.SessionObject {
		return ("(*CommonStateValue) Differs() - a.SessionObject != b.SessionObject")
	}

	return ""
}

type CommonStateSeidMap map[pfcp.SEID]CommonStateKey

type CommonState struct {
	l sync.Mutex
	m map[CommonStateKey]CommonStateValue
	SessionService
	sgwSeidMap, pgwSeidMap CommonStateSeidMap
}

func newCommonStateMap(SessionService SessionService) *CommonState {
	return &CommonState{
		l:              sync.Mutex{},
		m:              map[CommonStateKey]CommonStateValue{},
		SessionService: SessionService,
		sgwSeidMap:     map[pfcp.SEID]CommonStateKey{},
		pgwSeidMap:     map[pfcp.SEID]CommonStateKey{},
	}
}

func (cs *CommonState) String() string {
	cs.l.Lock()
	defer cs.l.Unlock()
	var sb strings.Builder
	for k, v := range cs.m {
		fmt.Fprintf(&sb, "%s:%s\n", k, &v)
	}
	return sb.String()
}

/*

Common State Session Functions

Common state aggregates sgw and pgw session descriptors, based on a common key which is the shared tunnel between the sgw and pgw

The inputs come from the two independent state machines for sgw and pgw.
There are two input types - insert/update, and remove.
The input sources have no awareness of the aggregate state, or of actions which arise from it.

The common state is responsible for signalling aggregate session events, via a 5g smf function, to a 5g upf.
As long as a aggregate session state exists, i.e. both sgw and pgw have current entries, then the aggregate session is maintained, even if the session is not 'viable'.
When either of sgw or pgw withdraw a half session entry then the entire session is removed from the SMF, but still held in common state.
A new replacement for a missing half session would build a new upf session, though that is not expected.

The sgw and pgw half sessions include their respective SEID, but this is used for consistency checking and reporting only.
A changed SEID would be logged but the change accepted.

A half session can be 'suspended' when the session state is insufficient to complete a SMF session.

Future / IE driven
The sgw and pgw parsers would isolate the PDR and FAR elements, and forward the respective elements for the external facing interfaces into the common state.
The internal PDRs/FARs define the 'bridging state', e.g. may apply buffering rules, which would need to be promoted to the external state.
In this model, the parse action could be defined as two stages - 1) separate IEs 2) process/parse only local IEs.

Insert action

Insert receives sgw or pgw half sessions.
The existing state may or may not have an entry corresponding.
Such an entry is equivalent to the offered parameter values.
The key is the common part, by definition they agree.
The other parts are the value (nullable), and SEID (validation only).
The primary function on receiving updates is to supply session messages to SMF.
In general, the SMF action corresponds to an initial transition from partial session state to complete, and the SMF action is obvious.
Other actions are transitions from complete to partial, or complete to complete, where some state changed.
These are all 'session modify events', where the required action is a distinct policy choice.
In contrast, a remove event is irreversible, and the SMF action is well defined.
Remaining common state can only proceed via a fresh SMF session.
The modification event is processed in the SMF function rather than in common state.
The initial SMF session formation opens a SMF session, subsequent actions are sent/processed by the SMF session object.
This means that the session reference needs to be stored in common state, and is used as the indicator to distinguish an initial session formation from later modifications.

Remove action

The remove action returns common state to the equivalent state when only one half session has been signalled.
It also invokes the delete function on an existing session.
If there is no existing session then no delete is needed.

Feedback to pgw / sgw

In this model, the sgw and pgw accept sessions regardless of underlying smf/upf session actions.
In the event that a SMF signalled session fails there is no simple mechanism to convey the failure.
An alternate mode would return outcome indications for one or both session actions after SMF session request completes.
The model in which no reply is sent to either SGWc or PGWc is possible but risks deadlock.
Deferring the second request of either SGW or PGW is more likely to be safe.
A way to do this is to add an optional callback/reply channel to the common state request.
The switches to control the behaviour would require the common state to expose to SGW/PGW the state of the other party.

Common state API
a request is {seid, CommonStateKey, *XSessionState}
CommonState.insert(request)
CommonState.remove(request)
 func (csm *CommonStateMap) insert(seid pfcp.SEID, commonStateKey CommonStateKey ,xgwSessionState XgwSessionState) {
	log.Tracef("(CommonStateMap) insert(seid %s, commonStateKey %s ,xgwSessionState %s)",seid,commonStateKey,xgwSessionState)
 }
func (csm *CommonStateMap) deleteSessionState(csk CommonStateKey, role pfcpRole) (actionRequired bool, err error) {


Implied SMF API
The SMF API accepts create, modify and delete inputs.
Modify can be add/change/remove of either sgw or pgw parameters.
In future it could include buffering flags.
The transformation of these events into signalled requests is required to drive a single session state system
Ideally, the change could be isolated to specific FAR and/or PDR - the expected change in SGW FAR being the obvious case.

In order to make the SMF stateless, the SMF API requires the prior and new state to decide the required action.

The API is
association.create(sgw state, pgw state) -> session object
session.modify(old sgw state, old pgw state, old sgw state, old pgw state)
session.delete()
*/

func (csm *CommonState) update(seid pfcp.SEID, commonStateKey CommonStateKey, xgwSessionState XgwSessionState) {
	log.Tracef("(CommonStateMap) update(seid %s, commonStateKey %s ,xgwSessionState %s)", seid, commonStateKey, xgwSessionState)

	csm.l.Lock()

	var mapRef CommonStateSeidMap

	switch xgwSessionState.(type) {
	case *SgwSessionState:
		mapRef = csm.sgwSeidMap
	case *PgwSessionState:
		mapRef = csm.pgwSeidMap
	default:
		panic("shti hnappened")
	}

	cssSeid, cssSeidPresent := mapRef[seid]
	if cssSeidPresent && cssSeid != commonStateKey {
		log.Errorf("seid mismatch for css %s != %s seid %s", cssSeid, commonStateKey, seid)
	} else if !cssSeidPresent {
		mapRef[seid] = commonStateKey
	}

	oldCsv, csvPresent := csm.m[commonStateKey]

	if !csvPresent {
		log.Trace("no state exists, building a new CSV")
	} else {
		log.Trace("CSV state found")
	}

	newCsv := oldCsv

	switch xs := xgwSessionState.(type) {
	case *SgwSessionState:
		newCsv.SgwSessionState = xs
		newCsv.sgwSeid = seid
	case *PgwSessionState:
		newCsv.PgwSessionState = xs
		newCsv.pgwSeid = seid
	default:
		panic("shti hnappened")
	}

	csm.m[commonStateKey] = newCsv
	csm.l.Unlock()
	csm.sessionEvent(commonStateKey, &oldCsv, &newCsv)
}

func (csm *CommonState) remove(seid pfcp.SEID, xgwSessionState XgwSessionState) {
	log.Tracef("(CommonStateMap) remove(seid %s ,xgwSessionState %s)", seid, xgwSessionState)

	csm.l.Lock()

	var mapRef CommonStateSeidMap

	switch xgwSessionState.(type) {
	case SgwSessionState:
		mapRef = csm.sgwSeidMap
	case PgwSessionState:
		mapRef = csm.pgwSeidMap
	}

	if css, present := mapRef[seid]; !present {
		csm.l.Unlock()
		log.Errorf("logic error, deleting common state for SEID, css state not found for SEID %s", seid)
	} else {
		delete(mapRef, seid)
		if csv, present := csm.m[css]; !present {
			csm.l.Unlock()
			log.Errorf("logic error, deleting common state for SEID, cs state not found for SEID %s,css %s", seid, css)
		} else {
			log.Tracef("deleting common state for SEID %s css %s state %s", seid, css, &csv)
			csvOnEntry := csv

			switch xgwSessionState.(type) {
			case SgwSessionState:
				csv.SgwSessionState = nil
				csv.sgwSeid = 0
			case PgwSessionState:
				csv.PgwSessionState = nil
				csv.pgwSeid = 0
			}

			if csv.PgwSessionState == nil && csv.SgwSessionState == nil {
				log.Tracef("removing all csm state for %s", css)
				delete(csm.m, css)
			} else {
				log.Tracef("not removing all csm state for %s", css)
				csm.m[css] = csv
			}

			csm.l.Unlock()
			csm.sessionEvent(css, &csvOnEntry, &csv)

		}
	}
}

func (csm *CommonState) sessionEvent(commonStateKey CommonStateKey, oldCsv, newCsv *CommonStateValue) {
	diff := (oldCsv).Differs(newCsv)

	if len(diff) == 0 {
		log.Trace("(*CommonState) sessionEvent() - null event")
	} else {
		log.Tracef("(*CommonState) sessionEvent() - state change detected - %s", diff)
	}

	if oldCsv.SessionObject == nil {
		if newCsv.PgwSessionState != nil && newCsv.SgwSessionState != nil {
			log.Trace("new session required")
			csm.sessionCreate(commonStateKey, newCsv)
		} else {
			log.Trace("no new session required, state incomplete")
		}
	} else {
		if newCsv.PgwSessionState != nil && newCsv.SgwSessionState != nil {
			log.Trace("session modify required")
			csm.sessionModify(commonStateKey, oldCsv, newCsv)
		} else {
			log.Trace("session delete required")
			csm.sessionDelete(commonStateKey)
		}
	}
}

func (csm *CommonState) sessionCreate(commonStateKey CommonStateKey, commonStateValue *CommonStateValue) {
	log.Tracef("(*CommonState) sessionCreate(commonStateKey %s, commonStateValue %s)", commonStateKey, commonStateValue)
	csm.l.Lock()
	csv, present := csm.m[commonStateKey]
	csm.l.Unlock()

	// TODO *****
	// There should be a CSK level lock - two threads should not try to do session changes on the same target unsynchronised!
	// The lock acquiistion could handle map level locks...

	if !present {
		log.Errorf("cannot create session for missing CSV (%s)", commonStateKey)
		return
	} else if csv.SessionObject != nil {
		log.Errorf("session already present for create session request (%s)", commonStateKey)
		csv.SessionObject.delete()
	} else {
		log.Tracef("got CSV for create session request (%s)", commonStateKey)
	}

	sessionObject := csm.SessionService.create(commonStateValue)
	csv.SessionObject = sessionObject
	csm.l.Lock()
	csm.m[commonStateKey] = csv
	csm.l.Unlock()
}

func (csm *CommonState) sessionDelete(commonStateKey CommonStateKey) {
	log.Tracef("(*CommonState) sessionDelete(commonStateKey %s)", commonStateKey)
	csm.l.Lock()
	csv, present := csm.m[commonStateKey]
	csm.l.Unlock()

	if !present {
		log.Errorf("cannot delete session for missing CSV (%s)", commonStateKey)
		return
	} else if csv.SessionObject == nil {
		log.Errorf("session not present for delete session request (%s)", commonStateKey)
		return
	} else {
		log.Tracef("got CSV for delete session request (%s)", commonStateKey)
		csv.SessionObject.delete()
		csv.SessionObject = nil
		csm.l.Lock()
		csm.m[commonStateKey] = csv
		csm.l.Unlock()
	}
}

func (csm *CommonState) sessionModify(commonStateKey CommonStateKey, oldCommonStateValue, newCommonStateValue *CommonStateValue) {
	log.Tracef(
		"(*CommonState) sessionModify(commonStateKey %s, oldCommonStateValue %s, newCommonStateValue %s)",
		commonStateKey,
		oldCommonStateValue,
		newCommonStateValue,
	)

	csm.l.Lock()
	csv, present := csm.m[commonStateKey]
	csm.l.Unlock()

	if !present {
		log.Errorf("cannot modify session for missing CSV (%s)", commonStateKey)
		return
	} else if csv.SessionObject == nil {
		log.Errorf("session not present for modify session request (%s)", commonStateKey)
		return
	} else {
		log.Tracef("got CSV for modify session request (%s)", commonStateKey)
		csv.SessionObject.modify(oldCommonStateValue, newCommonStateValue)
		csv.SessionObject = nil
	}
}
