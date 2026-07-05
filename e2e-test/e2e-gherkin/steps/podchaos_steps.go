// Copyright 2026 Chaos Mesh Authors.
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

package steps

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/util"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/pkg/fixture"
)

func (tc *TestContext) RegisterPodChaosSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^a single pod named "([^"]*)" is running$`, tc.aSinglePodNamedIsRunning)
	ctx.Step(`^a "([^"]*)" chaos named "([^"]*)" is applied to pods with label "([^"]*)"$`, tc.aChaosNamedIsAppliedToPodsWithLabel)
	ctx.Step(`^the pod named "([^"]*)" should eventually not be found$`, tc.thePodNamedShouldEventuallyNotBeFound)
	ctx.Step(`^a deployment named "([^"]*)" with (\d+) replicas is running$`, tc.aDeploymentNamedWithReplicasIsRunning)
	ctx.Step(`^the initial pod UIDs are recorded$`, tc.theInitialPodUIDsAreRecorded)
	ctx.Step(`^at least one pod should be replaced with a different UID$`, tc.atLeastOnePodShouldBeReplacedWithDifferentUID)
	ctx.Step(`^the chaos experiment "([^"]*)" is paused$`, tc.theChaosExperimentIsPaused)
	ctx.Step(`^no further pods should be killed within (\d+) minute$`, tc.noFurtherPodsShouldBeKilledWithinMinutes)
}

func (tc *TestContext) aSinglePodNamedIsRunning(name string) error {
	pod := fixture.NewCommonNginxPod(name, tc.Namespace)
	_, err := tc.KubeCli.CoreV1().Pods(tc.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return tc.waitPodRunning(name)
}

func (tc *TestContext) parsePodChaosAction(action string) (v1alpha1.PodChaosAction, error) {
	switch strings.ToLower(action) {
	case "podkill", "pod-kill":
		return v1alpha1.PodKillAction, nil
	case "podfailure", "pod-failure":
		return v1alpha1.PodFailureAction, nil
	case "containerkill", "container-kill":
		return v1alpha1.ContainerKillAction, nil
	default:
		return "", fmt.Errorf("unsupported pod chaos action: %s", action)
	}
}

func (tc *TestContext) aChaosNamedIsAppliedToPodsWithLabel(action, name, labelKeyVal string) error {
	parts := strings.Split(labelKeyVal, "=")
	if len(parts) != 2 {
		return fmt.Errorf("invalid label selector format: %s", labelKeyVal)
	}
	labelKey := parts[0]
	labelVal := parts[1]

	normalizedAction, err := tc.parsePodChaosAction(action)
	if err != nil {
		return err
	}

	podKillChaos := &v1alpha1.PodChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: tc.Namespace,
		},
		Spec: v1alpha1.PodChaosSpec{
			Action: normalizedAction,
			ContainerSelector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Selector: v1alpha1.PodSelectorSpec{
						GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
							Namespaces: []string{tc.Namespace},
							LabelSelectors: map[string]string{
								labelKey: labelVal,
							},
						},
					},
					Mode: v1alpha1.OneMode,
				},
			},
		},
	}
	return tc.Client.Create(context.TODO(), podKillChaos)
}

func (tc *TestContext) thePodNamedShouldEventuallyNotBeFound(name string) error {
	return wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		_, err = tc.KubeCli.CoreV1().Pods(tc.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil && apierrors.IsNotFound(err) {
			return true, nil
		}
		return false, nil
	})
}

func (tc *TestContext) aDeploymentNamedWithReplicasIsRunning(name string, replicas int) error {
	nd := fixture.NewCommonNginxDeployment(name, tc.Namespace, int32(replicas))
	_, err := tc.KubeCli.AppsV1().Deployments(tc.Namespace).Create(context.TODO(), nd, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return util.WaitDeploymentReady(name, tc.Namespace, tc.KubeCli)
}

func (tc *TestContext) theInitialPodUIDsAreRecorded() error {
	listOption := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app": "nginx",
		}).String(),
	}
	pods, err := tc.KubeCli.CoreV1().Pods(tc.Namespace).List(context.TODO(), listOption)
	if err != nil {
		return err
	}
	tc.InitialPods = pods.Items
	return nil
}

func (tc *TestContext) atLeastOnePodShouldBeReplacedWithDifferentUID() error {
	listOption := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app": "nginx",
		}).String(),
	}
	return wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		newPods, err := tc.KubeCli.CoreV1().Pods(tc.Namespace).List(context.TODO(), listOption)
		if err != nil {
			return false, nil
		}
		return !fixture.HaveSameUIDs(tc.InitialPods, newPods.Items), nil
	})
}

func (tc *TestContext) theChaosExperimentIsPaused(name string) error {
	ctx := context.TODO()
	chaos := &v1alpha1.PodChaos{}
	err := tc.Client.Get(ctx, client.ObjectKey{Namespace: tc.Namespace, Name: name}, chaos)
	if err != nil {
		return err
	}
	err = util.PauseChaos(ctx, tc.Client, chaos)
	if err != nil {
		return err
	}

	isOneShot := chaos.Spec.Action == v1alpha1.PodKillAction || chaos.Spec.Action == v1alpha1.ContainerKillAction

	var pollTimeout time.Duration
	if isOneShot {
		pollTimeout = 5 * time.Second
	} else {
		pollTimeout = 5 * time.Minute
	}

	err = wait.Poll(1*time.Second, pollTimeout, func() (done bool, err error) {
		err = tc.Client.Get(ctx, client.ObjectKey{Namespace: tc.Namespace, Name: name}, chaos)
		if err != nil {
			return false, err
		}
		if chaos.Status.Experiment.DesiredPhase == v1alpha1.StoppedPhase {
			return true, nil
		}
		return false, nil
	})

	if isOneShot {
		if err == wait.ErrWaitTimeout {
			return nil
		}
		if err == nil {
			return fmt.Errorf("expected timeout since one-shot chaos shouldn't enter stopped phase, but it did")
		}
		return err
	}

	return err
}

func (tc *TestContext) noFurtherPodsShouldBeKilledWithinMinutes(duration int) error {
	listOption := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app": "nginx",
		}).String(),
	}
	pods, err := tc.KubeCli.CoreV1().Pods(tc.Namespace).List(context.TODO(), listOption)
	if err != nil {
		return err
	}

	err = wait.Poll(5*time.Second, time.Duration(duration)*time.Minute, func() (done bool, err error) {
		newPods, err := tc.KubeCli.CoreV1().Pods(tc.Namespace).List(context.TODO(), listOption)
		if err != nil {
			return false, nil
		}
		if !fixture.HaveSameUIDs(pods.Items, newPods.Items) {
			return true, fmt.Errorf("a pod was killed during the pause period")
		}
		return false, nil
	})
	if err == wait.ErrWaitTimeout {
		return nil
	}
	if err == nil {
		return fmt.Errorf("expected no pods to be killed, but a pod was replaced")
	}
	return err
}

func (tc *TestContext) waitPodRunning(name string) error {
	return wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		pod, err := tc.KubeCli.CoreV1().Pods(tc.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		if pod.Status.Phase != corev1.PodRunning {
			return false, nil
		}
		return true, nil
	})
}
