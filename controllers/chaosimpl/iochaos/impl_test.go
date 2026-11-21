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

package iochaos

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
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/iochaos/podiochaosmanager"
	impltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
)

func TestIOChaos(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IO Chaos")
}

var _ = Describe("IO chaos synchronization", func() {
	const source = "test/experiment"
	const target = "test/app/main"
	key := client.ObjectKey{Namespace: "test", Name: "app"}
	ctx := context.Background()
	var (
		kube        client.Client
		impl        impltypes.ChaosImpl
		chaos       *v1alpha1.IOChaos
		pod         *corev1.Pod
		record      *v1alpha1.Record
		conflict    bool
		updateError error
	)

	BeforeEach(func() {
		conflict, updateError = false, nil
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		pod = &corev1.Pod{
			TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Pod"},
			ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: key.Name, UID: "pod-uid"},
			Status:     corev1.PodStatus{Phase: corev1.PodRunning},
		}
		kube = fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod).
			WithStatusSubresource(&v1alpha1.PodIOChaos{}).
			WithInterceptorFuncs(interceptor.Funcs{
				Create: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
					obj.SetGeneration(1)
					return c.Create(ctx, obj, opts...)
				},
				Update: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
					if updateError != nil {
						return updateError
					}
					if conflict {
						conflict = false
						latest := &v1alpha1.PodIOChaos{}
						Expect(c.Get(ctx, key, latest)).To(Succeed())
						latest.Spec.Actions = append(latest.Spec.Actions, v1alpha1.IOChaosAction{Source: "test/other"})
						latest.Generation++
						Expect(c.Update(ctx, latest)).To(Succeed())
						return apierrors.NewConflict(v1alpha1.SchemeGroupVersion.WithResource("podiochaos").GroupResource(), key.Name, errors.New("concurrent update"))
					}
					obj.SetGeneration(obj.GetGeneration() + 1)
					return c.Update(ctx, obj, opts...)
				},
			}).Build()
		logger := logr.Discard()
		builder := podiochaosmanager.NewBuilder(podiochaosmanager.Params{Logger: logger, Client: kube, Reader: kube, Scheme: scheme})
		impl = NewImpl(kube, builder, logger).Impl
		chaos = &v1alpha1.IOChaos{
			ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "experiment"},
			Spec: v1alpha1.IOChaosSpec{
				Action: v1alpha1.IoFaults, VolumePath: "/data", Path: "**/records", Percent: 25, Errno: 5,
				Methods: []v1alpha1.IoMethod{v1alpha1.Read},
			},
		}
		record = &v1alpha1.Record{Id: target, Phase: v1alpha1.NotInjected}
	})

	It("replaces only its own action on retry and waits for both applied and recovered generations", func() {
		records := []*v1alpha1.Record{record}
		Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(waitForApplySync))
		child := &v1alpha1.PodIOChaos{}
		Expect(kube.Get(ctx, key, child)).To(Succeed())
		Expect(child.OwnerReferences).To(ConsistOf(metav1.OwnerReference{
			APIVersion: "v1", Kind: "Pod", Name: key.Name, UID: pod.UID,
		}))
		Expect(child.Spec.Container).To(HaveValue(Equal("main")))
		Expect(child.Spec.VolumeMountPath).To(Equal("/data"))
		wantAction := v1alpha1.IOChaosAction{
			Type: v1alpha1.IoFaults, Filter: v1alpha1.Filter{Path: "**/records", Percent: 25, Methods: []v1alpha1.IoMethod{v1alpha1.Read}},
			Faults: []v1alpha1.IoFault{{Errno: 5, Weight: 1}}, Source: source,
		}
		Expect(child.Spec.Actions).To(Equal([]v1alpha1.IOChaosAction{wantAction}))
		Expect(chaos.Status.Instances).To(HaveKeyWithValue(target, int64(1)))

		conflict = true
		Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(waitForApplySync))
		Expect(kube.Get(ctx, key, child)).To(Succeed())
		other := v1alpha1.IOChaosAction{Source: "test/other"}
		Expect(child.Spec.Actions).To(Equal([]v1alpha1.IOChaosAction{other, wantAction}))
		Expect(chaos.Status.Instances).To(HaveKeyWithValue(target, int64(3)))

		record.Phase = waitForApplySync
		Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(waitForApplySync))
		child.Status.ObservedGeneration = child.Generation
		Expect(kube.Status().Update(ctx, child)).To(Succeed())
		Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(v1alpha1.Injected))

		record.Phase = v1alpha1.Injected
		Expect(impl.Recover(ctx, 0, records, chaos)).To(Equal(waitForRecoverSync))
		Expect(kube.Get(ctx, key, child)).To(Succeed())
		Expect(child.Spec.Actions).To(Equal([]v1alpha1.IOChaosAction{other}))
		Expect(chaos.Status.Instances).To(HaveKeyWithValue(target, int64(4)))
		record.Phase = waitForRecoverSync
		Expect(impl.Recover(ctx, 0, records, chaos)).To(Equal(waitForRecoverSync))
		child.Status.ObservedGeneration = child.Generation
		Expect(kube.Status().Update(ctx, child)).To(Succeed())
		Expect(impl.Recover(ctx, 0, records, chaos)).To(Equal(v1alpha1.NotInjected))
	})

	It("surfaces daemon failures and distinguishes disappearance during synchronization", func() {
		records := []*v1alpha1.Record{record}
		Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(waitForApplySync))
		child := &v1alpha1.PodIOChaos{}
		Expect(kube.Get(ctx, key, child)).To(Succeed())
		child.Status.ObservedGeneration = child.Generation
		child.Status.FailedMessage = "daemon did not apply the requested actions"
		Expect(kube.Status().Update(ctx, child)).To(Succeed())
		record.Phase = waitForApplySync
		phase, err := impl.Apply(ctx, 0, records, chaos)
		Expect(phase).To(Equal(waitForApplySync))
		Expect(err).To(MatchError(child.Status.FailedMessage))
		record.Phase = waitForRecoverSync
		phase, err = impl.Recover(ctx, 0, records, chaos)
		Expect(phase).To(Equal(waitForRecoverSync))
		Expect(err).To(MatchError(child.Status.FailedMessage))
		Expect(kube.Delete(ctx, child)).To(Succeed())
		Expect(impl.Recover(ctx, 0, records, chaos)).To(Equal(v1alpha1.NotInjected))
		record.Phase = waitForApplySync
		Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(v1alpha1.NotInjected))
	})

	It("keeps the recorded generation when recovery cannot be persisted", func() {
		records := []*v1alpha1.Record{record}
		Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(waitForApplySync))
		record.Phase = v1alpha1.Injected
		updateError = errors.New("update unavailable")
		phase, err := impl.Recover(ctx, 0, records, chaos)
		Expect(phase).To(Equal(v1alpha1.Injected))
		Expect(err).To(MatchError(updateError))
		Expect(chaos.Status.Instances).To(HaveKeyWithValue(target, int64(1)))
		Expect(kube.Delete(ctx, pod)).To(Succeed())
		Expect(impl.Recover(ctx, 0, records, chaos)).To(Equal(v1alpha1.NotInjected))
	})
	It("reports an invalid target when requesting or waiting for recovery", func() {
		record.Id = "invalid"
		for _, expected := range []v1alpha1.Phase{v1alpha1.Injected, waitForRecoverSync} {
			record.Phase = expected
			phase, err := impl.Recover(ctx, 0, []*v1alpha1.Record{record}, chaos)
			Expect(err).To(HaveOccurred())
			Expect(phase).To(Equal(expected))
		}
	})

})
