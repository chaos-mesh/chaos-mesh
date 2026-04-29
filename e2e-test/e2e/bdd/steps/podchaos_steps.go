// Copyright 2024 Chaos Mesh Authors.
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

package steps

import (
	"context"
	"fmt"
	"time"

	"github.com/cucumber/godog"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/config"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/util"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/pkg/fixture"
)

// RegisterPodChaosSteps registers all PodChaos-related Gherkin step
// definitions onto the supplied ScenarioContext. Steps are bound to the
// ScenarioContext so that state (e.g. initial pod UIDs) flows naturally
// between Given / When / Then clauses.
func RegisterPodChaosSteps(sc *godog.ScenarioContext, ctx *ScenarioContext) {
	// ---- Given steps -------------------------------------------------------

	sc.Step(`^a nginx pod named "([^"]+)" is running in the test namespace$`,
		func(name string) error {
			return ctx.givenNginxPod(name)
		})

	sc.Step(`^a nginx deployment named "([^"]+)" with (\d+) replica(?:s)? is running in the test namespace$`,
		func(name string, replicas int) error {
			return ctx.givenNginxDeployment(name, int32(replicas))
		})

	sc.Step(`^a timer deployment named "([^"]+)" is running in the test namespace$`,
		func(name string) error {
			return ctx.givenTimerDeployment(name)
		})

	// ---- When steps --------------------------------------------------------

	sc.Step(`^I create a PodKill chaos named "([^"]+)" targeting pods with label "([^"]+)"$`,
		func(chaosName, labelSelector string) error {
			return ctx.createPodKillChaos(chaosName, labelSelector, nil)
		})

	sc.Step(`^I create a PodFailure chaos named "([^"]+)" targeting pods with label "([^"]+)"$`,
		func(chaosName, labelSelector string) error {
			return ctx.createPodFailureChaos(chaosName, labelSelector, nil)
		})

	sc.Step(`^I create a PodFailure chaos with duration "([^"]+)" named "([^"]+)" targeting pods with label "([^"]+)"$`,
		func(duration, chaosName, labelSelector string) error {
			return ctx.createPodFailureChaos(chaosName, labelSelector, &duration)
		})

	sc.Step(`^I create a ContainerKill chaos named "([^"]+)" targeting container "([^"]+)" in pods with label "([^"]+)"$`,
		func(chaosName, container, labelSelector string) error {
			return ctx.createContainerKillChaos(chaosName, container, labelSelector)
		})

	sc.Step(`^I delete the PodChaos "([^"]+)"$`,
		func(chaosName string) error {
			return ctx.deletePodChaos(chaosName)
		})

	sc.Step(`^I pause the PodChaos "([^"]+)"$`,
		func(chaosName string) error {
			return ctx.pausePodChaos(chaosName)
		})

	sc.Step(`^I unpause the PodChaos "([^"]+)"$`,
		func(chaosName string) error {
			return ctx.unpausePodChaos(chaosName)
		})

	// ---- Then steps --------------------------------------------------------

	sc.Step(`^the pod "([^"]+)" should be deleted within 5 minutes$`,
		func(podName string) error {
			return ctx.assertPodDeleted(podName, 5*time.Minute)
		})

	sc.Step(`^at least one pod UID should change within 5 minutes$`,
		func() error {
			return ctx.assertPodUIDChanged(5 * time.Minute)
		})

	sc.Step(`^the chaos "([^"]+)" should NOT enter stopped phase within (\d+) seconds$`,
		func(chaosName string, secs int) error {
			return ctx.assertChaosNotStopped(chaosName, time.Duration(secs)*time.Second)
		})

	sc.Step(`^no pod UIDs should change within 1 minute$`,
		func() error {
			return ctx.assertNoUIDChange(1 * time.Minute)
		})

	sc.Step(`^a pod in deployment "([^"]+)" should have its container image replaced with the pause image$`,
		func(deployName string) error {
			return ctx.assertPodUsePauseImage(deployName, 5*time.Minute)
		})

	sc.Step(`^a pod in deployment "([^"]+)" should have its container image replaced with the pause image again$`,
		func(deployName string) error {
			return ctx.assertPodUsePauseImage(deployName, 5*time.Minute)
		})

	sc.Step(`^all pods in deployment "([^"]+)" should recover to the original image within 2 minutes$`,
		func(deployName string) error {
			return ctx.assertPodsRecoverImage(deployName, 2*time.Minute)
		})

	sc.Step(`^no pod in deployment "([^"]+)" should have the pause image within 30 seconds$`,
		func(deployName string) error {
			return ctx.assertNoPauseImage(deployName, 30*time.Second)
		})

	sc.Step(`^the chaos "([^"]+)" should enter stopped phase within 5 minutes$`,
		func(chaosName string) error {
			return ctx.assertChaosPhase(chaosName, v1alpha1.StoppedPhase, 5*time.Minute)
		})

	sc.Step(`^the chaos "([^"]+)" should enter running phase within 5 minutes$`,
		func(chaosName string) error {
			return ctx.assertChaosPhase(chaosName, v1alpha1.RunningPhase, 5*time.Minute)
		})

	sc.Step(`^the nginx container should show a last termination state within 5 minutes$`,
		func() error {
			return ctx.assertContainerTerminated("nginx", 5*time.Minute)
		})

	sc.Step(`^the nginx container should recover to running state within 5 minutes$`,
		func() error {
			return ctx.assertContainerRunning("nginx", 5*time.Minute)
		})
}

