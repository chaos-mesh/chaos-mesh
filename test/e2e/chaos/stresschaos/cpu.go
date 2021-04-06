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

package stresschaos

import (
	"context"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubernetes/test/e2e/framework"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestcaseCPUStressInjectionOnceThenRecover(
	ns string,
	cli client.Client,
	peers []*corev1.Pod,
	ports []uint16,
	c http.Client,
) {
	ctx := context.Background()
	By("create cpu stress chaos CRD objects")
	cpuStressChaos := makeCPUStressChaos(ns, "cpu-stress", ns, "stress-peer-0", 1, 100)
	err := cli.Create(ctx, cpuStressChaos.DeepCopy())
	framework.ExpectNoError(err, "create stresschaos error")

	lastCPUTime := make([]uint64, 2)
	diff := make([]uint64, 2)
	By("waiting for assertion some pods are experiencing cpu stress ")
	err = wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		conditions, err := probeStressCondition(c, peers, ports)
		if err != nil {
			return false, err
		}

		diff[0] = conditions[0].CpuTime - lastCPUTime[0]
		diff[1] = conditions[1].CpuTime - lastCPUTime[1]
		lastCPUTime[0] = conditions[0].CpuTime
		lastCPUTime[1] = conditions[1].CpuTime
		framework.Logf("get CPU: [%d, %d]", diff[0], diff[1])
		// diff means the increasing CPU time (in nanosecond)
		// just pick two threshold, 5e8 is a little shorter than one second
		if diff[0] > 5e8 && diff[1] < 5e6 {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "cpu stress failed")
	By("delete pod failure chaos CRD objects")

	err = cli.Delete(ctx, cpuStressChaos.DeepCopy())
	framework.ExpectNoError(err, "delete stresschaos error")
	By("waiting for assertion recovering")
	lastCPUTime = make([]uint64, 2)
	diff = make([]uint64, 2)
	err = wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		conditions, err := probeStressCondition(c, peers, ports)
		if err != nil {
			return false, err
		}

		diff[0] = conditions[0].CpuTime - lastCPUTime[0]
		diff[1] = conditions[1].CpuTime - lastCPUTime[1]
		lastCPUTime[0] = conditions[0].CpuTime
		lastCPUTime[1] = conditions[1].CpuTime
		framework.Logf("get CPU: [%d, %d]", diff[0], diff[1])
		// diff means the increasing CPU time (in nanosecond)
		// just pick two threshold, they are both much shorter than 1 second
		if diff[0] < 1e7 && diff[1] < 5e6 {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "fail to recover from cpu stress")
}
