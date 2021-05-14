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

package sidecar

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/util"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/pkg/fixture"
)

func TestcaseNoTemplate(
	ns string,
	cmNamespace string,
	cmName string,
	kubeCli kubernetes.Interface,
	cli client.Client,
) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := createInjectionConfig(ctx, cli, ns, cmName,
		map[string]string{
			"chaosfs-io": `name: chaosfs-io
selector:
  labelSelectors:
    app: io`})
	framework.ExpectNoError(err, "failed to create injection config")

	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app.kubernetes.io/component": "controller-manager",
		}).String(),
	}
	pods, err := kubeCli.CoreV1().Pods(cmNamespace).List(listOptions)
	framework.ExpectNoError(err, "get chaos mesh controller pods error")

	err = wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
		newPods, err := kubeCli.CoreV1().Pods(cmNamespace).List(listOptions)
		framework.ExpectNoError(err, "get chaos mesh controller pods error")
		if !fixture.HaveSameUIDs(pods.Items, newPods.Items) {
			return true, nil
		}
		if newPods.Items[0].Status.ContainerStatuses[0].RestartCount > 0 {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectError(err, "wait chaos mesh not dies")
	framework.ExpectEqual(err.Error(), wait.ErrWaitTimeout.Error())

	err = enableWebhook(ns)
	framework.ExpectNoError(err, "enable webhook on ns error")
	nd := fixture.NewIOTestDeployment("io-test", ns)
	_, err = kubeCli.AppsV1().Deployments(ns).Create(nd)
	framework.ExpectNoError(err, "create io-test deployment error")
	err = util.WaitDeploymentReady("io-test", ns, kubeCli)
	framework.ExpectNoError(err, "wait io-test deployment ready error")

	// cleanup

}

func TestcaseNoTemplateArgs(
	ns string,
	cmNamespace string,
	cmName string,
	kubeCli kubernetes.Interface,
	cli client.Client,
) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := createInjectionConfig(ctx, cli, ns, cmName,
		map[string]string{
			"chaosfs-io": `name: chaosfs-io
template: chaosfs-sidecar
selector:
  labelSelectors:
    app: io`})
	framework.ExpectNoError(err, "failed to create injection config")

	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app.kubernetes.io/component": "controller-manager",
		}).String(),
	}
	pods, err := kubeCli.CoreV1().Pods(cmNamespace).List(listOptions)
	framework.ExpectNoError(err, "get chaos mesh controller pods error")

	err = wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
		newPods, err := kubeCli.CoreV1().Pods(cmNamespace).List(listOptions)
		framework.ExpectNoError(err, "get chaos mesh controller pods error")
		if !fixture.HaveSameUIDs(pods.Items, newPods.Items) {
			return true, nil
		}
		if newPods.Items[0].Status.ContainerStatuses[0].RestartCount > 0 {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectError(err, "wait chaos mesh not dies")
	framework.ExpectEqual(err.Error(), wait.ErrWaitTimeout.Error())

	err = enableWebhook(ns)
	framework.ExpectNoError(err, "enable webhook on ns error")
	nd := fixture.NewIOTestDeployment("io-test", ns)
	_, err = kubeCli.AppsV1().Deployments(ns).Create(nd)
	framework.ExpectNoError(err, "create io-test deployment error")
	err = util.WaitDeploymentReady("io-test", ns, kubeCli)
	framework.ExpectNoError(err, "wait io-test deployment ready error")
}
