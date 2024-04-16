// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package pfcp

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

// validation operations
// validation use the same iteration of the parse tree as does serialisation
// there should be a abstraction which avoids duplication, but, this is plain golang only....
// Generics might work but their limitations and complexity are to painful to consider trying.

// For initial use, use a string buffer to record each failing case.
// Success is an empty buffer

// func (thisIe ieNode) validate(sb *strings.Builder) {
// 	// currently, validation of plain IEs always succeeds
// }

func (thisNode *IeNode) validate(sb *strings.Builder) {
	if !thisNode.IeTypeCode.isGroupIe() {
	} else if attributeSet, present := groupIeAttributeSets[thisNode.IeTypeCode]; !present {
		fmt.Fprintf(sb, "missing IE value in groupIeAttributeSets %d\n", thisNode.IeTypeCode)
	} else {
		// first validate the local IE set
		thisNode.IeID = validateAttributeSet(sb, &thisNode.ies, thisNode.IeTypeCode, attributeSet)
		// then, recursively validate
		for i := range thisNode.ies {
			thisNode.ies[i].validate(sb)
		}
	}
}

func validateAttributeSet(sb *strings.Builder, ies *[]IeNode, source pfcpTypeCode, attributeSet groupIeAttributeSet) (groupId *IeID) {
	// source parameter is used for context in reports/logs
	countRequired, idRequired := attributeSet.properties()
	requiredMap := make(map[IeTypeCode]struct{})
	uniqueMap := make(map[IeTypeCode]struct{})

	for i := range *ies {
		thisTypeCode := (*ies)[i].IeTypeCode
		if attributes, found := attributeSet[thisTypeCode]; !found {
			fmt.Fprintf(sb, "unallowed IE %d (%s) in IE set '%s'\n", thisTypeCode, thisTypeCode, source)
		} else {
			if attributes.required {
				requiredMap[thisTypeCode] = struct{}{}
			}
			if !attributes.multiple {
				if _, found := uniqueMap[thisTypeCode]; found {
					fmt.Fprintf(sb, "illegal duplicate IE (type %d - %s) found in group IE %s\n", thisTypeCode, thisTypeCode, source)
				} else {
					uniqueMap[thisTypeCode] = struct{}{}
				}
			}
			if attributes.isID {
				if !thisTypeCode.isGroupIe() {
					idValue := (*ies)[i].ParseID()
					groupId = &idValue
				} else {
					fmt.Fprintf(sb, "illegal ID IE (type %d), the ID IE cannot be a group IE!  found in group IE %s\n", thisTypeCode, source)
				}
			}
		}
	}

	// Only now can check required list, as all IEs have been visited
	if int(countRequired) != len(requiredMap) {
		missing := setRemove(attributeSet.required(), requiredMap)

		fmt.Fprintf(sb, "missing required elements in %s:", source)
		for _, ieCodes := range missing {
			fmt.Fprintf(sb, " %s (%d)", ieCodes, ieCodes)
		}
		fmt.Fprintf(sb, "\n")
	}

	if idRequired && groupId == nil {
		fmt.Fprintf(sb, "expected ID IE not found in group IE %s\n", source)
	}
	return
}

func (thisIe *IeNode) ParseID() IeID {
	// tdod implement specific length checks for the distinct IE types
	switch thisIe.IeTypeCode {
	case PDR_ID, FAR_ID, BAR_ID, URR_ID, QER_ID:
		return IeID(thisIe.deserialiseUint())
	}

	// could be any of a protocol error in the peer, or calling program logic error, or missing decoder type here
	log.Errorf("unimplemented ID %d in parseID", thisIe.IeTypeCode)
	return IEInvalid
}
