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

package v1alpha1

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// These tests are written in BDD-style using Ginkgo framework. Refer to
// http://onsi.github.io/ginkgo to learn more.

var _ = Describe("PhysicalMachineChaos", func() {
	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context("Create API", func() {
		It("should create an object successfully", func() {
			testCases := []struct {
				physicalMachineChaos *PhysicalMachineChaos
				key                  types.NamespacedName
			}{
				{
					physicalMachineChaos: &PhysicalMachineChaos{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "foo",
							Namespace: "default",
						},
						Spec: PhysicalMachineChaosSpec{
							Action: "stress-mem",
							PhysicalMachineSelector: PhysicalMachineSelector{
								Address: []string{"123.123.123.123.123"},
								Mode:    OneMode,
							},
							ExpInfo: ExpInfo{
								StressMemory: &StressMemorySpec{
									Size: "10MB",
								},
							},
						},
					},
					key: types.NamespacedName{
						Name:      "foo",
						Namespace: "default",
					},
				}, {
					physicalMachineChaos: &PhysicalMachineChaos{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "foo1",
							Namespace: "default",
						},
						Spec: PhysicalMachineChaosSpec{
							Action: "stress-mem",
							PhysicalMachineSelector: PhysicalMachineSelector{
								Selector: PhysicalMachineSelectorSpec{
									GenericSelectorSpec: GenericSelectorSpec{
										LabelSelectors: map[string]string{
											"foo1": "bar",
										},
									},
								},
								Mode: OneMode,
							},
							ExpInfo: ExpInfo{
								StressMemory: &StressMemorySpec{
									Size: "10MB",
								},
							},
						},
					},
					key: types.NamespacedName{
						Name:      "foo1",
						Namespace: "default",
					},
				},
			}

			for _, testCase := range testCases {
				By("creating an API obj")
				Expect(k8sClient.Create(context.TODO(), testCase.physicalMachineChaos)).To(Succeed())

				fetched := &PhysicalMachineChaos{}
				Expect(k8sClient.Get(context.TODO(), testCase.key, fetched)).To(Succeed())
				Expect(fetched).To(Equal(testCase.physicalMachineChaos))

				By("deleting the created object")
				Expect(k8sClient.Delete(context.TODO(), testCase.physicalMachineChaos)).To(Succeed())
				Expect(k8sClient.Get(context.TODO(), testCase.key, testCase.physicalMachineChaos)).ToNot(Succeed())
			}
		})
	})
})
