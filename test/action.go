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
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	aggregatorclientset "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"

	e2eutil "github.com/chaos-mesh/chaos-mesh/test/e2e/util"
)

const (
	operatorChartName = "chaos-mesh"
	// this hardcoded template is used for replacing images we use in E2E
	sidecarTemplate = `initContainers:
- name: inject-scripts
  image: %s
  imagePullPolicy: IfNotPresent
  command: ["sh", "-c", "/scripts/init.sh -d {{.DataPath}} -f {{.MountPath}}/fuse-data"]
containers:
- name: chaosfs
  image: %s
  imagePullPolicy: IfNotPresent
  ports:
  - containerPort: 65534
  securityContext:
    privileged: true
  command:
    - /usr/local/bin/chaosfs
    - -addr=:65534
    - -pidfile=/tmp/fuse/pid
    - -original={{.MountPath}}/fuse-data
    - -mountpoint={{.DataPath}}
  volumeMounts:
    - name: {{.VolumeName}}
      mountPath: {{.MountPath}}
      mountPropagation: Bidirectional
volumeMounts:
- name: {{.VolumeName}}
  mountPath: {{.MountPath}}
  mountPropagation: HostToContainer
- name: scripts
  mountPath: /tmp/scripts
- name: fuse
  mountPath: /tmp/fuse
volumes:
- name: scripts
  emptyDir: {}
- name: fuse
  emptyDir: {}
postStart:
  {{.ContainerName}}:
    command:
      - /tmp/scripts/wait-fuse.sh
`
)

// OperatorAction describe the common operation during test (e2e/stability/etc..)
type OperatorAction interface {
	CleanCRDOrDie()
	DeployOperator(config OperatorConfig) error
	InstallCRD(config OperatorConfig) error
	InstallTemplate(config OperatorConfig) error
}

// NewOperatorAction create an OperatorAction interface instance
func NewOperatorAction(
	kubeCli kubernetes.Interface,
	aggrCli aggregatorclientset.Interface,
	apiExtCli apiextensionsclientset.Interface,
	cfg *Config) OperatorAction {

	oa := &operatorAction{
		kubeCli:   kubeCli,
		aggrCli:   aggrCli,
		apiExtCli: apiExtCli,
		cfg:       cfg,
	}
	return oa
}

func (oa *operatorAction) DeployOperator(info OperatorConfig) error {
	klog.Infof("deploying chaos-mesh:%v", info.ReleaseName)
	cmd := fmt.Sprintf(`helm install %s --name %s --namespace %s --set-string %s`,
		oa.operatorChartPath(info.Tag),
		info.ReleaseName,
		info.Namespace,
		info.operatorHelmSetString())
	klog.Info(cmd)
	res, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to deploy operator: %v, %s", err, string(res))
	}
	klog.Infof("start to waiting chaos-mesh ready")
	err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {

		ls := &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app.kubernetes.io/instance": "chaos-mesh",
			},
		}
		l, err := metav1.LabelSelectorAsSelector(ls)
		if err != nil {
			klog.Errorf("failed to get selector, err:%v", err)
			return false, nil
		}
		pods, err := oa.kubeCli.CoreV1().Pods(info.Namespace).List(metav1.ListOptions{LabelSelector: l.String()})
		if err != nil {
			klog.Errorf("failed to get chaos-mesh pods, err:%v", err)
			return false, nil
		}
		for _, pod := range pods.Items {
			if pod.Status.Phase != corev1.PodRunning {
				return false, nil
			}
		}
		return true, nil
	})
	if err != nil {
		return err
	}
	return e2eutil.WaitForAPIServicesAvailable(oa.aggrCli, labels.Everything())
}

func (oa *operatorAction) InstallCRD(info OperatorConfig) error {
	klog.Infof("deploying chaos-mesh crd :%v", info.ReleaseName)
	oa.runKubectlOrDie("apply", "-f", oa.manifestPath("e2e/crd.yaml"))
	e2eutil.WaitForCRDsEstablished(oa.apiExtCli, labels.Everything())
	// workaround for https://github.com/kubernetes/kubernetes/issues/65517
	klog.Infof("force sync kubectl cache")
	cmdArgs := []string{"sh", "-c", "rm -rf ~/.kube/cache ~/.kube/http-cache"}
	_, err := exec.Command(cmdArgs[0], cmdArgs[1:]...).CombinedOutput()
	if err != nil {
		klog.Fatalf("Failed to run '%s': %v", strings.Join(cmdArgs, " "), err)
	}
	return nil
}

func (oa *operatorAction) InstallTemplate(info OperatorConfig) error {
	klog.Infof("deploying chaos-mesh sidecar template :%v", info.ReleaseName)
	templateCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: info.Namespace,
			Name:      "chaosfs-sidecar",
			Labels: map[string]string{
				"app.kubernetes.io/component": "template",
			},
		},
		Data: map[string]string{
			"data": fmt.Sprintf(sidecarTemplate, oa.cfg.ChaosScriptsImage, oa.cfg.ChaosFSImage),
		},
	}
	if _, err := oa.kubeCli.CoreV1().ConfigMaps(info.Namespace).
		Create(templateCM); err != nil {
		return err
	}

	return wait.PollImmediate(5*time.Second, 3*time.Minute, func() (done bool, err error) {
		list, err := oa.kubeCli.CoreV1().ConfigMaps(info.Namespace).List(
			metav1.ListOptions{LabelSelector: "app.kubernetes.io/component=template"},
		)
		if err != nil {
			return false, err
		}
		if len(list.Items) > 0 {
			return true, nil
		}
		return false, nil
	})
}

func (oa *operatorAction) CleanCRDOrDie() {
	oa.runKubectlOrDie("delete", "crds", "--all")
}
