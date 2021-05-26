// Copyright 2020 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// These tests are written in BDD-style using Ginkgo framework. Refer to
// http://onsi.github.io/ginkgo to learn more.
var _ = Describe("StressChaos", func() {

	Context("CRUD API", func() {
		var (
			key              types.NamespacedName
			created, fetched *StressChaos
		)

		It("Should create an object successfully", func() {
			key = types.NamespacedName{
				Name:      "foo",
				Namespace: "default",
			}
			created = &StressChaos{
				ObjectMeta: v1.ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				Spec: StressChaosSpec{
					ContainerSelector: ContainerSelector{
						PodSelector: PodSelector{
							Mode: OnePodMode,
						},
					},
					Stressors: &Stressors{MemoryStressor: &MemoryStressor{Stressor: Stressor{Workers: 1}}},
				},
			}
			By("creating an API object")
			Expect(k8sClient.Create(context.TODO(), created)).To(Succeed())

			fetched = &StressChaos{}
			Expect(k8sClient.Get(context.TODO(), key, fetched)).To(Succeed())
			Expect(fetched).To(Equal(created))

			By("deleting the created object")
			Expect(k8sClient.Delete(context.TODO(), created)).To(Succeed())
			Expect(k8sClient.Get(context.TODO(), key, created)).NotTo(Succeed())
		})
	})

})
