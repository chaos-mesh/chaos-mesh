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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/pointer"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/util"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/pkg/fixture"
)

// RegisterNetworkChaosSteps registers all NetworkChaos-related Gherkin step
// definitions onto the supplied ScenarioContext.
func RegisterNetworkChaosSteps(sc *godog.ScenarioContext, ctx *ScenarioContext) {
	// ---- Given steps -------------------------------------------------------

	sc.Step(`^the network peers are ready and all connections are good$`,
		func() error {
			return ctx.givenNetworkPeersReady()
		})

	sc.Step(`^a hostNetwork deployment named "([^"]+)" is running in the test namespace$`,
		func(name string) error {
			return ctx.givenHostNetworkDeployment(name)
		})

	// ---- When steps --------------------------------------------------------

	sc.Step(`^I create a NetworkDelay chaos named "([^"]+)" with latency "([^"]+)" from "([^"]+)" in direction "([^"]+)"$`,
		func(chaosName, latency, fromApp, direction string) error {
			return ctx.createNetworkDelayChaos(chaosName, latency, fromApp, "", "", direction, false)
		})

	sc.Step(`^I create a NetworkDelay chaos named "([^"]+)" with latency "([^"]+)" from "([^"]+)" to "([^"]+)" in direction "([^"]+)"$`,
		func(chaosName, latency, fromApp, toApp, direction string) error {
			return ctx.createNetworkDelayChaos(chaosName, latency, fromApp, toApp, "", direction, false)
		})

	sc.Step(`^I create a NetworkDelay chaos named "([^"]+)" with latency "([^"]+)" from "([^"]+)" to partition "([^"]+)" in direction "([^"]+)"$`,
		func(chaosName, latency, fromApp, partition, direction string) error {
			return ctx.createNetworkDelayChaos(chaosName, latency, fromApp, "", partition, direction, false)
		})

	sc.Step(`^I inject both-direction delay between partition "([^"]+)" and partition "([^"]+)"$`,
		func(fromPartition, toPartition string) error {
			return ctx.createCrossoverDelayChaos(fromPartition, toPartition)
		})

	sc.Step(`^I create a NetworkPartition chaos named "([^"]+)" from "([^"]+)" to "([^"]+)" in direction "([^"]+)"$`,
		func(chaosName, fromApp, toApp, direction string) error {
			return ctx.createNetworkPartitionChaos(chaosName, fromApp, toApp, "", direction, false)
		})

	sc.Step(`^I create a NetworkPartition chaos named "([^"]+)" from "([^"]+)" to partition "([^"]+)" with all pods in direction "([^"]+)"$`,
		func(chaosName, fromApp, partition, direction string) error {
			return ctx.createNetworkPartitionChaos(chaosName, fromApp, "", partition, direction, false)
		})

	sc.Step(`^I create a NetworkPartition chaos named "([^"]+)" from "([^"]+)" with no target in direction "([^"]+)"$`,
		func(chaosName, fromApp, direction string) error {
			return ctx.createNetworkPartitionChaos(chaosName, fromApp, "", "", direction, true)
		})

	sc.Step(`^I delete the NetworkChaos "([^"]+)"$`,
		func(chaosName string) error {
			return ctx.deleteNetworkChaos(chaosName)
		})

	// ---- Then steps --------------------------------------------------------

	sc.Step(`^all connections should recover within 15 seconds$`,
		func() error {
			return ctx.assertAllConnectionsGood(15 * time.Second)
		})

	sc.Step(`^slow connections should be peer pairs (\[\[.*\]\]) within 15 seconds$`,
		func(pairsJSON string) error {
			pairs, err := parsePairList(pairsJSON)
			if err != nil {
				return err
			}
			return ctx.assertSlowConnections(pairs, 15*time.Second, false)
		})

	sc.Step(`^slow connections should be peer pairs (\[\[.*\]\]) within 15 seconds bidirectional$`,
		func(pairsJSON string) error {
			pairs, err := parsePairList(pairsJSON)
			if err != nil {
				return err
			}
			return ctx.assertSlowConnections(pairs, 15*time.Second, true)
		})

	sc.Step(`^blocked connections should be peer pairs (\[\[.*\]\]) within 15 seconds$`,
		func(pairsJSON string) error {
			pairs, err := parsePairList(pairsJSON)
			if err != nil {
				return err
			}
			return ctx.assertBlockedConnections(pairs, 15*time.Second)
		})

	sc.Step(`^the chaos "([^"]+)" should not inject into "([^"]+)" pods within 1 minute$`,
		func(chaosName, appName string) error {
			return ctx.assertChaosNotInjected(chaosName, appName, time.Minute)
		})
}

// ---------------------------------------------------------------------------
// Given helpers
// ---------------------------------------------------------------------------

