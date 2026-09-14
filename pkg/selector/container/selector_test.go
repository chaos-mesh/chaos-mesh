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

package container

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
)

func newPodWithContainers(name string, containerNames ...string) v1.Pod {
	var containers []v1.Container
	for _, containerName := range containerNames {
		containers = append(containers, v1.Container{Name: containerName})
	}

	return v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: v1.PodSpec{
			Containers: containers,
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}
}

func TestSelectContainers(t *testing.T) {
	g := NewGomegaWithT(t)

	multiContainerPod := newPodWithContainers("multi-container-pod", "consul", "java-app")
	singleContainerPod := newPodWithContainers("single-container-pod", "java-app")

	impl := &SelectImpl{
		c: fake.NewClientBuilder().WithRuntimeObjects(&multiContainerPod, &singleContainerPod).Build(),
		Option: generic.Option{
			ClusterScoped: true,
		},
	}

	type TestCase struct {
		name          string
		selector      v1alpha1.ContainerSelector
		expectedIds   []string
		expectedError string
	}

	tcs := []TestCase{
		{
			name: "select the container specified by containerNames",
			selector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Selector: v1alpha1.PodSelectorSpec{
						Pods: map[string][]string{metav1.NamespaceDefault: {"multi-container-pod"}},
					},
					Mode: v1alpha1.AllMode,
				},
				ContainerNames: []string{"java-app"},
			},
			expectedIds: []string{"default/multi-container-pod/java-app"},
		},
		{
			name: "select the only container when containerNames is not set",
			selector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Selector: v1alpha1.PodSelectorSpec{
						Pods: map[string][]string{metav1.NamespaceDefault: {"single-container-pod"}},
					},
					Mode: v1alpha1.AllMode,
				},
			},
			expectedIds: []string{"default/single-container-pod/java-app"},
		},
		{
			name: "reject a pod with several containers when containerNames is not set",
			selector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Selector: v1alpha1.PodSelectorSpec{
						Pods: map[string][]string{metav1.NamespaceDefault: {"multi-container-pod"}},
					},
					Mode: v1alpha1.AllMode,
				},
			},
			expectedError: "pod default/multi-container-pod has 2 containers, please specify the container name in containerNames",
		},
	}

	for _, tc := range tcs {
		containers, err := impl.Select(context.Background(), &tc.selector)
		if len(tc.expectedError) != 0 {
			g.Expect(err).To(HaveOccurred(), tc.name)
			g.Expect(err.Error()).To(ContainSubstring(tc.expectedError), tc.name)
			continue
		}

		g.Expect(err).ShouldNot(HaveOccurred(), tc.name)

		var ids []string
		for _, container := range containers {
			ids = append(ids, container.Id())
		}
		g.Expect(ids).To(Equal(tc.expectedIds), tc.name)
	}
}
