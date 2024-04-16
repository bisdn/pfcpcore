// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package main

import (
	"pfcpcore/pfcp"

	log "github.com/sirupsen/logrus"
)

// SgwPgwApplication is instantiated twice, once each for the pgw and sgw role
type SgwPgwApplication struct {
	role pfcpRole
	*SgwPgwState
}

func newSgwPgwApplication(role pfcpRole, sgwPgwState *SgwPgwState) *SgwPgwApplication {
	return &SgwPgwApplication{
		role:        role,
		SgwPgwState: sgwPgwState,
	}
}

/*
application interface usage note:

CallbackSessionEstablishmentRequest() and CallbackSessionDeletionRequest() are called by pfcp FSM.
The CallbackSessionEstablishmentRequest() is called for every SER and SMR received, but CallbackSessionDeletionRequest() can be called only one per session.
However, CallbackSessionDeletionRequest() may be called even if we have not created a full session due to missing elements in the session state.

The SEID provided in each case is the same, and happens to be the locally generated version.  But that should not be relevant to the operation of this code.

While SMR support does not demand a local state indexed by SEID, SDR does so require.
The state structure and function is identical for SGW and PGW.
For instantiated sessions, that state should store the common state key.
*/

/*
In a pgw or sgw session lifecycle input events are multiple session offers, and at most one session withdraw.
The offers are correlated by an SEID, or possibly in future a custom state.
The recognition of new vs. changed offer is not explicit but can be inferred from the SEID/state.

While session state exists (between initial successful session parse, and session delete), the shared state must be synchronised with session state changes.
It is expected that changes in the shared state key values are less frequent than in shared state non-key values (value values!).

For such changes, the shared key is fixed over the lives of the component sessions, and so the shared state should continue to be valid and associated with the same SEID/sessions for as long as all persist.

However, during initial session state building, the 'glue' common state is not immediately created.  So, initial session state is likely not eligible for installation in common state.

Outline approach:
the local pgw/sgw session state is lightweight/null.  When a state with valid key is presented then a common state action is taken.  Either the 'value' is valid, or not.  An invalid value is still posted, but cannot be used to form a downstream session, though it might modify one.

If a key changes then the prior key entry is invalidated.  The possibilities, based on the stored state, received SEID, and parsed session state are:

initial (new SEID - not in local map) - session parses for common key ->

	enter in common state, save in local map

initial (new SEID - not in local map) - session does not parse for common key ->

	save in local map

follow on (existing SEID - already in local map) - session parses key, key agrees with map ->

	update common state

follow on (existing SEID - already in local map) - session does not parse key, or, parses, but does not agree with map ->

	the state should be removed in common, and if a new key, inserted under the new key

Key points - the common state holds the 'value', even if null, and in future possibly the 'buffer' flag.
This code is not concerned with implications of state change, only keeping it current and correct.
So, it need not 'know' what the prior value in common state is...

order of operation
state: local map, key is seid, value is parse result (common+pgw/sgw)
common API - insert/update and remove operations - all local state is given - SEID is for sanity check! - the pgw/sgw switch is interface/type  based
*/
func (st *SgwPgwApplication) CallbackSessionEstablishmentRequest(seid pfcp.SEID, ser *pfcp.IeNode) (uint8, error) {
	log.Tracef("CallbackSessionEstablishmentRequest() role:%s seid:%s", st.role, seid)

	switch st.role {
	case roleSgw:
		if css, pgs := ParseSgwSER(ser); css != nil {
			st.SgwPgwState.update(seid, *css, pgs)
		}
	case rolePgw:
		if css, pgs := ParsePgwSER(ser); css != nil {
			st.SgwPgwState.update(seid, *css, pgs)
		}
	}
	return 0, nil
}

func (st *SgwPgwApplication) CallbackSessionDeletionRequest(seid pfcp.SEID) (uint8, error) {
	log.Tracef("CallbackSessionDeletionRequest() role:%s seid:%s", st.role, seid)

	switch st.role {
	case roleSgw:
		st.SgwPgwState.remove(seid, SgwSessionState{})

	case rolePgw:
		st.SgwPgwState.remove(seid, PgwSessionState{})
	}

	return 0, nil
}
