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

	. "github.com/onsi/ginkgo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/e2econst"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/util"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/pkg/fixture"
)

func TestcasePodFailureOnceThenDelete(ns string, kubeCli kubernetes.Interface, cli client.Client) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	By("preparing experiment pods")
	appName := "timer-pod-failure1"
	nd := fixture.NewTimerDeployment(appName, ns)
	_, err := kubeCli.AppsV1().Deployments(ns).Create(nd)
	framework.ExpectNoError(err, "create timer deployment error")
	err = util.WaitDeploymentReady(appName, ns, kubeCli)
	framework.ExpectNoError(err, "wait timer deployment ready error")

	By("create pod failure chaos CRD objects")
	listOption := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app": appName,
		}).String(),
	}
	podFailureChaos := &v1alpha1.PodChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "timer-failure1",
			Namespace: ns,
		},
		Spec: v1alpha1.PodChaosSpec{
			Action: v1alpha1.PodFailureAction,
			ContainerSelector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Selector: v1alpha1.PodSelectorSpec{
						Namespaces: []string{
							ns,
						},
						LabelSelectors: map[string]string{
							"app": appName,
						},
					},
					Mode: v1alpha1.OnePodMode,
				},
			},
		},
	}

	err = cli.Create(ctx, podFailureChaos)
	framework.ExpectNoError(err, "create pod failure chaos error")

	By("waiting for assertion some pod fall into failure")
	err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		pods, err := kubeCli.CoreV1().Pods(ns).List(listOption)
		if err != nil {
			return false, nil
		}
		if len(pods.Items) != 1 {
			return false, nil
		}
		pod := pods.Items[0]
		for _, c := range pod.Spec.Containers {
			if c.Image == e2econst.PauseImage {
				return true, nil
			}
		}
		return false, nil
	})
	framework.ExpectNoError(err, "failed to verify PodFailure")

	By("delete pod failure chaos CRD objects")
	err = cli.Delete(ctx, podFailureChaos)
	framework.ExpectNoError(err, "failed to delete pod failure chaos")

	By("waiting for assertion recovering")
	err = wait.Poll(5*time.Second, 2*time.Minute, func() (done bool, err error) {
		pods, err := kubeCli.CoreV1().Pods(ns).List(listOption)
		if err != nil {
			return false, nil
		}
		if len(pods.Items) != 1 {
			return false, nil
		}
		pod := pods.Items[0]
		for _, c := range pod.Spec.Containers {
			if c.Image == nd.Spec.Template.Spec.Containers[0].Image {
				return true, nil
			}
		}
		return false, nil
	})
	framework.ExpectNoError(err, "pod failure recover failed")
}

func TestcasePodFailurePauseThenUnPause(ns string, kubeCli kubernetes.Interface, cli client.Client) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	By("preparing experiment pods")
	appName := "timer-pod-failure2"
	nd := fixture.NewTimerDeployment(appName, ns)
	_, err := kubeCli.AppsV1().Deployments(ns).Create(nd)
	framework.ExpectNoError(err, "create timer deployment error")
	err = util.WaitDeploymentReady(appName, ns, kubeCli)
	framework.ExpectNoError(err, "wait timer deployment ready error")

	By("create pod failure chaos CRD objects")
	var pods *corev1.PodList
	listOption := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app": appName,
		}).String(),
	}

	podFailureChaos := &v1alpha1.PodChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "timer-failure2",
			Namespace: ns,
		},
		Spec: v1alpha1.PodChaosSpec{
			Action:   v1alpha1.PodFailureAction,
			Duration: pointer.StringPtr("9m"),
			ContainerSelector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Selector: v1alpha1.PodSelectorSpec{
						Namespaces: []string{
							ns,
						},
						LabelSelectors: map[string]string{
							"app": appName,
						},
					},
					Mode: v1alpha1.OnePodMode,
				},
			},
		},
	}
	err = cli.Create(ctx, podFailureChaos)
	framework.ExpectNoError(err, "create pod failure chaos error")
	chaosKey := types.NamespacedName{
		Namespace: ns,
		Name:      "timer-failure2",
	}

	By("waiting for assertion some pod fall into failure")
	// check whether the pod failure chaos succeeded or not
	err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		pods, err := kubeCli.CoreV1().Pods(ns).List(listOption)
		if err != nil {
			return false, nil
		}
		pod := pods.Items[0]
		for _, c := range pod.Spec.Containers {
			if c.Image == e2econst.PauseImage {
				return true, nil
			}
		}
		return false, nil
	})
	framework.ExpectNoError(err, "image not update to pause")

	// pause experiment
	By("pause pod failure chaos")
	err = util.PauseChaos(ctx, cli, podFailureChaos)
	framework.ExpectNoError(err, "pause chaos error")

	By("waiting for assertion about chaos experiment paused")
	err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		chaos := &v1alpha1.PodChaos{}
		err = cli.Get(ctx, chaosKey, chaos)
		framework.ExpectNoError(err, "get pod chaos error")
		if chaos.Status.Experiment.DesiredPhase == v1alpha1.StoppedPhase {
			return true, nil
		}
		return false, err
	})
	framework.ExpectNoError(err, "check paused chaos failed")

	By("wait for 30 seconds and no pod failure")
	pods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
	framework.ExpectNoError(err, "get timer pod error")
	err = wait.Poll(5*time.Second, 30*time.Second, func() (done bool, err error) {
		pods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
		framework.ExpectNoError(err, "get timer pod error")
		pod := pods.Items[0]
		for _, c := range pod.Spec.Containers {
			if c.Image == e2econst.PauseImage {
				return false, nil
			}
		}

		return true, nil
	})
	framework.ExpectNoError(err, "check paused chaos failed")

	By("resume paused chaos experiment")
	err = util.UnPauseChaos(ctx, cli, podFailureChaos)
	framework.ExpectNoError(err, "resume chaos error")

	By("waiting for assertion about pod failure happens again")
	err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		chaos := &v1alpha1.PodChaos{}
		err = cli.Get(ctx, chaosKey, chaos)
		framework.ExpectNoError(err, "get pod chaos error")
		if chaos.Status.Experiment.DesiredPhase == v1alpha1.RunningPhase {
			return true, nil
		}
		return false, err
	})
	framework.ExpectNoError(err, "check resumed chaos failed")

	By("waiting for assert pod failure happens again")
	err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		pods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
		framework.ExpectNoError(err, "get timer pod error")
		pod := pods.Items[0]
		for _, c := range pod.Spec.Containers {
			if c.Image == e2econst.PauseImage {
				return true, nil
			}
		}
		return false, nil
	})
	framework.ExpectNoError(err, "wait pod failure failed")
}
