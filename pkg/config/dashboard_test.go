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

import (
	"testing"

	"github.com/kelseyhightower/envconfig"
)

func TestChaosDashboardConfig(t *testing.T) {
	config := ChaosDashboardConfig{}
	err := envconfig.Process("", &config)
	if err != nil {
		t.Fatal("Error parsing empty ChaosDashboardConfig", err)
	}

	if config.ListenHost != "0.0.0.0" {
		t.Error("ListenHost is not set")
	}

	if config.ListenPort != 2333 {
		t.Error("ListenPort is not set")
	}

	if config.ClusterScoped != true {
		t.Error("ClusterScoped is not set")
	}

	if config.EnableFilterNamespace != false {
		t.Error("EnableFilterNamespace is not set")
	}

	if config.SecurityMode != true {
		t.Error("SecurityMode is not set")
	}

	if config.DNSServerCreate != true {
		t.Error("DNSServerCreate is not set")
	}
}

func TestTTLConfigWithStringTime(t *testing.T) {
	config := TTLConfigWithStringTime{}
	err := envconfig.Process("", &config)
	if err != nil {
		t.Fatal("Error parsing empty TTLConfigWithStringTime", err)
	}

	if config.ResyncPeriod != "12h" {
		t.Error("ResyncPeriod is not set")
	}

	if config.EventTTL != "168h" {
		t.Error("EventTTL is not set")
	}

	if config.ExperimentTTL != "336h" {
		t.Error("ExperimentTTL is not set")
	}

	if config.ScheduleTTL != "336h" {
		t.Error("ScheduleTTL is not set")
	}

	if config.WorkflowTTL != "336h" {
		t.Error("WorkflowTTL is not set")
	}

	parsed, err := config.Parse()

	if err != nil {
		t.Fatal("Error parsing config", err)
	}

	if parsed.ResyncPeriod.Hours() != 12 {
		t.Errorf("ResyncPeriod is not 12h, but %v", parsed.ResyncPeriod)
	}

	if parsed.EventTTL.Hours() != 168 {
		t.Errorf("EventTTL is not 168h, but %v", parsed.EventTTL)
	}

	if parsed.ExperimentTTL.Hours() != 336 {
		t.Errorf("ExperimentTTL is not 336h, but %v", parsed.ExperimentTTL)
	}

	if parsed.ScheduleTTL.Hours() != 336 {
		t.Errorf("ScheduleTTL is not 336h, but %v", parsed.ScheduleTTL)
	}

	if parsed.WorkflowTTL.Hours() != 336 {
		t.Errorf("WorkflowTTL is not 336h, but %v", parsed.WorkflowTTL)
	}
}
