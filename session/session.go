// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package session

import (
	"fmt"

	"pfcpcore/pfcp"
)

func AccessSER(ser *pfcp.PfcpMessage) {
	fmt.Printf("success, now attempting to access message\n")

	fmt.Printf("\n\n----------------------\ndemonstrate access by simple means\n----------------------\n")
	rootNode := ser.Node()

	fmt.Printf("reading PDR, id=0\n")
	pdr := rootNode.Getter().GetById(pfcp.Create_PDR, 0)
	pdr.Inner().Print()

	pdi := rootNode.Getter().GetById(pfcp.Create_PDR, 0).GetByTc(pfcp.PDI)
	pdi.Inner().Print()

	sourceInterface := rootNode.Getter().GetById(pfcp.Create_PDR, 0).GetByTc(pfcp.PDI).GetByTc(pfcp.Source_Interface)
	sourceInterface.Inner().Print()
	if sourceInterface.Inner().DeserialiseEnumInterface() == pfcp.EnumAccess {
		fmt.Println("found the access PDR by knowing the PDR ID ")
	}
	fmt.Println("\ninspecting PDRs by ID")

	for _, pdrId := range []pfcp.IeID{0, 32768} {
		fmt.Printf("getting PDR, ID=%s\n", &pdrId)

		if rootNode.Getter().GetById(pfcp.Create_PDR, pdrId).GetByTc(pfcp.PDI).GetByTc(pfcp.Source_Interface).Inner().DeserialiseEnumInterface() == pfcp.EnumAccess {
			fmt.Println("confirm the PDR is access")
		} else {
			rootNode.Getter().GetById(pfcp.Create_PDR, pdrId).Print()
			fmt.Println("is not the access PDR!")
			fmt.Printf("the source interface is %s\n",
				rootNode.Getter().GetById(pfcp.Create_PDR, pdrId).GetByTc(pfcp.PDI).GetByTc(pfcp.Source_Interface).Inner().DeserialiseEnumInterface())
		}
		fmt.Println("")
	}

	// badForm := rootNode.Getter().GetById(pfcp.Create_PDR, 0).GetByTc(pfcp.PDI).GetByTc(pfcp.Source_IP_Address)
	// badForm.Inner().Print()
	fmt.Printf("\n\n----------------------\ndemonstrate access by predicate (finding PDRs)\n----------------------\n")

	// could define a customer getter here, e.g.
	// pdrIsCore := pdrHasInterfaceType(pfcp.EnumCore)

	accessPdr := rootNode.Getter().GetByPredicate(pfcp.Create_PDR, pdrIsAccess)
	// accessPdr := rootNode.Getter().GetByPredicate(pfcp.Create_PDR, pdrHasInterfaceType(pfcp.EnumAccess))
	corePdr := rootNode.Getter().GetByPredicate(pfcp.Create_PDR, pdrHasInterfaceType(pfcp.EnumCore))

	fmt.Printf("found the access PDR again:\n%s\n", accessPdr.Dump())
	fmt.Printf("found the core PDR again:\n%s\n", corePdr.Dump())
	fmt.Printf("all done (getter)\n")
}

type Session struct {
	Seid                                    pfcp.SEID
	UeIpAdress, EnbIpAddress, CoreIpAddress pfcp.IpV4
	UplinkTeid, DownlinkTeid                pfcp.TEID
}

func (session Session) String() string {
	return fmt.Sprintf(
		"SEID:%8x UE: %s\neNb: %s\nUPF: %s\nUplink TEID: %d\nDownlink TEID: %d\n",
		session.Seid,
		session.UeIpAdress,
		session.EnbIpAddress,
		session.CoreIpAddress,
		session.UplinkTeid,
		session.DownlinkTeid,
	)
}

func ParseSERSeid(ser *pfcp.IeNode) (*pfcp.FSeid, error) {
	return ser.Getter().GetByTc(pfcp.F_SEID).DeserialiseFSeid()
}

