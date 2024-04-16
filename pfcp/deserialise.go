// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package pfcp

import "fmt"

// TODO make cause code a type
func (node *IeNode) ReadCauseCode() (uint8, error) {
	root := node.Getter()
	return root.GetByTc(Cause).DeserialiseU8()
}

func (thisIe IeNode) DeserialiseOuterHeader() OuterHeaderCreate {
	return getOuterHeaderCreate(thisIe.bytes)
}

// temporary kludge,UE IP address is a more complex thing than this
func (thisIe IeNode) DeserialiseUeIPAddress() IpV4 {
	return getUeIpAddress(thisIe.bytes)
}

func (thisIe IeNode) DeserialiseFTeid() FTeid {
	if fteid, err := getFteid(thisIe.bytes); err == nil {
		return *fteid
	} else {
		panic("invalid format in FTEID")
	}
}

// the below functions are mostly systematically derived from the unsafe versions above or elsewhere

func (get Get) ParseID() (IeID, error) {
	if len(get.err) != 0 {
		return IEInvalid, (get.Error())
	} else if id := get.IeNode.ParseID(); id == IEInvalid {
		return IEInvalid, fmt.Errorf("problem parsing an ID IE")
	} else {
		return id, nil
	}
}

func (get Get) DeserialiseFSeid() (*FSeid, error) {
	if len(get.err) == 0 {
		return getFSeid(get.IeNode.bytes)
	} else {
		return new(FSeid), (get.Error())
	}
}

func (get Get) DeserialiseFTeid() (*FTeid, error) {
	if len(get.err) == 0 {
		return getFteid(get.IeNode.bytes)
	} else {
		return new(FTeid), (get.Error())
	}
}

func (get Get) DeserialiseUeIPAddress() (IpV4, error) {
	if len(get.err) == 0 {
		return get.IeNode.DeserialiseUeIPAddress(), nil
	} else {
		return *new(IpV4), (get.Error())
	}
}

func (get Get) DeserialiseOuterHeader() (OuterHeaderCreate, error) {
	if len(get.err) == 0 {
		return get.IeNode.DeserialiseOuterHeader(), nil
	} else {
		return *new(OuterHeaderCreate), (get.Error())
	}
}

func (get Get) DeserialiseNodeIdString() (string, error) {
	if len(get.err) == 0 {
		return showNodeIdString(get.IeNode.bytes)
	} else {
		return "", (get.Error())
	}
}

func showNodeIdString(bytes []byte) (string, error) {
	if len(bytes) == 0 {
		return "", fmt.Errorf("empty Ie")
	} else {
		switch bytes[0] {
		case 0:
			return ReadIpV4(bytes[1:5]).String(), nil
		case 1:
			return showBytes(bytes[1:]), nil
		case 2:
			return showFQNString(bytes[1:]), nil
		default:
			return "", fmt.Errorf("invalid Node-id format ID")
		}
	}
}
