// Copyright 2020 Chaos Mesh Authors.
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

package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/chaos-mesh/chaos-mesh/pkg/webhook/config/watcher"
)

// TLSConfig defines the configuration for chaos-daemon tls client
type TLSConfig struct {
	// ChaosDaemonClientCert is the path of chaos daemon certificate
	ChaosDaemonClientCert string `envconfig:"CHAOS_DAEMON_CLIENT_CERT" default:""`
	// ChaosDaemonClientKey is the path of chaos daemon certificate key
	ChaosDaemonClientKey string `envconfig:"CHAOS_DAEMON_CLIENT_KEY" default:""`
	// ChaosMeshCACert is the path of chaos mesh ca cert
	ChaosMeshCACert string `envconfig:"CHAOS_MESH_CA_CERT" default:""`
}

// ChaosControllerConfig defines the configuration for Chaos Controller
type ChaosControllerConfig struct {
	// ChaosDaemonPort is the port which grpc server listens on
	ChaosDaemonPort int `envconfig:"CHAOS_DAEMON_SERVICE_PORT" default:"31767"`

	TLSConfig

	// The QPS config for kubernetes client
	QPS float32 `envconfig:"QPS" default:"30"`
	// The Burst config for kubernetes client
	Burst int `envconfig:"BURST" default:"50"`

	// BPFKIPort is the port which BFFKI grpc server listens on
	BPFKIPort int `envconfig:"BPFKI_PORT" default:"50051"`
	// MetricsAddr is the address the metric endpoint binds to
	MetricsAddr string `envconfig:"METRICS_ADDR" default:":10080"`
	// PprofAddr is the address the pprof endpoint binds to.
	PprofAddr string `envconfig:"PPROF_ADDR" default:"0"`
	// EnableLeaderElection enables leader election for controller manager
	// Enabling this will ensure there is only one active controller manager
	EnableLeaderElection bool `envconfig:"ENABLE_LEADER_ELECTION" default:"false"`
	// EnableFilterNamespace will filter namespace with annotation. Only the pods/containers in namespace
	// annotated with `chaos-mesh.org/inject=enabled` will be injected
	EnableFilterNamespace bool `envconfig:"ENABLE_FILTER_NAMESPACE" default:"false"`
	// CertsDir is the directory for storing certs key file and cert file
	CertsDir string `envconfig:"CERTS_DIR" default:"/etc/webhook/certs"`
	// RPCTimeout is timeout of RPC between controllers and chaos-operator
	RPCTimeout    time.Duration `envconfig:"RPC_TIMEOUT" default:"1m"`
	WatcherConfig *watcher.Config
	// ClusterScoped means control Chaos Object in cluster level(all namespace),
	ClusterScoped bool `envconfig:"CLUSTER_SCOPED" default:"true"`
	// TargetNamespace is the target namespace to injecting chaos.
	// It only works with ClusterScoped is false;
	TargetNamespace string `envconfig:"TARGET_NAMESPACE" default:""`

	// DNSServiceName is the name of DNS service, which is used for DNS chaos
	DNSServiceName string `envconfig:"CHAOS_DNS_SERVICE_NAME" default:""`
	DNSServicePort int    `envconfig:"CHAOS_DNS_SERVICE_PORT" default:""`

	// SecurityMode is used for enable authority validation in admission webhook
	SecurityMode bool `envconfig:"SECURITY_MODE" default:"true" json:"security_mode"`

	// Namespace is the namespace which the controller manager run in
	Namespace string `envconfig:"NAMESPACE" default:""`

	// AllowHostNetworkTesting removes the restriction on chaos testing pods with `hostNetwork` set to true
	AllowHostNetworkTesting bool `envconfig:"ALLOW_HOST_NETWORK_TESTING" default:"false"`

	// PodFailurePauseImage is used to set a custom image for pod failure
	PodFailurePauseImage string `envconfig:"POD_FAILURE_PAUSE_IMAGE" default:"gcr.io/google-containers/pause:latest"`
}

// EnvironChaosController returns the settings from the environment.
func EnvironChaosController() (ChaosControllerConfig, error) {
	cfg := ChaosControllerConfig{}
	err := envconfig.Process("", &cfg)
	return cfg, err
}
