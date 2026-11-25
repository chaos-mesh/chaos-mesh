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

package podfailure

import (
	"context"
	"errors"
	"testing"

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
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/annotation"
)

func TestPodFailure(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pod Failure")
}

var _ = Describe("Pod failure lifecycle", func() {
	ctx := context.Background()
	var (
		pod           *corev1.Pod
		chaos         *v1alpha1.PodChaos
		records       []*v1alpha1.Record
		kube          client.WithWatch
		impl          *Impl
		patchFailure  error
		lookupFailure error
	)

	BeforeEach(func() {
		pod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "workload"},
			Spec: corev1.PodSpec{
				Containers:     []corev1.Container{{Name: "app", Image: "app:v1"}},
				InitContainers: []corev1.Container{{Name: "setup", Image: "setup:v2"}},
			},
		}
		chaos = &v1alpha1.PodChaos{
			ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "failure"},
			Spec:       v1alpha1.PodChaosSpec{Action: v1alpha1.PodFailureAction},
		}
		records = []*v1alpha1.Record{{Id: "test/workload"}}
		patchFailure, lookupFailure = nil, nil

		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		kube = fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod).Build()
		impl = NewImpl(interceptor.NewClient(kube, interceptor.Funcs{
			Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
				if lookupFailure != nil {
					return lookupFailure
				}
				return c.Get(ctx, key, obj, opts...)
			},
			Patch: func(ctx context.Context, c client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
				if patchFailure != nil {
					return patchFailure
				}
				return c.Patch(ctx, obj, patch, opts...)
			},
		}))
	})

	It("preserves original images across repeated injection and can retry recovery", func() {
		key := client.ObjectKeyFromObject(pod)
		appImage := annotation.GenKeyForImage(chaos, "app", false)
		initImage := annotation.GenKeyForImage(chaos, "setup", true)
		originalImages := map[string]string{appImage: "app:v1", initImage: "setup:v2"}
		for range 2 {
			Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(v1alpha1.Injected))
			Expect(kube.Get(ctx, key, pod)).To(Succeed())
			Expect(pod.Spec.Containers[0].Image).To(Equal(config.ControllerCfg.PodFailurePauseImage))
			Expect(pod.Spec.InitContainers[0].Image).To(Equal(config.ControllerCfg.PodFailurePauseImage))
			Expect(pod.Annotations).To(Equal(originalImages))
		}

		pod.Annotations["owner"] = "team"
		Expect(kube.Update(ctx, pod)).To(Succeed())
		patchFailure = errors.New("patch unavailable")
		phase, err := impl.Recover(ctx, 0, records, chaos)
		Expect(err).To(MatchError(patchFailure))
		Expect(phase).To(Equal(v1alpha1.Injected))
		Expect(kube.Get(ctx, key, pod)).To(Succeed())
		Expect(pod.Annotations).To(HaveKeyWithValue(appImage, "app:v1"))
		Expect(pod.Spec.Containers[0].Image).To(Equal(config.ControllerCfg.PodFailurePauseImage))

		patchFailure = nil
		for range 2 {
			Expect(impl.Recover(ctx, 0, records, chaos)).To(Equal(v1alpha1.NotInjected))
			Expect(kube.Get(ctx, key, pod)).To(Succeed())
			Expect(pod.Spec.Containers[0].Image).To(Equal("app:v1"))
			Expect(pod.Spec.InitContainers[0].Image).To(Equal("setup:v2"))
			Expect(pod.Annotations).To(Equal(map[string]string{"owner": "team"}))
		}
	})

	It("leaves the pod unchanged when injection cannot be persisted", func() {
		patchFailure = errors.New("patch unavailable")
		phase, err := impl.Apply(ctx, 0, records, chaos)
		Expect(err).To(MatchError(patchFailure))
		Expect(phase).To(Equal(v1alpha1.NotInjected))
		Expect(kube.Get(ctx, client.ObjectKeyFromObject(pod), pod)).To(Succeed())
		Expect(pod.Spec.Containers[0].Image).To(Equal("app:v1"))
		Expect(pod.Spec.InitContainers[0].Image).To(Equal("setup:v2"))
		Expect(pod.Annotations).To(BeEmpty())
	})

	It("only considers recovery complete when the target is gone", func() {
		lookupFailure = errors.New("lookup unavailable")
		phase, err := impl.Recover(ctx, 0, records, chaos)
		Expect(err).To(MatchError(lookupFailure))
		Expect(phase).To(Equal(v1alpha1.Injected))

		lookupFailure = nil
		Expect(kube.Delete(ctx, pod)).To(Succeed())
		phase, err = impl.Apply(ctx, 0, records, chaos)
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
		Expect(phase).To(Equal(v1alpha1.NotInjected))
		Expect(impl.Recover(ctx, 0, records, chaos)).To(Equal(v1alpha1.NotInjected))
	})

	It("rejects an invalid target record during injection and recovery", func() {
		records[0].Id = "workload"
		for _, operation := range []func(context.Context, int, []*v1alpha1.Record, v1alpha1.InnerObject) (v1alpha1.Phase, error){impl.Apply, impl.Recover} {
			phase, err := operation(ctx, 0, records, chaos)
			Expect(err).To(HaveOccurred())
			Expect(phase).To(Equal(v1alpha1.NotInjected))
		}
	})
})
