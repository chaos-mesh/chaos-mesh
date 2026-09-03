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

package podfailure

import (
	"context"
	"strings"
	"testing"

	"github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/annotation"
)

const testPauseImage = "registry.k8s.io/pause:3.9"

func newTestPod(t *testing.T) (client.Client, *v1.Pod, []*v1alpha1.Record) {
	t.Helper()
	g := gomega.NewWithT(t)
	oldPauseImage := config.ControllerCfg.PodFailurePauseImage
	config.ControllerCfg.PodFailurePauseImage = testPauseImage
	t.Cleanup(func() { config.ControllerCfg.PodFailurePauseImage = oldPauseImage })
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "applications", Name: "app-0", UID: "pod-uid",
			Annotations: map[string]string{"example.com/keep": "value"},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{Name: "app", Image: "app:v1"},
				{Name: "sidecar", Image: "sidecar:v1"},
			},
			InitContainers: []v1.Container{{Name: "init", Image: "init:v1"}},
		},
	}
	scheme := runtime.NewScheme()
	g.Expect(v1.AddToScheme(scheme)).To(gomega.Succeed())
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod.DeepCopy()).Build()
	return c, pod, []*v1alpha1.Record{{Id: "applications/app-0"}}
}

func testChaos(namespace, name, uid string) *v1alpha1.PodChaos {
	return &v1alpha1.PodChaos{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name, UID: types.UID(uid)},
		Spec:       v1alpha1.PodChaosSpec{Action: v1alpha1.PodFailureAction},
	}
}

func expectImages(t *testing.T, c client.Client, original *v1.Pod, paused bool) *v1.Pod {
	t.Helper()
	g := gomega.NewWithT(t)
	actual := &v1.Pod{}
	g.Expect(c.Get(context.Background(), client.ObjectKeyFromObject(original), actual)).To(gomega.Succeed())
	for i, container := range original.Spec.Containers {
		expected := container.Image
		if paused {
			expected = testPauseImage
		}
		g.Expect(actual.Spec.Containers[i].Image).To(gomega.Equal(expected))
	}
	for i, container := range original.Spec.InitContainers {
		expected := container.Image
		if paused {
			expected = testPauseImage
		}
		g.Expect(actual.Spec.InitContainers[i].Image).To(gomega.Equal(expected))
	}
	g.Expect(actual.Annotations["example.com/keep"]).To(gomega.Equal("value"))
	return actual
}

func TestPodFailureOverlappingExperiments(t *testing.T) {
	for _, names := range []struct{ firstNamespace, firstName, secondNamespace, secondName string }{
		{"experiments", "first", "experiments", "second"},
		{"team-a", "same-name", "team-b", "same-name"},
		{"experiments", strings.Repeat("a", 63), "experiments", strings.Repeat("b", 63)},
	} {
		for _, reverse := range []bool{false, true} {
			name := names.firstNamespace + "/" + names.firstName + "-" + names.secondNamespace + "/" + names.secondName
			if reverse {
				name += "-reverse"
			}
			t.Run(name, func(t *testing.T) {
				g := gomega.NewWithT(t)
				c, original, records := newTestPod(t)
				impl := NewImpl(c)
				a := testChaos(names.firstNamespace, names.firstName, "first-uid")
				b := testChaos(names.secondNamespace, names.secondName, "second-uid")
				for _, chaos := range []*v1alpha1.PodChaos{a, a, b} {
					phase, err := impl.Apply(context.Background(), 0, records, chaos)
					g.Expect(err).NotTo(gomega.HaveOccurred())
					g.Expect(phase).To(gomega.Equal(v1alpha1.Injected))
				}
				expectImages(t, c, original, true)
				if reverse {
					a, b = b, a
				}
				for _, chaos := range []*v1alpha1.PodChaos{testChaos("unrelated", "other", "other-uid"), a, a} {
					phase, err := impl.Recover(context.Background(), 0, records, chaos)
					g.Expect(err).NotTo(gomega.HaveOccurred())
					g.Expect(phase).To(gomega.Equal(v1alpha1.NotInjected))
					expectImages(t, c, original, true)
				}
				phase, err := impl.Recover(context.Background(), 0, records, b)
				g.Expect(err).NotTo(gomega.HaveOccurred())
				g.Expect(phase).To(gomega.Equal(v1alpha1.NotInjected))
				actual := expectImages(t, c, original, false)
				g.Expect(actual.Annotations).To(gomega.Equal(original.Annotations))
				_, err = impl.Recover(context.Background(), 0, records, b)
				g.Expect(err).NotTo(gomega.HaveOccurred())
				expectImages(t, c, original, false)
			})
		}
	}
}

