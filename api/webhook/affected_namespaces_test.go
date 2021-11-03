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

package webhook

import (
	"testing"

	"github.com/onsi/gomega"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func TestAffectedNamespaces(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	_, namespaces := affectedNamespaces(&v1alpha1.Schedule{
		Spec: v1alpha1.ScheduleSpec{
			ScheduleItem: v1alpha1.ScheduleItem{
				EmbedChaos: v1alpha1.EmbedChaos{
					PodChaos: &v1alpha1.PodChaosSpec{
						ContainerSelector: v1alpha1.ContainerSelector{
							PodSelector: v1alpha1.PodSelector{
								Selector: v1alpha1.PodSelectorSpec{
									GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
										Namespaces: []string{"ns1", "ns2"},
									},
								},
							},
						},
					},
				},
			},
		},
	})
	g.Expect(namespaces).To(gomega.Equal(map[string]struct{}{
		"ns1": {},
		"ns2": {},
	}))

	_, namespaces = affectedNamespaces(&v1alpha1.Workflow{
		Spec: v1alpha1.WorkflowSpec{
			Templates: []v1alpha1.Template{
				{
					EmbedChaos: &v1alpha1.EmbedChaos{
						NetworkChaos: &v1alpha1.NetworkChaosSpec{
							Target: &v1alpha1.PodSelector{
								Selector: v1alpha1.PodSelectorSpec{
									GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
										Namespaces: []string{"ns1", "ns2"},
									},
								},
							},
						},
					},
				},
			},
		},
	})
	g.Expect(namespaces).To(gomega.Equal(map[string]struct{}{
		"ns1": {},
		"ns2": {},
	}))

	clusterScoped, _ := affectedNamespaces(&v1alpha1.NetworkChaos{
		Spec: v1alpha1.NetworkChaosSpec{
			Target: &v1alpha1.PodSelector{},
		},
	})
	g.Expect(clusterScoped).To(gomega.BeTrue())

	clusterScoped, _ = affectedNamespaces(&v1alpha1.NetworkChaos{})
	g.Expect(clusterScoped).To(gomega.BeTrue())

	_, namespaces = affectedNamespaces(&v1alpha1.Workflow{
		Spec: v1alpha1.WorkflowSpec{
			Templates: []v1alpha1.Template{
				{
					EmbedChaos: &v1alpha1.EmbedChaos{
						NetworkChaos: &v1alpha1.NetworkChaosSpec{
							Target: &v1alpha1.PodSelector{
								Selector: v1alpha1.PodSelectorSpec{
									GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
										Namespaces: []string{"ns1", "ns2"},
									},
								},
							},
						},
					},
				},
				{
					EmbedChaos: &v1alpha1.EmbedChaos{
						NetworkChaos: &v1alpha1.NetworkChaosSpec{
							Target: &v1alpha1.PodSelector{
								Selector: v1alpha1.PodSelectorSpec{
									GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
										Namespaces: []string{"ns3", "ns4"},
									},
								},
							},
						},
					},
				},
			},
		},
	})
	g.Expect(namespaces).To(gomega.Equal(map[string]struct{}{
		"ns1": {},
		"ns2": {},
		"ns3": {},
		"ns4": {},
	}))

	_, namespaces = affectedNamespaces(&v1alpha1.NetworkChaos{
		Spec: v1alpha1.NetworkChaosSpec{
			Target: &v1alpha1.PodSelector{
				Selector: v1alpha1.PodSelectorSpec{
					Pods: map[string][]string{
						"ns1": {"pod1", "pod2"},
						"ns2": {"pod3", "pod4"},
					},
				},
			},
		},
	})
	g.Expect(namespaces).To(gomega.Equal(map[string]struct{}{
		"ns1": {},
		"ns2": {},
	}))
}
