// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package events

// For each chaos resource, the controller manager will record events
// once the chaos is started or completed or failed. The starting and
// completing events should be of type "Normal". And the failed events should
// be of type "Warning". The reasons are defined as following.
const (
	// The chaos just started
	ChaosInjected string = "ChaosInjected"

	// The chaos just failed when injecting. The message should include detailed error
	ChaosInjectFailed string = "ChaosInjectFailed"

	// The chaos just failed when recovering. The message should include detailed error
	ChaosRecoverFailed string = "ChaosRecoverFailed"

	// The chaos just completed
	ChaosRecovered string = "ChaosRecovered"
)
