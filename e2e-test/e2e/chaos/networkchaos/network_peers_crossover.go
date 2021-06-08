// Copyright 2021 Chaos Mesh Authors.
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
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/util"
)

// This test case is for https://github.com/chaos-mesh/chaos-mesh/issues/1450
// For example, if the source is A, B, and the target is C, D, and the direction is both,
// now the connection between A and B will also be affected by this chaos, this is unexpected.
func TestcasePeersCrossover(
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
	)

	By("injecting network chaos between partition 0 and 1")
	networkDelay := makeNetworkDelayChaos(
		ns, "network-chaos-1",
		map[string]string{"partition": "0"},
		map[string]string{"partition": "1"},
		v1alpha1.AllPodMode,
		v1alpha1.AllPodMode,
		v1alpha1.Both,
		testDelayTcParam,
		nil,
	)
	// that's important
	networkDelay.Spec.Direction = v1alpha1.Both

	By("Injecting delay between partition 0 (peer 0,2) with partition 1 (peer 1,3)")
	err := cli.Create(ctx, networkDelay.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")

	err = wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		result = probeNetworkCondition(c, networkPeers, ports, false)
		if len(result[networkConditionBlocked]) != 0 || len(result[networkConditionSlow]) != 4 {
			return false, nil
		}
		return true, nil
	})

	framework.ExpectNoError(err, "failed to waiting condition for chaos injection")
	framework.ExpectEqual(len(result[networkConditionBlocked]), 0)
	framework.ExpectEqual(result[networkConditionSlow], [][]int{{0, 1}, {0, 3}, {1, 2}, {2, 3}})

	By("recover")
	err = cli.Delete(ctx, networkDelay.DeepCopy())
	framework.ExpectNoError(err, "delete network chaos error")

	err = wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		result = probeNetworkCondition(c, networkPeers, ports, false)
		if len(result[networkConditionBlocked]) != 0 || len(result[networkConditionSlow]) != 0 {
			return false, nil
		}
		return true, nil
	})

	framework.ExpectNoError(err, "failed to waiting condition for chaos recover")
	framework.ExpectEqual(len(result[networkConditionBlocked]), 0)
	framework.ExpectEqual(len(result[networkConditionSlow]), 0)

}
