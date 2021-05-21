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

package schedule

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var _ = Describe("Schedule", func() {

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context(("Schedule basic"), func() {
		It(("Should be created and deleted successfully"), func() {
			key := types.NamespacedName{
				Name:      "foo0",
				Namespace: "default",
			}
			duration := "100m"
			schedule := &v1alpha1.Schedule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo0",
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
					Type:              v1alpha1.TypeTimeChaos,
				},
				Status: v1alpha1.ScheduleStatus{
					LastScheduleTime: metav1.NewTime(time.Time{}),
				},
			}

			By("creating an API obj")
			Expect(k8sClient.Create(context.TODO(), schedule)).To(Succeed())

			fetched := &v1alpha1.Schedule{}
			Expect(k8sClient.Get(context.TODO(), key, fetched)).To(Succeed())
			Expect(fetched).To(Equal(schedule))

			By("deleting the created object")
			Expect(k8sClient.Delete(context.TODO(), schedule)).To(Succeed())
			Expect(k8sClient.Get(context.TODO(), key, schedule)).ToNot(Succeed())
		})
	})

	Context("Schedule cron", func() {
		It("should create non-concurrent chaos", func() {
			key := types.NamespacedName{
				Name:      "foo1",
				Namespace: "default",
			}
			duration := "100s"
			schedule := &v1alpha1.Schedule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo1",
					Namespace: "default",
				},
				Spec: v1alpha1.ScheduleSpec{
					Schedule: "@every 5s",
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
					HistoryLimit:      2,
					Type:              v1alpha1.TypeTimeChaos,
				},
				Status: v1alpha1.ScheduleStatus{
					LastScheduleTime: metav1.NewTime(time.Now()),
				},
			}

			By("creating a schedule obj")
			{
				Expect(k8sClient.Create(context.TODO(), schedule)).To(Succeed())
			}

			By("Reconciling the created schedule obj")
			{
				err := wait.Poll(time.Second*5, time.Minute*1, func() (ok bool, err error) {
					err = k8sClient.Get(context.TODO(), key, schedule)
					if err != nil {
						return false, err
					}
					return len(schedule.Status.Active) > 0, nil
				})
				Expect(err).ToNot(HaveOccurred())
			}

			By("Disallow concurrency")
			{
				time.Sleep(5 * time.Second)
				err := k8sClient.Get(context.TODO(), key, schedule)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(schedule.Status.Active)).To(Equal(1))
			}

			By("deleting the created object")
			{
				Expect(k8sClient.Delete(context.TODO(), schedule)).To(Succeed())
				Expect(k8sClient.Get(context.TODO(), key, schedule)).ToNot(Succeed())
			}
		})
		It("should create concurrent chaos", func() {
			key := types.NamespacedName{
				Name:      "foo2",
				Namespace: "default",
			}
			duration := "100s"
			schedule := &v1alpha1.Schedule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo2",
					Namespace: "default",
				},
				Spec: v1alpha1.ScheduleSpec{
					Schedule: "@every 5s",
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
					ConcurrencyPolicy: v1alpha1.AllowConcurrent,
					HistoryLimit:      2,
					Type:              v1alpha1.TypeTimeChaos,
				},
				Status: v1alpha1.ScheduleStatus{
					LastScheduleTime: metav1.NewTime(time.Now()),
				},
			}

			By("creating a schedule obj")
			{
				Expect(k8sClient.Create(context.TODO(), schedule)).To(Succeed())
			}

			By("Allowing concurrency and skip deleting running chaos")
			{
				err := wait.Poll(5*time.Second, 1*time.Minute, func() (done bool, err error) {
					err = k8sClient.Get(context.TODO(), key, schedule)
					if err != nil {
						return false, err
					}
					ctrl.Log.Info("active chaos", "size", len(schedule.Status.Active))
					return len(schedule.Status.Active) >= 4, nil
				})
				Expect(err).ToNot(HaveOccurred())
			}

			By("deleting the created object")
			{
				Expect(k8sClient.Delete(context.TODO(), schedule)).To(Succeed())
				Expect(k8sClient.Get(context.TODO(), key, schedule)).ToNot(Succeed())
			}
		})
		It("should collect garbage", func() {
			key := types.NamespacedName{
				Name:      "foo3",
				Namespace: "default",
			}
			duration := "3s"
			schedule := &v1alpha1.Schedule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo3",
					Namespace: "default",
				},
				Spec: v1alpha1.ScheduleSpec{
					Schedule: "@every 5s",
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
					ConcurrencyPolicy: v1alpha1.AllowConcurrent,
					HistoryLimit:      2,
					Type:              v1alpha1.TypeTimeChaos,
				},
				Status: v1alpha1.ScheduleStatus{
					LastScheduleTime: metav1.NewTime(time.Now()),
				},
			}

			By("creating a schedule obj")
			{
				Expect(k8sClient.Create(context.TODO(), schedule)).To(Succeed())
			}

			By("deleting outdated chaos")
			{
				time.Sleep(time.Minute * 1)
				err := wait.Poll(5*time.Second, 1*time.Minute, func() (done bool, err error) {
					err = k8sClient.Get(context.TODO(), key, schedule)
					if err != nil {
						return false, err
					}
					ctrl.Log.Info("active chaos", "size", len(schedule.Status.Active))
					return len(schedule.Status.Active) == 2, nil
				})
				Expect(err).ToNot(HaveOccurred())
			}

			By("deleting the created object")
			{
				Expect(k8sClient.Delete(context.TODO(), schedule)).To(Succeed())
				Expect(k8sClient.Get(context.TODO(), key, schedule)).ToNot(Succeed())
			}
		})
	})
})
