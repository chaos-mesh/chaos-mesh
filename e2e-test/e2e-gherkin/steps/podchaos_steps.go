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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/config"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/util"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/pkg/fixture"
)

func (tc *TestContext) RegisterPodChaosSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^a single pod named "([^"]*)" is running$`, tc.aSinglePodNamedIsRunning)
	ctx.Step(`^a "([^"]*)" chaos named "([^"]*)" with mode "([^"]*)" is applied to pods with label "([^"]*)"$`, tc.aChaosNamedWithModeIsAppliedToPodsWithLabel)
	ctx.Step(`^the pod named "([^"]*)" should eventually not be found$`, tc.thePodNamedShouldEventuallyNotBeFound)
	ctx.Step(`^a deployment named "([^"]*)" with (\d+) replicas is running$`, tc.aDeploymentNamedWithReplicasIsRunning)
	ctx.Step(`^the initial pod UIDs are recorded$`, tc.theInitialPodUIDsAreRecorded)
	ctx.Step(`^at least one pod should be replaced with a different UID$`, tc.atLeastOnePodShouldBeReplacedWithDifferentUID)
	ctx.Step(`^the chaos experiment "([^"]*)" is paused$`, tc.theChaosExperimentIsPaused)
	ctx.Step(`^no further pods should be killed within (\d+) minute$`, tc.noFurtherPodsShouldBeKilledWithinMinutes)
	ctx.Step(`^the pods with label "([^"]*)" should eventually have their container image changed to the pause image$`, tc.podsHavePauseImage)
	ctx.Step(`^the chaos experiment "([^"]*)" is deleted$`, tc.deleteChaos)
	ctx.Step(`^the pods with label "([^"]*)" should eventually recover their original container image$`, tc.podsHaveOriginalImage)
	ctx.Step(`^the chaos experiment "([^"]*)" is unpaused$`, tc.unpauseChaos)
	ctx.Step(`^a "ContainerKill" chaos named "([^"]*)" with mode "([^"]*)" targeting container "([^"]*)" is applied to pods with label "([^"]*)"$`, tc.applyContainerKill)
	ctx.Step(`^the container "([^"]*)" in pods with label "([^"]*)" should eventually be terminated$`, tc.containerTerminated)
	ctx.Step(`^the container "([^"]*)" in pods with label "([^"]*)" should eventually be running and ready$`, tc.containerRunningAndReady)
	ctx.Step(`^the container ID of container "([^"]*)" in pods with label "([^"]*)" is recorded$`, tc.recordContainerID)
	ctx.Step(`^the container ID should change$`, tc.containerIDChanged)
	ctx.Step(`^the container ID of container "([^"]*)" in pods with label "([^"]*)" is recorded again$`, tc.recordContainerID)
	ctx.Step(`^the container ID should not change within (\d+) minute$`, tc.containerIDNotChangedMinutes)
	ctx.Step(`^the container ID should not change within (\d+) seconds$`, tc.containerIDNotChangedSeconds)
}

func (tc *TestContext) aSinglePodNamedIsRunning(name string) error {
	pod := fixture.NewCommonNginxPod(name, tc.Namespace)
	if len(pod.Spec.Containers) > 0 {
		tc.OriginalImage = pod.Spec.Containers[0].Image
	}
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

func (tc *TestContext) aChaosNamedWithModeIsAppliedToPodsWithLabel(action, name, modeStr, labelKeyVal string) error {
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

	mode, err := tc.parseChaosMode(modeStr)
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
					Mode: mode,
				},
			},
		},
	}
	return tc.Client.Create(context.TODO(), podKillChaos)
}

func (tc *TestContext) thePodNamedShouldEventuallyNotBeFound(name string) error {
	return wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 5*time.Minute, true, func(ctx context.Context) (done bool, err error) {
		_, err = tc.KubeCli.CoreV1().Pods(tc.Namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil && apierrors.IsNotFound(err) {
			return true, nil
		}
		return false, nil
	})
}

