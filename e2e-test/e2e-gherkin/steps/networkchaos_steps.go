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
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	e2econfig "github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/config"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/util"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/pkg/fixture"
	"github.com/chaos-mesh/chaos-mesh/pkg/log"
	"github.com/chaos-mesh/chaos-mesh/pkg/portforward"
)

func (tc *TestContext) RegisterNetworkChaosSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^4 network peers are running$`, tc.prepareNetworkPeers)
	ctx.Step(`^a "([^"]*)" chaos named "([^"]*)" is applied on pods with label "([^"]*)"$`, tc.applyNetworkDelay)
	ctx.Step(`^the network condition between peers should eventually be slow$`, tc.verifyConnectionsSlow)
	ctx.Step(`^the network condition between peers should eventually recover$`, tc.verifyConnectionsGood)
	ctx.Step(`^a "([^"]*)" chaos named "([^"]*)" with target "([^"]*)" is applied on pods with label "([^"]*)"$`, tc.applyNetworkDelayWithTarget)
	ctx.Step(`^the network condition from "([^"]*)" to "([^"]*)" should eventually be slow$`, tc.verifyConnectionFromPeerToPeerSlow)
	ctx.Step(`^a "([^"]*)" chaos named "([^"]*)" with target partition "([^"]*)" is applied on pods with label "([^"]*)"$`, tc.applyNetworkDelayWithTargetPartition)
	ctx.Step(`^the network condition from "([^"]*)" to partition "([^"]*)" should eventually be slow$`, tc.verifyConnectionFromPeerToPartitionSlow)
	ctx.Step(`^a "([^"]*)" chaos named "([^"]*)" with target partition "([^"]*)" in both directions is applied on pods with label "([^"]*)"$`, tc.applyNetworkDelayWithTargetPartitionBothDirections)
	ctx.Step(`^the network condition in both directions between "([^"]*)" and partition "([^"]*)" should eventually be slow$`, tc.verifyConnectionBothDirectionsBetweenPeerAndPartitionSlow)
	ctx.Step(`^a "([^"]*)" chaos named "([^"]*)" with single external target "([^"]*)" is applied on pods with label "([^"]*)"$`, tc.applyNetworkDelayWithSingleExternalTarget)
	ctx.Step(`^a "([^"]*)" chaos named "([^"]*)" with external targets "([^"]*)" and "([^"]*)" is applied on pods with label "([^"]*)"$`, tc.applyNetworkDelayWithMultipleExternalTargets)
	ctx.Step(`^the network condition from "([^"]*)" to "([^"]*)" and "([^"]*)" should eventually be slow$`, tc.verifyMultipleConnectionsSlow)
}

func (tc *TestContext) prepareNetworkPeers() error {
	tc.HTTPClient = http.Client{Timeout: 10 * time.Second}

	if tc.FWCancel != nil {
		tc.FWCancel()
		tc.FWCancel = nil
	}
	clientRawConfig, err := e2econfig.LoadClientRawConfig()
	if err != nil {
		return fmt.Errorf("failed to load raw config: %w", err)
	}
	logger, err := log.NewDefaultZapLogger()
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	tc.FWCancel = cancel
	fw, err := portforward.NewPortForwarder(ctx, e2econfig.NewSimpleRESTClientGetter(clientRawConfig), true, logger)
	if err != nil {
		return fmt.Errorf("failed to create port forwarder: %w", err)
	}
	tc.FW = fw

	tc.Ports = []uint16{}
	tc.PFCancels = []context.CancelFunc{}
	tc.NetworkPeers = []*corev1.Pod{}

	for index := 0; index < 4; index++ {
		name := fmt.Sprintf("network-peer-%d", index)
		svc := fixture.NewE2EService(name, tc.Namespace)
		if _, err := tc.KubeCli.CoreV1().Services(tc.Namespace).Create(context.TODO(), svc, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("failed to create service %s: %w", name, err)
		}

		nd := fixture.NewNetworkTestDeployment(name, tc.Namespace, map[string]string{"partition": strconv.Itoa(index % 2)})
		if _, err := tc.KubeCli.AppsV1().Deployments(tc.Namespace).Create(context.TODO(), nd, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("failed to create deployment %s: %w", name, err)
		}

		if err := util.WaitDeploymentReady(name, tc.Namespace, tc.KubeCli); err != nil {
			return fmt.Errorf("failed to wait for deployment %s ready: %w", name, err)
		}

		pod, err := tc.getPeerPod(name)
		if err != nil {
			return fmt.Errorf("failed to select pod for %s: %w", name, err)
		}
		tc.NetworkPeers = append(tc.NetworkPeers, pod)

		_, port, pfCancel, err := portforward.ForwardOnePort(tc.FW, tc.Namespace, "svc/"+svc.Name, 8080)
		if err != nil {
			return fmt.Errorf("failed to port-forward for service %s: %w", svc.Name, err)
		}
		tc.Ports = append(tc.Ports, port)
		tc.PFCancels = append(tc.PFCancels, pfCancel)

		if err := util.WaitE2EHelperReady(tc.HTTPClient, port); err != nil {
			return fmt.Errorf("e2e helper not ready on port %d for %s: %w", port, name, err)
		}
	}

	return nil
}

