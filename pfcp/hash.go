// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package pfcp

import (
	"bytes"
	"encoding/gob"
	"hash/fnv"
)

func Hash64[K any](s K) uint64 {
	var b bytes.Buffer
	gob.NewEncoder(&b).Encode(s)
	fnv64 := fnv.New64a()
	fnv64.Write(b.Bytes())
	return fnv64.Sum64()
}