func (tc *TestContext) aDeploymentNamedWithReplicasIsRunning(name string, replicas int) error {
	nd := fixture.NewCommonNginxDeployment(name, tc.Namespace, int32(replicas))
	if len(nd.Spec.Template.Spec.Containers) > 0 {
		tc.OriginalImage = nd.Spec.Template.Spec.Containers[0].Image
	}
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
	return wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 5*time.Minute, true, func(ctx context.Context) (done bool, err error) {
		newPods, err := tc.KubeCli.CoreV1().Pods(tc.Namespace).List(ctx, listOption)
		if err != nil {
			return false, nil
		}
		return !fixture.HaveSameUIDs(tc.InitialPods, newPods.Items), nil
	})
}

func (tc *TestContext) theChaosExperimentIsPaused(ctx context.Context, name string) error {
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

	err = wait.PollUntilContextTimeout(ctx, 1*time.Second, pollTimeout, true, func(innerCtx context.Context) (done bool, err error) {
		err = tc.Client.Get(innerCtx, client.ObjectKey{Namespace: tc.Namespace, Name: name}, chaos)
		if err != nil {
			// API errors here are usually transient (or caused by the poll
			// deadline being close); keep polling instead of aborting.
			return false, nil
		}
		if chaos.Status.Experiment.DesiredPhase == v1alpha1.StoppedPhase {
			return true, nil
		}
		return false, nil
	})

	if isOneShot {
		// wait.Interrupted covers every way the poll can end because of its
		// deadline: context.DeadlineExceeded, the legacy ErrWaitTimeout, and
		// wrapped forms like the client-go rate limiter error that surfaces
		// when the remaining deadline is shorter than the next request slot.
		if wait.Interrupted(err) {
			return nil
		}
		if err == nil {
			return fmt.Errorf("expected timeout since one-shot chaos shouldn't enter stopped phase, but it did")
		}
		return err
	}

	return err
}

func (tc *TestContext) noFurtherPodsShouldBeKilledWithinMinutes(ctx context.Context, duration int) error {
	listOption := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app": "nginx",
		}).String(),
	}
	pods, err := tc.KubeCli.CoreV1().Pods(tc.Namespace).List(ctx, listOption)
	if err != nil {
		return err
	}

	err = wait.PollUntilContextTimeout(ctx, 5*time.Second, time.Duration(duration)*time.Minute, true, func(innerCtx context.Context) (done bool, err error) {
		newPods, err := tc.KubeCli.CoreV1().Pods(tc.Namespace).List(innerCtx, listOption)
		if err != nil {
			return false, nil
		}
		if !fixture.HaveSameUIDs(pods.Items, newPods.Items) {
			return true, fmt.Errorf("a pod was killed during the pause period")
		}
		return false, nil
	})
	if wait.Interrupted(err) {
		return nil
	}
	if err == nil {
		return fmt.Errorf("expected no pods to be killed, but a pod was replaced")
	}
	return err
}

func (tc *TestContext) waitPodRunning(name string) error {
	return wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 5*time.Minute, true, func(ctx context.Context) (done bool, err error) {
		pod, err := tc.KubeCli.CoreV1().Pods(tc.Namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		if pod.Status.Phase != corev1.PodRunning {
			return false, nil
		}
		return true, nil
	})
}

func (tc *TestContext) podsHavePauseImage(labelKeyVal string) error {
	parts := strings.Split(labelKeyVal, "=")
	if len(parts) != 2 {
		return fmt.Errorf("invalid label selector format: %s", labelKeyVal)
	}
	listOption := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			parts[0]: parts[1],
		}).String(),
	}
	return wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 5*time.Minute, true, func(ctx context.Context) (done bool, err error) {
		pods, err := tc.KubeCli.CoreV1().Pods(tc.Namespace).List(ctx, listOption)
		if err != nil {
			return false, nil
		}
		if len(pods.Items) != 1 {
			return false, nil
		}
		pod := pods.Items[0]
		for _, c := range pod.Spec.Containers {
			if c.Image == config.TestConfig.PauseImage {
				return true, nil
			}
		}
		return false, nil
	})
}

