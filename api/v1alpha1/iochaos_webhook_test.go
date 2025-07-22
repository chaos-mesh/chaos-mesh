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
)

var _ = Describe("iochaos_webhook", func() {
	Context("Defaulter", func() {
		It("set default namespace selector", func() {
			iochaos := &IOChaos{
				ObjectMeta: metav1.ObjectMeta{Namespace: metav1.NamespaceDefault},
			}
			iochaos.Default(context.Background(), iochaos)
			Expect(iochaos.Spec.Selector.Namespaces[0]).To(Equal(metav1.NamespaceDefault))
		})
	})
	Context("webhook.Validator of iochaos", func() {
		It("Validate", func() {

			type TestCase struct {
				name    string
				chaos   IOChaos
				execute func(chaos *IOChaos) error
				expect  string
			}
			errorDuration := "400S"

			tcs := []TestCase{
				{
					name: "simple ValidateCreate",
					chaos: IOChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo1",
						},
					},
					execute: func(chaos *IOChaos) error {
						_, err := chaos.ValidateCreate(context.Background(), chaos)
						return err
					},
					expect: "",
				},
				{
					name: "simple ValidateUpdate",
					chaos: IOChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo2",
						},
					},
					execute: func(chaos *IOChaos) error {
						_, err := chaos.ValidateUpdate(context.Background(), chaos, chaos)
						return err
					},
					expect: "",
				},
				{
					name: "simple ValidateDelete",
					chaos: IOChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo3",
						},
					},
					execute: func(chaos *IOChaos) error {
						_, err := chaos.ValidateDelete(context.Background(), chaos)
						return err
					},
					expect: "",
				},
				{
					name: "parse the duration error",
					chaos: IOChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo6",
						},
						Spec: IOChaosSpec{
							Duration: &errorDuration,
						},
					},
					execute: func(chaos *IOChaos) error {
						_, err := chaos.ValidateCreate(context.Background(), chaos)
						return err
					},
					expect: "error",
				},
				{
					name: "validate value with FixedPercentMode",
					chaos: IOChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo7",
						},
						Spec: IOChaosSpec{
							ContainerSelector: ContainerSelector{
								PodSelector: PodSelector{
									Value: "0",
									Mode:  FixedMode,
								},
							},
						},
					},
					execute: func(chaos *IOChaos) error {
						_, err := chaos.ValidateCreate(context.Background(), chaos)
						return err
					},
					expect: "error",
				},
				{
					name: "validate value with FixedPercentMode, parse value error",
					chaos: IOChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo8",
						},
						Spec: IOChaosSpec{
							ContainerSelector: ContainerSelector{
								PodSelector: PodSelector{
									Value: "num",
									Mode:  FixedMode,
								},
							},
						},
					},
					execute: func(chaos *IOChaos) error {
						_, err := chaos.ValidateCreate(context.Background(), chaos)
						return err
					},
					expect: "error",
				},
				{
					name: "validate value with RandomMaxPercentMode",
					chaos: IOChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo9",
						},
						Spec: IOChaosSpec{
							ContainerSelector: ContainerSelector{
								PodSelector: PodSelector{
									Value: "0",
									Mode:  RandomMaxPercentMode,
								},
							},
						},
					},
					execute: func(chaos *IOChaos) error {
						_, err := chaos.ValidateCreate(context.Background(), chaos)
						return err
					},
					expect: "error",
				},
				{
					name: "validate value with RandomMaxPercentMode ,parse value error",
					chaos: IOChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo10",
						},
						Spec: IOChaosSpec{
							ContainerSelector: ContainerSelector{
								PodSelector: PodSelector{
									Value: "num",
									Mode:  RandomMaxPercentMode,
								},
							},
						},
					},
					execute: func(chaos *IOChaos) error {
						_, err := chaos.ValidateCreate(context.Background(), chaos)
						return err
					},
					expect: "error",
				},
				{
					name: "validate value with FixedPercentMode",
					chaos: IOChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo11",
						},
						Spec: IOChaosSpec{
							ContainerSelector: ContainerSelector{
								PodSelector: PodSelector{
									Value: "101",
									Mode:  FixedPercentMode,
								},
							},
						},
					},
					execute: func(chaos *IOChaos) error {
						_, err := chaos.ValidateCreate(context.Background(), chaos)
						return err
					},
					expect: "error",
				},
				{
					name: "validate delay",
					chaos: IOChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo12",
						},
						Spec: IOChaosSpec{
							Delay:  "1S",
							Action: IoLatency,
						},
					},
					execute: func(chaos *IOChaos) error {
						_, err := chaos.ValidateCreate(context.Background(), chaos)
						return err
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
	})
})
