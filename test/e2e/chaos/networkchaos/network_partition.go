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
	"github.com/chaos-mesh/chaos-mesh/test/pkg/fixture"
	. "github.com/onsi/ginkgo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/utils/pointer"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

// TestcaseForbidHostNetwork We do NOT allow that inject chaos on a pod which uses hostNetwork
func TestcaseForbidHostNetwork(
	ns string,
	kubeCli kubernetes.Interface,
	cli client.Client,
) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	By("preparing experiment pods")
	name := "network-peer-4"
	nd := fixture.NewNetworkTestDeployment(name, ns, map[string]string{"partition": "0"})
	nd.Spec.Template.Spec.HostNetwork = true
	_, err := kubeCli.AppsV1().Deployments(ns).Create(nd)
	framework.ExpectNoError(err, "create network-peer deployment error")
	err = util.WaitDeploymentReady(name, ns, kubeCli)
	framework.ExpectNoError(err, "wait network-peer deployment ready error")

	By("create network partition chaos CRD objects")
	networkPartition := &v1alpha1.NetworkChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "network-chaos-1",
			Namespace: ns,
		},
		Spec: v1alpha1.NetworkChaosSpec{
			Action: v1alpha1.PartitionAction,
			Selector: v1alpha1.SelectorSpec{
				Namespaces:     []string{ns},
				LabelSelectors: map[string]string{"app": "network-peer-4"},
			},
			Mode:      v1alpha1.OnePodMode,
			Direction: v1alpha1.To,
			Target: &v1alpha1.Target{
				TargetSelector: v1alpha1.SelectorSpec{
					Namespaces:     []string{ns},
					LabelSelectors: map[string]string{"app": "network-peer-1"},
				},
				TargetMode: v1alpha1.OnePodMode,
			},
			Duration: pointer.StringPtr("9m"),
			Scheduler: &v1alpha1.SchedulerSpec{
				Cron: "@every 10m",
			},
		},
	}

	err = cli.Create(ctx, networkPartition.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")

	By("waiting for rejecting for network chaos with hostNetwork")
	err = wait.Poll(5*time.Second, 1*time.Minute, func() (done bool, err error) {
		err = cli.Get(ctx, types.NamespacedName{
			Namespace: ns,
			Name:      "network-chaos-1",
		}, networkPartition)
		if err != nil {
			return false, err
		}
		experimentPhase := networkPartition.Status.ChaosStatus.Experiment.Phase
		klog.Infof("current chaos phase: %s", experimentPhase)
		if experimentPhase == v1alpha1.ExperimentPhaseFailed {
			return true, nil
		}
		return false, nil
	})

	framework.ExpectNoError(err, "failed to waiting on ExperimentPhaseFailed state with chaos")
	framework.ExpectEqual(networkPartition.Status.ChaosStatus.Experiment.Phase, v1alpha1.ExperimentPhaseFailed)
	framework.ExpectEqual(strings.Contains(networkPartition.Status.ChaosStatus.FailedMessage, "it's dangerous to inject network chaos on a pod"), true)
}

