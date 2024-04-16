// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package main

import (
	"fmt"

	"pfcpcore/pfcp"
)

type PgwSessionState struct {
	accessPdr, coreFar pfcp.FTeid
	corePdr            pfcp.IpV4
}

func (PgwSessionState PgwSessionState) String() string {
	return fmt.Sprintf("pgw: access: %s -> |, core: %s -> %s",
		PgwSessionState.accessPdr,
		PgwSessionState.corePdr,
		PgwSessionState.coreFar,
	)
}
