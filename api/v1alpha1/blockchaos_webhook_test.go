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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("blockchaos_webhook", func() {
	Context("webhook.Validator of blockchaos", func() {
		It("Validate", func() {

			type TestCase struct {
				name    string
				chaos   BlockChaos
				execute func(chaos *BlockChaos) error
				expect  string
			}
			errorDuration := "400S"

			tcs := []TestCase{
				{
					name: "simple ValidateCreate",
					chaos: BlockChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo1",
						},
					},
					execute: func(chaos *BlockChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "",
				},
				{
					name: "simple ValidateUpdate",
					chaos: BlockChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo2",
						},
					},
					execute: func(chaos *BlockChaos) error {
						return chaos.ValidateUpdate(chaos)
					},
					expect: "",
				},
				{
					name: "simple ValidateDelete",
					chaos: BlockChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo3",
						},
					},
					execute: func(chaos *BlockChaos) error {
						return chaos.ValidateDelete()
					},
					expect: "",
				},
				{
					name: "parse the duration error",
					chaos: BlockChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo6",
						},
						Spec: BlockChaosSpec{
							Duration: &errorDuration,
						},
					},
					execute: func(chaos *BlockChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate value with FixedPercentMode",
					chaos: BlockChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo7",
						},
						Spec: BlockChaosSpec{
							ContainerNodeVolumePathSelector: ContainerNodeVolumePathSelector{
								ContainerSelector: ContainerSelector{
									PodSelector: PodSelector{
										Value: "0",
										Mode:  FixedMode,
									},
								},
								VolumeName: "",
							},
						},
					},
					execute: func(chaos *BlockChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate value with FixedPercentMode, parse value error",
					chaos: BlockChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo8",
						},
						Spec: BlockChaosSpec{
							ContainerNodeVolumePathSelector: ContainerNodeVolumePathSelector{
								ContainerSelector: ContainerSelector{
									PodSelector: PodSelector{
										Value: "num",
										Mode:  FixedMode,
									},
								},
								VolumeName: "",
							},
						},
					},
					execute: func(chaos *BlockChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate value with RandomMaxPercentMode",
					chaos: BlockChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo9",
						},
						Spec: BlockChaosSpec{
							ContainerNodeVolumePathSelector: ContainerNodeVolumePathSelector{
								ContainerSelector: ContainerSelector{
									PodSelector: PodSelector{
										Value: "0",
										Mode:  RandomMaxPercentMode,
									},
								},
								VolumeName: "",
							},
						},
					},
					execute: func(chaos *BlockChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate value with RandomMaxPercentMode ,parse value error",
					chaos: BlockChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo10",
						},
						Spec: BlockChaosSpec{
							ContainerNodeVolumePathSelector: ContainerNodeVolumePathSelector{
								ContainerSelector: ContainerSelector{
									PodSelector: PodSelector{
										Value: "num",
										Mode:  RandomMaxPercentMode,
									},
								},
							},
						},
					},
					execute: func(chaos *BlockChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate value with FixedPercentMode",
					chaos: BlockChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo11",
						},
						Spec: BlockChaosSpec{
							ContainerNodeVolumePathSelector: ContainerNodeVolumePathSelector{
								ContainerSelector: ContainerSelector{
									PodSelector: PodSelector{
										Value: "101",
										Mode:  FixedPercentMode,
									},
								},
							},
						},
					},
					execute: func(chaos *BlockChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate delay",
					chaos: BlockChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo12",
						},
						Spec: BlockChaosSpec{
							Action: BlockDelay,
							Delay:  nil,
						},
					},
					execute: func(chaos *BlockChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate delay",
					chaos: BlockChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo13",
						},
						Spec: BlockChaosSpec{
							Action: BlockDelay,
							Delay: &BlockDelaySpec{
								Latency: "1SSS",
							},
						},
					},
					execute: func(chaos *BlockChaos) error {
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
})
