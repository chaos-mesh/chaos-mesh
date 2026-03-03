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

package physicalmachine

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/log"
	. "github.com/chaos-mesh/chaos-mesh/pkg/testutils"
)

func TestSelectPhysicalMachines(t *testing.T) {
	g := NewGomegaWithT(t)

	objects, physicalMachines := GenerateNPhysicalMachines("p", 5, PhysicalMachineArg{Labels: map[string]string{"l1": "l1"}})
	objects2, physicalMachines2 := GenerateNPhysicalMachines("s", 2, PhysicalMachineArg{Namespace: "test-s", Labels: map[string]string{"l2": "l2"}})

	objects = append(objects, objects2...)
	physicalMachines = append(physicalMachines, physicalMachines2...)

	err := v1alpha1.SchemeBuilder.AddToScheme(scheme.Scheme)
	g.Expect(err).NotTo(HaveOccurred())

	c := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithRuntimeObjects(objects...).
		WithStatusSubresource(&v1alpha1.PodNetworkChaos{}).
		Build()
	var r client.Reader

	type TestCase struct {
		name     string
		selector v1alpha1.PhysicalMachineSelectorSpec
		expected []v1alpha1.PhysicalMachine
	}

	tcs := []TestCase{
		{
			name: "filter specified physical machines",
			selector: v1alpha1.PhysicalMachineSelectorSpec{
				PhysicalMachines: map[string][]string{
					metav1.NamespaceDefault: {"p3", "p4"},
					"test-s":                {"s1"},
				},
			},
			expected: []v1alpha1.PhysicalMachine{physicalMachines[3], physicalMachines[4], physicalMachines[6]},
		},
		{
			name: "filter labels physical machines",
			selector: v1alpha1.PhysicalMachineSelectorSpec{
				GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
					LabelSelectors: map[string]string{"l2": "l2"},
				},
			},
			expected: []v1alpha1.PhysicalMachine{physicalMachines[5], physicalMachines[6]},
		},
		{
			name: "filter physicalMachines by label expressions",
			selector: v1alpha1.PhysicalMachineSelectorSpec{
				GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
					ExpressionSelectors: []metav1.LabelSelectorRequirement{
						{
							Key:      "l2",
							Operator: metav1.LabelSelectorOpIn,
							Values:   []string{"l2"},
						},
					},
				},
			},
			expected: []v1alpha1.PhysicalMachine{physicalMachines[5], physicalMachines[6]},
		},
		{
			name: "filter physicalMachines by label selectors and expression selectors",
			selector: v1alpha1.PhysicalMachineSelectorSpec{
				GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
					LabelSelectors: map[string]string{"l1": "l1"},
					ExpressionSelectors: []metav1.LabelSelectorRequirement{
						{
							Key:      "l2",
							Operator: metav1.LabelSelectorOpIn,
							Values:   []string{"l2"},
						},
					},
				},
			},
			expected: nil,
		},
		{
			name: "filter namespace and labels",
			selector: v1alpha1.PhysicalMachineSelectorSpec{
				GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
					Namespaces:     []string{"test-s"},
					LabelSelectors: map[string]string{"l2": "l2"},
				},
			},
			expected: []v1alpha1.PhysicalMachine{physicalMachines[5], physicalMachines[6]},
		},
		{
			name: "filter namespace and labels",
			selector: v1alpha1.PhysicalMachineSelectorSpec{
				GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
					Namespaces:     []string{metav1.NamespaceDefault},
					LabelSelectors: map[string]string{"l2": "l2"},
				},
			},
			expected: nil,
		},
	}

	var (
		testCfgClusterScoped   = true
		testCfgTargetNamespace = ""
	)

	logger, _ := log.NewDefaultZapLogger()

	for _, tc := range tcs {
		filtered, err := SelectPhysicalMachines(context.Background(), c, r, tc.selector, testCfgClusterScoped, testCfgTargetNamespace, false, logger)
		g.Expect(err).ShouldNot(HaveOccurred(), tc.name)
		g.Expect(len(filtered)).To(Equal(len(tc.expected)), tc.name)
	}
}
