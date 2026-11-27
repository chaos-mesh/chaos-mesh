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
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/iochaos/podiochaosmanager"
)

func TestIOChaos(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IO Chaos")
}

var _ = Describe("IO chaos synchronization", func() {
	It("sets the Pod owner reference even if the client omits type metadata", func(ctx SpecContext) {
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "app", UID: "pod-uid"},
			Status:     corev1.PodStatus{Phase: corev1.PodRunning},
		}
		kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod).
			WithInterceptorFuncs(interceptor.Funcs{
				Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
					if err := c.Get(ctx, key, obj, opts...); err != nil {
						return err
					}
					if pod, ok := obj.(*corev1.Pod); ok {
						pod.TypeMeta = metav1.TypeMeta{}
					}
					return nil
				},
			}).Build()
		log := logr.Discard()
		builder := podiochaosmanager.NewBuilder(podiochaosmanager.Params{
			Logger: log, Client: kube, Reader: kube, Scheme: scheme,
		})
		impl := NewImpl(kube, builder, log).Impl
		chaos := &v1alpha1.IOChaos{
			ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "experiment"},
			Spec: v1alpha1.IOChaosSpec{
				Action: v1alpha1.IoFaults, VolumePath: "/data", Path: "**/records", Percent: 25, Errno: 5,
				Methods: []v1alpha1.IoMethod{v1alpha1.Read},
			},
		}
		records := []*v1alpha1.Record{{Id: "test/app/main", Phase: v1alpha1.NotInjected}}
		Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(waitForApplySync))

		child := &v1alpha1.PodIOChaos{}
		Expect(kube.Get(ctx, client.ObjectKeyFromObject(pod), child)).To(Succeed())
		Expect(child.OwnerReferences).To(ConsistOf(metav1.OwnerReference{
			APIVersion: "v1", Kind: "Pod", Name: pod.Name, UID: pod.UID,
		}))
	})
	It("reports an invalid target when requesting or waiting for recovery", func(ctx SpecContext) {
		impl := &Impl{Client: fake.NewClientBuilder().Build()}
		chaos := &v1alpha1.IOChaos{}
		record := &v1alpha1.Record{Id: "invalid"}
		for _, expected := range []v1alpha1.Phase{v1alpha1.Injected, waitForRecoverSync} {
			record.Phase = expected
			phase, err := impl.Recover(ctx, 0, []*v1alpha1.Record{record}, chaos)
			Expect(err).To(HaveOccurred())
			Expect(phase).To(Equal(expected))
		}
	})
})
