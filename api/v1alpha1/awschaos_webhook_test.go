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

var _ = Describe("awschaos_webhook", func() {
	Context("webhook.Validator of awschaos", func() {
		It("Validate", func() {

			type TestCase struct {
				name    string
				chaos   AwsChaos
				execute func(chaos *AwsChaos) error
				expect  string
			}
			testDeviceName := "testDeviceName"
			testEbsVolume := "testEbsVolume"
			tcs := []TestCase{
				{
					name: "simple ValidateCreate for DetachVolume",
					chaos: AwsChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo1",
						},
						Spec: AwsChaosSpec{
							Action: DetachVolume,
						},
					},
					execute: func(chaos *AwsChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "unknow action",
					chaos: AwsChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo6",
						},
					},
					execute: func(chaos *AwsChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate the DetachVolume without EbsVolume",
					chaos: AwsChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo7",
						},
						Spec: AwsChaosSpec{
							Action: DetachVolume,
							AwsSelector: AwsSelector{
								DeviceName: &testDeviceName,
							},
						},
					},
					execute: func(chaos *AwsChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate the DetachVolume without DeviceName",
					chaos: AwsChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo7",
						},
						Spec: AwsChaosSpec{
							Action: DetachVolume,
							AwsSelector: AwsSelector{
								EbsVolume: &testEbsVolume,
							},
						},
					},
					execute: func(chaos *AwsChaos) error {
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
