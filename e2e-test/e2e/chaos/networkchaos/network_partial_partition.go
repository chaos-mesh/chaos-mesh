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

package networkchaos

import (
	"context"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/util"
)

const (
	partialDropSamples = 200
	partialDropPercent = "50"
	// With 200 samples at p=0.5 the standard deviation is ~7 packets, so the
	// band below is roughly +/- 4.5 sigma.
	partialDropMinFailures = 70
	partialDropMaxFailures = 130
)

// countSendFailures sends UDP packets from one peer to another and counts how
// many were rejected on the sender. Packets dropped by an OUTPUT chain fail
// immediately with "operation not permitted", so no receive round-trip is needed.
func countSendFailures(c http.Client, sourcePort uint16, targetIP string, samples int) int {
	failures := 0
	for i := 0; i < samples; i++ {
		if err := sendUDPPacket(c, sourcePort, targetIP); err != nil {
			failures++
		}
	}
	return failures
}

func TestcaseNetworkPartialPartition(
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
	gomega.Expect(countSendFailures(c, ports[0], networkPeers[1].Status.PodIP, 10)).To(gomega.BeZero())

	By("drop half of the packets from peer-0 to peer-1")
	partialPartition := makeNetworkPartitionChaos(
		ns, "network-chaos-partial",
		map[string]string{"app": "network-peer-0"},
		map[string]string{"app": "network-peer-1"},
		v1alpha1.OneMode,
		v1alpha1.OneMode,
		v1alpha1.To,
		pointer.String("9m"),
	)
	partialPartition.Spec.Partition = &v1alpha1.PartitionSpec{DropPercent: partialDropPercent}
	err := cli.Create(ctx, partialPartition.DeepCopy())
	framework.ExpectNoError(err, "create network chaos error")

	err = wait.Poll(time.Second, 30*time.Second, func() (bool, error) {
		return countSendFailures(c, ports[0], networkPeers[1].Status.PodIP, 20) > 0, nil
	})
	framework.ExpectNoError(err, "wait for partial partition to be injected")

	failures := countSendFailures(c, ports[0], networkPeers[1].Status.PodIP, partialDropSamples)
	klog.Infof("partial partition dropped %d of %d packets", failures, partialDropSamples)
	gomega.Expect(failures).To(gomega.BeNumerically(">=", partialDropMinFailures))
	gomega.Expect(failures).To(gomega.BeNumerically("<=", partialDropMaxFailures))

	By("leave the reverse direction and other peers untouched")
	gomega.Expect(countSendFailures(c, ports[1], networkPeers[0].Status.PodIP, 20)).To(gomega.BeZero())
	gomega.Expect(countSendFailures(c, ports[0], networkPeers[2].Status.PodIP, 20)).To(gomega.BeZero())

	By("recover")
	err = cli.Delete(ctx, partialPartition.DeepCopy())
	framework.ExpectNoError(err, "delete network chaos error")

	err = wait.Poll(time.Second, 30*time.Second, func() (bool, error) {
		return countSendFailures(c, ports[0], networkPeers[1].Status.PodIP, 20) == 0, nil
	})
	framework.ExpectNoError(err, "wait for partial partition to be recovered")
}
