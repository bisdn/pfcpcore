// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package main

import (
	"fmt"
	"net/netip"
	"strings"
)

func parseUpfParameter(localAndPeerName string) (localAddr, peerAddr netip.AddrPort, err error) {
	if sx := strings.Split(localAndPeerName, ","); len(sx) != 2 {
		err = fmt.Errorf("failed to parse upf=\"%s\" as a pair of addr/ports", localAndPeerName)
	} else if localAddr, err = netip.ParseAddrPort(sx[0]); err != nil {
		err = fmt.Errorf("failed to parse upf local value (%s) as addr/port", sx[0])
	} else if peerAddr, err = netip.ParseAddrPort(sx[1]); err != nil {
		err = fmt.Errorf("failed to parse upf peer value (%s) as addr/port", sx[1])
	}
	return
}