func (ctx *ScenarioContext) givenNetworkPeersReady() error {
	for i, port := range ctx.Ports {
		if err := util.WaitE2EHelperReady(ctx.HTTPClient, port); err != nil {
			return fmt.Errorf("wait e2e helper ready for peer %d: %w", i, err)
		}
	}
	// Verify baseline: no blocked, no slow connections.
	result := probeNetwork(ctx, false)
	if len(result[networkConditionBlocked]) != 0 || len(result[networkConditionSlow]) != 0 {
		return fmt.Errorf("baseline network not clean: blocked=%v slow=%v",
			result[networkConditionBlocked], result[networkConditionSlow])
	}
	return nil
}

func (ctx *ScenarioContext) givenHostNetworkDeployment(name string) error {
	nd := fixture.NewNetworkTestDeployment(name, ctx.Namespace, map[string]string{"partition": "0"})
	nd.Spec.Template.Spec.HostNetwork = true
	_, err := ctx.KubeCli.AppsV1().Deployments(ctx.Namespace).Create(
		context.TODO(), nd, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("create hostNetwork deployment: %w", err)
	}
	return util.WaitDeploymentReady(name, ctx.Namespace, ctx.KubeCli)
}

// ---------------------------------------------------------------------------
// When helpers
// ---------------------------------------------------------------------------

func directionFromString(s string) v1alpha1.Direction {
	switch strings.ToLower(s) {
	case "from":
		return v1alpha1.From
	case "both":
		return v1alpha1.Both
	default:
		return v1alpha1.To
	}
}

func (ctx *ScenarioContext) createNetworkDelayChaos(
	chaosName, latency, fromApp, toApp, toPartition, directionStr string,
	noTarget bool,
) error {
	dir := directionFromString(directionStr)
	duration := pointer.String("9m")

	var toLabels map[string]string
	toMode := v1alpha1.OneMode
	if toApp != "" {
		toLabels = map[string]string{"app": toApp}
	} else if toPartition != "" {
		toLabels = map[string]string{"partition": toPartition}
		toMode = v1alpha1.AllMode
	}

	tcparam := v1alpha1.TcParameter{
		Delay: &v1alpha1.DelaySpec{
			Latency:     latency,
			Correlation: "25",
			Jitter:      "0ms",
		},
	}

	chaos := makeNetworkDelayChaos(
		ctx.Namespace, chaosName,
		map[string]string{"app": fromApp}, toLabels,
		v1alpha1.OneMode, toMode,
		dir, tcparam, duration,
	)
	bctx, _ := ctx.Background()
	return ctx.Cli.Create(bctx, chaos.DeepCopy())
}

func (ctx *ScenarioContext) createCrossoverDelayChaos(fromPartition, toPartition string) error {
	tcparam := v1alpha1.TcParameter{
		Delay: &v1alpha1.DelaySpec{
			Latency:     "200ms",
			Correlation: "25",
			Jitter:      "0ms",
		},
	}
	chaos := makeNetworkDelayChaos(
		ctx.Namespace, "network-chaos-1",
		map[string]string{"partition": fromPartition},
		map[string]string{"partition": toPartition},
		v1alpha1.AllMode, v1alpha1.AllMode,
		v1alpha1.Both, tcparam, nil,
	)
	chaos.Spec.Direction = v1alpha1.Both
	bctx, _ := ctx.Background()
	return ctx.Cli.Create(bctx, chaos.DeepCopy())
}

func (ctx *ScenarioContext) createNetworkPartitionChaos(
	chaosName, fromApp, toApp, toPartition, directionStr string,
	noTarget bool,
) error {
	dir := directionFromString(directionStr)
	duration := pointer.String("9m")

	var toLabels map[string]string
	toMode := v1alpha1.OneMode
	if toApp != "" {
		toLabels = map[string]string{"app": toApp}
	} else if toPartition != "" {
		toLabels = map[string]string{"partition": toPartition}
		toMode = v1alpha1.AllMode
	}

	var fromLabels map[string]string
	if fromApp != "" {
		fromLabels = map[string]string{"app": fromApp}
	}

	chaos := makeNetworkPartitionChaos(
		ctx.Namespace, chaosName,
		fromLabels, toLabels,
		v1alpha1.OneMode, toMode,
		dir, duration,
	)
	bctx, _ := ctx.Background()
	return ctx.Cli.Create(bctx, chaos.DeepCopy())
}

func (ctx *ScenarioContext) deleteNetworkChaos(chaosName string) error {
	chaos := &v1alpha1.NetworkChaos{}
	chaos.Name = chaosName
	chaos.Namespace = ctx.Namespace
	bctx, _ := ctx.Background()
	return ctx.Cli.Delete(bctx, chaos)
}

// ---------------------------------------------------------------------------
// Then helpers – assertions
// ---------------------------------------------------------------------------

func probeNetwork(ctx *ScenarioContext, bidirection bool) map[string][][]int {
	return probeNetworkCondition(ctx.HTTPClient, ctx.NetworkPeers, ctx.Ports, bidirection)
}

func (ctx *ScenarioContext) assertAllConnectionsGood(timeout time.Duration) error {
	return wait.PollUntilContextTimeout(context.TODO(), time.Second, timeout, true,
		func(_ context.Context) (bool, error) {
			result := probeNetwork(ctx, false)
			return len(result[networkConditionBlocked]) == 0 && len(result[networkConditionSlow]) == 0, nil
		})
}

