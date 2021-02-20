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
	"github.com/chaos-mesh/chaos-mesh/test/e2e/util"
	"github.com/chaos-mesh/chaos-mesh/test/pkg/fixture"
)

func TestcaseContainerKillOnceThenDelete(ns string, kubeCli kubernetes.Interface, cli client.Client) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nd := fixture.NewCommonNginxDeployment("nginx", ns, 1)
	_, err := kubeCli.AppsV1().Deployments(ns).Create(nd)
	framework.ExpectNoError(err, "create nginx deployment error")
	err = util.WaitDeploymentReady("nginx", ns, kubeCli)
	framework.ExpectNoError(err, "wait nginx deployment ready error")

	containerKillChaos := &v1alpha1.PodChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-container-kill",
			Namespace: ns,
		},
		Spec: v1alpha1.PodChaosSpec{
			Selector: v1alpha1.SelectorSpec{
				Namespaces: []string{
					ns,
				},
				LabelSelectors: map[string]string{
					"app": "nginx",
				},
			},
			Action:        v1alpha1.ContainerKillAction,
			Mode:          v1alpha1.OnePodMode,
			ContainerName: "nginx",
			Scheduler: &v1alpha1.SchedulerSpec{
				Cron: "@every 10s",
			},
		},
	}
	err = cli.Create(ctx, containerKillChaos)
	framework.ExpectNoError(err, "create container kill chaos error")

	err = wait.PollImmediate(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		listOption := metav1.ListOptions{
			LabelSelector: labels.SelectorFromSet(map[string]string{
				"app": "nginx",
			}).String(),
		}
		pods, err := kubeCli.CoreV1().Pods(ns).List(listOption)
		if err != nil {
			return false, nil
		}
		if len(pods.Items) != 1 {
			return false, nil
		}
		pod := pods.Items[0]
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.Name == "nginx" && !cs.Ready && cs.LastTerminationState.Terminated != nil {
				return true, nil
			}
		}
		return false, nil
	})
	framework.ExpectNoError(err, "container kill apply failed")

	err = cli.Delete(ctx, containerKillChaos)
	framework.ExpectNoError(err, "failed to delete container kill chaos")

	By("success to perform container kill")
	err = wait.PollImmediate(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		listOption := metav1.ListOptions{
			LabelSelector: labels.SelectorFromSet(map[string]string{
				"app": "nginx",
			}).String(),
		}
		pods, err := kubeCli.CoreV1().Pods(ns).List(listOption)
		if err != nil {
			return false, nil
		}
		if len(pods.Items) != 1 {
			return false, nil
		}
		pod := pods.Items[0]
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.Name == "nginx" && cs.Ready && cs.State.Running != nil {
				return true, nil
			}
		}
		return false, nil
	})
	framework.ExpectNoError(err, "container kill recover failed")

}

func TestcaseContainerKillPauseThenUnPause(ns string, kubeCli kubernetes.Interface, cli client.Client) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	nd := fixture.NewCommonNginxDeployment("nginx", ns, 1)
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

	// Get the running nginx container ID
	containerID := pods.Items[0].Status.ContainerStatuses[0].ContainerID

	containerKillChaos := &v1alpha1.PodChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-container-kill",
			Namespace: ns,
		},
		Spec: v1alpha1.PodChaosSpec{
			Selector: v1alpha1.SelectorSpec{
				Namespaces: []string{
					ns,
				},
				LabelSelectors: map[string]string{
					"app": "nginx",
				},
			},
			Action:        v1alpha1.ContainerKillAction,
			Mode:          v1alpha1.OnePodMode,
			ContainerName: "nginx",
			Duration:      pointer.StringPtr("9m"),
			Scheduler: &v1alpha1.SchedulerSpec{
				Cron: "@every 10m",
			},
		},
	}
	err = cli.Create(ctx, containerKillChaos)
	framework.ExpectNoError(err, "create container kill chaos error")

	chaosKey := types.NamespacedName{
		Namespace: ns,
		Name:      "nginx-container-kill",
	}

	// nginx container is killed as expected
	err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		newPods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
		framework.ExpectNoError(err, "get nginx pods error")
		return containerID != newPods.Items[0].Status.ContainerStatuses[0].ContainerID, nil
	})
	framework.ExpectNoError(err, "wait container kill failed")

	// pause experiment
	err = util.PauseChaos(ctx, cli, containerKillChaos)
	framework.ExpectNoError(err, "pause chaos error")

	err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		chaos := &v1alpha1.PodChaos{}
		err = cli.Get(ctx, chaosKey, chaos)
		framework.ExpectNoError(err, "get pod chaos error")
		if chaos.Status.Experiment.Phase == v1alpha1.ExperimentPhasePaused {
			return true, nil
		}
		return false, err
	})
	framework.ExpectNoError(err, "check paused chaos failed")

	// wait for 1 minutes and check whether nginx container will be killed or not
	pods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
	framework.ExpectNoError(err, "get nginx pods error")
	containerID = pods.Items[0].Status.ContainerStatuses[0].ContainerID
	err = wait.Poll(5*time.Second, 1*time.Minute, func() (done bool, err error) {
		newPods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
		framework.ExpectNoError(err, "get nginx pods error")
		return containerID != newPods.Items[0].Status.ContainerStatuses[0].ContainerID, nil
	})
	framework.ExpectError(err, "wait container not killed failed")
	framework.ExpectEqual(err.Error(), wait.ErrWaitTimeout.Error())

	// resume experiment
	err = util.UnPauseChaos(ctx, cli, containerKillChaos)
	framework.ExpectNoError(err, "resume chaos error")

	err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		chaos := &v1alpha1.PodChaos{}
		err = cli.Get(ctx, chaosKey, chaos)
		framework.ExpectNoError(err, "get pod chaos error")
		if chaos.Status.Experiment.Phase == v1alpha1.ExperimentPhaseRunning {
			return true, nil
		}
		return false, err
	})
	framework.ExpectNoError(err, "check resumed chaos failed")

	// nginx container is killed by resumed experiment
	pods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
	framework.ExpectNoError(err, "get nginx pods error")
	containerID = pods.Items[0].Status.ContainerStatuses[0].ContainerID
	err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		newPods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
		framework.ExpectNoError(err, "get nginx pods error")
		return containerID != newPods.Items[0].Status.ContainerStatuses[0].ContainerID, nil
	})
	framework.ExpectNoError(err, "wait container killed failed")

}