func (tc *TestContext) getPeerPod(name string) (*corev1.Pod, error) {
	listOption := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", name),
	}
	podList, err := tc.KubeCli.CoreV1().Pods(tc.Namespace).List(context.TODO(), listOption)
	if err != nil {
		return nil, err
	}
	if len(podList.Items) == 0 {
		return nil, fmt.Errorf("no pod found for app=%s", name)
	}
	return &podList.Items[0], nil
}

func parseLabelKeyVal(labelKeyVal string) (map[string]string, error) {
	parts := strings.Split(labelKeyVal, "=")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid label selector format: %s", labelKeyVal)
	}
	return map[string]string{parts[0]: parts[1]}, nil
}

func (tc *TestContext) applyNetworkDelay(action, name, labelKeyVal string) error {
	labelMap, err := parseLabelKeyVal(labelKeyVal)
	if err != nil {
		return err
	}

	chaos := tc.makeNetworkDelaySpec(name, labelMap, nil, v1alpha1.OneMode, v1alpha1.OneMode, v1alpha1.To, nil)
	return tc.Client.Create(context.TODO(), chaos)
}

func (tc *TestContext) applyNetworkDelayWithTarget(action, name, targetLabelKeyVal, sourceLabelKeyVal string) error {
	sourceMap, err := parseLabelKeyVal(sourceLabelKeyVal)
	if err != nil {
		return err
	}
	targetMap, err := parseLabelKeyVal(targetLabelKeyVal)
	if err != nil {
		return err
	}

	chaos := tc.makeNetworkDelaySpec(name, sourceMap, targetMap, v1alpha1.OneMode, v1alpha1.OneMode, v1alpha1.To, nil)
	return tc.Client.Create(context.TODO(), chaos)
}

func (tc *TestContext) applyNetworkDelayWithTargetPartition(action, name, partitionKeyVal, sourceLabelKeyVal string) error {
	sourceMap, err := parseLabelKeyVal(sourceLabelKeyVal)
	if err != nil {
		return err
	}
	targetMap, err := parseLabelKeyVal(partitionKeyVal)
	if err != nil {
		return err
	}

	chaos := tc.makeNetworkDelaySpec(name, sourceMap, targetMap, v1alpha1.OneMode, v1alpha1.AllMode, v1alpha1.To, nil)
	return tc.Client.Create(context.TODO(), chaos)
}

func (tc *TestContext) applyNetworkDelayWithTargetPartitionBothDirections(action, name, partitionKeyVal, sourceLabelKeyVal string) error {
	sourceMap, err := parseLabelKeyVal(sourceLabelKeyVal)
	if err != nil {
		return err
	}
	targetMap, err := parseLabelKeyVal(partitionKeyVal)
	if err != nil {
		return err
	}

	chaos := tc.makeNetworkDelaySpec(name, sourceMap, targetMap, v1alpha1.OneMode, v1alpha1.AllMode, v1alpha1.Both, nil)
	return tc.Client.Create(context.TODO(), chaos)
}

func (tc *TestContext) applyNetworkDelayWithSingleExternalTarget(action, name, targetPeerName, sourceLabelKeyVal string) error {
	sourceMap, err := parseLabelKeyVal(sourceLabelKeyVal)
	if err != nil {
		return err
	}

	pod, err := tc.findPeerByName(targetPeerName)
	if err != nil {
		return err
	}

	chaos := tc.makeNetworkDelaySpec(name, sourceMap, nil, v1alpha1.OneMode, v1alpha1.OneMode, v1alpha1.To, []string{pod.Status.PodIP})
	return tc.Client.Create(context.TODO(), chaos)
}

