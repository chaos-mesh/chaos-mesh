// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

// For each chaos resource, the controller manager will record events
// once the chaos is started or completed or failed. The starting and
// completing events should be of type "Normal". And the failed events should
// be of type "Warning". The reasons are defined as following.
const (
	// The chaos just started
	EventChaosStarted string = "ChaosStarted"

	// The chaos just failed when injecting. The message should include detailed error
	EventChaosInjectFailed string = "ChaosInjectFailed"

	// The chaos just failed when recovering. The message should include detailed error
	EventChaosRecoverFailed string = "ChaosRecoverFailed"

	// The chaos just completed
	EventChaosCompleted string = "ChaosCompleted"
)
