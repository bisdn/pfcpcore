// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package pfcp

var (
	HeartBeatRequest  PfcpMessage = *NewNodeMessage(PFCP_Heartbeat_Request, IE_RecoveryTimeStamp(1010101))
	HeartBeatResponse PfcpMessage = *NewNodeMessage(PFCP_Heartbeat_Response, IE_RecoveryTimeStamp(2020202))
)
