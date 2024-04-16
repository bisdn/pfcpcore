// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package main

import (
	"fmt"

	"pfcpcore/pfcp"
)

func pdrHasInterfaceType(interfaceType pfcp.EnumInterface) pfcp.Predicate {
	return func(get *pfcp.Get) bool {
		return get.GetByTc(pfcp.PDI).GetByTc(pfcp.Source_Interface).Inner().DeserialiseEnumInterface() == interfaceType
	}
}

func ParseUpfSER(ser *pfcp.IeNode) (*PgwSessionState, error) {
	pgwSessionState := new(PgwSessionState)
	root := ser.Getter()
	accessPdr := root.GetByPredicate(pfcp.Create_PDR, pdrHasInterfaceType(pfcp.EnumAccess))
	corePdr := root.GetByPredicate(pfcp.Create_PDR, pdrHasInterfaceType(pfcp.EnumCore))

	if coreFarId, err := corePdr.GetByTc(pfcp.FAR_ID).ParseID(); err != nil {
		return nil, err
	} else {
		coreFar := root.GetById(pfcp.Create_FAR, coreFarId)
		if pgwSessionState.corePdr, err = corePdr.GetByTc(pfcp.PDI).GetByTc(pfcp.UE_IP_Address).DeserialiseUeIPAddress(); err != nil {
			return nil, err
		} else if outerHeader, err := coreFar.GetByTc(pfcp.Forwarding_Parameters).GetByTc(pfcp.Outer_Header_Creation).DeserialiseOuterHeader(); err != nil {
			return nil, err
		} else if accessPdr, err := accessPdr.GetByTc(pfcp.PDI).GetByTc(pfcp.F_TEID).DeserialiseFTeid(); err != nil {
			return nil, err
		} else if accessPdr.IpV4 == nil {
			return nil, fmt.Errorf("missing ipv4 in FTEID")
		} else if accessPdr.Teid == nil {
			return nil, fmt.Errorf("missing teid in FTEID")
		} else {
			pgwSessionState.coreFar = outerHeader.FTeid()
			pgwSessionState.accessPdr = *accessPdr
		}
	}
	return pgwSessionState, nil
}
