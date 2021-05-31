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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("stresschaos_webhook", func() {
	Context("Defaulter", func() {
		It("set default namespace selector", func() {
			stresschaos := &StressChaos{
				ObjectMeta: metav1.ObjectMeta{Namespace: metav1.NamespaceDefault},
			}
			stresschaos.Default()
			Expect(stresschaos.Spec.Selector.Namespaces[0]).To(Equal(metav1.NamespaceDefault))
		})
	})
	Context("webhook.Validator of stresschaos", func() {
		It("Validate StressChaos", func() {

			type TestCase struct {
				name    string
				chaos   StressChaos
				execute func(chaos *StressChaos) error
				expect  string
			}
			stressors := &Stressors{
				MemoryStressor: &MemoryStressor{
					Stressor: Stressor{Workers: 1},
				},
			}
			tcs := []TestCase{
				{
					name: "simple ValidateCreate",
					chaos: StressChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo1",
						},
						Spec: StressChaosSpec{
							Stressors: stressors,
						},
					},
					execute: func(chaos *StressChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "",
				},
				{
					name: "simple ValidateUpdate",
					chaos: StressChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo2",
						},
						Spec: StressChaosSpec{
							Stressors: stressors,
						},
					},
					execute: func(chaos *StressChaos) error {
						return chaos.ValidateUpdate(chaos)
					},
					expect: "",
				},
				{
					name: "simple ValidateDelete",
					chaos: StressChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo3",
						},
						Spec: StressChaosSpec{
							Stressors: stressors,
						},
					},
					execute: func(chaos *StressChaos) error {
						return chaos.ValidateDelete()
					},
					expect: "",
				},
				{
					name: "missing stressors",
					chaos: StressChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo5",
						},
					},
					execute: func(chaos *StressChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
			}

			for _, tc := range tcs {
				err := tc.execute(&tc.chaos)
				if tc.expect == "error" {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
			}
		})

		It("Validate Stressors", func() {
			type TestCase struct {
				name     string
				stressor Validateable
				errs     int
			}
			tcs := []TestCase{
				{
					name:     "missing workers",
					stressor: &Stressor{},
					errs:     1,
				},
				{
					name: "default MemoryStressor",
					stressor: &MemoryStressor{
						Stressor: Stressor{Workers: 1},
					},
					errs: 0,
				},
				{
					name: "default CPUStressor",
					stressor: &CPUStressor{
						Stressor: Stressor{Workers: 1},
					},
					errs: 0,
				},
			}
			parent := field.NewPath("parent")
			for _, tc := range tcs {
				Expect(tc.stressor.Validate(parent)).To(HaveLen(tc.errs))
			}
		})

		It("Parse MemoryStressor fields", func() {
			vm := MemoryStressor{}
			incorrectBytes := []string{"-1", "-1%", "101%", "x%", "-1Kb"}
			for _, b := range incorrectBytes {
				vm.Size = b
				Expect(vm.tryParseBytes()).Should(HaveOccurred())
			}
			correctBytes := []string{"", "1%", "100KB", "100B"}
			for _, b := range correctBytes {
				vm.Size = b
				Expect(vm.tryParseBytes()).ShouldNot(HaveOccurred())
			}
		})

	})

})
