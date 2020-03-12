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

package test

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

const (
	operartorChartName = "chaos-mesh"
)

type OperatorAction interface {
	DeployOperator(config OperatorConfig) error
	InstallCRD(config OperatorConfig) error
}

func NewOperatorAction(
	kubeCli kubernetes.Interface,
	cfg *Config,
) OperatorAction {

	oa := &operatorAction{
		kubeCli: kubeCli,
		cfg:     cfg,
	}
	return oa
}

func (oa *operatorAction) DeployOperator(info OperatorConfig) error {
	klog.Infof("deploying chaos-mesh:%v", info.ReleaseName)
	cmd := fmt.Sprintf(`helm install %s --name %s --namespace %s --set-string %s`,
		oa.operatorChartPath(info.Tag),
		info.ReleaseName,
		info.Namespace,
		info.OperatorHelmSetString())
	klog.Info(cmd)
	res, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to deploy operator: %v, %s", err, string(res))
	}
	return nil
}

func (oa *operatorAction) InstallCRD(info OperatorConfig) error {
	klog.Infof("deploying chaos-mesh crd :%v", info.ReleaseName)
	oa.runKubectlOrDie("apply", "-f", oa.manifestPath("e2e/crd.yaml"))

	// workaround for https://github.com/kubernetes/kubernetes/issues/65517
	klog.Infof("force sync kubectl cache")
	cmdArgs := []string{"sh", "-c", "rm -rf ~/.kube/cache ~/.kube/http-cache"}
	_, err := exec.Command(cmdArgs[0], cmdArgs[1:]...).CombinedOutput()
	if err != nil {
		klog.Fatalf("Failed to run '%s': %v", strings.Join(cmdArgs, " "), err)
	}
	return nil
}

func (oi *OperatorConfig) OperatorHelmSetString() string {
	set := map[string]string{
		"controllerManager.image":           fmt.Sprintf("%s:%s", oi.Manager.Image, oi.Manager.Tag),
		"controllerManager.imagePullPolicy": oi.Manager.ImagePullPolicy,
		"chaosDaemon.image":                 fmt.Sprintf("%s:%s", oi.Daemon.Image, oi.Daemon.Tag),
		"chaosDaemon.runtime":               oi.Daemon.Runtime,
		"chaosDaemon.socketPath":            oi.Daemon.SocketPath,
	}
	arr := make([]string, 0, len(set))
	for k, v := range set {
		arr = append(arr, fmt.Sprintf("%s=%s", k, v))
	}
	return fmt.Sprintf("\"%s\"", strings.Join(arr, ","))
}

func (oa *operatorAction) operatorChartPath(tag string) string {
	return oa.chartPath(operartorChartName, tag)
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
