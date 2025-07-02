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
							Name:      "foo",
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
						_, err := schedule.ValidateCreate(context.Background(), schedule)
						return err
					},
					expect: "error",
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
						_, err := schedule.ValidateCreate(context.Background(), schedule)
						return err
					},
					expect: "",
				},
				{
					name: "validation for cron with second",
					schedule: Schedule{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo3",
						},
						Spec: ScheduleSpec{
							ScheduleItem: ScheduleItem{Workflow: &WorkflowSpec{}},
							Type:         ScheduleTypeWorkflow,
							Schedule:     "*/1 * * * * *",
						},
					},
					execute: func(schedule *Schedule) error {
						_, err := schedule.ValidateCreate(context.Background(), schedule)
						return err
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

	Context("webhook.Default of schedule", func() {
		s := Schedule{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: metav1.NamespaceDefault,
				Name:      "foo3",
			},
			Spec: ScheduleSpec{
				ScheduleItem: ScheduleItem{Workflow: &WorkflowSpec{}},
				Type:         ScheduleTypeWorkflow,
				Schedule:     "*/1 * * * * *",
			},
		}
		s.Default(context.Background(), &s)
		Expect(s.Spec.ConcurrencyPolicy).To(Equal(ForbidConcurrent))
	})
})