func ParseSER(ser *pfcp.IeNode) (uint64, error) {
	root := ser.Getter()
	// accessPdr := root.GetByPredicate(pfcp.Create_PDR, pdrHasInterfaceType(pfcp.EnumAccess))
	// corePdr := root.GetByPredicate(pfcp.Create_PDR, pdrHasInterfaceType(pfcp.EnumCore))
	// accessPdrId := accessPdr.GetByTc(pfcp.PDR_ID).Inner().ParseID()
	// accessFarId := accessPdr.GetByTc(pfcp.FAR_ID).Inner().ParseID()
	// corePdrId := corePdr.GetByTc(pfcp.PDR_ID).Inner().ParseID()
	// coreFarId := corePdr.GetByTc(pfcp.FAR_ID).Inner().ParseID()

	// coreFar := root.GetById(pfcp.Create_FAR, coreFarId)

	// // accessFar not needed in simple example
	// // accessFar := root.GetById(pfcp.Create_FAR, accessFarId)

	// fmt.Printf("located PDRs:\naccess: %d FAR: %d\ncore: %d FAR: %d\n", accessPdrId, accessFarId, corePdrId, coreFarId)

	// seid, _ := root.GetByTc(pfcp.F_SEID).DeserialiseU64()
	// session.Seid = pfcp.SEID(seid)

	return root.GetByTc(pfcp.F_SEID).DeserialiseU64()
}

func ParseUpfSER(ser *pfcp.IeNode) (session Session) {
	fmt.Printf("now attempting to parse message\n")
	root := ser.Getter()
	accessPdr := root.GetByPredicate(pfcp.Create_PDR, pdrHasInterfaceType(pfcp.EnumAccess))
	corePdr := root.GetByPredicate(pfcp.Create_PDR, pdrHasInterfaceType(pfcp.EnumCore))
	accessPdrId := accessPdr.GetByTc(pfcp.PDR_ID).Inner().ParseID()
	accessFarId := accessPdr.GetByTc(pfcp.FAR_ID).Inner().ParseID()
	corePdrId := corePdr.GetByTc(pfcp.PDR_ID).Inner().ParseID()
	coreFarId := corePdr.GetByTc(pfcp.FAR_ID).Inner().ParseID()

	coreFar := root.GetById(pfcp.Create_FAR, coreFarId)

	// accessFar not needed in simple example
	// accessFar := root.GetById(pfcp.Create_FAR, accessFarId)

	fmt.Printf("located PDRs:\naccess: %d FAR: %d\ncore: %d FAR: %d\n", accessPdrId, accessFarId, corePdrId, coreFarId)

	seid, _ := root.GetByTc(pfcp.F_SEID).DeserialiseU64()
	session.Seid = pfcp.SEID(seid)

	session.UeIpAdress = corePdr.GetByTc(pfcp.PDI).GetByTc(pfcp.UE_IP_Address).Inner().DeserialiseUeIPAddress()

	outerHeader := coreFar.
		GetByTc(pfcp.Forwarding_Parameters).
		GetByTc(pfcp.Outer_Header_Creation).
		Inner().DeserialiseOuterHeader()

	session.EnbIpAddress = outerHeader.IpV4
	session.DownlinkTeid = outerHeader.Teid

	uplinkFTeid := accessPdr.GetByTc(pfcp.PDI).
		GetByTc(pfcp.F_TEID).
		Inner().DeserialiseFTeid()

	session.UplinkTeid = *uplinkFTeid.Teid
	session.CoreIpAddress = *uplinkFTeid.IpV4
	return
}

// example custom predicate (the only known use is for tracking down a PDR based on role)
func pdrIsAccess(get *pfcp.Get) bool {
	return get.GetByTc(pfcp.PDI).GetByTc(pfcp.Source_Interface).Inner().DeserialiseEnumInterface() == pfcp.EnumAccess
}

// example of creating a more flexible predicate, used below to get the core interface
func pdrHasInterfaceType(interfaceType pfcp.EnumInterface) pfcp.Predicate {
	return func(get *pfcp.Get) bool {
		return get.GetByTc(pfcp.PDI).GetByTc(pfcp.Source_Interface).Inner().DeserialiseEnumInterface() == interfaceType
	}
}