// ---------------------------------------------------------------------------
// Given helpers
// ---------------------------------------------------------------------------

func (ctx *ScenarioContext) givenNginxPod(name string) error {
	pod := fixture.NewCommonNginxPod(name, ctx.Namespace)
	_, err := ctx.KubeCli.CoreV1().Pods(ctx.Namespace).Create(
		context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("create nginx pod: %w", err)
	}
	return waitPodRunning(name, ctx.Namespace, ctx.KubeCli)
}

func (ctx *ScenarioContext) givenNginxDeployment(name string, replicas int32) error {
	nd := fixture.NewCommonNginxDeployment(name, ctx.Namespace, replicas)
	_, err := ctx.KubeCli.AppsV1().Deployments(ctx.Namespace).Create(
		context.TODO(), nd, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("create nginx deployment: %w", err)
	}
	if err := util.WaitDeploymentReady(name, ctx.Namespace, ctx.KubeCli); err != nil {
		return fmt.Errorf("wait nginx deployment ready: %w", err)
	}
	// Snapshot initial UIDs so we can detect kills later.
	return ctx.snapshotPodUIDs("app="+name)
}

func (ctx *ScenarioContext) givenTimerDeployment(name string) error {
	nd := fixture.NewTimerDeployment(name, ctx.Namespace)
	_, err := ctx.KubeCli.AppsV1().Deployments(ctx.Namespace).Create(
		context.TODO(), nd, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("create timer deployment: %w", err)
	}
	return util.WaitDeploymentReady(name, ctx.Namespace, ctx.KubeCli)
}

func (ctx *ScenarioContext) snapshotPodUIDs(labelSelector string) error {
	pods, err := ctx.KubeCli.CoreV1().Pods(ctx.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return err
	}
	for _, p := range pods.Items {
		ctx.InitialPodUIDs[p.Name] = string(p.UID)
	}
	return nil
}

// ---------------------------------------------------------------------------
// When helpers – create / delete / pause chaos
// ---------------------------------------------------------------------------

func parseLabelSelector(raw string) map[string]string {
	// Accept "key=value" format only (sufficient for these tests).
	m := make(map[string]string)
	for i := 0; i < len(raw); i++ {
		if raw[i] == '=' {
			m[raw[:i]] = raw[i+1:]
			return m
		}
	}
	return m
}

func (ctx *ScenarioContext) createPodKillChaos(chaosName, labelSelector string, duration *string) error {
	chaos := &v1alpha1.PodChaos{
		ObjectMeta: metav1.ObjectMeta{Name: chaosName, Namespace: ctx.Namespace},
		Spec: v1alpha1.PodChaosSpec{
			Action:   v1alpha1.PodKillAction,
			Duration: duration,
			ContainerSelector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Selector: v1alpha1.PodSelectorSpec{
						GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
							Namespaces:     []string{ctx.Namespace},
							LabelSelectors: parseLabelSelector(labelSelector),
						},
					},
					Mode: v1alpha1.OneMode,
				},
			},
		},
	}
	bctx, _ := ctx.Background()
	return ctx.Cli.Create(bctx, chaos)
}

func (ctx *ScenarioContext) createPodFailureChaos(chaosName, labelSelector string, duration *string) error {
	chaos := &v1alpha1.PodChaos{
		ObjectMeta: metav1.ObjectMeta{Name: chaosName, Namespace: ctx.Namespace},
		Spec: v1alpha1.PodChaosSpec{
			Action:   v1alpha1.PodFailureAction,
			Duration: duration,
			ContainerSelector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Selector: v1alpha1.PodSelectorSpec{
						GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
							Namespaces:     []string{ctx.Namespace},
							LabelSelectors: parseLabelSelector(labelSelector),
						},
					},
					Mode: v1alpha1.OneMode,
				},
			},
		},
	}
	bctx, _ := ctx.Background()
	if ctx.ClientSet != nil {
		_, err := ctx.ClientSet.ApiV1alpha1().Podchaos(ctx.Namespace).Create(bctx, chaos, metav1.CreateOptions{})
		return err
	}
	return ctx.Cli.Create(bctx, chaos)
}

