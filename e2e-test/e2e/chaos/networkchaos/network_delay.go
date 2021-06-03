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

package networkchaos

import (
	"context"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/util"
)

func TestcaseNetworkDelay(
	ns string,
	cli client.Client,
	networkPeers []*corev1.Pod,
	ports []uint16,
	c http.Client,
) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	By("prepare experiment playground")
	for index := range networkPeers {
		err := util.WaitE2EHelperReady(c, ports[index])

		framework.ExpectNoError(err, "wait e2e helper ready error")
	}

	result := probeNetworkCondition(c, networkPeers, ports, false)
	framework.ExpectEqual(len(result[networkConditionBlocked]), 0)
	framework.ExpectEqual(len(result[networkConditionSlow]), 0)

	var (
		testDelayTcParam = v1alpha1.TcParameter{
			Delay: &v1alpha1.DelaySpec{
				Latency:     "200ms",
				Correlation: "25",
				Jitter:      "0ms",
			},
		}
		testDelayTcParamEvenMoreComplicate = v1alpha1.TcParameter{
			Delay: &v1alpha1.DelaySpec{
				Latency:     "200ms",
				Correlation: "25",
				Jitter:      "0ms",
			},
			Loss: &v1alpha1.LossSpec{
				Loss:        "25",
				Correlation: "25",
			},
			Duplicate: &v1alpha1.DuplicateSpec{
				Duplicate:   "25",
				Correlation: "25",
			},
			Corrupt: &v1alpha1.CorruptSpec{
				Corrupt:     "25",
				Correlation: "25",
			},
		}
		testDelayDuration = pointer.StringPtr("9m")
	)

	By("normal delay chaos")
	networkDelay := makeNetworkDelayChaos(
		ns, "network-chaos-1",
		map[string]string{"app": "network-peer-0"},
		nil, // no target specified
		v1alpha1.OnePodMode,
		v1alpha1.OnePodMode,
		v1alpha1.To,
		testDelayTcParam,
		testDelayDuration,
	)
	By("Injecting delay for 0")
	err := cli.Create(ctx, networkDelay.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")

	wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		result = probeNetworkCondition(c, networkPeers, ports, false)
		if len(result[networkConditionBlocked]) != 0 || len(result[networkConditionSlow]) != 3 {
			return false, nil
		}
		return true, nil
	})

	framework.ExpectEqual(len(result[networkConditionBlocked]), 0)
	framework.ExpectEqual(result[networkConditionSlow], [][]int{{0, 1}, {0, 2}, {0, 3}})

	By("recover")
	err = cli.Delete(ctx, networkDelay.DeepCopy())
	framework.ExpectNoError(err, "delete network chaos error")

	wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		result = probeNetworkCondition(c, networkPeers, ports, false)
		if len(result[networkConditionBlocked]) != 0 || len(result[networkConditionSlow]) != 0 {
			return false, nil
		}
		return true, nil
	})

	framework.ExpectEqual(len(result[networkConditionBlocked]), 0)
	framework.ExpectEqual(len(result[networkConditionSlow]), 0)

	networkDelayWithTarget := makeNetworkDelayChaos(
		ns, "network-chaos-1",
		map[string]string{"app": "network-peer-0"},
		map[string]string{"app": "network-peer-1"}, // 0 -> 1 add delays
		v1alpha1.OnePodMode,
		v1alpha1.OnePodMode,
		v1alpha1.To,
		testDelayTcParam,
		testDelayDuration,
	)

	By("Injecting delay for 0 -> 1")
	err = cli.Create(ctx, networkDelayWithTarget.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")

	wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		result = probeNetworkCondition(c, networkPeers, ports, false)
		if len(result[networkConditionBlocked]) != 0 || len(result[networkConditionSlow]) != 1 {
			return false, nil
		}
		return true, nil
	})

	framework.ExpectEqual(len(result[networkConditionBlocked]), 0)
	framework.ExpectEqual(result[networkConditionSlow], [][]int{{0, 1}})

	err = cli.Delete(ctx, networkDelayWithTarget.DeepCopy())
	framework.ExpectNoError(err, "delete network chaos error")

	wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		result = probeNetworkCondition(c, networkPeers, ports, false)
		if len(result[networkConditionBlocked]) != 0 || len(result[networkConditionSlow]) != 0 {
			return false, nil
		}
		return true, nil
	})

	framework.ExpectEqual(len(result[networkConditionBlocked]), 0)
	framework.ExpectEqual(len(result[networkConditionSlow]), 0)

	evenNetworkDelay := makeNetworkDelayChaos(
		ns, "network-chaos-2",
		map[string]string{"app": "network-peer-0"},
		map[string]string{"partition": "0"}, // 0 -> even its partition (idx % 2)
		v1alpha1.OnePodMode,
		v1alpha1.AllPodMode,
		v1alpha1.To,
		testDelayTcParam,
		testDelayDuration,
	)
	By("Injecting delay for 0 -> even partition")
	err = cli.Create(ctx, evenNetworkDelay.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")

	wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		result = probeNetworkCondition(c, networkPeers, ports, false)
		if len(result[networkConditionBlocked]) != 0 || len(result[networkConditionSlow]) != 1 {
			return false, nil
		}
		return true, nil
	})
	framework.ExpectEqual(len(result[networkConditionBlocked]), 0)
	framework.ExpectEqual(result[networkConditionSlow], [][]int{{0, 2}})

	By("Injecting delay for 0 -> 1")
	err = cli.Create(ctx, networkDelayWithTarget.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")

	wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		result = probeNetworkCondition(c, networkPeers, ports, false)
		if len(result[networkConditionBlocked]) != 0 || len(result[networkConditionSlow]) != 2 {
			return false, nil
		}
		return true, nil
	})
	framework.ExpectEqual(len(result[networkConditionBlocked]), 0)
	framework.ExpectEqual(result[networkConditionSlow], [][]int{{0, 1}, {0, 2}})

	err = cli.Delete(ctx, networkDelayWithTarget.DeepCopy())
	framework.ExpectNoError(err, "delete network chaos error")

	wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		result = probeNetworkCondition(c, networkPeers, ports, false)
		if len(result[networkConditionBlocked]) != 0 || len(result[networkConditionSlow]) != 1 {
			return false, nil
		}
		return true, nil
	})
	framework.ExpectEqual(len(result[networkConditionBlocked]), 0)
	framework.ExpectEqual(result[networkConditionSlow], [][]int{{0, 2}})

	err = cli.Delete(ctx, evenNetworkDelay.DeepCopy())
	framework.ExpectNoError(err, "delete network chaos error")

	wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		result = probeNetworkCondition(c, networkPeers, ports, false)
		if len(result[networkConditionBlocked]) != 0 || len(result[networkConditionSlow]) != 0 {
			return false, nil
		}
		return true, nil
	})
	framework.ExpectEqual(len(result[networkConditionBlocked]), 0)
	framework.ExpectEqual(len(result[networkConditionSlow]), 0)

	complicateNetem := makeNetworkDelayChaos(
		ns, "network-chaos-3",
		map[string]string{"app": "network-peer-0"},
		nil, // no target specified
		v1alpha1.OnePodMode,
		v1alpha1.OnePodMode,
		v1alpha1.To,
		testDelayTcParamEvenMoreComplicate,
		testDelayDuration,
	)
	By("Injecting complicate chaos for 0")
	err = cli.Create(ctx, complicateNetem.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")
	wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		result = probeNetworkCondition(c, networkPeers, ports, false)
		if len(result[networkConditionBlocked]) != 0 || len(result[networkConditionSlow]) != 3 {
			return false, nil
		}
		return true, nil
	})
	framework.ExpectEqual(len(result[networkConditionBlocked]), 0)
	framework.ExpectEqual(result[networkConditionSlow], [][]int{{0, 1}, {0, 2}, {0, 3}})

	By("recover")
	err = cli.Delete(ctx, complicateNetem.DeepCopy())
	framework.ExpectNoError(err, "delete network chaos error")

	wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		result = probeNetworkCondition(c, networkPeers, ports, false)
		if len(result[networkConditionBlocked]) != 0 || len(result[networkConditionSlow]) != 0 {
			return false, nil
		}
		return true, nil
	})
	framework.ExpectEqual(len(result[networkConditionBlocked]), 0)
	framework.ExpectEqual(len(result[networkConditionSlow]), 0)

	bothDirectionNetem := makeNetworkDelayChaos(
		ns, "network-chaos-4",
		map[string]string{"app": "network-peer-0"},
		map[string]string{"partition": "0"}, // 0 -> even its partition (idx % 2)
		v1alpha1.OnePodMode,
		v1alpha1.AllPodMode,
		v1alpha1.Both,
		testDelayTcParam,
		testDelayDuration,
	)
	By("Injecting both direction chaos for 0")
	err = cli.Create(ctx, bothDirectionNetem.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")
	wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		result = probeNetworkCondition(c, networkPeers, ports, true)
		if len(result[networkConditionBlocked]) != 0 || len(result[networkConditionSlow]) != 2 {
			return false, nil
		}
		return true, nil
	})
	framework.ExpectEqual(len(result[networkConditionBlocked]), 0)
	framework.ExpectEqual(result[networkConditionSlow], [][]int{{0, 2}, {2, 0}})

	By("recover")
	err = cli.Delete(ctx, bothDirectionNetem.DeepCopy())
	framework.ExpectNoError(err, "delete network chaos error")

	wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		result = probeNetworkCondition(c, networkPeers, ports, true)
		if len(result[networkConditionBlocked]) != 0 || len(result[networkConditionSlow]) != 0 {
			return false, nil
		}
		return true, nil
	})
	framework.ExpectEqual(len(result[networkConditionBlocked]), 0)
	framework.ExpectEqual(len(result[networkConditionSlow]), 0)
}