func (tc *TestContext) deleteChaos(ctx context.Context, name string) error {
	chaos := &v1alpha1.PodChaos{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: tc.Namespace,
			Name:      name,
		},
	}
	if err := tc.Client.Delete(ctx, chaos); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}

func (tc *TestContext) podsHaveOriginalImage(labelKeyVal string) error {
	parts := strings.Split(labelKeyVal, "=")
	if len(parts) != 2 {
		return fmt.Errorf("invalid label selector format: %s", labelKeyVal)
	}
	listOption := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			parts[0]: parts[1],
		}).String(),
	}
	return wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 5*time.Minute, true, func(ctx context.Context) (done bool, err error) {
		pods, err := tc.KubeCli.CoreV1().Pods(tc.Namespace).List(ctx, listOption)
		if err != nil {
			return false, nil
		}
		if len(pods.Items) != 1 {
			return false, nil
		}
		pod := pods.Items[0]
		for _, c := range pod.Spec.Containers {
			expectedImage := "nginx:latest"
			if tc.OriginalImage != "" {
				expectedImage = tc.OriginalImage
			}
			if c.Image == expectedImage {
				return true, nil
			}
		}
		return false, nil
	})
}

func (tc *TestContext) unpauseChaos(ctx context.Context, name string) error {
	chaos := &v1alpha1.PodChaos{}
	err := tc.Client.Get(ctx, client.ObjectKey{Namespace: tc.Namespace, Name: name}, chaos)
	if err != nil {
		return err
	}
	err = util.UnPauseChaos(ctx, tc.Client, chaos)
	if err != nil {
		return err
	}

	chaosKey := types.NamespacedName{
		Namespace: tc.Namespace,
		Name:      name,
	}
	return wait.PollUntilContextTimeout(ctx, 5*time.Second, 5*time.Minute, false, func(innerCtx context.Context) (done bool, err error) {
		err = tc.Client.Get(innerCtx, chaosKey, chaos)
		if err != nil {
			return false, err
		}
		if chaos.Status.Experiment.DesiredPhase == v1alpha1.RunningPhase {
			return true, nil
		}
		return false, err
	})
}

func (tc *TestContext) applyContainerKill(name, modeStr, containerName, labelKeyVal string) error {
	parts := strings.Split(labelKeyVal, "=")
	if len(parts) != 2 {
		return fmt.Errorf("invalid label selector format: %s", labelKeyVal)
	}
	labelKey := parts[0]
	labelVal := parts[1]

	mode, err := tc.parseChaosMode(modeStr)
	if err != nil {
		return err
	}

	podChaos := &v1alpha1.PodChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: tc.Namespace,
		},
		Spec: v1alpha1.PodChaosSpec{
			Action: v1alpha1.ContainerKillAction,
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
					Mode: mode,
				},
				ContainerNames: []string{containerName},
			},
		},
	}
	return tc.Client.Create(context.TODO(), podChaos)
}

func (tc *TestContext) containerTerminated(containerName, labelKeyVal string) error {
	parts := strings.Split(labelKeyVal, "=")
	if len(parts) != 2 {
		return fmt.Errorf("invalid label selector format: %s", labelKeyVal)
	}
	listOption := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			parts[0]: parts[1],
		}).String(),
	}
	return wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 5*time.Minute, true, func(ctx context.Context) (done bool, err error) {
		pods, err := tc.KubeCli.CoreV1().Pods(tc.Namespace).List(ctx, listOption)
		if err != nil {
			return false, nil
		}
		if len(pods.Items) != 1 {
			return false, nil
		}
		pod := pods.Items[0]
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.Name == containerName && cs.LastTerminationState.Terminated != nil {
				return true, nil
			}
		}
		return false, nil
	})
}

