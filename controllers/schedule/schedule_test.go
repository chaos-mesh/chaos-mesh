// Copyright 2019 Chaos Mesh Authors.
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

package schedule

import (
	"context"
	"time"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var _ = Describe("Schedule", func() {
	var (
		key              types.NamespacedName
		created, fetched *v1alpha1.Schedule
	)

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context(("Schedule basic"), func() {
		It(("Should be created and deleted successfully"), func() {
			key = types.NamespacedName{
				Name:      "foo",
				Namespace: "default",
			}
			duration := "100s"
			created = &v1alpha1.Schedule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				Spec: v1alpha1.ScheduleSpec{
					Schedule: "@every 10s",
					EmbedChaos: v1alpha1.EmbedChaos{
						TimeChaos: &v1alpha1.TimeChaosSpec{
							TimeOffset: "100ms",
							ClockIds:   []string{"CLOCK_REALTIME"},
							Duration:   &duration,
							ContainerSelector: v1alpha1.ContainerSelector{
								PodSelector: v1alpha1.PodSelector{
									Mode: v1alpha1.OnePodMode,
								},
							},
						},
					},
					ConcurrencyPolicy: v1alpha1.ForbidConcurrent,
					HistoryLimit:      5,
					Type:              v1alpha1.TypeTask,
				},
			}

			By("creating an API obj")
			Expect(k8sClient.Create(context.TODO(), created)).To(Succeed())

			fetched = &v1alpha1.Schedule{}
			Expect(k8sClient.Get(context.TODO(), key, fetched)).To(Succeed())
			Expect(fetched).To(Equal(created))

			By("deleting the created object")
			Expect(k8sClient.Delete(context.TODO(), created)).To(Succeed())
			Expect(k8sClient.Get(context.TODO(), key, created)).ToNot(Succeed())
		})
	})

	Context("Schedule cron", func() {
		It("should create chaos", func() {
			key = types.NamespacedName{
				Name:      "foo",
				Namespace: "default",
			}
			duration := "100s"
			schedule := &v1alpha1.Schedule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				Spec: v1alpha1.ScheduleSpec{
					Schedule: "@every 10s",
					EmbedChaos: v1alpha1.EmbedChaos{
						TimeChaos: &v1alpha1.TimeChaosSpec{
							TimeOffset: "100ms",
							ClockIds:   []string{"CLOCK_REALTIME"},
							Duration:   &duration,
							ContainerSelector: v1alpha1.ContainerSelector{
								PodSelector: v1alpha1.PodSelector{
									Mode: v1alpha1.OnePodMode,
								},
							},
						},
					},
					ConcurrencyPolicy: v1alpha1.ForbidConcurrent,
					HistoryLimit:      5,
					Type:              v1alpha1.TypeTask,
				},
			}

			By("creating an API obj")
			Expect(k8sClient.Create(context.TODO(), schedule)).To(Succeed())

			err := wait.Poll(time.Second*5, time.Minute*1, func() (ok bool, err error) {
				err = k8sClient.Get(context.TODO(), key, schedule)
				ok = len(schedule.Status.Active) > 0
				return
			})
			Expect(err).ToNot(HaveOccurred())

			By("deleting the created object")
			Expect(k8sClient.Delete(context.TODO(), schedule)).To(Succeed())
			Expect(k8sClient.Get(context.TODO(), key, schedule)).ToNot(Succeed())
		})
	})
})