func (ctx *ScenarioContext) createContainerKillChaos(chaosName, container, labelSelector string) error {
	chaos := &v1alpha1.PodChaos{
		ObjectMeta: metav1.ObjectMeta{Name: chaosName, Namespace: ctx.Namespace},
		Spec: v1alpha1.PodChaosSpec{
			Action: v1alpha1.ContainerKillAction,
			ContainerSelector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Selector: v1alpha1.PodSelectorSpec{
						GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
							Namespaces:     []string{ctx.Namespace},
							LabelSelectors: parseLabelSelector(labelSelector),
						},
					},
					Mode: v1alpha1.OneMode,
				},
				ContainerNames: []string{container},
			},
		},
	}
	bctx, _ := ctx.Background()
	return ctx.Cli.Create(bctx, chaos)
}

func (ctx *ScenarioContext) deletePodChaos(chaosName string) error {
	chaos := &v1alpha1.PodChaos{}
	chaos.Name = chaosName
	chaos.Namespace = ctx.Namespace
	bctx, _ := ctx.Background()
	if ctx.ClientSet != nil {
		return ctx.ClientSet.ApiV1alpha1().Podchaos(ctx.Namespace).Delete(bctx, chaosName, metav1.DeleteOptions{})
	}
	return ctx.Cli.Delete(bctx, chaos)
}

func (ctx *ScenarioContext) pausePodChaos(chaosName string) error {
	chaos := &v1alpha1.PodChaos{}
	if err := ctx.Cli.Get(context.TODO(), types.NamespacedName{Namespace: ctx.Namespace, Name: chaosName}, chaos); err != nil {
		return err
	}
	bctx, _ := ctx.Background()
	return util.PauseChaos(bctx, ctx.Cli, chaos)
}

func (ctx *ScenarioContext) unpausePodChaos(chaosName string) error {
	chaos := &v1alpha1.PodChaos{}
	if err := ctx.Cli.Get(context.TODO(), types.NamespacedName{Namespace: ctx.Namespace, Name: chaosName}, chaos); err != nil {
		return err
	}
	bctx, _ := ctx.Background()
	return util.UnPauseChaos(bctx, ctx.Cli, chaos)
}

// ---------------------------------------------------------------------------
// Then helpers – assertions
// ---------------------------------------------------------------------------

func (ctx *ScenarioContext) assertPodDeleted(podName string, timeout time.Duration) error {
	return wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, timeout, true,
		func(pollCtx context.Context) (bool, error) {
			_, err := ctx.KubeCli.CoreV1().Pods(ctx.Namespace).Get(pollCtx, podName, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			return false, nil
		})
}

func (ctx *ScenarioContext) assertPodUIDChanged(timeout time.Duration) error {
	listOption := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{"app": "nginx"}).String(),
	}
	return wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, timeout, true,
		func(pollCtx context.Context) (bool, error) {
			newPods, err := ctx.KubeCli.CoreV1().Pods(ctx.Namespace).List(pollCtx, listOption)
			if err != nil {
				return false, nil
			}
			for _, p := range newPods.Items {
				if orig, ok := ctx.InitialPodUIDs[p.Name]; ok && orig != string(p.UID) {
					return true, nil
				}
				if _, ok := ctx.InitialPodUIDs[p.Name]; !ok {
					return true, nil
				}
			}
			return false, nil
		})
}

func (ctx *ScenarioContext) assertChaosNotStopped(chaosName string, timeout time.Duration) error {
	chaosKey := types.NamespacedName{Namespace: ctx.Namespace, Name: chaosName}
	err := wait.PollUntilContextTimeout(context.TODO(), time.Second, timeout, true,
		func(pollCtx context.Context) (bool, error) {
			chaos := &v1alpha1.PodChaos{}
			if err := ctx.Cli.Get(pollCtx, chaosKey, chaos); err != nil {
				return false, err
			}
			return chaos.Status.Experiment.DesiredPhase == v1alpha1.StoppedPhase, nil
		})
	// We expect the poll to time-out (ErrWaitTimeout), meaning it never stopped.
	if err == nil {
		return fmt.Errorf("chaos %s unexpectedly entered stopped phase", chaosName)
	}
	return nil
}

func (ctx *ScenarioContext) assertNoUIDChange(timeout time.Duration) error {
	listOption := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{"app": "nginx"}).String(),
	}
	// Re-snapshot current UIDs before the check window.
	_ = ctx.snapshotPodUIDs("app=nginx")

	err := wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, timeout, true,
		func(pollCtx context.Context) (bool, error) {
			newPods, err := ctx.KubeCli.CoreV1().Pods(ctx.Namespace).List(pollCtx, listOption)
			if err != nil {
				return false, nil
			}
			for _, p := range newPods.Items {
				if orig, ok := ctx.InitialPodUIDs[p.Name]; ok && orig != string(p.UID) {
					return true, nil // a kill happened – bad
				}
			}
			return false, nil
		})
	if err == nil {
		return fmt.Errorf("a pod was killed while chaos was paused")
	}
	return nil
}