func (tc *TestContext) applyNetworkDelayWithMultipleExternalTargets(action, name, target1, target2, sourceLabelKeyVal string) error {
	sourceMap, err := parseLabelKeyVal(sourceLabelKeyVal)
	if err != nil {
		return err
	}

	pod1, err := tc.findPeerByName(target1)
	if err != nil {
		return err
	}
	pod2, err := tc.findPeerByName(target2)
	if err != nil {
		return err
	}

	chaos := tc.makeNetworkDelaySpec(name, sourceMap, nil, v1alpha1.OneMode, v1alpha1.OneMode, v1alpha1.To, []string{pod1.Status.PodIP, pod2.Status.PodIP})
	return tc.Client.Create(context.TODO(), chaos)
}

func (tc *TestContext) findPeerByName(name string) (*corev1.Pod, error) {
	for _, p := range tc.NetworkPeers {
		if p.Labels["app"] == name {
			return p, nil
		}
	}
	return nil, fmt.Errorf("peer pod with app label %s not found in recorded peers", name)
}

func (tc *TestContext) makeNetworkDelaySpec(
	name string,
	fromLabel, toLabel map[string]string,
	fromMode, toMode v1alpha1.SelectorMode,
	direction v1alpha1.Direction,
	externalTargets []string,
) *v1alpha1.NetworkChaos {
	var target *v1alpha1.PodSelector
	if toLabel != nil {
		target = &v1alpha1.PodSelector{
			Selector: v1alpha1.PodSelectorSpec{
				GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
					Namespaces:     []string{tc.Namespace},
					LabelSelectors: toLabel,
				},
			},
			Mode: toMode,
		}
	}

	return &v1alpha1.NetworkChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: tc.Namespace,
		},
		Spec: v1alpha1.NetworkChaosSpec{
			Action: v1alpha1.DelayAction,
			TcParameter: v1alpha1.TcParameter{
				Delay: &v1alpha1.DelaySpec{
					Latency:     "200ms",
					Correlation: "25",
					Jitter:      "0ms",
				},
			},
			Duration:        pointer.String("9m"),
			Target:          target,
			Direction:       direction,
			ExternalTargets: externalTargets,
			PodSelector: v1alpha1.PodSelector{
				Selector: v1alpha1.PodSelectorSpec{
					GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
						Namespaces:     []string{tc.Namespace},
						LabelSelectors: fromLabel,
					},
				},
				Mode: fromMode,
			},
		},
	}
}