func injectLegacy(t *testing.T, c client.Client, original *v1.Pod, chaos *v1alpha1.PodChaos) {
	t.Helper()
	g := gomega.NewWithT(t)
	pod := &v1.Pod{}
	g.Expect(c.Get(context.Background(), client.ObjectKeyFromObject(original), pod)).To(gomega.Succeed())
	for i, container := range pod.Spec.Containers {
		pod.Annotations[annotation.GenKeyForImage(chaos, container.Name, false)] = container.Image
		pod.Spec.Containers[i].Image = testPauseImage
	}
	for i, container := range pod.Spec.InitContainers {
		pod.Annotations[annotation.GenKeyForImage(chaos, container.Name, true)] = container.Image
		pod.Spec.InitContainers[i].Image = testPauseImage
	}
	g.Expect(c.Update(context.Background(), pod)).To(gomega.Succeed())
}

func TestPodFailureLegacyRecoveryAndRetry(t *testing.T) {
	g := gomega.NewWithT(t)
	c, original, records := newTestPod(t)
	impl := NewImpl(c)
	a := testChaos("experiments", "legacy", "legacy-uid")
	injectLegacy(t, c, original, a)
	before := expectImages(t, c, original, true)
	phase, err := impl.Apply(context.Background(), 0, records, a)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(phase).To(gomega.Equal(v1alpha1.Injected))
	after := expectImages(t, c, original, true)
	g.Expect(after.Annotations).To(gomega.Equal(before.Annotations))
	phase, err = impl.Recover(context.Background(), 0, records, a)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(phase).To(gomega.Equal(v1alpha1.NotInjected))
	g.Expect(expectImages(t, c, original, false).Annotations).To(gomega.Equal(original.Annotations))
}

func TestPodFailureDoesNotOverlapLegacyInjection(t *testing.T) {
	g := gomega.NewWithT(t)
	c, original, records := newTestPod(t)
	a := testChaos("experiments", "legacy", "legacy-uid")
	injectLegacy(t, c, original, a)
	before := expectImages(t, c, original, true)
	phase, err := NewImpl(c).Apply(context.Background(), 0, records, testChaos("experiments", "new", "new-uid"))
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(phase).To(gomega.Equal(v1alpha1.NotInjected))
	g.Expect(expectImages(t, c, original, true).Annotations).To(gomega.Equal(before.Annotations))
}

type interleavingClient struct {
	client.Client
	beforePatch func()
}

func (c *interleavingClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	if c.beforePatch != nil {
		beforePatch := c.beforePatch
		c.beforePatch = nil
		beforePatch()
	}
	return c.Client.Patch(ctx, obj, patch, opts...)
}

func TestPodFailureConcurrentApply(t *testing.T) {
	g := gomega.NewWithT(t)
	c, original, records := newTestPod(t)
	a := testChaos("experiments", "a", "a-uid")
	b := testChaos("experiments", "b", "b-uid")
	interleaved := &interleavingClient{Client: c, beforePatch: func() {
		_, err := NewImpl(c).Apply(context.Background(), 0, records, b)
		g.Expect(err).NotTo(gomega.HaveOccurred())
	}}
	_, err := NewImpl(interleaved).Apply(context.Background(), 0, records, a)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	_, err = NewImpl(c).Recover(context.Background(), 0, records, a)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	expectImages(t, c, original, true)
	_, err = NewImpl(c).Recover(context.Background(), 0, records, b)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	expectImages(t, c, original, false)
}