func (ctx *ScenarioContext) assertSlowConnections(expected [][]int, timeout time.Duration, bidirection bool) error {
	var result map[string][][]int
	err := wait.PollUntilContextTimeout(context.TODO(), time.Second, timeout, true,
		func(_ context.Context) (bool, error) {
			result = probeNetwork(ctx, bidirection)
			return pairsEqual(result[networkConditionSlow], expected) &&
				len(result[networkConditionBlocked]) == 0, nil
		})
	if err != nil {
		return fmt.Errorf("expected slow=%v blocked=[] got slow=%v blocked=%v",
			expected, result[networkConditionSlow], result[networkConditionBlocked])
	}
	return nil
}

func (ctx *ScenarioContext) assertBlockedConnections(expected [][]int, timeout time.Duration) error {
	var result map[string][][]int
	err := wait.PollUntilContextTimeout(context.TODO(), time.Second, timeout, true,
		func(_ context.Context) (bool, error) {
			result = probeNetwork(ctx, false)
			return pairsEqual(result[networkConditionBlocked], expected) &&
				len(result[networkConditionSlow]) == 0, nil
		})
	if err != nil {
		return fmt.Errorf("expected blocked=%v slow=[] got blocked=%v slow=%v",
			expected, result[networkConditionBlocked], result[networkConditionSlow])
	}
	return nil
}

func (ctx *ScenarioContext) assertChaosNotInjected(chaosName, appName string, timeout time.Duration) error {
	chaosKey := types.NamespacedName{Namespace: ctx.Namespace, Name: chaosName}
	return wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, timeout, true,
		func(pollCtx context.Context) (bool, error) {
			chaos := &v1alpha1.NetworkChaos{}
			if err := ctx.Cli.Get(pollCtx, chaosKey, chaos); err != nil {
				return false, err
			}
			for _, record := range chaos.Status.ChaosStatus.Experiment.Records {
				if strings.Contains(record.Id, appName) && record.Phase == v1alpha1.Injected {
					return false, nil // unexpectedly injected
				}
			}
			return true, nil
		})
}

// ---------------------------------------------------------------------------
// Helpers re-used from the original networkchaos package
// ---------------------------------------------------------------------------

// These functions mirror the unexported helpers in
// e2e-test/e2e/chaos/networkchaos/misc.go so the BDD layer can build chaos
// objects the same way without importing an internal package.

func makeNetworkDelayChaos(
	namespace, name string,
	fromLabels, toLabels map[string]string,
	fromMode, toMode v1alpha1.SelectorMode,
	direction v1alpha1.Direction,
	tcparam v1alpha1.TcParameter,
	duration *string,
) *v1alpha1.NetworkChaos {
	var target *v1alpha1.PodSelector
	if toLabels != nil {
		target = &v1alpha1.PodSelector{
			Selector: v1alpha1.PodSelectorSpec{
				GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
					Namespaces:     []string{namespace},
					LabelSelectors: toLabels,
				},
			},
			Mode: toMode,
		}
	}
	return &v1alpha1.NetworkChaos{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: v1alpha1.NetworkChaosSpec{
			Action:      v1alpha1.DelayAction,
			TcParameter: tcparam,
			Duration:    duration,
			Target:      target,
			Direction:   direction,
			PodSelector: v1alpha1.PodSelector{
				Selector: v1alpha1.PodSelectorSpec{
					GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
						Namespaces:     []string{namespace},
						LabelSelectors: fromLabels,
					},
				},
				Mode: fromMode,
			},
		},
	}
}

func makeNetworkPartitionChaos(
	namespace, name string,
	fromLabels, toLabels map[string]string,
	fromMode, toMode v1alpha1.SelectorMode,
	direction v1alpha1.Direction,
	duration *string,
) *v1alpha1.NetworkChaos {
	var target *v1alpha1.PodSelector
	if toLabels != nil {
		target = &v1alpha1.PodSelector{
			Selector: v1alpha1.PodSelectorSpec{
				GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
					Namespaces:     []string{namespace},
					LabelSelectors: toLabels,
				},
			},
			Mode: toMode,
		}
	}
	return &v1alpha1.NetworkChaos{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: v1alpha1.NetworkChaosSpec{
			Action:    v1alpha1.PartitionAction,
			Direction: direction,
			Target:    target,
			Duration:  duration,
			PodSelector: v1alpha1.PodSelector{
				Selector: v1alpha1.PodSelectorSpec{
					GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
						Namespaces:     []string{namespace},
						LabelSelectors: fromLabels,
					},
				},
				Mode: fromMode,
			},
		},
	}
}

// ---------------------------------------------------------------------------
// Utility helpers
// ---------------------------------------------------------------------------

func parsePairList(raw string) ([][]int, error) {
	var pairs [][]int
	if err := json.Unmarshal([]byte(raw), &pairs); err != nil {
		return nil, fmt.Errorf("parse pair list %q: %w", raw, err)
	}
	return pairs, nil
}

func pairsEqual(a, b [][]int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}
		for j := range a[i] {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}
	return true
}
