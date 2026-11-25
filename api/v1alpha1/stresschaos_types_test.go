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
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"

	"github.com/chaos-mesh/chaos-mesh/api/genericwebhook"
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
				Name:      "foo",
				Namespace: "default",
				Spec: StressChaosSpec{
					ContainerSelector: ContainerSelector{
						PodSelector: PodSelector{
							Mode: OneMode,
						},
					},
					Stressors: &Stressors{MemoryStressor: &MemoryStressor{Workers: 1, Size: "64MiB"}},
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

var _ = Describe("MemoryStressor time", func() {
	DescribeTable("passes time to memStress",
		func(stressor MemoryStressor, want []string) {
			stressors := &Stressors{MemoryStressor: &stressor}
			_, memory, err := stressors.Normalize()
			Expect(err).NotTo(HaveOccurred())
			Expect(memory).To(Equal(strings.Join(want, " ")))
		},
		Entry("no time", MemoryStressor{Stressor: Stressor{Workers: 1}, Size: "1GiB"},
			[]string{"--workers", "1", "--size", "1GiB"}),
		Entry("time with typed size", MemoryStressor{Stressor: Stressor{Workers: 2}, Size: "50%", Time: "5m"},
			[]string{"--workers", "2", "--size", "50%", "--time", "5m"}),
		Entry("time with legacy size option", MemoryStressor{Stressor: Stressor{Workers: 1}, Time: "30s", Options: []string{"--size", "64MiB"}},
			[]string{"--workers", "1", "--time", "30s", "--size", "64MiB"}),
	)

	DescribeTable("validates time as a duration",
		func(time string, invalid bool) {
			obj := &StressChaos{
				Spec: StressChaosSpec{
					ContainerSelector: ContainerSelector{PodSelector: PodSelector{Mode: AllMode}},
					Stressors: &Stressors{MemoryStressor: &MemoryStressor{
						Stressor: Stressor{Workers: 1},
						Size:     "64MiB",
						Time:     time,
					}},
				},
			}
			errs := genericwebhook.Validate(obj)
			if invalid {
				Expect(errs).To(ConsistOf(HaveField("Field", "Spec.Stressors.MemoryStressor.Time")))
			} else {
				Expect(errs).To(BeEmpty())
			}
		},
		Entry("empty", "", false),
		Entry("seconds", "10s", false),
		Entry("compound", "1m30s", false),
		Entry("words", "ten minutes", true),
		Entry("no unit", "10", true),
	)
})
