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

package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/ttlcontroller"
)

// ChaosDashboardConfig defines the configuration for Chaos Dashboard
type ChaosDashboardConfig struct {
	ListenHost           string            `envconfig:"LISTEN_HOST" default:"0.0.0.0" json:"listen_host"`
	ListenPort           int               `envconfig:"LISTEN_PORT" default:"2333" json:"listen_port"`
	MetricHost           string            `envconfig:"METRIC_HOST" default:"0.0.0.0" json:"-"`
	MetricPort           int               `envconfig:"METRIC_PORT" default:"2334" json:"-"`
	EnableLeaderElection bool              `envconfig:"ENABLE_LEADER_ELECTION" json:"-"`
	Database             *DatabaseConfig   `json:"-"`
	PersistTTL           *PersistTTLConfig `json:"-"`
	// ClusterScoped means control Chaos Object in cluster level(all namespace).
	ClusterScoped bool `envconfig:"CLUSTER_SCOPED" default:"true" json:"cluster_mode"`
	// TargetNamespace is the target namespace to injecting chaos.
	// It only works with ClusterScoped is false.
	TargetNamespace string `envconfig:"TARGET_NAMESPACE" default:"" json:"target_namespace"`
	// EnableFilterNamespace will filter namespace with annotation. Only the pods/containers in namespace
	// annotated with `chaos-mesh.org/inject=enabled` will be injected.
	EnableFilterNamespace bool `envconfig:"ENABLE_FILTER_NAMESPACE" default:"false"`
	// SecurityMode will use the token login by the user if set to true
	SecurityMode bool `envconfig:"SECURITY_MODE" default:"true" json:"security_mode"`
	// GcpSecurityMode will use the gcloud authentication to login to GKE user
	GcpSecurityMode bool   `envconfig:"GCP_SECURITY_MODE" default:"false" json:"gcp_security_mode"`
	GcpClientId     string `envconfig:"GCP_CLIENT_ID" default:"" json:"-"`
	GcpClientSecret string `envconfig:"GCP_CLIENT_SECRET" default:"" json:"-"`

	RootUrl string `envconfig:"ROOT_URL" default:"http://localhost:2333" json:"root_path"`

	// enableProfiling is a flag to enable pprof in controller-manager and chaos-daemon
	EnableProfiling bool `envconfig:"ENABLE_PROFILING" default:"true"`

	DNSServerCreate bool   `envconfig:"DNS_SERVER_CREATE" default:"false" json:"dns_server_create"`
	Version         string `json:"version"`

	// The QPS config for kubernetes client
	QPS float32 `envconfig:"QPS" default:"200"`
	// The Burst config for kubernetes client
	Burst int `envconfig:"BURST" default:"300"`
}

// PersistTTLConfig defines the configuration of ttl
type PersistTTLConfig struct {
	SyncPeriod string `envconfig:"CLEAN_SYNC_PERIOD" default:"12h"`
	Event      string `envconfig:"TTL_EVENT"         default:"168h"` // one week
	Experiment string `envconfig:"TTL_EXPERIMENT"    default:"336h"` // two weeks
	Schedule   string `envconfig:"TTL_SCHEDULE"      default:"336h"` // two weeks
	Workflow   string `envconfig:"TTL_WORKFLOW"      default:"336h"` // two weeks
}

// DatabaseConfig defines the configuration for databases
type DatabaseConfig struct {
	Driver     string `envconfig:"DATABASE_DRIVER"     default:"sqlite3"`
	Datasource string `envconfig:"DATABASE_DATASOURCE" default:"core.sqlite"`
}

// GetChaosDashboardEnv gets all env variables related to dashboard.
func GetChaosDashboardEnv() (*ChaosDashboardConfig, error) {
	cfg := ChaosDashboardConfig{}
	err := envconfig.Process("", &cfg)
	return &cfg, err
}

// ParsePersistTTLConfig parse PersistTTLConfig to persistTTLConfigParsed.
func ParsePersistTTLConfig(config *PersistTTLConfig) (*ttlcontroller.TTLConfig, error) {
	syncPeriod, err := time.ParseDuration(config.SyncPeriod)
	if err != nil {
		return nil, errors.Wrap(err, "parse configuration sync period")
	}

	event, err := time.ParseDuration(config.Event)
	if err != nil {
		return nil, errors.Wrap(err, "parse configuration TTL for event")
	}

	experiment, err := time.ParseDuration(config.Experiment)
	if err != nil {
		return nil, errors.Wrap(err, "parse configuration TTL for experiment")
	}

	schedule, err := time.ParseDuration(config.Schedule)
	if err != nil {
		return nil, errors.Wrap(err, "parse configuration TTL for schedule")
	}

	workflow, err := time.ParseDuration(config.Workflow)
	if err != nil {
		return nil, errors.Wrap(err, "parse configuration TTL for workflow")
	}

	return &ttlcontroller.TTLConfig{
		DatabaseTTLResyncPeriod: syncPeriod,
		EventTTL:                event,
		ArchiveTTL:              experiment,
		ScheduleTTL:             schedule,
		WorkflowTTL:             workflow,
	}, nil
}
