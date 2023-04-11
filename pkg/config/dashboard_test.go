// Copyright 2023 Chaos Mesh Authors.
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

package config

import "testing"

func TestTTLConfigWithStringTime(t *testing.T) {
	config := TTLConfigWithStringTime{
		ResyncPeriod:  "12h",
		EventTTL:      "168h",
		ExperimentTTL: "168h",
		ScheduleTTL:   "168h",
		WorkflowTTL:   "168h",
	}

	if config.ResyncPeriod != "12h" {
		t.Error("ResyncPeriod is not set")
	}

	if config.EventTTL != "168h" {
		t.Error("EventTTL is not set")
	}

	if config.ExperimentTTL != "168h" {
		t.Error("ExperimentTTL is not set")
	}

	if config.ScheduleTTL != "168h" {
		t.Error("ScheduleTTL is not set")
	}

	if config.WorkflowTTL != "168h" {
		t.Error("WorkflowTTL is not set")
	}

	parsed, err := config.Parse()

	if err != nil {
		t.Error("Error parsing config")
	}

	if parsed.ResyncPeriod.Hours() != 12 {
		t.Errorf("ResyncPeriod is not 12h, but %v", parsed.ResyncPeriod)
	}

	if parsed.EventTTL.Hours() != 168 {
		t.Errorf("EventTTL is not 168h, but %v", parsed.EventTTL)
	}

	if parsed.ExperimentTTL.Hours() != 168 {
		t.Errorf("ExperimentTTL is not 168h, but %v", parsed.ExperimentTTL)
	}

	if parsed.ScheduleTTL.Hours() != 168 {
		t.Errorf("ScheduleTTL is not 168h, but %v", parsed.ScheduleTTL)
	}

	if parsed.WorkflowTTL.Hours() != 168 {
		t.Errorf("WorkflowTTL is not 168h, but %v", parsed.WorkflowTTL)
	}
}