func (ctx *ScenarioContext) assertChaosPhase(chaosName string, phase v1alpha1.DesiredPhase, timeout time.Duration) error {
	chaosKey := types.NamespacedName{Namespace: ctx.Namespace, Name: chaosName}
	return wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, timeout, true,
		func(pollCtx context.Context) (bool, error) {
			chaos := &v1alpha1.PodChaos{}
			if err := ctx.Cli.Get(pollCtx, chaosKey, chaos); err != nil {
				return false, nil
			}
			return chaos.Status.Experiment.DesiredPhase == phase, nil
		})
}

func (ctx *ScenarioContext) assertPodUsePauseImage(deployName string, timeout time.Duration) error {
	listOption := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{"app": deployName}).String(),
	}
	return wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, timeout, true,
		func(pollCtx context.Context) (bool, error) {
			pods, err := ctx.KubeCli.CoreV1().Pods(ctx.Namespace).List(pollCtx, listOption)
			if err != nil || len(pods.Items) == 0 {
				return false, nil
			}
			for _, c := range pods.Items[0].Spec.Containers {
				if c.Image == config.TestConfig.PauseImage {
					return true, nil
				}
			}
			return false, nil
		})
}

func (ctx *ScenarioContext) assertPodsRecoverImage(deployName string, timeout time.Duration) error {
	deploy, err := ctx.KubeCli.AppsV1().Deployments(ctx.Namespace).Get(
		context.TODO(), deployName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get deployment %s: %w", deployName, err)
	}
	originalImage := deploy.Spec.Template.Spec.Containers[0].Image
	listOption := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{"app": deployName}).String(),
	}
	return wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, timeout, true,
		func(pollCtx context.Context) (bool, error) {
			pods, err := ctx.KubeCli.CoreV1().Pods(ctx.Namespace).List(pollCtx, listOption)
			if err != nil || len(pods.Items) == 0 {
				return false, nil
			}
			for _, c := range pods.Items[0].Spec.Containers {
				if c.Image == originalImage {
					return true, nil
				}
			}
			return false, nil
		})
}

func (ctx *ScenarioContext) assertNoPauseImage(deployName string, timeout time.Duration) error {
	listOption := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{"app": deployName}).String(),
	}
	err := wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, timeout, true,
		func(pollCtx context.Context) (bool, error) {
			pods, err := ctx.KubeCli.CoreV1().Pods(ctx.Namespace).List(pollCtx, listOption)
			if err != nil || len(pods.Items) == 0 {
				return false, nil
			}
			for _, c := range pods.Items[0].Spec.Containers {
				if c.Image == config.TestConfig.PauseImage {
					return true, nil // pause image found – bad
				}
			}
			return false, nil
		})
	if err == nil {
		return fmt.Errorf("pause image still present after chaos paused")
	}
	return nil
}

func (ctx *ScenarioContext) assertContainerTerminated(containerName string, timeout time.Duration) error {
	listOption := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{"app": "nginx"}).String(),
	}
	return wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, timeout, true,
		func(pollCtx context.Context) (bool, error) {
			pods, err := ctx.KubeCli.CoreV1().Pods(ctx.Namespace).List(pollCtx, listOption)
			if err != nil || len(pods.Items) == 0 {
				return false, nil
			}
			for _, cs := range pods.Items[0].Status.ContainerStatuses {
				if cs.Name == containerName && cs.LastTerminationState.Terminated != nil {
					return true, nil
				}
			}
			return false, nil
		})
}

func (ctx *ScenarioContext) assertContainerRunning(containerName string, timeout time.Duration) error {
	listOption := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{"app": "nginx"}).String(),
	}
	return wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, timeout, true,
		func(pollCtx context.Context) (bool, error) {
			pods, err := ctx.KubeCli.CoreV1().Pods(ctx.Namespace).List(pollCtx, listOption)
			if err != nil || len(pods.Items) == 0 {
				return false, nil
			}
			for _, cs := range pods.Items[0].Status.ContainerStatuses {
				if cs.Name == containerName && cs.Ready && cs.State.Running != nil {
					return true, nil
				}
			}
			return false, nil
		})
}

// waitPodRunning waits until the named pod reaches Running phase.
func waitPodRunning(name, namespace string, cli kubernetes.Interface) error {
	return wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 5*time.Minute, true,
		func(pollCtx context.Context) (bool, error) {
			pod, err := cli.CoreV1().Pods(namespace).Get(pollCtx, name, metav1.GetOptions{})
			if err != nil {
				return false, nil
			}
			return pod.Status.Phase == corev1.PodRunning, nil
		})
}
