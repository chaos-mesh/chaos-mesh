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

package podchaos

import (
	"context"
	"time"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/util"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/pkg/fixture"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestcasePodKillOnceThenDelete(ns string, kubeCli kubernetes.Interface, cli client.Client) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pod := fixture.NewCommonNginxPod("nginx", ns)
	_, err := kubeCli.CoreV1().Pods(ns).Create(pod)
	framework.ExpectNoError(err, "create nginx pod error")
	err = waitPodRunning("nginx", ns, kubeCli)
	framework.ExpectNoError(err, "wait nginx running error")

	podKillChaos := &v1alpha1.PodChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-kill",
			Namespace: ns,
		},
		Spec: v1alpha1.PodChaosSpec{
			Action: v1alpha1.PodKillAction,
			ContainerSelector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Selector: v1alpha1.PodSelectorSpec{
						Namespaces: []string{
							ns,
						},
						LabelSelectors: map[string]string{
							"app": "nginx",
						},
					},
					Mode: v1alpha1.OnePodMode,
				},
			},
		},
	}
	err = cli.Create(ctx, podKillChaos)
	framework.ExpectNoError(err, "create pod chaos error")

	err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		_, err = kubeCli.CoreV1().Pods(ns).Get("nginx", metav1.GetOptions{})
		if err != nil && apierrors.IsNotFound(err) {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "Pod kill chaos perform failed")

}
func TestcasePodKillPauseThenUnPause(ns string, kubeCli kubernetes.Interface, cli client.Client) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nd := fixture.NewCommonNginxDeployment("nginx", ns, 3)
	_, err := kubeCli.AppsV1().Deployments(ns).Create(nd)
	framework.ExpectNoError(err, "create nginx deployment error")
	err = util.WaitDeploymentReady("nginx", ns, kubeCli)
	framework.ExpectNoError(err, "wait nginx deployment ready error")

	var pods *corev1.PodList
	var newPods *corev1.PodList
	listOption := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app": "nginx",
		}).String(),
	}
	pods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
	framework.ExpectNoError(err, "get nginx pods error")

	podKillChaos := &v1alpha1.PodChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-kill",
			Namespace: ns,
		},
		Spec: v1alpha1.PodChaosSpec{
			Action:   v1alpha1.PodKillAction,
			Duration: pointer.StringPtr("9m"),
			ContainerSelector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Selector: v1alpha1.PodSelectorSpec{
						Namespaces: []string{
							ns,
						},
						LabelSelectors: map[string]string{
							"app": "nginx",
						},
					},
					Mode: v1alpha1.OnePodMode,
				},
			},
		},
	}
	err = cli.Create(ctx, podKillChaos)
	framework.ExpectNoError(err, "create pod chaos error")

	chaosKey := types.NamespacedName{
		Namespace: ns,
		Name:      "nginx-kill",
	}

	// some pod is killed as expected
	err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		newPods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
		framework.ExpectNoError(err, "get nginx pods error")
		return !fixture.HaveSameUIDs(pods.Items, newPods.Items), nil
	})
	framework.ExpectNoError(err, "wait pod killed failed")

	// pause experiment
	err = util.PauseChaos(ctx, cli, podKillChaos)
	framework.ExpectNoError(err, "pause chaos error")

	err = wait.Poll(1*time.Second, 5*time.Second, func() (done bool, err error) {
		chaos := &v1alpha1.PodChaos{}
		err = cli.Get(ctx, chaosKey, chaos)
		framework.ExpectNoError(err, "get pod chaos error")
		if chaos.Status.Experiment.DesiredPhase == v1alpha1.StoppedPhase {
			return true, nil
		}
		return false, err
	})
	framework.ExpectError(err, "chaos shouldn't enter stopped phase")

	// wait for 1 minutes and no pod is killed
	pods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
	framework.ExpectNoError(err, "get nginx pods error")
	err = wait.Poll(5*time.Second, 1*time.Minute, func() (done bool, err error) {
		newPods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
		framework.ExpectNoError(err, "get nginx pods error")
		return !fixture.HaveSameUIDs(pods.Items, newPods.Items), nil
	})
	framework.ExpectError(err, "wait pod not killed failed")
	framework.ExpectEqual(err.Error(), wait.ErrWaitTimeout.Error())

}
