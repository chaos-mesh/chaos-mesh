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

package test

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	aggregatorclientset "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
	"k8s.io/kubernetes/test/e2e/framework"

	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/e2econst"
	e2eutil "github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/util"
)

const (
	operatorChartName = "chaos-mesh"
)

// OperatorAction describe the common operation during test (e2e/stability/etc..)
type OperatorAction interface {
	CleanCRDOrDie()
	DeployOperator(config *OperatorConfig) error
	UpgradeOperator(config *OperatorConfig) error
	RestartDaemon(info *OperatorConfig) error
	RestartControllerManager(info *OperatorConfig) error
	InstallCRD(config *OperatorConfig) error
}

// BuildOperatorAction build an OperatorAction interface instance or fail
func BuildOperatorActionAndCfg(cfg *Config) (OperatorAction, *OperatorConfig, error) {
	// Get clients
	config, err := framework.LoadConfig()
	if err != nil {
		return nil, nil, errors.Wrap(err, "load config")
	}
	kubeCli, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, errors.Wrap(err, "create kube client")
	}
	aggrCli, err := aggregatorclientset.NewForConfig(config)
	if err != nil {
		return nil, nil, errors.Wrap(err, "create aggr client")
	}
	apiExtCli, err := apiextensionsclientset.NewForConfig(config)
	if err != nil {
		return nil, nil, errors.Wrap(err, "create apiExt clientset")
	}
	oa := NewOperatorAction(kubeCli, aggrCli, apiExtCli, cfg)
	ocfg := NewDefaultOperatorConfig()
	ocfg.Manager.ImageRegistry = cfg.ManagerImageRegistry
	ocfg.Manager.ImageRepository = cfg.ManagerImage
	ocfg.Manager.ImageTag = cfg.ManagerTag
	ocfg.Daemon.ImageRegistry = cfg.DaemonImageRegistry
	ocfg.Daemon.ImageRepository = cfg.DaemonImage
	ocfg.Daemon.ImageTag = cfg.DaemonTag
	ocfg.DNSImage = cfg.ChaosCoreDNSImage
	ocfg.EnableDashboard = cfg.EnableDashboard

	return oa, &ocfg, nil
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

func (oa *operatorAction) DeployOperator(info *OperatorConfig) error {
	klog.Infof("create namespace chaos-mesh")
	cmd := fmt.Sprintf(`kubectl create ns %s`, e2econst.ChaosMeshNamespace)
	klog.Infof(cmd)
	output, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return errors.Errorf("failed to create namespace chaos-mesh: %v %s", err, string(output))
	}
	return oa.UpgradeOperator(info)
}

func (oa *operatorAction) UpgradeOperator(info *OperatorConfig) error {
	klog.Infof("deploying chaos-mesh:%v", info.ReleaseName)
	cmd := fmt.Sprintf(`helm upgrade --install %s %s --namespace %s --set %s --skip-crds`,
		info.ReleaseName,
		oa.operatorChartPath(info.Tag),
		info.Namespace,
		info.operatorHelmSetValue())
	klog.Info(cmd)
	res, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return errors.Errorf("failed to deploy operator: %v, %s", err, string(res))
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
		pods, err := oa.kubeCli.CoreV1().Pods(info.Namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: l.String()})
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

func (oa *operatorAction) InstallCRD(info *OperatorConfig) error {
	klog.Infof("deploying chaos-mesh crd :%v", info.ReleaseName)
	oa.runKubectlOrDie("create", "-f", oa.manifestPath("e2e/crd.yaml"), "--validate=false")

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

func (oa *operatorAction) RestartDaemon(info *OperatorConfig) error {
	return oa.restartComponent(info, "chaos-daemon-")
}

func (oa *operatorAction) RestartControllerManager(info *OperatorConfig) error {
	return oa.restartComponent(info, "chaos-controller-manager-")
}

func (oa *operatorAction) restartComponent(info *OperatorConfig, prefix string) error {
	klog.Infof("klling component %v", prefix)
	ls := &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app.kubernetes.io/instance": "chaos-mesh",
		},
	}
	l, err := metav1.LabelSelectorAsSelector(ls)
	if err != nil {
		return errors.Wrap(err, "get selector")
	}

	pods, err := oa.kubeCli.CoreV1().Pods(info.Namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: l.String()})
	if err != nil {
		return errors.Wrap(err, "select pods")
	}

	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, prefix) {
			err = oa.kubeCli.CoreV1().Pods(info.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
			if err != nil {
				return errors.Wrapf(err, "delete pod(%s)", pod.Name)
			}
		}
	}

	klog.Infof("start to waiting chaos-mesh ready")
	err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		pods, err := oa.kubeCli.CoreV1().Pods(info.Namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: l.String()})
		if err != nil {
			klog.Errorf("get chaos-mesh pods: %v", err)
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

func (oa *operatorAction) CleanCRDOrDie() {
	oa.runKubectlOrDie("delete", "crds", "--all")
}
