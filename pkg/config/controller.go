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
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/mcuadros/go-defaults"
	"github.com/spf13/viper"
	"os"
	"regexp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/chaos-mesh/chaos-mesh/pkg/webhook/config/watcher"
)

// ChaosControllerConfig defines the configuration for Chaos Controller
type ChaosControllerConfig struct {
	// ChaosDaemonPort is the port which grpc server listens on
	ChaosDaemonPort int `mapstructure:"chaos_daemon_port" default:"31767"`
	// BPFKIPort is the port which BFFKI grpc server listens on
	BPFKIPort int `mapstructure:"bpfki_port" default:"50051"`
	// MetricsAddr is the address the metric endpoint binds to
	MetricsAddr string `mapstructure:"metrics_addr" default:":10080"`
	// PprofAddr is the address the pprof endpoint binds to.
	PprofAddr string `mapstructure:"pprof_addr" default:"0"`
	// EnableLeaderElection is enable leader election for controller manager
	// Enabling this will ensure there is only one active controller manager
	EnableLeaderElection bool `mapstructure:"enableLeader_election" default:"false"`
	// CertsDir is the directory for storing certs key file and cert file
	CertsDir string `mapstructure:"certs_dir" default:"/etc/webhook/certs"`
	// AllowedNamespaces is a regular expression, and matching namespace will allow the chaos task to be performed
	AllowedNamespaces string `mapstructure:"allowed_namespaces" default:""`
	// IgnoredNamespaces is a regular expression, and the chaos task will be ignored by a matching namespace
	IgnoredNamespaces string `mapstructure:"ignored_namespaces" default:""`
	// RPCTimeout is timeout of RPC between controllers and chaos-operator
	RPCTimeout    time.Duration `mapstructure:"rpc_timeout" default:"1m"`
	WatcherConfig *watcher.Config `mapstructure:"watcher_config"`
	// ClusterScoped means control Chaos Object in cluster level(all namespace),
	ClusterScoped bool `mapstructure:"cluster_scoped" default:"true"`
	// TargetNamespace is the target namespace to injecting chaos.
	// It only works with ClusterScoped is false;
	TargetNamespace string `mapstructure:"target_namespace" default:""`

	// DNSServiceName is the name of DNS service, which is used for DNS chaos
	DNSServiceName string `mapstructure:"chaos_dns_service_name" default:""`
	DNSServicePort int    `mapstructure:"chaos_dns_service_port" default:""`

	// Namespace is the namespace which the controller manager run in
	Namespace string `mapstructure:"namespace" default:""`

	// AllowHostNetworkTesting removes the restriction on chaos testing pods with `hostNetwork` set to true
	AllowHostNetworkTesting bool `mapstructure:"allow_host_network_testing" default:"false"`
}

var ControllerCfg *ChaosControllerConfig
var log = ctrl.Log.WithName("config")

func init() {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	v := viper.NewWithOptions(viper.KeyDelimiter("::"))

	v.AutomaticEnv()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("/etc/chaos-mesh/")
	err := v.ReadInConfig()
	if err != nil {
		log.Error(err, "fail to read config")
		os.Exit(1)
	}

	err = updateConfig(v)
	if err != nil {
		log.Error(err, "fail to load config")
		os.Exit(1)
	}

	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {
		err := updateConfig(v)
		if err != nil {
			log.Error(err, "fail to reload config")
		}
	})
}

func updateConfig(v *viper.Viper) error {
	newCfg := &ChaosControllerConfig{}
	defaults.SetDefaults(newCfg)

	err := v.Unmarshal(&newCfg)
	if err != nil {
		return err
	}
	err = validate(newCfg)
	if err != nil {
		return err
	}

	oldPtr := (*unsafe.Pointer)(unsafe.Pointer(&ControllerCfg))
	newPtr := unsafe.Pointer(newCfg)
	atomic.StorePointer(oldPtr, newPtr)

	log.Info("config loaded", "config", ControllerCfg)
	return nil
}

func validate(config *ChaosControllerConfig) error {

	if config.WatcherConfig == nil {
		return fmt.Errorf("required WatcherConfig is missing")
	}

	if config.ClusterScoped != config.WatcherConfig.ClusterScoped {
		return fmt.Errorf("K8sConfigMapWatcher config ClusterScoped is not same with controller-manager ClusterScoped. k8s configmap watcher: %t, controller manager: %t", config.WatcherConfig.ClusterScoped, config.ClusterScoped)
	}

	if !config.ClusterScoped {
		if strings.TrimSpace(config.TargetNamespace) == "" {
			return fmt.Errorf("no target namespace specified with namespace scoped mode")
		}
		if !IsAllowedNamespaces(config.TargetNamespace, config.AllowedNamespaces, config.IgnoredNamespaces) {
			return fmt.Errorf("target namespace %s is not allowed with filter, please check config AllowedNamespaces and IgnoredNamespaces", config.TargetNamespace)
		}

		if config.TargetNamespace != config.WatcherConfig.TargetNamespace {
			return fmt.Errorf("K8sConfigMapWatcher config TargertNamespace is not same with controller-manager TargetNamespace. k8s configmap watcher: %s, controller manager: %s", config.WatcherConfig.TargetNamespace, config.TargetNamespace)
		}
	}

	return nil
}

// IsAllowedNamespaces returns whether namespace allows the execution of a chaos task
func IsAllowedNamespaces(namespace string, allowedNamespaces, ignoredNamespaces string) bool {
	if allowedNamespaces != "" {
		matched, err := regexp.MatchString(allowedNamespaces, namespace)
		if err != nil {
			return false
		}
		return matched
	}

	if ignoredNamespaces != "" {
		matched, err := regexp.MatchString(ignoredNamespaces, namespace)
		if err != nil {
			return false
		}
		return !matched
	}

	return true
}
