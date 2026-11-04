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

func (tc *TestContext) aSinglePodNamedIsRunning(ctx context.Context, name string) error {
	pod := fixture.NewCommonNginxPod(name, tc.Namespace)
	_, err := tc.KubeCli.CoreV1().Pods(tc.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return tc.waitPodRunning(ctx, name)
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

func (tc *TestContext) aChaosNamedWithModeIsAppliedToPodsWithLabel(action, name, mode, selector string) error {
	normalizedAction, err := tc.parsePodChaosAction(action)
	if err != nil {
		return err
	}
	chaos, err := tc.newPodChaos(name, mode, selector, normalizedAction, nil)
	if err != nil {
		return err
	}
	return tc.Client.Create(context.TODO(), chaos)
}

func (tc *TestContext) newPodChaos(name, modeStr, selector string, action v1alpha1.PodChaosAction, containers []string) (*v1alpha1.PodChaos, error) {
	selectorLabels, err := parseLabelSelector(selector)
	if err != nil {
		return nil, err
	}
	mode, err := tc.parseChaosMode(modeStr)
	if err != nil {
		return nil, err
	}
	return &v1alpha1.PodChaos{
		Name: name, Namespace: tc.Namespace,
		Spec: v1alpha1.PodChaosSpec{
			Action: action,
			ContainerSelector: v1alpha1.ContainerSelector{
				ContainerNames: containers,
				PodSelector: v1alpha1.PodSelector{
					Mode: mode,
					Selector: v1alpha1.PodSelectorSpec{GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
						Namespaces: []string{tc.Namespace}, LabelSelectors: selectorLabels,
					}},
				},
			},
		},
	}, nil
}

func (tc *TestContext) thePodNamedShouldEventuallyNotBeFound(ctx context.Context, name string) error {
	return wait.PollUntilContextTimeout(ctx, 5*time.Second, 5*time.Minute, true, func(ctx context.Context) (done bool, err error) {
		_, err = tc.KubeCli.CoreV1().Pods(tc.Namespace).Get(ctx, name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
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

func (tc *TestContext) theInitialPodUIDsAreRecorded(ctx context.Context) error {
	pods, err := tc.podsWithLabel(ctx, "app=nginx")
	if err != nil {
		return err
	}
	if len(pods.Items) == 0 {
		return fmt.Errorf("no pods to record")
	}
	tc.InitialPods = pods.Items
	return nil
}

func (tc *TestContext) atLeastOnePodShouldBeReplacedWithDifferentUID(ctx context.Context) error {
	if len(tc.InitialPods) == 0 {
		return fmt.Errorf("initial pod UIDs have not been recorded")
	}
	return wait.PollUntilContextTimeout(ctx, 5*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
		pods, err := tc.podsWithLabel(ctx, "app=nginx")
		if err != nil {
			return false, nil
		}
		for _, pod := range pods.Items {
			found := false
			for _, initial := range tc.InitialPods {
				if pod.UID == initial.UID {
					found = true
					break
				}
			}
			if !found {
				return true, nil
			}
		}
		return false, nil
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
	if !isOneShot {
		return wait.PollUntilContextTimeout(ctx, time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
			if err := tc.Client.Get(ctx, client.ObjectKeyFromObject(chaos), chaos); err != nil {
				return false, nil
			}
			return chaos.Status.Experiment.DesiredPhase == v1alpha1.StoppedPhase, nil
		})
	}
	observation, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	deadline, _ := observation.Deadline()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := tc.Client.Get(observation, client.ObjectKeyFromObject(chaos), chaos); err != nil {
			return err
		}
		if chaos.Status.Experiment.DesiredPhase == v1alpha1.StoppedPhase {
			return fmt.Errorf("one-shot chaos unexpectedly entered stopped phase")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-observation.Done():
			return ctx.Err()
		case <-ticker.C:
			if observation.Err() != nil || !time.Now().Before(deadline) {
				if parentDeadline, ok := ctx.Deadline(); ok && !time.Now().Before(parentDeadline) && ctx.Err() == nil {
					return context.DeadlineExceeded
				}
				return ctx.Err()
			}
		}
	}
}

func (tc *TestContext) noFurtherPodsShouldBeKilledWithinMinutes(ctx context.Context, duration int) error {
	observation, cancel := context.WithTimeout(ctx, time.Duration(duration)*time.Minute)
	defer cancel()
	deadline, _ := observation.Deadline()
	pods, err := tc.podsWithLabel(observation, "app=nginx")
	if err != nil {
		return err
	}
	if len(pods.Items) == 0 {
		return fmt.Errorf("no pods to observe")
	}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		current, err := tc.podsWithLabel(observation, "app=nginx")
		if err != nil {
			return err
		}
		if !fixture.HaveSameUIDs(pods.Items, current.Items) {
			return fmt.Errorf("a pod was replaced during the pause period")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-observation.Done():
			return ctx.Err()
		case <-ticker.C:
			if observation.Err() != nil || !time.Now().Before(deadline) {
				if parentDeadline, ok := ctx.Deadline(); ok && !time.Now().Before(parentDeadline) && ctx.Err() == nil {
					return context.DeadlineExceeded
				}
				return ctx.Err()
			}
		}
	}
}

func (tc *TestContext) waitPodRunning(ctx context.Context, name string) error {
	return wait.PollUntilContextTimeout(ctx, 5*time.Second, 5*time.Minute, true, func(ctx context.Context) (done bool, err error) {
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

func (tc *TestContext) podsHavePauseImage(ctx context.Context, selector string) error {
	return tc.waitPodsImage(ctx, selector, config.TestConfig.PauseImage)
}

func (tc *TestContext) waitPodsImage(ctx context.Context, selector, image string) error {
	if image == "" {
		return fmt.Errorf("expected container image has not been recorded")
	}
	return wait.PollUntilContextTimeout(ctx, 5*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
		pods, err := tc.podsWithLabel(ctx, selector)
		if err != nil {
			return false, nil
		}
		if len(pods.Items) != 1 {
			return false, nil
		}
		for _, container := range pods.Items[0].Spec.Containers {
			if container.Image == image {
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

func (tc *TestContext) podsHaveOriginalImage(ctx context.Context, selector string) error {
	return tc.waitPodsImage(ctx, selector, tc.OriginalImage)
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

func (tc *TestContext) applyContainerKill(name, mode, containerName, selector string) error {
	chaos, err := tc.newPodChaos(name, mode, selector, v1alpha1.ContainerKillAction, []string{containerName})
	if err != nil {
		return err
	}
	return tc.Client.Create(context.TODO(), chaos)
}

type recordedContainer struct {
	Selector      string
	PodName       string
	PodUID        types.UID
	ContainerName string
	ContainerID   string
	RestartCount  int32
}

func (tc *TestContext) recordContainerID(ctx context.Context, containerName, selector string) error {
	tc.ContainerBaseline = nil
	pods, err := tc.podsWithLabel(ctx, selector)
	if err != nil {
		return err
	}
	if len(pods.Items) != 1 {
		return fmt.Errorf("expected 1 pod, found %d", len(pods.Items))
	}
	pod := &pods.Items[0]
	if pod.UID == "" {
		return fmt.Errorf("pod %s has no UID", pod.Name)
	}
	for _, status := range pod.Status.ContainerStatuses {
		if status.Name != containerName {
			continue
		}
		if status.ContainerID == "" {
			return fmt.Errorf("container %s has empty container ID", containerName)
		}
		if status.State.Running == nil {
			return fmt.Errorf("container %s is not running", containerName)
		}
		tc.ContainerBaseline = &recordedContainer{Selector: selector, PodName: pod.Name, PodUID: pod.UID, ContainerName: containerName, ContainerID: status.ContainerID, RestartCount: status.RestartCount}
		return nil
	}
	return fmt.Errorf("container %s not found in pod %s", containerName, pod.Name)
}

func (tc *TestContext) recordedContainerStatus(ctx context.Context) (*corev1.ContainerStatus, error) {
	baseline := tc.ContainerBaseline
	if baseline == nil {
		return nil, fmt.Errorf("container baseline has not been recorded")
	}
	pods, err := tc.podsWithLabel(ctx, baseline.Selector)
	if err != nil {
		return nil, err
	}
	return baseline.status(pods)
}

func (baseline *recordedContainer) status(pods *corev1.PodList) (*corev1.ContainerStatus, error) {
	if len(pods.Items) != 1 {
		return nil, fmt.Errorf("expected 1 pod, found %d", len(pods.Items))
	}
	pod := &pods.Items[0]
	if pod.Name != baseline.PodName || pod.UID != baseline.PodUID {
		return nil, fmt.Errorf("recorded pod %s was replaced", baseline.PodName)
	}
	for i := range pod.Status.ContainerStatuses {
		if pod.Status.ContainerStatuses[i].Name == baseline.ContainerName {
			return &pod.Status.ContainerStatuses[i], nil
		}
	}
	return nil, fmt.Errorf("container %s not found in pod %s", baseline.ContainerName, pod.Name)
}

func (tc *TestContext) checkRecordedTarget(containerName, selector string) error {
	if tc.ContainerBaseline == nil {
		return fmt.Errorf("container baseline has not been recorded")
	}
	if tc.ContainerBaseline.ContainerName != containerName || tc.ContainerBaseline.Selector != selector {
		return fmt.Errorf("container target differs from the recorded baseline")
	}
	return nil
}

func (tc *TestContext) containerTerminated(ctx context.Context, containerName, selector string) error {
	if err := tc.checkRecordedTarget(containerName, selector); err != nil {
		return err
	}
	return tc.containerIDChanged(ctx)
}

func (tc *TestContext) containerRunningAndReady(ctx context.Context, containerName, selector string) error {
	if err := tc.checkRecordedTarget(containerName, selector); err != nil {
		return err
	}
	return wait.PollUntilContextTimeout(ctx, 5*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
		pods, err := tc.podsWithLabel(ctx, tc.ContainerBaseline.Selector)
		if err != nil {
			return false, nil
		}
		status, err := tc.ContainerBaseline.status(pods)
		if err != nil {
			return false, err
		}
		return tc.ContainerBaseline.restarted(status) && status.Ready && status.State.Running != nil, nil
	})
}

func (baseline *recordedContainer) restarted(status *corev1.ContainerStatus) bool {
	return status.ContainerID != "" && status.ContainerID != baseline.ContainerID && status.RestartCount > baseline.RestartCount
}

func (tc *TestContext) containerIDChanged(ctx context.Context) error {
	if tc.ContainerBaseline == nil {
		return fmt.Errorf("container baseline has not been recorded")
	}
	return wait.PollUntilContextTimeout(ctx, 5*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
		pods, err := tc.podsWithLabel(ctx, tc.ContainerBaseline.Selector)
		if err != nil {
			return false, nil
		}
		status, err := tc.ContainerBaseline.status(pods)
		if err != nil {
			return false, err
		}
		return tc.ContainerBaseline.restarted(status), nil
	})
}

func (tc *TestContext) containerIDNotChanged(ctx context.Context, interval, duration time.Duration) error {
	observation, cancel := context.WithTimeout(ctx, duration)
	defer cancel()
	deadline, _ := observation.Deadline()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		status, err := tc.recordedContainerStatus(observation)
		if err != nil {
			return err
		}
		if status.ContainerID == "" {
			return fmt.Errorf("container %s has empty container ID", status.Name)
		}
		if status.ContainerID != tc.ContainerBaseline.ContainerID || status.RestartCount != tc.ContainerBaseline.RestartCount {
			return fmt.Errorf("container %s restarted during the observation period", status.Name)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-observation.Done():
			return ctx.Err()
		case <-ticker.C:
			if observation.Err() != nil || !time.Now().Before(deadline) {
				if parentDeadline, ok := ctx.Deadline(); ok && !time.Now().Before(parentDeadline) && ctx.Err() == nil {
					return context.DeadlineExceeded
				}
				return ctx.Err()
			}
		}
	}
}

func (tc *TestContext) containerIDNotChangedMinutes(ctx context.Context, duration int) error {
	return tc.containerIDNotChanged(ctx, 5*time.Second, time.Duration(duration)*time.Minute)
}

func (tc *TestContext) containerIDNotChangedSeconds(ctx context.Context, duration int) error {
	return tc.containerIDNotChanged(ctx, time.Second, time.Duration(duration)*time.Second)
}

func (tc *TestContext) podsWithLabel(ctx context.Context, selector string) (*corev1.PodList, error) {
	selectorLabels, err := parseLabelSelector(selector)
	if err != nil {
		return nil, err
	}
	return tc.KubeCli.CoreV1().Pods(tc.Namespace).List(ctx, metav1.ListOptions{LabelSelector: labels.SelectorFromSet(selectorLabels).String()})
}
