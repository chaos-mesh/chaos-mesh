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
)

var _ = Describe("networkchaos_webhook", func() {
	Context("Defaulter", func() {
		It("set default namespace selector", func() {
			networkchaos := &NetworkChaos{
				ObjectMeta: metav1.ObjectMeta{Namespace: metav1.NamespaceDefault},
			}
			networkchaos.Default()
			Expect(networkchaos.Spec.Selector.Namespaces[0]).To(Equal(metav1.NamespaceDefault))
		})

		It("set default DelaySpec", func() {
			networkchaos := &NetworkChaos{
				ObjectMeta: metav1.ObjectMeta{Namespace: metav1.NamespaceDefault},
				Spec: NetworkChaosSpec{
					TcParameter: TcParameter{
						Delay: &DelaySpec{
							Latency: "90ms",
						},
					},
				},
			}
			networkchaos.Default()
			Expect(networkchaos.Spec.Delay.Correlation).To(Equal(DefaultCorrelation))
			Expect(networkchaos.Spec.Delay.Jitter).To(Equal(DefaultJitter))
		})
	})
	Context("webhook.Validator of networkchaos", func() {
		It("Validate", func() {

			type TestCase struct {
				name    string
				chaos   NetworkChaos
				execute func(chaos *NetworkChaos) error
				expect  string
			}
			tcs := []TestCase{
				{
					name: "simple ValidateCreate",
					chaos: NetworkChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo1",
						},
					},
					execute: func(chaos *NetworkChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "",
				},
				{
					name: "simple ValidateUpdate",
					chaos: NetworkChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo2",
						},
					},
					execute: func(chaos *NetworkChaos) error {
						return chaos.ValidateUpdate(chaos)
					},
					expect: "",
				},
				{
					name: "simple ValidateDelete",
					chaos: NetworkChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo3",
						},
					},
					execute: func(chaos *NetworkChaos) error {
						return chaos.ValidateDelete()
					},
					expect: "",
				},
				{
					name: "validate the delay",
					chaos: NetworkChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo6",
						},
						Spec: NetworkChaosSpec{
							TcParameter: TcParameter{
								Delay: &DelaySpec{
									Latency:     "1S",
									Jitter:      "1S",
									Correlation: "num",
								},
							},
						},
					},
					execute: func(chaos *NetworkChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate the reorder",
					chaos: NetworkChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo7",
						},
						Spec: NetworkChaosSpec{
							TcParameter: TcParameter{
								Delay: &DelaySpec{
									Reorder: &ReorderSpec{
										Reorder:     "num",
										Correlation: "num",
									},
								},
							},
						},
					},
					execute: func(chaos *NetworkChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate the loss",
					chaos: NetworkChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo8",
						},
						Spec: NetworkChaosSpec{
							TcParameter: TcParameter{
								Loss: &LossSpec{
									Loss:        "num",
									Correlation: "num",
								},
							},
						},
					},
					execute: func(chaos *NetworkChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate the duplicate",
					chaos: NetworkChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo9",
						},
						Spec: NetworkChaosSpec{
							TcParameter: TcParameter{
								Duplicate: &DuplicateSpec{
									Duplicate:   "num",
									Correlation: "num",
								},
							},
						},
					},
					execute: func(chaos *NetworkChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate the corrupt",
					chaos: NetworkChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo10",
						},
						Spec: NetworkChaosSpec{
							TcParameter: TcParameter{
								Corrupt: &CorruptSpec{
									Corrupt:     "num",
									Correlation: "num",
								},
							},
						},
					},
					execute: func(chaos *NetworkChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate the bandwidth",
					chaos: NetworkChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo11",
						},
						Spec: NetworkChaosSpec{
							TcParameter: TcParameter{
								Bandwidth: &BandwidthSpec{
									Rate: "10",
								},
							},
						},
					},
					execute: func(chaos *NetworkChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate the target",
					chaos: NetworkChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo12",
						},
						Spec: NetworkChaosSpec{
							Target: &PodSelector{
								Mode:  FixedPodMode,
								Value: "0",
							},
						},
					},
					execute: func(chaos *NetworkChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate direction and externalTargets",
					chaos: NetworkChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo12",
						},
						Spec: NetworkChaosSpec{
							Direction:       From,
							ExternalTargets: []string{"8.8.8.8"},
						},
					},
					execute: func(chaos *NetworkChaos) error {
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
	})
	Context("convertUnitToBytes", func() {
		It("should convert number with unit successfully", func() {
			n, err := ConvertUnitToBytes("  10   mbPs  ")
			Expect(err).Should(Succeed())
			Expect(n).To(Equal(uint64(10 * 1024 * 1024)))
		})

		It("should return error with invalid unit", func() {
			n, err := ConvertUnitToBytes(" 10 cpbs")
			Expect(err).Should(HaveOccurred())
			Expect(n).To(Equal(uint64(0)))
		})
	})
})
