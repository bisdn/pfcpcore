// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package pfcp

import (
	"fmt"
	"strings"
)

type Get struct {
	*IeNode         // although a pointer,it is not allowed that it be invalid, code need not check it explicitly (panic allowed)
	err      string // Perhaps, this should change to a slice of strings, since it is copied at each parse step.  But,only when the parse failed is it not empty.
	contexts []IeTypeCode
}

func (get Get) Return() (*IeNode, error) {
	if len(get.err) == 0 {
		return get.IeNode, nil
	} else {
		return nil, get.Error()
	}
}

func (get *Get) Error() error {
	if len(get.err) == 0 {
		return nil
	} else {
		sb := new(strings.Builder)
		for i := range get.contexts {
			fmt.Fprintf(sb, "%d ", get.contexts[i])
		}
		fmt.Fprintf(sb, "%s", get.err)
		return fmt.Errorf(sb.String())
	}
}

func (get Get) Inner() *IeNode {
	if len(get.err) == 0 {
		return get.IeNode
	} else {
		panic(get.Error())
	}
}

func (node *IeNode) Getter() Get {
	return Get{
		IeNode: node,
	}
}

func (node *IeNode) getByTc(tc IeTypeCode) *IeNode {
	for i := range node.ies {
		if tc == node.ies[i].IeTypeCode && node.ies[i].IeID == nil { // note: grouped IEs cannot be got by this method, if needed, write getFirst, otherwise panic perhaps?
			return &node.ies[i]
		}
	}
	return nil
}

// **********************************
//	IMPORTANT NOTE!!!!!
// **********************************
//
// The below methods 'func (get *Get) GetXXX' MUST take Get not *Get as their receiver, because they return object different to their inputs,
// and the strategy to instantiate the returned value is to use the input parameter as a (new,copy) value.
// A pointer receiver would have to explicitly copy the input if it were not a value parameter

func (get Get) GetByTc(tc IeTypeCode) Get {
	// need a wrapper to simplify this test...?
	if len(get.err) == 0 {
		if ie := get.IeNode.getByTc(tc); ie != nil {
			get.IeNode = ie
		} else {
			// local fail
			get.err = fmt.Sprintf("GetByTc, not found, %s", tc)
		}
	} else {
		// prior fail
		get.contexts = append(get.contexts, get.IeNode.IeTypeCode)
	}
	return get
}

func (node *IeNode) getById(tc IeTypeCode, id IeID) *IeNode {
	for i := range node.ies {
		if tc == node.ies[i].IeTypeCode && node.ies[i].IeID != nil && *node.ies[i].IeID == id {
			return &node.ies[i]
		}
	}
	return nil
}

func (get Get) GetById(tc IeTypeCode, id IeID) Get {
	// need a wrapper to simplify this test...?
	if len(get.err) == 0 {
		if ie := get.IeNode.getById(tc, id); ie != nil {
			get.IeNode = ie
		} else {
			// local fail
			get.err = fmt.Sprintf("GetById, not found, %s %d", tc, id)
		}
	} else {
		// prior fail
		get.contexts = append(get.contexts, get.IeNode.IeTypeCode)
	}
	return get
}

type Predicate = func(*Get) bool // TODO extend predicate so that it has a name which can be used for context logs

func (get Get) GetByPredicate(tc IeTypeCode, predicate Predicate) Get {
	if len(get.err) != 0 {
		// prior fail
		// not clear how this should be handled....
		// but obviously it fails safe
		// get.contexts = append(get.contexts, get.ieNode.IeTypeCode)
	} else {
		proxyGet := &Get{IeNode: nil, err: "", contexts: nil}
		for i := range get.IeNode.ies {
			if tc == get.IeNode.ies[i].IeTypeCode {
				proxyGet.IeNode = &get.IeNode.ies[i]
				if predicate(proxyGet) {
					get.IeNode = proxyGet.IeNode
					break
				}
			}
		}
	}
	return get
}