func (tc *TestContext) containerRunningAndReady(containerName, labelKeyVal string) error {
	parts := strings.Split(labelKeyVal, "=")
	if len(parts) != 2 {
		return fmt.Errorf("invalid label selector format: %s", labelKeyVal)
	}
	listOption := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			parts[0]: parts[1],
		}).String(),
	}
	return wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 5*time.Minute, true, func(ctx context.Context) (done bool, err error) {
		pods, err := tc.KubeCli.CoreV1().Pods(tc.Namespace).List(ctx, listOption)
		if err != nil {
			return false, nil
		}
		if len(pods.Items) != 1 {
			return false, nil
		}
		pod := pods.Items[0]
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.Name == containerName && cs.Ready && cs.State.Running != nil {
				return true, nil
			}
		}
		return false, nil
	})
}

func (tc *TestContext) recordContainerID(containerName, labelKeyVal string) error {
	parts := strings.Split(labelKeyVal, "=")
	if len(parts) != 2 {
		return fmt.Errorf("invalid label selector format: %s", labelKeyVal)
	}
	listOption := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			parts[0]: parts[1],
		}).String(),
	}
	pods, err := tc.KubeCli.CoreV1().Pods(tc.Namespace).List(context.TODO(), listOption)
	if err != nil {
		return err
	}
	if len(pods.Items) != 1 {
		return fmt.Errorf("expected 1 pod, found %d", len(pods.Items))
	}
	for _, cs := range pods.Items[0].Status.ContainerStatuses {
		if cs.Name == containerName {
			tc.LastContainerID = cs.ContainerID
			return nil
		}
	}
	return fmt.Errorf("container %s not found in pod %s", containerName, pods.Items[0].Name)
}

func (tc *TestContext) containerIDChanged() error {
	listOption := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app": "nginx",
		}).String(),
	}
	return wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 5*time.Minute, true, func(ctx context.Context) (done bool, err error) {
		pods, err := tc.KubeCli.CoreV1().Pods(tc.Namespace).List(ctx, listOption)
		if err != nil {
			return false, nil
		}
		if len(pods.Items) != 1 {
			return false, nil
		}
		for _, cs := range pods.Items[0].Status.ContainerStatuses {
			if cs.Name == "nginx" {
				return tc.LastContainerID != cs.ContainerID, nil
			}
		}
		return false, nil
	})
}

func (tc *TestContext) containerIDNotChangedMinutes(duration int) error {
	listOption := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app": "nginx",
		}).String(),
	}
	err := wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, time.Duration(duration)*time.Minute, true, func(ctx context.Context) (done bool, err error) {
		pods, err := tc.KubeCli.CoreV1().Pods(tc.Namespace).List(ctx, listOption)
		if err != nil {
			return false, nil
		}
		if len(pods.Items) != 1 {
			return false, nil
		}
		for _, cs := range pods.Items[0].Status.ContainerStatuses {
			if cs.Name == "nginx" {
				if tc.LastContainerID != cs.ContainerID {
					return true, fmt.Errorf("container ID changed from %s to %s", tc.LastContainerID, cs.ContainerID)
				}
			}
		}
		return false, nil
	})
	if wait.Interrupted(err) {
		return nil
	}
	if err == nil {
		return fmt.Errorf("expected container ID not to change, but no error occurred")
	}
	return err
}

func (tc *TestContext) containerIDNotChangedSeconds(duration int) error {
	listOption := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app": "nginx",
		}).String(),
	}
	err := wait.PollUntilContextTimeout(context.TODO(), 1*time.Second, time.Duration(duration)*time.Second, true, func(ctx context.Context) (done bool, err error) {
		pods, err := tc.KubeCli.CoreV1().Pods(tc.Namespace).List(ctx, listOption)
		if err != nil {
			return false, nil
		}
		if len(pods.Items) != 1 {
			return false, nil
		}
		for _, cs := range pods.Items[0].Status.ContainerStatuses {
			if cs.Name == "nginx" {
				if tc.LastContainerID != cs.ContainerID {
					return true, fmt.Errorf("container ID changed from %s to %s", tc.LastContainerID, cs.ContainerID)
				}
			}
		}
		return false, nil
	})
	if wait.Interrupted(err) {
		return nil
	}
	if err == nil {
		return fmt.Errorf("expected container ID not to change, but no error occurred")
	}
	return err
}
