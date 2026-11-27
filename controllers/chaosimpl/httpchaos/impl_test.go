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

package httpchaos

import (
	"context"
	"errors"
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
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/httpchaos/podhttpchaosmanager"
	impltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
)

func TestHTTPChaos(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HTTP Chaos")
}

var _ = Describe("HTTP chaos synchronization", func() {
	ctx := context.Background()
	key := client.ObjectKey{Namespace: "test", Name: "app"}
	const targetID = "test/app"
	const source = "test/experiment"
	var (
		kube          client.Client
		impl          impltypes.ChaosImpl
		chaos         *v1alpha1.HTTPChaos
		record        *v1alpha1.Record
		records       []*v1alpha1.Record
		updateFailure error
	)

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: key.Name, UID: "pod-uid"},
			Status:     corev1.PodStatus{Phase: corev1.PodRunning},
		}
		kube = fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod).
			WithStatusSubresource(&v1alpha1.PodHttpChaos{}).
			WithInterceptorFuncs(interceptor.Funcs{
				Update: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
					if updateFailure != nil {
						return updateFailure
					}
					// The fake Kubernetes client does not increment generation on a spec update.
					obj.SetGeneration(obj.GetGeneration() + 1)
					return c.Update(ctx, obj, opts...)
				},
			}).Build()
		logger := logr.Discard()
		builder := podhttpchaosmanager.NewBuilder(podhttpchaosmanager.Params{
			Logger: logger, Client: kube, Reader: kube, Scheme: scheme,
		})
		impl = NewImpl(kube, builder, logger).Impl
		chaos = &v1alpha1.HTTPChaos{
			ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: "experiment"},
			Spec:       v1alpha1.HTTPChaosSpec{Port: 8080, Target: "Request"},
		}
		record = &v1alpha1.Record{Id: targetID, Phase: v1alpha1.NotInjected}
		records = []*v1alpha1.Record{record}
		updateFailure = nil
	})

	It("waits for both generations and preserves rules owned by other experiments", func() {
		otherRule := v1alpha1.PodHttpChaosRule{Source: "test/other", Port: 9090}
		child := &v1alpha1.PodHttpChaos{
			ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: key.Name, Generation: 3},
			Spec:       v1alpha1.PodHttpChaosSpec{Rules: []v1alpha1.PodHttpChaosRule{otherRule}},
		}
		Expect(kube.Create(ctx, child)).To(Succeed())
		chaos.Spec.TLS = &v1alpha1.PodHttpChaosTLS{}
		Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(waitForApplySync))
		Expect(kube.Get(ctx, key, child)).To(Succeed())
		applied := child.Generation
		Expect(applied).To(Equal(int64(4)))
		Expect(chaos.Status.Instances).To(HaveKeyWithValue(targetID, applied))
		Expect(child.Spec.Rules).To(HaveLen(2))
		Expect(child.Spec.Rules[0]).To(Equal(otherRule))
		Expect(child.Spec.Rules[1].Source).To(Equal(source))
		Expect(child.Spec.Rules[1].Port).To(Equal(int32(8080)))
		Expect(child.Spec.TLS).NotTo(BeNil())

		record.Phase = waitForApplySync
		Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(waitForApplySync))
		child.Status.FailedMessage = "unable to apply"
		Expect(kube.Status().Update(ctx, child)).To(Succeed())
		phase, err := impl.Apply(ctx, 0, records, chaos)
		Expect(err).To(MatchError("unable to apply"))
		Expect(phase).To(Equal(waitForApplySync))
		child.Status = v1alpha1.PodHttpChaosStatus{ObservedGeneration: applied}
		Expect(kube.Status().Update(ctx, child)).To(Succeed())
		Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(v1alpha1.Injected))

		record.Phase = v1alpha1.Injected
		updateFailure = errors.New("unable to update")
		phase, err = impl.Recover(ctx, 0, records, chaos)
		Expect(err).To(MatchError(updateFailure))
		Expect(phase).To(Equal(v1alpha1.Injected))
		Expect(kube.Get(ctx, key, child)).To(Succeed())
		Expect(child.Spec.Rules).To(HaveLen(2))

		updateFailure = nil
		Expect(impl.Recover(ctx, 0, records, chaos)).To(Equal(waitForRecoverSync))
		Expect(kube.Get(ctx, key, child)).To(Succeed())
		recovered := child.Generation
		Expect(recovered).To(Equal(applied + 1))
		Expect(chaos.Status.Instances).To(HaveKeyWithValue(targetID, recovered))
		Expect(child.Spec.Rules).To(Equal([]v1alpha1.PodHttpChaosRule{otherRule}))
		record.Phase = waitForRecoverSync
		Expect(impl.Recover(ctx, 0, records, chaos)).To(Equal(waitForRecoverSync))
		child.Status.ObservedGeneration = recovered
		Expect(kube.Status().Update(ctx, child)).To(Succeed())
		Expect(impl.Recover(ctx, 0, records, chaos)).To(Equal(v1alpha1.NotInjected))
	})

	It("creates a child owned by the target pod and accepts disappearance during recovery", func() {
		Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(waitForApplySync))
		child := &v1alpha1.PodHttpChaos{}
		Expect(kube.Get(ctx, key, child)).To(Succeed())
		Expect(child.OwnerReferences).To(HaveLen(1))
		Expect(child.OwnerReferences[0].UID).To(BeEquivalentTo("pod-uid"))
		Expect(child.Spec.Rules).To(HaveLen(1))
		Expect(child.Spec.Rules[0].Source).To(Equal(source))
		Expect(kube.Delete(ctx, child)).To(Succeed())
		record.Phase = waitForRecoverSync
		Expect(impl.Recover(ctx, 0, records, chaos)).To(Equal(v1alpha1.NotInjected))

		pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: key.Name}}
		Expect(kube.Delete(ctx, pod)).To(Succeed())
		record.Phase = v1alpha1.Injected
		Expect(impl.Recover(ctx, 0, records, chaos)).To(Equal(v1alpha1.NotInjected))
		record.Phase = v1alpha1.NotInjected
		phase, err := impl.Apply(ctx, 0, records, chaos)
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
		Expect(phase).To(Equal(v1alpha1.NotInjected))
	})
	It("finishes recovery when the target has stopped and there is no child configuration", func() {
		var pod corev1.Pod
		Expect(kube.Get(ctx, key, &pod)).To(Succeed())
		pod.Status.Phase = corev1.PodSucceeded
		Expect(kube.Status().Update(ctx, &pod)).To(Succeed())
		record.Phase = v1alpha1.Injected

		Expect(impl.Recover(ctx, 0, records, chaos)).To(Equal(v1alpha1.NotInjected))
		Expect(apierrors.IsNotFound(kube.Get(ctx, key, &v1alpha1.PodHttpChaos{}))).To(BeTrue())
	})

	It("reports a malformed target without marking it recovered", func() {
		record.Id = "invalid"
		record.Phase = v1alpha1.Injected
		phase, err := impl.Recover(ctx, 0, records, chaos)
		Expect(err).To(HaveOccurred())
		Expect(phase).To(Equal(v1alpha1.Injected))
	})

})
