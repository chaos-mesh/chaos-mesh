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

var _ = Describe("jvmchaos_webhook", func() {
	Context("Defaulter", func() {
		It("set default namespace selector", func() {
			jvmchaos := &JVMChaos{
				ObjectMeta: metav1.ObjectMeta{Namespace: metav1.NamespaceDefault},
			}
			jvmchaos.Default()
			Expect(jvmchaos.Spec.Selector.Namespaces[0]).To(Equal(metav1.NamespaceDefault))
		})
	})
	Context("webhook.Validator of jvmchaos", func() {
		It("Validate JVMChaos", func() {

			type TestCase struct {
				name    string
				chaos   JVMChaos
				execute func(chaos *JVMChaos) error
				expect  string
			}

			tcs := []TestCase{
				{
					name: "simple ValidateCreate",
					chaos: JVMChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo1",
						},
						Spec: JVMChaosSpec{
							Action: JVMLatencyAction,
							JVMParameter: JVMParameter{
								JVMClassMethodSpec: JVMClassMethodSpec{
									Class:  "Main",
									Method: "print",
								},
								LatencyDuration: 1000,
							},
						},
					},
					execute: func(chaos *JVMChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "",
				},
				{
					name: "simple ValidateUpdate",
					chaos: JVMChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo2",
						},
						Spec: JVMChaosSpec{
							Action: JVMLatencyAction,
							JVMParameter: JVMParameter{
								JVMClassMethodSpec: JVMClassMethodSpec{
									Class:  "Main",
									Method: "print",
								},
								LatencyDuration: 1000,
							},
						},
					},
					execute: func(chaos *JVMChaos) error {
						return chaos.ValidateUpdate(chaos)
					},
					expect: "",
				},
				{
					name: "simple ValidateDelete",
					chaos: JVMChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo3",
						},
						Spec: JVMChaosSpec{
							Action: JVMLatencyAction,
							JVMParameter: JVMParameter{
								JVMClassMethodSpec: JVMClassMethodSpec{
									Class:  "Main",
									Method: "print",
								},
								LatencyDuration: 1000,
							},
						},
					},
					execute: func(chaos *JVMChaos) error {
						return chaos.ValidateDelete()
					},
					expect: "",
				},
				{
					name: "missing latency",
					chaos: JVMChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo4",
						},
						Spec: JVMChaosSpec{
							Action: JVMLatencyAction,
							JVMParameter: JVMParameter{
								JVMClassMethodSpec: JVMClassMethodSpec{
									Class:  "Main",
									Method: "print",
								},
							},
						},
					},
					execute: func(chaos *JVMChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "missing value",
					chaos: JVMChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo5",
						},
						Spec: JVMChaosSpec{
							Action: JVMReturnAction,
							JVMParameter: JVMParameter{
								JVMClassMethodSpec: JVMClassMethodSpec{
									Class:  "Main",
									Method: "print",
								},
							},
						},
					},
					execute: func(chaos *JVMChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "missing exception",
					chaos: JVMChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo6",
						},
						Spec: JVMChaosSpec{
							Action: JVMExceptionAction,
							JVMParameter: JVMParameter{
								JVMClassMethodSpec: JVMClassMethodSpec{
									Class:  "Main",
									Method: "print",
								},
							},
						},
					},
					execute: func(chaos *JVMChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "missing class",
					chaos: JVMChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo7",
						},
						Spec: JVMChaosSpec{
							Action: JVMLatencyAction,
							JVMParameter: JVMParameter{
								JVMClassMethodSpec: JVMClassMethodSpec{
									Method: "print",
								},
								LatencyDuration: 1000,
							},
						},
					},
					execute: func(chaos *JVMChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "missing method",
					chaos: JVMChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo8",
						},
						Spec: JVMChaosSpec{
							Action: JVMLatencyAction,
							JVMParameter: JVMParameter{
								JVMClassMethodSpec: JVMClassMethodSpec{
									Class: "Main",
								},
								LatencyDuration: 1000,
							},
						},
					},
					execute: func(chaos *JVMChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "missing rule data",
					chaos: JVMChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo9",
						},
						Spec: JVMChaosSpec{
							Action: JVMRuleDataAction,
						},
					},
					execute: func(chaos *JVMChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "missing cpu-count and memory type",
					chaos: JVMChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo10",
						},
						Spec: JVMChaosSpec{
							Action: JVMStressAction,
						},
					},
					execute: func(chaos *JVMChaos) error {
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
