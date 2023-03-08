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

var _ = Describe("awschaos_webhook", func() {
	Context("webhook.Validator of awschaos", func() {
		It("Validate", func() {

			type TestCase struct {
				name    string
				chaos   AWSChaos
				execute func(chaos *AWSChaos) error
				expect  string
			}
			testDeviceName := "testDeviceName"
			testEbsVolume := "testEbsVolume"
			tcs := []TestCase{
				{
					name: "simple ValidateCreate for DetachVolume",
					chaos: AWSChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo1",
						},
						Spec: AWSChaosSpec{
							Action: DetachVolume,
						},
					},
					execute: func(chaos *AWSChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "unknow action",
					chaos: AWSChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo6",
						},
					},
					execute: func(chaos *AWSChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate the DetachVolume without EbsVolume",
					chaos: AWSChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo7",
						},
						Spec: AWSChaosSpec{
							Action: DetachVolume,
							AWSSelector: AWSSelector{
								DeviceName: &testDeviceName,
							},
						},
					},
					execute: func(chaos *AWSChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate the DetachVolume without DeviceName",
					chaos: AWSChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo7",
						},
						Spec: AWSChaosSpec{
							Action: DetachVolume,
							AWSSelector: AWSSelector{
								EbsVolume: &testEbsVolume,
							},
						},
					},
					execute: func(chaos *AWSChaos) error {
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
