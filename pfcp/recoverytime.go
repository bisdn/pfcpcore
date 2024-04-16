// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package pfcp

import "time"

var t0 = time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)

func GetRecoveryTime() uint32 {
	return uint32(time.Since(t0).Seconds())
}