func (tc *TestContext) testNetworkDelayProbe(fromPort uint16, targetIP string) (int64, error) {
	body := []byte(fmt.Sprintf("{\"targetIP\":\"%s\"}", targetIP))
	resp, err := tc.HTTPClient.Post(fmt.Sprintf("http://localhost:%d/network/ping", fromPort), "application/json", bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	parts := strings.Split(string(out), " ")
	if len(parts) != 2 || parts[0] != "OK" {
		return 0, fmt.Errorf("unexpected ping response: %s", string(out))
	}
	return strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
}

func (tc *TestContext) probeNetworkCondition(bidirection bool) map[string][][]int {
	result := make(map[string][][]int)

	for source := range tc.NetworkPeers {
		initialTarget := source + 1
		if bidirection {
			initialTarget = 0
		}
		for target := initialTarget; target < len(tc.NetworkPeers); target++ {
			if target == source {
				continue
			}

			delay, err := tc.testNetworkDelayProbe(tc.Ports[source], tc.NetworkPeers[target].Status.PodIP)
			if err != nil {
				klog.Errorf("probe error from peer %d to %d: %v", source, target, err)
				result["blocked"] = append(result["blocked"], []int{source, target})
				continue
			}

			if delay > 100*1e6 {
				result["slow"] = append(result["slow"], []int{source, target})
			} else {
				result["good"] = append(result["good"], []int{source, target})
			}
		}
	}

	return result
}

func parsePeerIndex(peerName string) (int, error) {
	var idx int
	if _, err := fmt.Sscanf(peerName, "network-peer-%d", &idx); err != nil {
		return -1, fmt.Errorf("invalid peer name: %s", peerName)
	}
	return idx, nil
}

func pairsEqual(a, b [][]int) bool {
	if len(a) != len(b) {
		return false
	}
	count := make(map[string]int)
	for _, p := range a {
		count[fmt.Sprintf("%d->%d", p[0], p[1])]++
	}
	for _, p := range b {
		key := fmt.Sprintf("%d->%d", p[0], p[1])
		count[key]--
		if count[key] < 0 {
			return false
		}
	}
	return true
}

func (tc *TestContext) verifyConnectionsSlow(ctx context.Context) error {
	expected := [][]int{{0, 1}, {0, 2}, {0, 3}}
	return wait.PollUntilContextTimeout(ctx, time.Second, 15*time.Second, false, func(ctx context.Context) (bool, error) {
		result := tc.probeNetworkCondition(false)
		return len(result["blocked"]) == 0 && pairsEqual(result["slow"], expected), nil
	})
}

func (tc *TestContext) verifyConnectionsGood(ctx context.Context) error {
	return wait.PollUntilContextTimeout(ctx, time.Second, 15*time.Second, false, func(ctx context.Context) (bool, error) {
		result := tc.probeNetworkCondition(false)
		return len(result["blocked"]) == 0 && len(result["slow"]) == 0, nil
	})
}

func (tc *TestContext) verifyConnectionFromPeerToPeerSlow(ctx context.Context, sourcePeer, targetPeer string) error {
	source, err := parsePeerIndex(sourcePeer)
	if err != nil {
		return err
	}
	target, err := parsePeerIndex(targetPeer)
	if err != nil {
		return err
	}

	expected := [][]int{{source, target}}
	return wait.PollUntilContextTimeout(ctx, time.Second, 15*time.Second, false, func(ctx context.Context) (bool, error) {
		result := tc.probeNetworkCondition(false)
		return len(result["blocked"]) == 0 && pairsEqual(result["slow"], expected), nil
	})
}

func (tc *TestContext) verifyConnectionFromPeerToPartitionSlow(ctx context.Context, sourcePeer, partitionStr string) error {
	source, err := parsePeerIndex(sourcePeer)
	if err != nil {
		return err
	}
	partition, err := strconv.Atoi(partitionStr)
	if err != nil {
		return fmt.Errorf("invalid partition: %s", partitionStr)
	}

	var expected [][]int
	for target := source + 1; target < len(tc.NetworkPeers); target++ {
		if target%2 == partition {
			expected = append(expected, []int{source, target})
		}
	}

	return wait.PollUntilContextTimeout(ctx, time.Second, 15*time.Second, false, func(ctx context.Context) (bool, error) {
		result := tc.probeNetworkCondition(false)
		return len(result["blocked"]) == 0 && pairsEqual(result["slow"], expected), nil
	})
}

func (tc *TestContext) verifyConnectionBothDirectionsBetweenPeerAndPartitionSlow(ctx context.Context, sourcePeer, partitionStr string) error {
	source, err := parsePeerIndex(sourcePeer)
	if err != nil {
		return err
	}
	partition, err := strconv.Atoi(partitionStr)
	if err != nil {
		return fmt.Errorf("invalid partition: %s", partitionStr)
	}

	var expected [][]int
	for target := 0; target < len(tc.NetworkPeers); target++ {
		if target == source {
			continue
		}
		if target%2 == partition {
			expected = append(expected, []int{source, target}, []int{target, source})
		}
	}

	return wait.PollUntilContextTimeout(ctx, time.Second, 15*time.Second, false, func(ctx context.Context) (bool, error) {
		result := tc.probeNetworkCondition(true)
		return len(result["blocked"]) == 0 && pairsEqual(result["slow"], expected), nil
	})
}

func (tc *TestContext) verifyMultipleConnectionsSlow(ctx context.Context, sourcePeer, target1, target2 string) error {
	source, err := parsePeerIndex(sourcePeer)
	if err != nil {
		return err
	}
	t1, err := parsePeerIndex(target1)
	if err != nil {
		return err
	}
	t2, err := parsePeerIndex(target2)
	if err != nil {
		return err
	}

	expected := [][]int{{source, t1}, {source, t2}}
	return wait.PollUntilContextTimeout(ctx, time.Second, 15*time.Second, false, func(ctx context.Context) (bool, error) {
		result := tc.probeNetworkCondition(false)
		return len(result["blocked"]) == 0 && pairsEqual(result["slow"], expected), nil
	})
}
