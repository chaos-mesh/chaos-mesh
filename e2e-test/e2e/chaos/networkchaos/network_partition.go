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
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/util"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/pkg/fixture"
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
	networkPartition := makeNetworkPartitionChaos(
		ns, "network-chaos-1",
		map[string]string{"app": "network-peer-4"},
		map[string]string{"app": "network-peer-1"},
		v1alpha1.OnePodMode,
		v1alpha1.OnePodMode,
		v1alpha1.To,
		pointer.StringPtr("9m"),
	)

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

		failed := true
		for _, record := range networkPartition.Status.ChaosStatus.Experiment.Records {
			klog.Infof("current chaos record %s phase: %s", record.Id, record.Phase)
			if strings.Contains(record.Id, "network-peer-4") && record.Phase == v1alpha1.Injected {
				failed = false
			}
		}
		return failed, nil
	})

	framework.ExpectNoError(err, "failed to waiting on not injected state with chaos")
	framework.ExpectEqual(networkPartition.Status.ChaosStatus.Experiment.DesiredPhase, v1alpha1.RunningPhase)
	// TODO: add failed event check
	//framework.ExpectEqual(strings.Contains(networkPartition.Status.ChaosStatus.FailedMessage, "it's dangerous to inject network chaos on a pod"), true)
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

	var result map[string][][]int
	wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		result = probeNetworkCondition(c, networkPeers, ports, false)
		if len(result[networkConditionBlocked]) != 0 || len(result[networkConditionSlow]) != 0 {
			return false, nil
		}
		return true, nil
	})
	framework.ExpectEqual(len(result[networkConditionBlocked]), 0)
	framework.ExpectEqual(len(result[networkConditionSlow]), 0)

	var (
		testDelayDuration = pointer.StringPtr("9m")
	)

	baseNetworkPartition := makeNetworkPartitionChaos(
		ns, "network-chaos-1",
		map[string]string{"app": "network-peer-0"},
		map[string]string{"app": "network-peer-1"},
		v1alpha1.OnePodMode,
		v1alpha1.OnePodMode,
		v1alpha1.To,
		testDelayDuration,
	)

	By("block from peer-0 to peer-1")
	err := cli.Create(ctx, baseNetworkPartition.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")

	wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		result = probeNetworkCondition(c, networkPeers, ports, false)
		if len(result[networkConditionBlocked]) != 1 || len(result[networkConditionSlow]) != 0 {
			return false, nil
		}
		return true, nil
	})
	framework.ExpectEqual(result[networkConditionBlocked], [][]int{{0, 1}})
	framework.ExpectEqual(len(result[networkConditionSlow]), 0)

	By("recover")
	err = cli.Delete(ctx, baseNetworkPartition.DeepCopy())
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

	By("block both from peer-0 to peer-1 and from peer-1 to peer-0")
	bothDirectionNetworkPartition := makeNetworkPartitionChaos(
		ns, "network-chaos-1",
		map[string]string{"app": "network-peer-0"},
		map[string]string{"app": "network-peer-1"},
		v1alpha1.OnePodMode,
		v1alpha1.OnePodMode,
		v1alpha1.Both,
		testDelayDuration,
	)
	err = cli.Create(ctx, bothDirectionNetworkPartition.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")

	wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		result = probeNetworkCondition(c, networkPeers, ports, false)
		if len(result[networkConditionBlocked]) != 2 || len(result[networkConditionSlow]) != 0 {
			return false, nil
		}
		return true, nil
	})
	framework.ExpectEqual(result[networkConditionBlocked], [][]int{{0, 1}, {1, 0}})
	framework.ExpectEqual(len(result[networkConditionSlow]), 0)

	By("recover")
	err = cli.Delete(ctx, bothDirectionNetworkPartition.DeepCopy())
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

	By("block from peer-1 to peer-0")
	fromDirectionNetworkPartition := makeNetworkPartitionChaos(
		ns, "network-chaos-1",
		map[string]string{"app": "network-peer-0"},
		map[string]string{"app": "network-peer-1"},
		v1alpha1.OnePodMode,
		v1alpha1.OnePodMode,
		v1alpha1.From,
		testDelayDuration,
	)

	err = cli.Create(ctx, fromDirectionNetworkPartition.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")

	wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		result = probeNetworkCondition(c, networkPeers, ports, false)
		if len(result[networkConditionBlocked]) != 1 || len(result[networkConditionSlow]) != 0 {
			return false, nil
		}
		return true, nil
	})
	framework.ExpectEqual(result[networkConditionBlocked], [][]int{{1, 0}})
	framework.ExpectEqual(len(result[networkConditionSlow]), 0)

	By("recover")
	err = cli.Delete(ctx, fromDirectionNetworkPartition.DeepCopy())
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

	By("network partition 1")

	bothDirectionWithPartitionNetworkPartition := makeNetworkPartitionChaos(
		ns, "network-chaos-1",
		map[string]string{"app": "network-peer-0"},
		map[string]string{"partition": "1"},
		v1alpha1.OnePodMode,
		v1alpha1.AllPodMode,
		v1alpha1.Both,
		testDelayDuration,
	)
	err = cli.Create(ctx, bothDirectionWithPartitionNetworkPartition.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")

	wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		result = probeNetworkCondition(c, networkPeers, ports, false)
		if len(result[networkConditionBlocked]) != 4 || len(result[networkConditionSlow]) != 0 {
			return false, nil
		}
		return true, nil
	})
	framework.ExpectEqual(result[networkConditionBlocked], [][]int{{0, 1}, {1, 0}, {0, 3}, {3, 0}})
	framework.ExpectEqual(len(result[networkConditionSlow]), 0)

	By("recover")
	err = cli.Delete(ctx, bothDirectionWithPartitionNetworkPartition.DeepCopy())
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

	By("multiple network partition chaos on peer-0")
	anotherNetworkPartition := makeNetworkPartitionChaos(
		ns, "network-chaos-2",
		map[string]string{"app": "network-peer-0"},
		map[string]string{"partition": "0"},
		v1alpha1.OnePodMode,
		v1alpha1.AllPodMode,
		v1alpha1.To,
		testDelayDuration,
	)
	err = cli.Create(ctx, bothDirectionWithPartitionNetworkPartition.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")
	err = cli.Create(ctx, anotherNetworkPartition.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")

	wait.Poll(time.Second, 30*time.Second, func() (done bool, err error) {
		result = probeNetworkCondition(c, networkPeers, ports, false)
		if len(result[networkConditionBlocked]) != 5 || len(result[networkConditionSlow]) != 0 {
			return false, nil
		}
		return true, nil
	})
	framework.ExpectEqual(result[networkConditionBlocked], [][]int{{0, 1}, {1, 0}, {0, 2}, {0, 3}, {3, 0}})
	framework.ExpectEqual(len(result[networkConditionSlow]), 0)

	By("recover")
	err = cli.Delete(ctx, bothDirectionWithPartitionNetworkPartition.DeepCopy())
	framework.ExpectNoError(err, "delete network chaos error")
	err = cli.Delete(ctx, anotherNetworkPartition.DeepCopy())
	framework.ExpectNoError(err, "delete network chaos error")

	wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		klog.Info("retry probeNetworkCondition")
		result = probeNetworkCondition(c, networkPeers, ports, false)
		if len(result[networkConditionBlocked]) != 0 || len(result[networkConditionSlow]) != 0 {
			return false, nil
		}
		return true, nil
	})
	framework.ExpectEqual(len(result[networkConditionBlocked]), 0)
	framework.ExpectEqual(len(result[networkConditionSlow]), 0)

	By("block from peer-0 to all")
	networkPartitionWithoutTarget := makeNetworkPartitionChaos(
		ns, "network-chaos-without-target",
		map[string]string{"app": "network-peer-0"},
		nil,
		v1alpha1.OnePodMode,
		v1alpha1.AllPodMode,
		v1alpha1.To,
		testDelayDuration,
	)
	err = cli.Create(ctx, networkPartitionWithoutTarget.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")

	wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		result = probeNetworkCondition(c, networkPeers, ports, false)
		if len(result[networkConditionBlocked]) != 5 || len(result[networkConditionSlow]) != 0 {
			return false, nil
		}
		return true, nil
	})
	// The expected behavior is to block only 0 -> 1, 0 -> 2 and 0 -> 3
	// Actually, 1 -> 0, 2 -> 0, 3 -> 0 are not blocked, but the `recvUDP`
	// request failed due to partition.
	framework.ExpectEqual(result[networkConditionBlocked], [][]int{{0, 1}, {1, 0}, {0, 2}, {2, 0}, {0, 3}, {3, 0}})
	framework.ExpectEqual(len(result[networkConditionSlow]), 0)

	By("recover")
	err = cli.Delete(ctx, networkPartitionWithoutTarget.DeepCopy())
	framework.ExpectNoError(err, "delete network chaos error")

	wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		klog.Info("retry probeNetworkCondition")
		result = probeNetworkCondition(c, networkPeers, ports, false)
		if len(result[networkConditionBlocked]) != 0 || len(result[networkConditionSlow]) != 0 {
			return false, nil
		}
		return true, nil
	})
	framework.ExpectEqual(len(result[networkConditionBlocked]), 0)
	framework.ExpectEqual(len(result[networkConditionSlow]), 0)

	By("block from peer-0 from all")
	networkPartitionWithoutTarget = makeNetworkPartitionChaos(
		ns, "network-chaos-without-target",
		map[string]string{"app": "network-peer-0"},
		nil,
		v1alpha1.OnePodMode,
		v1alpha1.AllPodMode,
		v1alpha1.From,
		testDelayDuration,
	)
	err = cli.Create(ctx, networkPartitionWithoutTarget.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")

	wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		result = probeNetworkCondition(c, networkPeers, ports, false)
		if len(result[networkConditionBlocked]) != 5 || len(result[networkConditionSlow]) != 0 {
			return false, nil
		}
		return true, nil
	})
	// The expected behavior is to block only 0 -> 1, 0 -> 2 and 0 -> 3
	// Actually, 1 -> 0, 2 -> 0, 3 -> 0 are not blocked, but the `recvUDP`
	// request failed due to partition.
	framework.ExpectEqual(result[networkConditionBlocked], [][]int{{0, 1}, {1, 0}, {0, 2}, {2, 0}, {0, 3}, {3, 0}})
	framework.ExpectEqual(len(result[networkConditionSlow]), 0)

	By("recover")
	err = cli.Delete(ctx, networkPartitionWithoutTarget.DeepCopy())
	framework.ExpectNoError(err, "delete network chaos error")

	wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		klog.Info("retry probeNetworkCondition")
		result = probeNetworkCondition(c, networkPeers, ports, false)
		if len(result[networkConditionBlocked]) != 0 || len(result[networkConditionSlow]) != 0 {
			return false, nil
		}
		return true, nil
	})
	framework.ExpectEqual(len(result[networkConditionBlocked]), 0)
	framework.ExpectEqual(len(result[networkConditionSlow]), 0)
}
