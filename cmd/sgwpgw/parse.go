// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package main

import (
	"pfcpcore/pfcp"

	log "github.com/sirupsen/logrus"
)

func ParseSgwSER(ser *pfcp.IeNode) (commonSessionState *CommonStateKey, sgwSessionState *SgwSessionState) {
	root := ser.Getter()

	getPdrFar := func(interfaceType pfcp.EnumInterface) (pdrSpFteid, farSpFteid *spFteid) {
		pdr := root.GetByPredicate(pfcp.Create_PDR, pdrHasInterfaceType(interfaceType))
		if farId, err := pdr.GetByTc(pfcp.FAR_ID).ParseID(); err != nil {
			log.Trace("missing FAR ID")
			return nil, nil
		} else {

			// first check for FTEID in PDR

			if fteid, err := pdr.GetByTc(pfcp.PDI).GetByTc(pfcp.F_TEID).DeserialiseFTeid(); err == nil {
				if spfteid, err := coerce(*fteid); err == nil {
					pdrSpFteid = &spfteid
				}
			}

			// independently, try to get FAR

			far := root.GetById(pfcp.Create_FAR, farId)

			if outerHeader, e := far.GetByTc(pfcp.Forwarding_Parameters).GetByTc(pfcp.Outer_Header_Creation).DeserialiseOuterHeader(); e != nil {
				return
			} else {
				farFteid := outerHeader.FTeid()
				if spfteid, err := coerce(farFteid); err == nil {
					farSpFteid = &spfteid
				}
				return
			}
		}
	}

	accessPdr, accessFar := getPdrFar(pfcp.EnumAccess)

	corePdr, coreFar := getPdrFar(pfcp.EnumCore)

	if accessPdr != nil {
		sgwSessionState = &SgwSessionState{
			accessPdr: *accessPdr,
			coreFar:   spFteidNull,
		}
		// allow incomplete SGW downlink FAR to be propagated
		if coreFar != nil {
			sgwSessionState.coreFar = *coreFar
		}
	}

	if accessFar != nil && corePdr != nil {
		commonSessionState = &CommonStateKey{
			uplink:   *accessFar,
			downlink: *corePdr,
		}
	}

	if commonSessionState == nil {
		log.Trace("ParseSgwSER - CommonSessionState not set")
	} else {
		log.Trace("ParseSgwSER - CommonSessionState set")
	}
	if sgwSessionState == nil {
		log.Trace("ParseSgwSER - SgwSessionState not set")
	} else {
		log.Trace("ParseSgwSER - SgwSessionState set")
	}

	return
}

func pdrHasInterfaceType(interfaceType pfcp.EnumInterface) pfcp.Predicate {
	return func(get *pfcp.Get) bool {
		return get.GetByTc(pfcp.PDI).GetByTc(pfcp.Source_Interface).Inner().DeserialiseEnumInterface() == interfaceType
	}
}

func ParsePgwSER(ser *pfcp.IeNode) (commonSessionState *CommonStateKey, pgwSessionState *PgwSessionState) {
	var accessPdrFteid, coreFarFteid pfcp.FTeid

	root := ser.Getter()

	corePdr := root.GetByPredicate(pfcp.Create_PDR, pdrHasInterfaceType(pfcp.EnumCore))
	if ueIpv4, err := corePdr.GetByTc(pfcp.PDI).GetByTc(pfcp.UE_IP_Address).DeserialiseUeIPAddress(); err == nil {
		pgwSessionState = &PgwSessionState{corePdr: ueIpv4}
	}

	if coreFarId, err := corePdr.GetByTc(pfcp.FAR_ID).ParseID(); err == nil {
		coreFar := root.GetById(pfcp.Create_FAR, coreFarId)
		if outerHeader, err := coreFar.GetByTc(pfcp.Forwarding_Parameters).GetByTc(pfcp.Outer_Header_Creation).DeserialiseOuterHeader(); err == nil {
			coreFarFteid = outerHeader.FTeid()
		}
	}

	accessPdr := root.GetByPredicate(pfcp.Create_PDR, pdrHasInterfaceType(pfcp.EnumAccess))
	if fteid, err := accessPdr.GetByTc(pfcp.PDI).GetByTc(pfcp.F_TEID).DeserialiseFTeid(); err == nil {
		accessPdrFteid = *fteid
	}

	if downlink, err := coerce(coreFarFteid); err == nil {
		if uplink, err := coerce(accessPdrFteid); err == nil {
			commonSessionState = &CommonStateKey{uplink: uplink, downlink: downlink}
		}
	}

	if commonSessionState == nil {
		log.Trace("ParsePgwSER - CommonSessionState not set")
	} else {
		log.Trace("ParsePgwSER - CommonSessionState set")
	}
	if pgwSessionState == nil {
		log.Trace("ParsePgwSER - PgwSessionState not set")
	} else {
		log.Trace("ParsePgwSER - PgwSessionState set")
	}

	return
}
