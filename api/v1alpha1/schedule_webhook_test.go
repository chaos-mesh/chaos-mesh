// Copyright 2021 Chaos Mesh Authors.
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

var _ = Describe("schedule_webhook", func() {
	Context("webhook.Validator of schedule", func() {
		It("Validate", func() {

			type TestCase struct {
				name     string
				schedule Schedule
				execute  func(schedule *Schedule) error
				expect   string
			}
			tcs := []TestCase{
				{
					name: "validation for normal chaos",
					schedule: Schedule{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo1",
						},
						Spec: ScheduleSpec{
							ScheduleItem: ScheduleItem{EmbedChaos: EmbedChaos{PodChaos: &PodChaosSpec{
								Action: ContainerKillAction,
							}}},
							Type:     ScheduleTypePodChaos,
							Schedule: "@every 5s",
						},
					},
					execute: func(schedule *Schedule) error {
						return schedule.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validation for schedule",
					schedule: Schedule{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo2",
						},
						Spec: ScheduleSpec{
							ScheduleItem: ScheduleItem{EmbedChaos: EmbedChaos{PodChaos: &PodChaosSpec{}}},
							Type:         ScheduleTypePodChaos,
							Schedule:     "@every -5s",
						},
					},
					execute: func(schedule *Schedule) error {
						return schedule.ValidateCreate()
					},
					expect: "",
				},
				{
					name: "validation for workflow",
					schedule: Schedule{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo3",
						},
						Spec: ScheduleSpec{
							ScheduleItem: ScheduleItem{Workflow: &WorkflowSpec{}},
							Type:         ScheduleTypeWorkflow,
							Schedule:     "@every 5s",
						},
					},
					execute: func(schedule *Schedule) error {
						return schedule.ValidateCreate()
					},
					expect: "",
				},
			}

			for _, tc := range tcs {
				err := tc.execute(&tc.schedule)
				if tc.expect == "error" {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
			}
		})
	})
})
