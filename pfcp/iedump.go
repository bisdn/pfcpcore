// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package pfcp

import (
	"fmt"
	"strings"
)

func pad(sb *strings.Builder, level int) {
	sb.WriteString(strings.Repeat(" ", level))
}

// dump is a reference consumer of a valid parse tree
// it visits each node and calls the dump function every sub-node
// note - an entry in the interface is required for every such function

func (node IeNode) dump(sb *strings.Builder, level int) {
	pad(sb, level)
	if node.IeID != nil {
		fmt.Fprintf(sb, "%s (%d)\n", node.IeTypeCode, *node.IeID)
	} else if node.IeTypeCode.isGroupIe() {
		fmt.Fprintf(sb, "%s\n", node.IeTypeCode)
	} else {
		fmt.Fprintf(sb, "%s: %s\n", node.IeTypeCode, node.show())
	}

	for i := range node.ies {
		node.ies[i].dump(sb, level+2)
	}
}

func (node IeNode) Dump() string {
	sb := new(strings.Builder)
	node.dump(sb, 0)
	return sb.String()
}

func (node IeNode) Print() {
	fmt.Println(node.Dump())
}
