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

var _ = Describe("iochaos_webhook", func() {
	Context("Defaulter", func() {
		It("set default namespace selector", func() {
			iochaos := &IoChaos{
				ObjectMeta: metav1.ObjectMeta{Namespace: metav1.NamespaceDefault},
			}
			iochaos.Default()
			Expect(iochaos.Spec.Selector.Namespaces[0]).To(Equal(metav1.NamespaceDefault))
		})
	})
	Context("ChaosValidator of iochaos", func() {
		It("Validate", func() {

			type TestCase struct {
				name    string
				chaos   IoChaos
				execute func(chaos *IoChaos) error
				expect  string
			}
			duration := "400s"
			errorDuration := "400S"

			tcs := []TestCase{
				{
					name: "simple ValidateCreate",
					chaos: IoChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo1",
						},
					},
					execute: func(chaos *IoChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "",
				},
				{
					name: "simple ValidateUpdate",
					chaos: IoChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo2",
						},
					},
					execute: func(chaos *IoChaos) error {
						return chaos.ValidateUpdate(chaos)
					},
					expect: "",
				},
				{
					name: "simple ValidateDelete",
					chaos: IoChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo3",
						},
					},
					execute: func(chaos *IoChaos) error {
						return chaos.ValidateDelete()
					},
					expect: "",
				},
				{
					name: "only define the Scheduler",
					chaos: IoChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo4",
						},
						Spec: IoChaosSpec{
							Scheduler: &SchedulerSpec{
								Cron: "@every 10m",
							},
						},
					},
					execute: func(chaos *IoChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "only define the Duration",
					chaos: IoChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo5",
						},
						Spec: IoChaosSpec{
							Duration: &duration,
						},
					},
					execute: func(chaos *IoChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "parse the duration and scheduler error",
					chaos: IoChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo6",
						},
						Spec: IoChaosSpec{
							Duration:  &errorDuration,
							Scheduler: &SchedulerSpec{Cron: "xx"},
						},
					},
					execute: func(chaos *IoChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate value with FixedPercentPodMode",
					chaos: IoChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo7",
						},
						Spec: IoChaosSpec{
							Value: "0",
							Mode:  FixedPodMode,
						},
					},
					execute: func(chaos *IoChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate value with FixedPercentPodMode, parse value error",
					chaos: IoChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo8",
						},
						Spec: IoChaosSpec{
							Value: "num",
							Mode:  FixedPodMode,
						},
					},
					execute: func(chaos *IoChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate value with RandomMaxPercentPodMode",
					chaos: IoChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo9",
						},
						Spec: IoChaosSpec{
							Value: "0",
							Mode:  RandomMaxPercentPodMode,
						},
					},
					execute: func(chaos *IoChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate value with RandomMaxPercentPodMode ,parse value error",
					chaos: IoChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo10",
						},
						Spec: IoChaosSpec{
							Value: "num",
							Mode:  RandomMaxPercentPodMode,
						},
					},
					execute: func(chaos *IoChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate value with FixedPercentPodMode",
					chaos: IoChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo11",
						},
						Spec: IoChaosSpec{
							Value: "101",
							Mode:  FixedPercentPodMode,
						},
					},
					execute: func(chaos *IoChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate delay",
					chaos: IoChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo12",
						},
						Spec: IoChaosSpec{
							Delay:  "1S",
							Action: IoLatency,
						},
					},
					execute: func(chaos *IoChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "parse the scheduler.cron error",
					chaos: IoChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo15",
						},
						Spec: IoChaosSpec{
							Duration:  &duration,
							Scheduler: &SchedulerSpec{Cron: "xx"},
						},
					},
					execute: func(chaos *IoChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate the duration and the scheduler.cron conflict",
					chaos: IoChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo16",
						},
						Spec: IoChaosSpec{
							Duration:  &duration,
							Scheduler: &SchedulerSpec{Cron: "@every 1m"},
						},
					},
					execute: func(chaos *IoChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate the duration and the scheduler.cron conflict",
					chaos: IoChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo16",
						},
						Spec: IoChaosSpec{
							Duration:  &duration,
							Percent:   101,
							Scheduler: &SchedulerSpec{Cron: "@every 1m"},
						},
					},
					execute: func(chaos *IoChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate the duration and the scheduler.cron conflict",
					chaos: IoChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo16",
						},
						Spec: IoChaosSpec{
							Duration:  &duration,
							Percent:   -100,
							Scheduler: &SchedulerSpec{Cron: "@every 1m"},
						},
					},
					execute: func(chaos *IoChaos) error {
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
