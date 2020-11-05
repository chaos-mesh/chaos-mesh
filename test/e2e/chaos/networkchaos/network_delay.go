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
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/test/e2e/util"
	. "github.com/onsi/ginkgo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/utils/pointer"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
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

	testDelay := func(from int, to int) int64 {
		delay, err := testNetworkDelay(c, ports[from], networkPeers[to].Status.PodIP)
		framework.ExpectNoError(err, "send request to test delay failed")

		return delay
	}
	allSlowConnection := func() [][]int {
		var result [][]int
		for source := 0; source < len(networkPeers); source++ {
			for target := source + 1; target < len(networkPeers); target++ {
				delay := testDelay(source, target)
				klog.Infof("delay from %d to %d: %d", source, target, delay)
				if delay > 100*1e6 {
					result = append(result, []int{source, target})
				}
			}
		}

		return result
	}

	framework.ExpectEqual(len(allSlowConnection()), 0)

	By("normal delay chaos")
	networkDelay := &v1alpha1.NetworkChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "network-chaos-1",
			Namespace: ns,
		},
		Spec: v1alpha1.NetworkChaosSpec{
			Action: v1alpha1.DelayAction,
			Selector: v1alpha1.SelectorSpec{
				Namespaces:     []string{ns},
				LabelSelectors: map[string]string{"app": "network-peer-0"},
			},
			Mode: v1alpha1.OnePodMode,
			TcParameter: v1alpha1.TcParameter{
				Delay: &v1alpha1.DelaySpec{
					Latency:     "200ms",
					Correlation: "25",
					Jitter:      "0ms",
				},
			},
			Duration: pointer.StringPtr("9m"),
			Scheduler: &v1alpha1.SchedulerSpec{
				Cron: "@every 10m",
			},
		},
	}
	By("Injecting delay for 0")
	err := cli.Create(ctx, networkDelay.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")
	time.Sleep(5 * time.Second)
	framework.ExpectEqual(allSlowConnection(), [][]int{{0, 1}, {0, 2}, {0, 3}})

	By("recover")
	err = cli.Delete(ctx, networkDelay.DeepCopy())
	framework.ExpectNoError(err, "delete network chaos error")
	time.Sleep(5 * time.Second)
	framework.ExpectEqual(len(allSlowConnection()), 0)

	networkDelay.Spec.Target = &v1alpha1.Target{
		TargetSelector: v1alpha1.SelectorSpec{
			Namespaces:     []string{ns},
			LabelSelectors: map[string]string{"app": "network-peer-1"},
		},
		TargetMode: v1alpha1.OnePodMode,
	}
	By("Injecting delay for 0 -> 1")
	err = cli.Create(ctx, networkDelay.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")
	time.Sleep(5 * time.Second)
	framework.ExpectEqual(allSlowConnection(), [][]int{{0, 1}})

	err = cli.Delete(ctx, networkDelay.DeepCopy())
	framework.ExpectNoError(err, "delete network chaos error")
	time.Sleep(5 * time.Second)
	framework.ExpectEqual(len(allSlowConnection()), 0)

	evenNetworkDelay := networkDelay.DeepCopy()
	evenNetworkDelay.Name = "network-chaos-2"
	evenNetworkDelay.Spec.Target.TargetSelector.LabelSelectors = map[string]string{"partition": "0"}
	evenNetworkDelay.Spec.Target.TargetMode = v1alpha1.AllPodMode
	By("Injecting delay for 0 -> even partition")
	err = cli.Create(ctx, evenNetworkDelay.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")
	time.Sleep(5 * time.Second)
	framework.ExpectEqual(allSlowConnection(), [][]int{{0, 2}})

	By("Injecting delay for 0 -> 1")
	err = cli.Create(ctx, networkDelay.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")
	time.Sleep(5 * time.Second)
	framework.ExpectEqual(allSlowConnection(), [][]int{{0, 1}, {0, 2}})

	err = cli.Delete(ctx, networkDelay.DeepCopy())
	framework.ExpectNoError(err, "delete network chaos error")
	time.Sleep(5 * time.Second)
	framework.ExpectEqual(allSlowConnection(), [][]int{{0, 2}})
	err = cli.Delete(ctx, evenNetworkDelay.DeepCopy())
	framework.ExpectNoError(err, "delete network chaos error")
	time.Sleep(5 * time.Second)
	framework.ExpectEqual(len(allSlowConnection()), 0)

	complicateNetem := &v1alpha1.NetworkChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "network-chaos-3",
			Namespace: ns,
		},
		Spec: v1alpha1.NetworkChaosSpec{
			Action: v1alpha1.DelayAction,
			Selector: v1alpha1.SelectorSpec{
				Namespaces:     []string{ns},
				LabelSelectors: map[string]string{"app": "network-peer-0"},
			},
			Mode: v1alpha1.OnePodMode,
			TcParameter: v1alpha1.TcParameter{
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
			},
			Duration: pointer.StringPtr("9m"),
			Scheduler: &v1alpha1.SchedulerSpec{
				Cron: "@every 10m",
			},
		},
	}
	By("Injecting complicate chaos for 0")
	err = cli.Create(ctx, complicateNetem.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")
	time.Sleep(5 * time.Second)
	framework.ExpectEqual(allSlowConnection(), [][]int{{0, 1}, {0, 2}, {0, 3}})
}
