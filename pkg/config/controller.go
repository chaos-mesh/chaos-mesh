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
)

// TLSConfig defines the configuration for chaos-daemon and chaosd tls client
type TLSConfig struct {
	// ChaosMeshCACert is the path of chaos daemon ca cert
	ChaosMeshCACert string `envconfig:"CHAOS_MESH_CA_CERT" default:""`
	// ChaosDaemonClientCert is the path of chaos daemon certificate
	ChaosDaemonClientCert string `envconfig:"CHAOS_DAEMON_CLIENT_CERT" default:""`
	// ChaosDaemonClientKey is the path of chaos daemon certificate key
	ChaosDaemonClientKey string `envconfig:"CHAOS_DAEMON_CLIENT_KEY" default:""`

	// ChaosdCACert is the path of chaosd ca cert
	ChaosdCACert string `envconfig:"CHAOSD_CA_CERT" default:""`
	// ChaosdClientCert is the path of chaosd certificate
	ChaosdClientCert string `envconfig:"CHAOSD_CLIENT_CERT" default:""`
	// ChaosdClientKey is the path of chaosd certificate key
	ChaosdClientKey string `envconfig:"CHAOSD_CLIENT_KEY" default:""`
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
	// WebhookHost and WebhookPort are combined into an address the webhook server bind to
	WebhookHost string `envconfig:"WEBHOOK_HOST" default:"0.0.0.0"`
	WebhookPort int    `envconfig:"WEBHOOK_PORT" default:"9443"`
	// MetricsHost and MetricsPort are combined into an address the metric endpoint binds to
	MetricsHost string `envconfig:"METRICS_HOST" default:"0.0.0.0"`
	MetricsPort int    `envconfig:"METRICS_PORT" default:"10080"`
	// PprofAddr is the address the pprof endpoint binds to.
	PprofAddr string `envconfig:"PPROF_ADDR" default:"0"`

	// CtrlAddr os the address the ctrlserver bind to
	CtrlAddr string `envconfig:"CTRL_ADDR"`

	// EnableLeaderElection enables leader election for controller manager
	// Enabling this will ensure there is only one active controller manager
	EnableLeaderElection bool `envconfig:"ENABLE_LEADER_ELECTION" default:"true"`
	// LeaderElectLeaseDuration is the duration that non-leader candidates will
	// wait to force acquire leadership. This is measured against time of
	// last observed ack. (default 15s)
	LeaderElectLeaseDuration time.Duration `envconfig:"LEADER_ELECT_LEASE_DURATION" default:"15s"`
	// LeaderElectRenewDeadline is the duration that the acting control-plane
	// will retry refreshing leadership before giving up. (default 10s)
	LeaderElectRenewDeadline time.Duration `envconfig:"LEADER_ELECT_RENEW_DEADLINE" default:"10s"`
	// LeaderElectRetryPeriod is the duration the LeaderElector clients should wait
	// between tries of actions. (default 2s)
	LeaderElectRetryPeriod time.Duration `envconfig:"LEADER_ELECT_RETRY_PERIOD" default:"2s"`

	// EnableFilterNamespace will filter namespace with annotation. Only the pods/containers in namespace
	// annotated with `chaos-mesh.org/inject=enabled` will be injected
	EnableFilterNamespace bool `envconfig:"ENABLE_FILTER_NAMESPACE" default:"false"`
	// CertsDir is the directory for storing certs key file and cert file
	CertsDir string `envconfig:"CERTS_DIR" default:"/etc/webhook/certs"`
	// RPCTimeout is timeout of RPC between controllers and chaos-operator
	RPCTimeout time.Duration `envconfig:"RPC_TIMEOUT" default:"1m"`
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

	// ChaosdSecurityMode is used for enable mTLS connection between chaos-controller-manager and chaod
	ChaosdSecurityMode bool `envconfig:"CHAOSD_SECURITY_MODE" default:"true" json:"chaosd_security_mode"`

	// Namespace is the namespace which the controller manager run in
	Namespace string `envconfig:"NAMESPACE" default:""`

	// AllowHostNetworkTesting removes the restriction on chaos testing pods with `hostNetwork` set to true
	AllowHostNetworkTesting bool `envconfig:"ALLOW_HOST_NETWORK_TESTING" default:"false"`

	// PodFailurePauseImage is used to set a custom image for pod failure
	PodFailurePauseImage string `envconfig:"POD_FAILURE_PAUSE_IMAGE" default:"gcr.io/google-containers/pause:latest"`

	EnabledControllers []string `envconfig:"ENABLED_CONTROLLERS" default:"*"`
	EnabledWebhooks    []string `envconfig:"ENABLED_WEBHOOKS" default:"*"`

	LocalHelmChartPath string `envconfig:"LOCAL_HELM_CHART_PATH" default:""`
}

// EnvironChaosController returns the settings from the environment.
func EnvironChaosController() (ChaosControllerConfig, error) {
	cfg := ChaosControllerConfig{}
	err := envconfig.Process("", &cfg)
	return cfg, err
}
