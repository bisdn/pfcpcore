// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package main

// =====================================
// spFteid{} is a fully complete 'FTEID', i.e. no pointers.
// Zero values might be 'invalid', but the assumption is that having an spFteid is an assertion that it can form a valid tunnel.
// use *spFteid where there is a possibility of that not being the case.
// Ideally, core pfcp should itself support some concrete type such as spFteid

import (
	"fmt"

	"pfcpcore/pfcp"
)

type spFteid struct {
	pfcp.TEID
	pfcp.IpV4
}

func (spFteid spFteid) String() string {
	return fmt.Sprintf("%s:%s", spFteid.IpV4, spFteid.TEID)
}

func (spFteid spFteid) PfcpFteid() pfcp.FTeid {
	return pfcp.FTeid{
		Teid: &spFteid.TEID,
		IpV4: &spFteid.IpV4,
	}
}

var (
	spFteidZero = spFteid{}
	spFteidNull = spFteid{
		TEID: 0xffffffff,
		IpV4: 0xffffffff,
	}
)

func coerce(pfcpFteid pfcp.FTeid) (spFteid, error) {
	if pfcpFteid.IpV4 == nil {
		return spFteidZero, fmt.Errorf("ipv4 field was nil")
	} else if pfcpFteid.Teid == nil {
		return spFteidZero, fmt.Errorf("teid field was nil")
	} else {
		return spFteid{
			TEID: *pfcpFteid.Teid,
			IpV4: *pfcpFteid.IpV4,
		}, nil
	}
}
