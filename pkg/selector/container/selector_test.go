// Copyright 2021 Chaos Mesh Authors.
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

package container

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
)

func TestContainerSelector(t *testing.T) {
	g := NewGomegaWithT(t)

	restartPolicyAlways := v1.ContainerRestartPolicyAlways

	// Create test pods with various container configurations
	pods := []client.Object{
		// Pod with only regular containers
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod-regular",
				Namespace: metav1.NamespaceDefault,
				Labels: map[string]string{
					"app": "test",
				},
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{Name: "nginx"},
					{Name: "sidecar"},
				},
			},
		},
		// Pod with init containers (with restartPolicy Always) and regular containers
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod-with-init",
				Namespace: metav1.NamespaceDefault,
				Labels: map[string]string{
					"app": "test",
				},
			},
			Spec: v1.PodSpec{
				InitContainers: []v1.Container{
					{Name: "init-setup", RestartPolicy: &restartPolicyAlways},
					{Name: "init-config", RestartPolicy: &restartPolicyAlways},
				},
				Containers: []v1.Container{
					{Name: "main-app"},
					{Name: "sidecar"},
				},
			},
		},
		// Pod with only init containers with restartPolicy Always
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod-only-init",
				Namespace: metav1.NamespaceDefault,
				Labels: map[string]string{
					"app": "test",
				},
			},
			Spec: v1.PodSpec{
				InitContainers: []v1.Container{
					{Name: "init-only", RestartPolicy: &restartPolicyAlways},
				},
			},
		},
		// Pod with init containers without restartPolicy Always (should be filtered out)
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod-with-init-no-restart",
				Namespace: metav1.NamespaceDefault,
				Labels: map[string]string{
					"app": "test",
				},
			},
			Spec: v1.PodSpec{
				InitContainers: []v1.Container{
					{Name: "init-once"}, // No restartPolicy - should not be selected
				},
				Containers: []v1.Container{
					{Name: "main-app"},
				},
			},
		},
	}

	c := fake.NewClientBuilder().
		WithObjects(pods...).
		Build()

	impl := &SelectImpl{
		c: c,
		r: c,
		Option: generic.Option{
			ClusterScoped:         true,
			TargetNamespace:       "",
			EnableFilterNamespace: false,
		},
	}

	t.Run("select first regular container when no names specified", func(t *testing.T) {
		selector := &v1alpha1.ContainerSelector{
			PodSelector: v1alpha1.PodSelector{
				Selector: v1alpha1.PodSelectorSpec{
					GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
						Namespaces: []string{metav1.NamespaceDefault},
						LabelSelectors: map[string]string{
							"app": "test",
						},
					},
				},
				Mode: v1alpha1.AllMode,
			},
			ContainerNames: []string{}, // No container names specified
		}

		containers, err := impl.Select(context.Background(), selector)
		g.Expect(err).ShouldNot(HaveOccurred())

		// Should select first regular container from each pod
		// pod-regular: nginx (first regular container)
		// pod-with-init: main-app (first regular container, NOT init-setup)
		// pod-only-init: no selection (no regular containers)
		// pod-with-init-no-restart: main-app (first regular container)
		g.Expect(len(containers)).To(Equal(3))

		containerNames := make(map[string]int)
		for _, c := range containers {
			containerNames[c.ContainerName]++
		}

		g.Expect(containerNames["nginx"]).To(Equal(1))
		g.Expect(containerNames["main-app"]).To(Equal(2)) // Found in both pod-with-init and pod-with-init-no-restart
		g.Expect(containerNames["init-setup"]).To(Equal(0))
		g.Expect(containerNames["init-config"]).To(Equal(0))
	})

	t.Run("select specific regular container by name", func(t *testing.T) {
		selector := &v1alpha1.ContainerSelector{
			PodSelector: v1alpha1.PodSelector{
				Selector: v1alpha1.PodSelectorSpec{
					GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
						Namespaces: []string{metav1.NamespaceDefault},
						LabelSelectors: map[string]string{
							"app": "test",
						},
					},
				},
				Mode: v1alpha1.AllMode,
			},
			ContainerNames: []string{"sidecar"},
		}

		containers, err := impl.Select(context.Background(), selector)
		g.Expect(err).ShouldNot(HaveOccurred())

		// Should find sidecar in both pod-regular and pod-with-init
		g.Expect(len(containers)).To(Equal(2))

		for _, c := range containers {
			g.Expect(c.ContainerName).To(Equal("sidecar"))
		}
	})

	t.Run("select init container by name", func(t *testing.T) {
		selector := &v1alpha1.ContainerSelector{
			PodSelector: v1alpha1.PodSelector{
				Selector: v1alpha1.PodSelectorSpec{
					GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
						Namespaces: []string{metav1.NamespaceDefault},
						LabelSelectors: map[string]string{
							"app": "test",
						},
					},
				},
				Mode: v1alpha1.AllMode,
			},
			ContainerNames: []string{"init-setup"},
		}

		containers, err := impl.Select(context.Background(), selector)
		g.Expect(err).ShouldNot(HaveOccurred())

		// Should find init-setup in pod-with-init only
		g.Expect(len(containers)).To(Equal(1))
		g.Expect(containers[0].ContainerName).To(Equal("init-setup"))
		g.Expect(containers[0].Pod.Name).To(Equal("pod-with-init"))
	})

	t.Run("select multiple init and regular containers by name", func(t *testing.T) {
		selector := &v1alpha1.ContainerSelector{
			PodSelector: v1alpha1.PodSelector{
				Selector: v1alpha1.PodSelectorSpec{
					GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
						Namespaces: []string{metav1.NamespaceDefault},
						LabelSelectors: map[string]string{
							"app": "test",
						},
					},
				},
				Mode: v1alpha1.AllMode,
			},
			ContainerNames: []string{"init-config", "main-app", "sidecar"},
		}

		containers, err := impl.Select(context.Background(), selector)
		g.Expect(err).ShouldNot(HaveOccurred())

		// Should find:
		// - init-config in pod-with-init
		// - main-app in pod-with-init and pod-with-init-no-restart
		// - sidecar in pod-regular and pod-with-init
		g.Expect(len(containers)).To(Equal(5))

		containerNames := make(map[string]int)
		for _, c := range containers {
			containerNames[c.ContainerName]++
		}

		g.Expect(containerNames["init-config"]).To(Equal(1))
		g.Expect(containerNames["main-app"]).To(Equal(2))
		g.Expect(containerNames["sidecar"]).To(Equal(2))
	})

	t.Run("select from specific pod with init containers", func(t *testing.T) {
		selector := &v1alpha1.ContainerSelector{
			PodSelector: v1alpha1.PodSelector{
				Selector: v1alpha1.PodSelectorSpec{
					GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
						Namespaces: []string{metav1.NamespaceDefault},
					},
					Pods: map[string][]string{
						metav1.NamespaceDefault: {"pod-with-init"},
					},
				},
				Mode: v1alpha1.AllMode,
			},
			ContainerNames: []string{"init-setup", "init-config", "main-app"},
		}

		containers, err := impl.Select(context.Background(), selector)
		g.Expect(err).ShouldNot(HaveOccurred())

		// Should find all three containers in pod-with-init
		g.Expect(len(containers)).To(Equal(3))

		containerNames := make(map[string]bool)
		for _, c := range containers {
			containerNames[c.ContainerName] = true
			g.Expect(c.Pod.Name).To(Equal("pod-with-init"))
		}

		g.Expect(containerNames["init-setup"]).To(BeTrue())
		g.Expect(containerNames["init-config"]).To(BeTrue())
		g.Expect(containerNames["main-app"]).To(BeTrue())
	})

	t.Run("no match for non-existent container name", func(t *testing.T) {
		selector := &v1alpha1.ContainerSelector{
			PodSelector: v1alpha1.PodSelector{
				Selector: v1alpha1.PodSelectorSpec{
					GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
						Namespaces: []string{metav1.NamespaceDefault},
						LabelSelectors: map[string]string{
							"app": "test",
						},
					},
				},
				Mode: v1alpha1.AllMode,
			},
			ContainerNames: []string{"non-existent"},
		}

		containers, err := impl.Select(context.Background(), selector)
		g.Expect(err).ShouldNot(HaveOccurred())
		g.Expect(len(containers)).To(Equal(0))
	})

	t.Run("init containers without restartPolicy Always are filtered out", func(t *testing.T) {
		selector := &v1alpha1.ContainerSelector{
			PodSelector: v1alpha1.PodSelector{
				Selector: v1alpha1.PodSelectorSpec{
					GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
						Namespaces: []string{metav1.NamespaceDefault},
						LabelSelectors: map[string]string{
							"app": "test",
						},
					},
				},
				Mode: v1alpha1.AllMode,
			},
			ContainerNames: []string{"init-once"},
		}

		containers, err := impl.Select(context.Background(), selector)
		g.Expect(err).ShouldNot(HaveOccurred())
		// Should not find init-once because it doesn't have restartPolicy Always
		g.Expect(len(containers)).To(Equal(0))
	})

	t.Run("only init containers with restartPolicy Always are selected", func(t *testing.T) {
		selector := &v1alpha1.ContainerSelector{
			PodSelector: v1alpha1.PodSelector{
				Selector: v1alpha1.PodSelectorSpec{
					GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
						Namespaces: []string{metav1.NamespaceDefault},
						LabelSelectors: map[string]string{
							"app": "test",
						},
					},
				},
				Mode: v1alpha1.AllMode,
			},
			ContainerNames: []string{"init-setup", "init-only"},
		}

		containers, err := impl.Select(context.Background(), selector)
		g.Expect(err).ShouldNot(HaveOccurred())

		// Should find init-setup in pod-with-init and init-only in pod-only-init
		// Both have restartPolicy Always
		g.Expect(len(containers)).To(Equal(2))

		containerNames := make(map[string]bool)
		for _, c := range containers {
			containerNames[c.ContainerName] = true
		}

		g.Expect(containerNames["init-setup"]).To(BeTrue())
		g.Expect(containerNames["init-only"]).To(BeTrue())
	})
}
