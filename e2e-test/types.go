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

package test

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	aggregatorclientset "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
	"k8s.io/kubernetes/test/e2e/framework"
)

const (
	imagePullPolicyIfNotPresent = "IfNotPresent"
)

// OperatorConfig describe the configuration during installing chaos-mesh
type OperatorConfig struct {
	Namespace       string
	ReleaseName     string
	Manager         ManagerConfig
	Daemon          DaemonConfig
	Tag             string
	DNSImage        string
	EnableDashboard bool
}

// ManagerConfig describe the chaos-operator configuration during installing chaos-mesh
type ManagerConfig struct {
	Image           string
	Tag             string
	ImagePullPolicy string
}

// DaemonConfig describe the chaos-daemon configuration during installing chaos-mesh
type DaemonConfig struct {
	Image           string
	Tag             string
	Runtime         string
	SocketPath      string
	ImagePullPolicy string
}

// NewDefaultOperatorConfig create the default configuration for chaos-mesh test
func NewDefaultOperatorConfig() OperatorConfig {
	return OperatorConfig{
		Namespace:   "chaos-testing",
		ReleaseName: "chaos-mesh",
		Tag:         "e2e",
		Manager: ManagerConfig{
			Image:           "localhost:5000/pingcap/chaos-mesh",
			Tag:             "latest",
			ImagePullPolicy: imagePullPolicyIfNotPresent,
		},
		Daemon: DaemonConfig{
			Image:           "localhost:5000/pingcap/chaos-daemon",
			Tag:             "latest",
			ImagePullPolicy: imagePullPolicyIfNotPresent,
			Runtime:         "containerd",
			SocketPath:      "/run/containerd/containerd.sock",
		},
		DNSImage: "pingcap/coredns:v0.2.0",
	}
}

type operatorAction struct {
	framework *framework.Framework
	kubeCli   kubernetes.Interface
	aggrCli   aggregatorclientset.Interface
	apiExtCli apiextensionsclientset.Interface
	cfg       *Config
}

func (oi *OperatorConfig) operatorHelmSetValue() string {
	set := map[string]string{
		"controllerManager.image":           fmt.Sprintf("%s:%s", oi.Manager.Image, oi.Manager.Tag),
		"controllerManager.imagePullPolicy": oi.Manager.ImagePullPolicy,
		"chaosDaemon.image":                 fmt.Sprintf("%s:%s", oi.Daemon.Image, oi.Daemon.Tag),
		"chaosDaemon.runtime":               oi.Daemon.Runtime,
		"chaosDaemon.socketPath":            oi.Daemon.SocketPath,
		"chaosDaemon.imagePullPolicy":       oi.Daemon.ImagePullPolicy,
		"dnsServer.create":                  "true",
		"dnsServer.image":                   oi.DNSImage,
		"dashboard.create":                  fmt.Sprintf("%t", oi.EnableDashboard),
	}
	arr := make([]string, 0, len(set))
	for k, v := range set {
		arr = append(arr, fmt.Sprintf("%s=%s", k, v))
	}
	return fmt.Sprintf("\"%s\"", strings.Join(arr, ","))
}

func (oa *operatorAction) operatorChartPath(tag string) string {
	return oa.chartPath(operatorChartName, tag)
}

func (oa *operatorAction) chartPath(name string, tag string) string {
	return filepath.Join(oa.cfg.ChartDir, tag, name)
}

func (oa *operatorAction) manifestPath(tag string) string {
	return filepath.Join(oa.cfg.ManifestDir, tag)
}

func (oa *operatorAction) runKubectlOrDie(args ...string) string {
	cmd := "kubectl"
	klog.Infof("Running '%s %s'", cmd, strings.Join(args, " "))
	out, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		klog.Fatalf("Failed to run '%s %s'\nCombined output: %q\nError: %v", cmd, strings.Join(args, " "), string(out), err)
	}
	klog.Infof("Combined output: %q", string(out))
	return string(out)
}

func (oa *operatorAction) apiVersions() []string {
	stdout := oa.runKubectlOrDie("api-versions")
	return strings.Split(stdout, "\n")
}