func TestcaseNetworkPartition(
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
	connect := func(source, target int) bool {
		err := sendUDPPacket(c, ports[source], networkPeers[target].Status.PodIP)
		if err != nil {
			klog.Infof("Error: %v", err)
			return false
		}

		data, err := recvUDPPacket(c, ports[target])
		if err != nil || data != "ping\n" {
			klog.Infof("Error: %v, Data: %s", err, data)
			return false
		}

		return true
	}
	allBlockedConnection := func() [][]int {
		var result [][]int
		for source := range networkPeers {
			for target := range networkPeers {
				if source == target {
					continue
				}

				if !connect(source, target) {
					result = append(result, []int{source, target})
				}
			}
		}

		return result
	}
	framework.ExpectEqual(len(allBlockedConnection()), 0)

	baseNetworkPartition := &v1alpha1.NetworkChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "network-chaos-1",
			Namespace: ns,
		},
		Spec: v1alpha1.NetworkChaosSpec{
			Action: v1alpha1.PartitionAction,
			Selector: v1alpha1.SelectorSpec{
				Namespaces:     []string{ns},
				LabelSelectors: map[string]string{"app": "network-peer-0"},
			},
			Mode:      v1alpha1.OnePodMode,
			Direction: v1alpha1.To,
			Target: &v1alpha1.Target{
				TargetSelector: v1alpha1.SelectorSpec{
					Namespaces:     []string{ns},
					LabelSelectors: map[string]string{"app": "network-peer-1"},
				},
				TargetMode: v1alpha1.OnePodMode,
			},
			Duration: pointer.StringPtr("9m"),
			Scheduler: &v1alpha1.SchedulerSpec{
				Cron: "@every 10m",
			},
		},
	}

	By("block from peer-0 to peer-1")
	err := cli.Create(ctx, baseNetworkPartition.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")
	time.Sleep(5 * time.Second)
	framework.ExpectEqual(allBlockedConnection(), [][]int{{0, 1}})

	By("recover")
	err = cli.Delete(ctx, baseNetworkPartition.DeepCopy())
	framework.ExpectNoError(err, "delete network chaos error")
	time.Sleep(5 * time.Second)
	framework.ExpectEqual(len(allBlockedConnection()), 0)

	By("block both from peer-0 to peer-1 and from peer-1 to peer-0")
	baseNetworkPartition.Spec.Direction = v1alpha1.Both
	err = cli.Create(ctx, baseNetworkPartition.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")
	time.Sleep(5 * time.Second)
	framework.ExpectEqual(allBlockedConnection(), [][]int{{0, 1}, {1, 0}})

	By("recover")
	err = cli.Delete(ctx, baseNetworkPartition.DeepCopy())
	framework.ExpectNoError(err, "delete network chaos error")
	time.Sleep(5 * time.Second)
	framework.ExpectEqual(len(allBlockedConnection()), 0)

	By("block from peer-1 to peer-0")
	baseNetworkPartition.Spec.Direction = v1alpha1.From
	err = cli.Create(ctx, baseNetworkPartition.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")
	time.Sleep(5 * time.Second)
	framework.ExpectEqual(allBlockedConnection(), [][]int{{1, 0}})

	By("recover")
	err = cli.Delete(ctx, baseNetworkPartition.DeepCopy())
	framework.ExpectNoError(err, "delete network chaos error")
	time.Sleep(5 * time.Second)
	framework.ExpectEqual(len(allBlockedConnection()), 0)

	By("network partition 1")
	baseNetworkPartition.Spec.Direction = v1alpha1.Both
	baseNetworkPartition.Spec.Target.TargetSelector.LabelSelectors = map[string]string{"partition": "1"}
	baseNetworkPartition.Spec.Target.TargetMode = v1alpha1.AllPodMode
	err = cli.Create(ctx, baseNetworkPartition.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")
	time.Sleep(5 * time.Second)
	framework.ExpectEqual(allBlockedConnection(), [][]int{{0, 1}, {0, 3}, {1, 0}, {3, 0}})

	By("recover")
	err = cli.Delete(ctx, baseNetworkPartition.DeepCopy())
	framework.ExpectNoError(err, "delete network chaos error")
	time.Sleep(5 * time.Second)
	framework.ExpectEqual(len(allBlockedConnection()), 0)

	By("multiple network partition chaos on peer-0")
	anotherNetworkPartition := baseNetworkPartition.DeepCopy()
	anotherNetworkPartition.Name = "network-chaos-2"
	anotherNetworkPartition.Spec.Direction = v1alpha1.To
	anotherNetworkPartition.Spec.Target.TargetSelector.LabelSelectors = map[string]string{"partition": "0"}
	anotherNetworkPartition.Spec.Target.TargetMode = v1alpha1.AllPodMode
	err = cli.Create(ctx, baseNetworkPartition.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")
	err = cli.Create(ctx, anotherNetworkPartition.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")
	time.Sleep(5 * time.Second)
	framework.ExpectEqual(allBlockedConnection(), [][]int{{0, 1}, {0, 2}, {0, 3}, {1, 0}, {3, 0}})

	By("recover")
	err = cli.Delete(ctx, baseNetworkPartition.DeepCopy())
	framework.ExpectNoError(err, "delete network chaos error")
	err = cli.Delete(ctx, anotherNetworkPartition.DeepCopy())
	framework.ExpectNoError(err, "delete network chaos error")
	time.Sleep(5 * time.Second)
	framework.ExpectEqual(len(allBlockedConnection()), 0)

}
