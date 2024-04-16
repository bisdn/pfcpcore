// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package main

import (
	"pfcpcore/smf"
)

// SgwPgwState is a singleton in the sgwpgw application
type SgwPgwState struct {
	*smf.Association
	*CommonState
}

func newSgwPgwState(association *smf.Association) *SgwPgwState {
	if association == nil {
		return &SgwPgwState{
			Association: association,
			CommonState: newCommonStateMap(NullSessionService{}),
		}
	} else {
		return &SgwPgwState{
			Association: association,
			CommonState: newCommonStateMap(newXgwSessionService(association)),
		}
	}
}
