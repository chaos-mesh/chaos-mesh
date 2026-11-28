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

package trafficcontrol

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
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/utils"
)

func TestTrafficControl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Network Traffic Control")
}

var _ = Describe("Traffic control rules", func() {
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

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		source := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: sourceKey.Name},
			Status: corev1.PodStatus{Phase: corev1.PodRunning, PodIP: "192.0.2.10"}}
		target := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: targetKey.Name},
			Status: corev1.PodStatus{Phase: corev1.PodRunning, PodIP: "192.0.2.11"}}
		kube = fake.NewClientBuilder().WithScheme(scheme).WithObjects(source, target).
			WithStatusSubresource(&v1alpha1.PodNetworkChaos{}).Build()
		log := logr.Discard()
		builder := podnetworkchaosmanager.NewBuilder(podnetworkchaosmanager.Params{
			Logger: log, Client: kube, Reader: kube, Scheme: scheme,
		})
		impl = NewImpl(kube, builder, log)
		chaos = &v1alpha1.NetworkChaos{
			ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: "traffic"},
			Spec: v1alpha1.NetworkChaosSpec{
				Action: v1alpha1.DelayAction, Direction: v1alpha1.To,
				Device: "net1", TargetDevice: "net2",
				TcParameter: v1alpha1.TcParameter{Delay: &v1alpha1.DelaySpec{Latency: "20ms"}},
			},
		}
		records = []*v1alpha1.Record{
			{Id: namespace + "/source", SelectorKey: ".", Phase: v1alpha1.NotInjected},
			{Id: namespace + "/target", SelectorKey: ".Target", Phase: v1alpha1.NotInjected},
		}
	})

	It("uses the source device and its peer filter for outbound changes", func() {
		Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(waitForApplySync))
		Expect(impl.Apply(ctx, 1, records, chaos)).To(Equal(v1alpha1.Injected))
		child := &v1alpha1.PodNetworkChaos{}
		Expect(apierrors.IsNotFound(kube.Get(ctx, targetKey, child))).To(BeTrue())
		Expect(kube.Get(ctx, sourceKey, child)).To(Succeed())
		Expect(child.Spec.TrafficControls).To(HaveLen(1))
		rule := child.Spec.TrafficControls[0]
		Expect(rule.Device).To(Equal("net1"))
		Expect(rule.Type).To(Equal(v1alpha1.Netem))
		Expect(rule.Delay).To(Equal(chaos.Spec.Delay))
		Expect(rule.IPSet).NotTo(BeEmpty())
		Expect(child.Spec.IPSets).To(ContainElement(HaveField("Name", rule.IPSet)))
		Expect(child.Spec.IPSets).To(ContainElement(HaveField("Cidrs", ContainElement("192.0.2.11/32"))))

		child.Status.FailedMessage = "daemon update failed"
		Expect(kube.Status().Update(ctx, child)).To(Succeed())
		records[0].Phase = waitForApplySync
		phase, err := impl.Apply(ctx, 0, records, chaos)
		Expect(err).To(MatchError("daemon update failed"))
		Expect(phase).To(Equal(waitForApplySync))
		child.Status.FailedMessage = ""
		Expect(kube.Status().Update(ctx, child)).To(Succeed())
		Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(v1alpha1.Injected))
	})

	It("uses the target device for inbound bandwidth control and removes the rule on recovery", func() {
		chaos.Spec.Action = v1alpha1.BandwidthAction
		chaos.Spec.Direction = v1alpha1.From
		chaos.Spec.TcParameter = v1alpha1.TcParameter{Bandwidth: &v1alpha1.BandwidthSpec{Rate: "1mbps"}}
		Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(v1alpha1.Injected))
		Expect(impl.Apply(ctx, 1, records, chaos)).To(Equal(waitForApplySync))
		child := &v1alpha1.PodNetworkChaos{}
		Expect(apierrors.IsNotFound(kube.Get(ctx, sourceKey, child))).To(BeTrue())
		Expect(kube.Get(ctx, targetKey, child)).To(Succeed())
		Expect(child.Spec.TrafficControls).To(HaveLen(1))
		rule := child.Spec.TrafficControls[0]
		Expect(rule.Device).To(Equal("net2"))
		Expect(rule.Type).To(Equal(v1alpha1.Bandwidth))
		Expect(rule.Bandwidth).To(Equal(chaos.Spec.Bandwidth))
		Expect(child.Spec.IPSets).To(ContainElement(HaveField("Cidrs", ContainElement("192.0.2.10/32"))))

		records[1].Phase = v1alpha1.Injected
		Expect(impl.Recover(ctx, 1, records, chaos)).To(Equal(waitForRecoverSync))
		Expect(kube.Get(ctx, targetKey, child)).To(Succeed())
		Expect(child.Spec.TrafficControls).To(BeEmpty())
		Expect(child.Spec.IPSets).To(BeEmpty())
		records[1].Phase = waitForRecoverSync
		Expect(impl.Recover(ctx, 1, records, chaos)).To(Equal(v1alpha1.NotInjected))
	})

	It("rejects unsupported actions before creating a child", func() {
		chaos.Spec.Action = v1alpha1.PartitionAction
		phase, err := impl.Apply(ctx, 0, records, chaos)
		Expect(err).To(MatchError(utils.ErrUnknownAction))
		Expect(phase).To(Equal(v1alpha1.NotInjected))
		Expect(apierrors.IsNotFound(kube.Get(ctx, sourceKey, &v1alpha1.PodNetworkChaos{}))).To(BeTrue())
	})
})
