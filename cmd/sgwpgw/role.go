// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package main

type pfcpRole uint8

const (
	roleSgw pfcpRole = iota + 1
	rolePgw
)

func (pfcpRole pfcpRole) String() string {
	switch pfcpRole {
	case roleSgw:
		return "sgw"
	case rolePgw:
		return "pgw"
	default:
		return "invalid"
	}
}