func TestPodFailureApplyDuringRecovery(t *testing.T) {
	g := gomega.NewWithT(t)
	c, original, records := newTestPod(t)
	a := testChaos("experiments", "a", "a-uid")
	b := testChaos("experiments", "b", "b-uid")
	_, err := NewImpl(c).Apply(context.Background(), 0, records, a)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	interleaved := &interleavingClient{Client: c, beforePatch: func() {
		_, err := NewImpl(c).Apply(context.Background(), 0, records, b)
		g.Expect(err).NotTo(gomega.HaveOccurred())
	}}
	_, err = NewImpl(interleaved).Recover(context.Background(), 0, records, a)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	expectImages(t, c, original, true)
	_, err = NewImpl(c).Recover(context.Background(), 0, records, b)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	expectImages(t, c, original, false)
}

func TestPodFailureInvalidStateDoesNotChangePod(t *testing.T) {
	for _, raw := range []string{
		"not-json",
		`{"owners":[]}`,
		`{"owners":["owner-uid","owner-uid"],"containers":{"app":"app:v1","sidecar":"sidecar:v1"},"initContainers":{"init":"init:v1"}}`,
		`{"owners":[""],"containers":{"app":"app:v1","sidecar":"sidecar:v1"},"initContainers":{"init":"init:v1"}}`,
		`{"owners":["owner-uid"],"containers":{"app":"","sidecar":"sidecar:v1"},"initContainers":{"init":"init:v1"}}`,
		`{"owners":["owner-uid"],"containers":{"app":"app:v1"}}`,
	} {
		t.Run(raw, func(t *testing.T) {
			g := gomega.NewWithT(t)
			c, original, records := newTestPod(t)
			pod := &v1.Pod{}
			g.Expect(c.Get(context.Background(), client.ObjectKeyFromObject(original), pod)).To(gomega.Succeed())
			pod.Annotations[stateAnnotation] = raw
			g.Expect(c.Update(context.Background(), pod)).To(gomega.Succeed())
			before := pod.DeepCopy()
			chaos := testChaos("experiments", "owner", "owner-uid")
			phase, err := NewImpl(c).Apply(context.Background(), 0, records, chaos)
			g.Expect(err).To(gomega.HaveOccurred())
			g.Expect(phase).To(gomega.Equal(v1alpha1.NotInjected))
			phase, err = NewImpl(c).Recover(context.Background(), 0, records, chaos)
			g.Expect(err).To(gomega.HaveOccurred())
			g.Expect(phase).To(gomega.Equal(v1alpha1.Injected))
			g.Expect(c.Get(context.Background(), client.ObjectKeyFromObject(original), pod)).To(gomega.Succeed())
			g.Expect(pod).To(gomega.Equal(before))
		})
	}
}

func TestPodFailureNewOwnerPausesUpdatedImages(t *testing.T) {
	g := gomega.NewWithT(t)
	c, original, records := newTestPod(t)
	a := testChaos("experiments", "a", "a-uid")
	b := testChaos("experiments", "b", "b-uid")
	_, err := NewImpl(c).Apply(context.Background(), 0, records, a)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	pod := expectImages(t, c, original, true)
	pod.Spec.Containers[0].Image = "app:v2"
	pod.Spec.InitContainers[0].Image = "init:v2"
	g.Expect(c.Update(context.Background(), pod)).To(gomega.Succeed())
	_, err = NewImpl(c).Apply(context.Background(), 0, records, b)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	expectImages(t, c, original, true)
	_, err = NewImpl(c).Recover(context.Background(), 0, records, a)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	expectImages(t, c, original, true)
	_, err = NewImpl(c).Recover(context.Background(), 0, records, b)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	expectImages(t, c, original, false)
}
