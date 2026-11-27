// Copyright Chaos Mesh Authors.
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

package partition

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/networkchaos/podnetworkchaosmanager"
)

func TestPartition(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Network Partition")
}

var _ = Describe("Partition rules", func() {
	ctx := context.Background()
	const namespace = "test"
	sourceKey := client.ObjectKey{Namespace: namespace, Name: "source"}
	targetKey := client.ObjectKey{Namespace: namespace, Name: "target"}
	var (
		kube    client.Client
		impl    *Impl
		chaos   *v1alpha1.NetworkChaos
		records []*v1alpha1.Record
	)

	workingPod := func(key client.ObjectKey, ip string) *corev1.Pod {
		return &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: key.Name},
			Status:     corev1.PodStatus{Phase: corev1.PodRunning, PodIP: ip},
		}
	}

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		kube = fake.NewClientBuilder().WithScheme(scheme).
			WithObjects(workingPod(sourceKey, "192.0.2.10"), workingPod(targetKey, "192.0.2.11")).
			WithStatusSubresource(&v1alpha1.PodNetworkChaos{}).Build()
		log := logr.Discard()
		builder := podnetworkchaosmanager.NewBuilder(podnetworkchaosmanager.Params{
			Logger: log, Client: kube, Reader: kube, Scheme: scheme,
		})
		impl = NewImpl(kube, builder, log)
		chaos = &v1alpha1.NetworkChaos{
			ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: "partition"},
			Spec: v1alpha1.NetworkChaosSpec{
				Action: v1alpha1.PartitionAction, Direction: v1alpha1.Both,
				Device: "net1", TargetDevice: "net2", PartitionBehavior: v1alpha1.RejectPartitionBehavior,
			},
		}
		records = []*v1alpha1.Record{
			{Id: namespace + "/source", SelectorKey: ".", Phase: v1alpha1.NotInjected},
			{Id: namespace + "/target", SelectorKey: ".Target", Phase: v1alpha1.NotInjected},
		}
	})

	It("installs filtered rules independently on both sides and does not duplicate them on retry", func() {
		for range 2 {
			Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(waitForApplySync))
			Expect(impl.Apply(ctx, 1, records, chaos)).To(Equal(waitForApplySync))
		}
		for _, side := range []struct {
			key        client.ObjectKey
			device     string
			remoteCIDR string
		}{
			{sourceKey, "net1", "192.0.2.11/32"},
			{targetKey, "net2", "192.0.2.10/32"},
		} {
			child := &v1alpha1.PodNetworkChaos{}
			Expect(kube.Get(ctx, side.key, child)).To(Succeed())
			Expect(child.Spec.Iptables).To(HaveLen(2))
			Expect(child.Spec.Iptables[0].Direction).To(Equal(v1alpha1.Output))
			Expect(child.Spec.Iptables[1].Direction).To(Equal(v1alpha1.Input))
			for _, rule := range child.Spec.Iptables {
				Expect(rule.Device).To(Equal(side.device))
				Expect(rule.PartitionBehavior).To(Equal(v1alpha1.RejectPartitionBehavior))
				Expect(rule.IPSets).To(HaveLen(1))
			}
			Expect(child.Spec.IPSets).To(ContainElement(HaveField("Cidrs", ContainElement(side.remoteCIDR))))
		}
	})

	It("preserves source rules when a pod matches both selectors", func() {
		records[1].Id = records[0].Id
		Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(waitForApplySync))
		Expect(impl.Apply(ctx, 1, records, chaos)).To(Equal(waitForApplySync))
		child := &v1alpha1.PodNetworkChaos{}
		Expect(kube.Get(ctx, sourceKey, child)).To(Succeed())
		Expect(child.Spec.Iptables).To(HaveLen(4))
		Expect(child.Spec.Iptables[0].Device).To(Equal("net1"))
		Expect(child.Spec.Iptables[2].Device).To(Equal("net2"))
	})

	It("clears only rules owned by this experiment during recovery", func() {
		otherRule := v1alpha1.RawIptables{RawRuleSource: v1alpha1.RawRuleSource{Source: "test/other"}, Device: "eth0"}
		child := &v1alpha1.PodNetworkChaos{
			ObjectMeta: metav1.ObjectMeta{Namespace: sourceKey.Namespace, Name: sourceKey.Name, Generation: 2},
			Spec:       v1alpha1.PodNetworkChaosSpec{Iptables: []v1alpha1.RawIptables{otherRule}},
		}
		Expect(kube.Create(ctx, child)).To(Succeed())
		Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(waitForApplySync))
		records[0].Phase = waitForApplySync
		Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(waitForApplySync))
		Expect(kube.Get(ctx, sourceKey, child)).To(Succeed())
		child.Status.ObservedGeneration = child.Generation
		Expect(kube.Status().Update(ctx, child)).To(Succeed())
		Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(v1alpha1.Injected))

		records[0].Phase = v1alpha1.Injected
		Expect(impl.Recover(ctx, 0, records, chaos)).To(Equal(waitForRecoverSync))
		Expect(kube.Get(ctx, sourceKey, child)).To(Succeed())
		Expect(child.Spec.Iptables).To(Equal([]v1alpha1.RawIptables{otherRule}))
		Expect(child.Spec.IPSets).To(BeEmpty())
		Expect(kube.Delete(ctx, child)).To(Succeed())
		records[0].Phase = waitForRecoverSync
		Expect(impl.Recover(ctx, 0, records, chaos)).To(Equal(v1alpha1.NotInjected))
	})

	It("does not persist a partial partition when a selected peer no longer exists", func() {
		records[1].Id = namespace + "/missing"
		phase, err := impl.Apply(ctx, 0, records, chaos)
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
		Expect(phase).To(Equal(v1alpha1.NotInjected))
		Expect(apierrors.IsNotFound(kube.Get(ctx, sourceKey, &v1alpha1.PodNetworkChaos{}))).To(BeTrue())
	})

	It("waits for the daemon to acknowledge the recovery generation after an error", func() {
		other := v1alpha1.RawIptables{RawRuleSource: v1alpha1.RawRuleSource{Source: "test/other"}, Device: "eth0"}
		child := &v1alpha1.PodNetworkChaos{
			ObjectMeta: metav1.ObjectMeta{Namespace: sourceKey.Namespace, Name: sourceKey.Name, Generation: 5},
			Spec: v1alpha1.PodNetworkChaosSpec{Iptables: []v1alpha1.RawIptables{
				other, {RawRuleSource: v1alpha1.RawRuleSource{Source: "test/partition"}, Device: "net1"},
			}},
		}
		Expect(kube.Create(ctx, child)).To(Succeed())
		records[0].Phase = v1alpha1.Injected
		Expect(impl.Recover(ctx, 0, records, chaos)).To(Equal(waitForRecoverSync))
		records[0].Phase = waitForRecoverSync
		Expect(kube.Get(ctx, sourceKey, child)).To(Succeed())
		wantGeneration := chaos.Status.Instances[records[0].Id]
		Expect(wantGeneration).To(BeNumerically(">", 0))
		Expect(child.Spec.Iptables).To(Equal([]v1alpha1.RawIptables{other}))

		child.Status.ObservedGeneration = wantGeneration
		child.Status.FailedMessage = "daemon could not commit the rules"
		Expect(kube.Status().Update(ctx, child)).To(Succeed())
		phase, err := impl.Recover(ctx, 0, records, chaos)
		Expect(err).To(MatchError("daemon could not commit the rules"))
		Expect(phase).To(Equal(waitForRecoverSync))

		child.Status.FailedMessage = ""
		child.Status.ObservedGeneration = wantGeneration - 1
		Expect(kube.Status().Update(ctx, child)).To(Succeed())
		Expect(impl.Recover(ctx, 0, records, chaos)).To(Equal(waitForRecoverSync))
		child.Status.ObservedGeneration = wantGeneration
		Expect(kube.Status().Update(ctx, child)).To(Succeed())
		Expect(impl.Recover(ctx, 0, records, chaos)).To(Equal(v1alpha1.NotInjected))

		Expect(kube.Get(ctx, sourceKey, child)).To(Succeed())
		Expect(child.Spec.Iptables).To(Equal([]v1alpha1.RawIptables{other}))
	})

	It("keeps an invalid target ID in its recovery phase", func() {
		records[0].Id = "invalid"
		for _, expected := range []v1alpha1.Phase{v1alpha1.Injected, waitForRecoverSync} {
			records[0].Phase = expected
			phase, err := impl.Recover(ctx, 0, records, chaos)
			Expect(err).To(HaveOccurred())
			Expect(phase).To(Equal(expected))
		}
	})
})
