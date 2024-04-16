// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package pfcp

import (
	"fmt"
)

type actionType uint8

var nullIe IeNode = IeNode{}

const (
	invalidAction actionType = iota
	insertAction  actionType = iota
	deleteAction  actionType = iota
	updateAction  actionType = iota
)

var actionNames = map[actionType]string{
	invalidAction: "invalid",
	insertAction:  "insert",
	deleteAction:  "delete",
	updateAction:  "update",
}

func (action actionType) String() string { return actionNames[action] }

func (target *PfcpMessage) Merge(update *PfcpMessage) (*PfcpMessage, error) {
	// for now assume that the Update message is Session Update (not, e.g., Association Update)

	sessionModificationAttributeSet := MessageIeAttributeSets[PFCP_Session_Modification_Request]

	if err := MergeIes(&target.iEnodes, &update.iEnodes, sessionModificationAttributeSet); err != nil {
		return nil, fmt.Errorf("Merge failed %s", err.Error())
	}
	seid := SEID(42)

	return &PfcpMessage{
		rawPfcpMessageHeader: rawPfcpMessageHeader{
			MessageTypeCode: PFCP_Session_Establishment_Request,
			SEID:            &seid,
			SequenceNumber:  42,
			Priority:        nil,
		},
		iEnodes: target.iEnodes,
	}, nil
}

func MergeIes(target, update *[]IeNode, attributeSet groupIeAttributeSet) error {
	// for now assume that the Update message is Session Update (not, e.g., Association Update)
	// Note - the clumsy multiple usages of '(*updatePtr)[i]' are needed to describe update-in-place of the structure
label1:
	for i := range *update {
		typeCode := (*update)[i].IeTypeCode
		targetIeCode := typeCode
		action := insertAction
		var targetID *IeID
		ieAttributes, found := attributeSet[typeCode]
		if found {
			if ieAttributes.isDelete || ieAttributes.isUpdate {
				targetIeCode = ieAttributes.baseIe
			}
			if ieAttributes.isDelete {
				action = deleteAction
			} else if ieAttributes.isUpdate {
				action = updateAction
			}
			if ieAttributes.multiple {
				if (*update)[i].IeID == nil {
					return fmt.Errorf(("no ID found for multiple IE type"))
				} else {
					targetID = (*update)[i].IeID
				}
			}

			// either a simple IE, or a an unknown or unmarked group IE.
			// it could still be a delete over a simple IE, if empty
		} else if (len((*update)[i].bytes) == 0) && (len((*update)[i].ies) == 0) {
			action = deleteAction
		}

		for j := range *target {
			if (*target)[j].IeTypeCode == targetIeCode && (targetID == nil || *targetID == *(*target)[j].IeID) {

				switch action {
				case updateAction: // this is the recursive case
					if nextAttributeSet, found := groupIeAttributeSets[typeCode]; found {
						if err := MergeIes(&((*target)[j].ies), &((*update)[i].ies), nextAttributeSet); err != nil {
							return err
						}
					} else {
						return fmt.Errorf("attribute set not found for: %s", typeCode)
					}
				case deleteAction:
					(*target)[j] = nullIe // note: requires nullIe to be removed on serialization and otherwise ignored during general processing
				case insertAction: // this is actually replace when the target exists
					(*target)[j] = (*update)[i]
				}
				continue label1
			}
		}
		// fall through to this point only when no match is found between the IE sets.
		if action == insertAction || action == updateAction {

			// when the source modify IE is Update rather than Create then this following action ensures the inserted IE is of the base type
			(*update)[i].IeTypeCode = targetIeCode

			*target = append((*target), (*update)[i])
		} else {
			return fmt.Errorf("invalid action when target not found: %s", &action)
		}
	}
	return nil
}
