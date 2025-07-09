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

var _ = Describe("timechaos_webhook", func() {
	Context("Defaulter", func() {
		It("set default namespace selector", func() {
			timechaos := &TimeChaos{
				ObjectMeta: metav1.ObjectMeta{Namespace: metav1.NamespaceDefault},
			}
			timechaos.Default(context.Background(), timechaos)
			Expect(timechaos.Spec.Selector.Namespaces[0]).To(Equal(metav1.NamespaceDefault))
			Expect(timechaos.Spec.ClockIds[0]).To(Equal("CLOCK_REALTIME"))
		})
	})
	Context("webhook.Validator of timechaos", func() {
		It("Validate", func() {

			type TestCase struct {
				name    string
				chaos   TimeChaos
				execute func(chaos *TimeChaos) error
				expect  string
			}
			tcs := []TestCase{
				{
					name: "simple ValidateCreate",
					chaos: TimeChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo1",
						},
						Spec: TimeChaosSpec{TimeOffset: "1s"},
					},
					execute: func(chaos *TimeChaos) error {
						_, err := chaos.ValidateCreate(context.Background(), chaos)
						return err
					},
					expect: "",
				},
				{
					name: "simple ValidateUpdate",
					chaos: TimeChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo2",
						},
						Spec: TimeChaosSpec{TimeOffset: "1s"},
					},
					execute: func(chaos *TimeChaos) error {
						_, err := chaos.ValidateUpdate(context.Background(), chaos, chaos)
						return err
					},
					expect: "",
				},
				{
					name: "simple ValidateDelete",
					chaos: TimeChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo3",
						},
						Spec: TimeChaosSpec{TimeOffset: "1s"},
					},
					execute: func(chaos *TimeChaos) error {
						_, err := chaos.ValidateDelete(context.Background(), chaos)
						return err
					},
					expect: "",
				},
				{
					name: "validate the timeOffset",
					chaos: TimeChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo6",
						},
						Spec: TimeChaosSpec{
							TimeOffset: "1S",
						},
					},
					execute: func(chaos *TimeChaos) error {
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
