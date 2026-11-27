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
	var (
		kube    client.Client
		impl    impltypes.ChaosImpl
		chaos   *v1alpha1.HTTPChaos
		record  *v1alpha1.Record
		records []*v1alpha1.Record
	)

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: key.Name, UID: "pod-uid"},
			Status:     corev1.PodStatus{Phase: corev1.PodRunning},
		}
		kube = fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod).Build()
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
