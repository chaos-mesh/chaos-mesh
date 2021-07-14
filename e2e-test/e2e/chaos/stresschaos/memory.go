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
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubernetes/test/e2e/framework"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestcaseMemoryStressInjectionOnceThenRecover(
	ns string,
	cli client.Client,
	peers []*corev1.Pod,
	ports []uint16,
	c http.Client,
) {
	// it will raise two pod: stress-peer-0 and stress-peer-1, and we only inject StressChaos into stress-peer-0

	// approximate equality within 5 MBytes (10% of injected 50M memory chaos)
	const allowedJitter = 5 * 1024 * 1024

	ctx := context.Background()
	By("create memory stress chaos CRD objects")
	memoryStressChaos := makeMemoryStressChaos(ns, "memory-stress", ns, "stress-peer-0", "50M", 1)
	err := cli.Create(ctx, memoryStressChaos.DeepCopy())
	framework.ExpectNoError(err, "create stresschaos error")

	By("waiting for assertion some pods are experiencing memory stress ")
	err = wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		conditions, err := probeStressCondition(c, peers, ports)
		if err != nil {
			return false, err
		}
		framework.Logf("get Memory: [%d, %d]", conditions[0].MemoryUsage, conditions[1].MemoryUsage)
		if int(conditions[0].MemoryUsage)-int(conditions[1].MemoryUsage) > allowedJitter {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "memory stress failed")
	By("delete pod failure chaos CRD objects")

	err = cli.Delete(ctx, memoryStressChaos.DeepCopy())
	framework.ExpectNoError(err, "delete stresschaos error")
	By("waiting for assertion recovering")
	err = wait.Poll(time.Second, 15*time.Second, func() (done bool, err error) {
		conditions, err := probeStressCondition(c, peers, ports)
		if err != nil {
			return false, err
		}
		By(fmt.Sprintf("get Memory: [%d, %d]", conditions[0].MemoryUsage, conditions[1].MemoryUsage))
		if conditions[0].MemoryUsage < conditions[1].MemoryUsage+allowedJitter {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "fail to recover from memory stress")
}
