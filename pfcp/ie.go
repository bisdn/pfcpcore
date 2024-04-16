// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package pfcp

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type IeTypeCode uint16

// TDOD? make the node structure polymorphic over the type code, to allow message+type code to be there too
// The cheaty/golang way is to use a wider field for the type code and differentiate IE, enterprise IE, and message type in the higher bits in the obvious way.
// Since we need to handle enterprise IE any way this is no big deal...
// The issue may be that it may be type unsafe if all of the static maps have keys which are unspecific - it is better to know that the index is one or the other...
type IeNode struct {
	IeTypeCode
	bytes []byte   // only present if the node IS NOT group
	*IeID          // only present if the node IS group
	ies   []IeNode // only present if the node IS group
}

func (IeNode *IeNode) Ies() *[]IeNode {
	return &IeNode.ies
}

func NewGroupNode(tc IeTypeCode, ies ...IeNode) *IeNode {
	return &IeNode{IeTypeCode: tc, ies: ies}
}

func NewIeNode(tc IeTypeCode, bytes []byte) *IeNode {
	return &IeNode{IeTypeCode: tc, bytes: bytes}
}

func (typeCode IeTypeCode) typeName() string { return typeCode.String() }

type pfcpTypeCode interface {
	// pfcpTypeCode can be either a pfcp message or an IE TypeCode
	typeName() string
}

// deserialisation (parsing) operations

/*
  func getNextIe(*[]byte) (*ie, error)
  func getNextIeClass([]byte) ([]ieClass, error)

  getNextIe: removes a single IE from a source byte slice, reducing the source byte slice as it does so
  returns an error only when the numbers don't add up (buffer underflow)

  getNextIeClass: uses getNextIe to recursive parse an entire byte slice -


  note(1) - getNextIe()N is destructive of its input, which it uses to hold the state of the traversal of the byte slice
          The caller should preserve the original slice if it needs to reuse it

  note(2) - getNextIe() could itself recurse in the case that its return value is a grouped IE
          Instead, the function recursiveGetNextIe() provides this action.
		  It is not expected that getNextIe() is used elsewhere, expect perhaps to construct a _validating_ inline parser.
          One might hope that the golang compiler makes the inline substitution
		  TDOD - verify this optimisation is not material
*/

func readIe(current *[]byte) (*IeNode, error) {
	if current == nil || *current == nil {
		return nil, fmt.Errorf("nil slice")
	} else if len(*current) < 4 {
		return nil, fmt.Errorf("short byte slice")
	} else if typeCode := binary.BigEndian.Uint16((*current)[0:2]); typeCode > 0x7fff {
		return nil, fmt.Errorf("vendor type code not supported")
	} else if length := binary.BigEndian.Uint16((*current)[2:4]); int(length+4) > len(*current) {
		return nil, fmt.Errorf("IE length overflows buffer")
	} else {
		rval := &IeNode{
			bytes:      (*current)[4 : 4+length],
			IeTypeCode: IeTypeCode(typeCode),
		}
		(*current) = (*current)[4+length:]
		return rval, nil
	}
}

func readIes(start []byte) ([]IeNode, error) {
	current := start
	var ies []IeNode
	for len(current) > 0 {
		nextIe, err := readIe(&current)
		if err != nil {
			return nil, err
		} else {
			if nextIe.IeTypeCode.isGroupIe() {
				if nextIes, err := readIes(nextIe.bytes); err != nil {
					return nil, err
				} else {
					nextIe.bytes = nil
					nextIe.ies = nextIes
				}
			}
			ies = append(ies, *nextIe)
		}
	}
	return ies, nil
}

//lint:file-ignore U1000 Ignore all unused code, it's for future use in IE parsers
// TODO - urgent!? - make all readonly actions on ieNode take pointers (*ieNode)

func (thisIe IeNode) DeserialiseU8() (uint8, error) {
	if len(thisIe.bytes) != 1 {
		return 0, fmt.Errorf("deserialiseU8 error wrong length (%d)", len(thisIe.bytes))
	} else {
		return thisIe.bytes[0], nil
	}
}

func (thisIe IeNode) DeserialiseU16() (uint16, error) {
	if len(thisIe.bytes) != 2 {
		return 0, fmt.Errorf("deserialiseU16 error wrong length (%d)", len(thisIe.bytes))
	} else {
		return binary.BigEndian.Uint16(thisIe.bytes[0:2]), nil
	}
}

func (thisIe IeNode) DeserialiseU32() (uint32, error) {
	if len(thisIe.bytes) != 4 {
		return 0, fmt.Errorf("deserialiseU32 error wrong length (%d)", len(thisIe.bytes))
	} else {
		return binary.BigEndian.Uint32(thisIe.bytes[0:4]), nil
	}
}

func (thisIe IeNode) DeserialiseU64() (uint64, error) {
	if len(thisIe.bytes) != 8 {
		return 0, fmt.Errorf("deserialiseU64 error wrong length (%d)", len(thisIe.bytes))
	} else {
		return binary.BigEndian.Uint64(thisIe.bytes[0:8]), nil
	}
}

func (thisIe IeNode) deserialiseUint() uint {
	switch len(thisIe.bytes) {
	case 0:
		return 0 // like default, this is an error condition, needs logging
	case 1:
		return uint(thisIe.bytes[0])
	case 2:
		return uint(binary.BigEndian.Uint16(thisIe.bytes[0:2]))
	case 4:
		return uint(binary.BigEndian.Uint32(thisIe.bytes[0:4]))
	case 8:
		return uint(binary.BigEndian.Uint64(thisIe.bytes[0:8]))
	default:
		panic(fmt.Errorf("deserialiseUint error bad length (%d)", len(thisIe.bytes)))
	}
}

// serialisation operations

func writeBEUint16(b *bytes.Buffer, u16 uint16) {
	b.WriteByte(uint8(u16 >> 8))
	b.WriteByte(uint8(u16))
}

func (thisIe IeNode) serialise(b *bytes.Buffer) {
	switch {
	// the code cannot cope with the internally invalid case that both ie child list and direct payload are none nil
	// possibly, panic is still a bad idea though, rather than just choose one.

	case thisIe.bytes != nil && thisIe.ies != nil:
		panic("assertion fail")

	case thisIe.ies != nil:
		// // NB assigning another buffer can be avoided by inserting a zero length TLV for this IE and saving the location of the length field,
		// // then, when the lower layer has filled the buffer, rewriting the length field to account for the added bytes.
		// // Once this simple scheme works the optimisation can be investigated.

		var nb bytes.Buffer
		serialise(&nb, thisIe.ies)
		ie := IeNode{
			IeTypeCode: thisIe.IeTypeCode,
			bytes:      nb.Bytes(),
		}
		ie.serialise(b)

	default: // includes option of a zero payload basic IE
		if len(thisIe.bytes) > 0xffff {
			panic("IE exceeds 16bit length limit")
		}
		writeBEUint16(b, uint16(thisIe.IeTypeCode))
		writeBEUint16(b, uint16(len(thisIe.bytes)))
		b.Write(thisIe.bytes)
	}
}

// this function is equally valid for serialising IEs in both a grouped IE and in a full PFCP message
func serialise(b *bytes.Buffer, ies []IeNode) {
	for _, ie := range ies {
		ie.serialise(b)
	}
}
