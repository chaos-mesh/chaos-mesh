// Copyright Chaos Mesh Authors.
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

package statuscheck

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var _ = Describe("StatusCheck", func() {

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context("Reconcile Synchronous StatusCheck", func() {
		It("success threshold exceed", func() {
			key := types.NamespacedName{
				Name:      "foo1",
				Namespace: "default",
			}
			statusCheck := &v1alpha1.StatusCheck{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo1",
					Namespace: "default",
				},
				Spec: v1alpha1.StatusCheckSpec{
					Mode:                v1alpha1.StatusCheckSynchronous,
					Type:                v1alpha1.TypeHTTP,
					IntervalSeconds:     1,
					TimeoutSeconds:      1,
					FailureThreshold:    3,
					SuccessThreshold:    1,
					RecordsHistoryLimit: 10,
					EmbedStatusCheck: &v1alpha1.EmbedStatusCheck{
						HTTPStatusCheck: &v1alpha1.HTTPStatusCheck{
							RequestUrl:  "http://123.123.123.123",
							RequestBody: "success",
							Criteria: v1alpha1.HTTPCriteria{
								StatusCode: "200",
							},
						},
					},
				},
			}

			By("creating a status check")
			{
				Expect(k8sClient.Create(context.TODO(), statusCheck)).To(Succeed())
			}

			By("reconciling status check")
			{
				Eventually(func() ([]v1alpha1.StatusCheckCondition, error) {
					err := k8sClient.Get(context.TODO(), key, statusCheck)
					if err != nil {
						return nil, err
					}
					return statusCheck.Status.Conditions, nil
				}, 5*time.Second, time.Second).Should(
					ConsistOf(
						MatchFields(IgnoreExtras, Fields{
							"Type":   Equal(v1alpha1.StatusCheckConditionCompleted),
							"Status": Equal(corev1.ConditionTrue),
						}),
						MatchFields(IgnoreExtras, Fields{
							"Type":   Equal(v1alpha1.StatusCheckConditionSuccessThresholdExceed),
							"Status": Equal(corev1.ConditionTrue),
						}),
						MatchFields(IgnoreExtras, Fields{
							"Type":   Equal(v1alpha1.StatusCheckConditionFailureThresholdExceed),
							"Status": Equal(corev1.ConditionFalse),
						}),
						MatchFields(IgnoreExtras, Fields{
							"Type":   Equal(v1alpha1.StatusCheckConditionDurationExceed),
							"Status": Equal(corev1.ConditionFalse),
						}),
					))
			}

			By("deleting the created object")
			{
				Expect(k8sClient.Delete(context.TODO(), statusCheck)).To(Succeed())
			}
		})
		It("failure threshold exceed", func() {
			key := types.NamespacedName{
				Name:      "foo1",
				Namespace: "default",
			}
			statusCheck := &v1alpha1.StatusCheck{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo1",
					Namespace: "default",
				},
				Spec: v1alpha1.StatusCheckSpec{
					Mode:                v1alpha1.StatusCheckSynchronous,
					Type:                v1alpha1.TypeHTTP,
					IntervalSeconds:     1,
					TimeoutSeconds:      1,
					FailureThreshold:    3,
					SuccessThreshold:    1,
					RecordsHistoryLimit: 10,
					EmbedStatusCheck: &v1alpha1.EmbedStatusCheck{
						HTTPStatusCheck: &v1alpha1.HTTPStatusCheck{
							RequestUrl:  "http://123.123.123.123",
							RequestBody: "failure",
							Criteria: v1alpha1.HTTPCriteria{
								StatusCode: "200",
							},
						},
					},
				},
			}

			By("creating a status check")
			{
				Expect(k8sClient.Create(context.TODO(), statusCheck)).To(Succeed())
			}

			By("reconciling status check")
			{
				Eventually(func() ([]v1alpha1.StatusCheckCondition, error) {
					err := k8sClient.Get(context.TODO(), key, statusCheck)
					if err != nil {
						return nil, err
					}
					return statusCheck.Status.Conditions, nil
				}, 10*time.Second, time.Second).Should(
					ConsistOf(
						MatchFields(IgnoreExtras, Fields{
							"Type":   Equal(v1alpha1.StatusCheckConditionCompleted),
							"Status": Equal(corev1.ConditionTrue),
						}),
						MatchFields(IgnoreExtras, Fields{
							"Type":   Equal(v1alpha1.StatusCheckConditionSuccessThresholdExceed),
							"Status": Equal(corev1.ConditionFalse),
						}),
						MatchFields(IgnoreExtras, Fields{
							"Type":   Equal(v1alpha1.StatusCheckConditionFailureThresholdExceed),
							"Status": Equal(corev1.ConditionTrue),
						}),
						MatchFields(IgnoreExtras, Fields{
							"Type":   Equal(v1alpha1.StatusCheckConditionDurationExceed),
							"Status": Equal(corev1.ConditionFalse),
						}),
					))
			}

			By("deleting the created object")
			{
				Expect(k8sClient.Delete(context.TODO(), statusCheck)).To(Succeed())
			}
		})
		It("duration exceed", func() {
			key := types.NamespacedName{
				Name:      "foo1",
				Namespace: "default",
			}
			duration := "100ms"
			statusCheck := &v1alpha1.StatusCheck{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo1",
					Namespace: "default",
				},
				Spec: v1alpha1.StatusCheckSpec{
					Mode:                v1alpha1.StatusCheckSynchronous,
					Type:                v1alpha1.TypeHTTP,
					IntervalSeconds:     1,
					TimeoutSeconds:      1,
					FailureThreshold:    3,
					SuccessThreshold:    3,
					RecordsHistoryLimit: 10,
					Duration:            &duration,
					EmbedStatusCheck: &v1alpha1.EmbedStatusCheck{
						HTTPStatusCheck: &v1alpha1.HTTPStatusCheck{
							RequestUrl:  "http://123.123.123.123",
							RequestBody: "success",
							Criteria: v1alpha1.HTTPCriteria{
								StatusCode: "200",
							},
						},
					},
				},
			}

			By("creating a status check")
			{
				Expect(k8sClient.Create(context.TODO(), statusCheck)).To(Succeed())
			}

			By("reconciling status check")
			{
				Eventually(func() ([]v1alpha1.StatusCheckCondition, error) {
					err := k8sClient.Get(context.TODO(), key, statusCheck)
					if err != nil {
						return nil, err
					}
					return statusCheck.Status.Conditions, nil
				}, 10*time.Second, time.Second).Should(
					ConsistOf(
						MatchFields(IgnoreExtras, Fields{
							"Type":   Equal(v1alpha1.StatusCheckConditionCompleted),
							"Status": Equal(corev1.ConditionTrue),
						}),
						MatchFields(IgnoreExtras, Fields{
							"Type":   Equal(v1alpha1.StatusCheckConditionSuccessThresholdExceed),
							"Status": Equal(corev1.ConditionFalse),
						}),
						MatchFields(IgnoreExtras, Fields{
							"Type":   Equal(v1alpha1.StatusCheckConditionFailureThresholdExceed),
							"Status": Equal(corev1.ConditionFalse),
						}),
						MatchFields(IgnoreExtras, Fields{
							"Type":   Equal(v1alpha1.StatusCheckConditionDurationExceed),
							"Status": Equal(corev1.ConditionTrue),
						}),
					))
			}

			By("deleting the created object")
			{
				Expect(k8sClient.Delete(context.TODO(), statusCheck)).To(Succeed())
			}
		})
		It("failure threshold exceed, execution timeout", func() {
			key := types.NamespacedName{
				Name:      "foo1",
				Namespace: "default",
			}
			statusCheck := &v1alpha1.StatusCheck{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo1",
					Namespace: "default",
				},
				Spec: v1alpha1.StatusCheckSpec{
					Mode:                v1alpha1.StatusCheckSynchronous,
					Type:                v1alpha1.TypeHTTP,
					IntervalSeconds:     1,
					TimeoutSeconds:      1,
					FailureThreshold:    3,
					SuccessThreshold:    1,
					RecordsHistoryLimit: 10,
					EmbedStatusCheck: &v1alpha1.EmbedStatusCheck{
						HTTPStatusCheck: &v1alpha1.HTTPStatusCheck{
							RequestUrl:  "http://123.123.123.123",
							RequestBody: "timeout",
							Criteria: v1alpha1.HTTPCriteria{
								StatusCode: "200",
							},
						},
					},
				},
			}

			By("creating a status check")
			{
				Expect(k8sClient.Create(context.TODO(), statusCheck)).To(Succeed())
			}

			By("reconciling status check")
			{
				Eventually(func() ([]v1alpha1.StatusCheckCondition, error) {
					err := k8sClient.Get(context.TODO(), key, statusCheck)
					if err != nil {
						return nil, err
					}
					return statusCheck.Status.Conditions, nil
				}, 10*time.Second, time.Second).Should(
					ConsistOf(
						MatchFields(IgnoreExtras, Fields{
							"Type":   Equal(v1alpha1.StatusCheckConditionCompleted),
							"Status": Equal(corev1.ConditionTrue),
						}),
						MatchFields(IgnoreExtras, Fields{
							"Type":   Equal(v1alpha1.StatusCheckConditionSuccessThresholdExceed),
							"Status": Equal(corev1.ConditionFalse),
						}),
						MatchFields(IgnoreExtras, Fields{
							"Type":   Equal(v1alpha1.StatusCheckConditionFailureThresholdExceed),
							"Status": Equal(corev1.ConditionTrue),
						}),
						MatchFields(IgnoreExtras, Fields{
							"Type":   Equal(v1alpha1.StatusCheckConditionDurationExceed),
							"Status": Equal(corev1.ConditionFalse),
						}),
					))
			}

			By("deleting the created object")
			{
				Expect(k8sClient.Delete(context.TODO(), statusCheck)).To(Succeed())
			}
		})

		Context("Reconcile Continuous StatusCheck", func() {
			It("success threshold exceed", func() {
				key := types.NamespacedName{
					Name:      "foo1",
					Namespace: "default",
				}
				duration := "10s"
				statusCheck := &v1alpha1.StatusCheck{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo1",
						Namespace: "default",
					},
					Spec: v1alpha1.StatusCheckSpec{
						Mode:                v1alpha1.StatusCheckContinuous,
						Type:                v1alpha1.TypeHTTP,
						Duration:            &duration,
						IntervalSeconds:     1,
						TimeoutSeconds:      1,
						FailureThreshold:    3,
						SuccessThreshold:    1,
						RecordsHistoryLimit: 10,
						EmbedStatusCheck: &v1alpha1.EmbedStatusCheck{
							HTTPStatusCheck: &v1alpha1.HTTPStatusCheck{
								RequestUrl:  "http://123.123.123.123",
								RequestBody: "success",
								Criteria: v1alpha1.HTTPCriteria{
									StatusCode: "200",
								},
							},
						},
					},
				}

				By("creating a status check")
				{
					Expect(k8sClient.Create(context.TODO(), statusCheck)).To(Succeed())
				}

				By("reconciling status check, success threshold exceed but not completed")
				{
					Eventually(func() ([]v1alpha1.StatusCheckCondition, error) {
						err := k8sClient.Get(context.TODO(), key, statusCheck)
						if err != nil {
							return nil, err
						}
						return statusCheck.Status.Conditions, nil
					}, 5*time.Second, time.Second).Should(
						ConsistOf(
							MatchFields(IgnoreExtras, Fields{
								"Type":   Equal(v1alpha1.StatusCheckConditionCompleted),
								"Status": Equal(corev1.ConditionFalse),
							}),
							MatchFields(IgnoreExtras, Fields{
								"Type":   Equal(v1alpha1.StatusCheckConditionSuccessThresholdExceed),
								"Status": Equal(corev1.ConditionTrue),
							}),
							MatchFields(IgnoreExtras, Fields{
								"Type":   Equal(v1alpha1.StatusCheckConditionFailureThresholdExceed),
								"Status": Equal(corev1.ConditionFalse),
							}),
							MatchFields(IgnoreExtras, Fields{
								"Type":   Equal(v1alpha1.StatusCheckConditionDurationExceed),
								"Status": Equal(corev1.ConditionFalse),
							}),
						))
				}

				By("reconciling status check, duration exceed and completed")
				{
					Eventually(func() ([]v1alpha1.StatusCheckCondition, error) {
						err := k8sClient.Get(context.TODO(), key, statusCheck)
						if err != nil {
							return nil, err
						}
						return statusCheck.Status.Conditions, nil
					}, 10*time.Second, time.Second).Should(
						ConsistOf(
							MatchFields(IgnoreExtras, Fields{
								"Type":   Equal(v1alpha1.StatusCheckConditionCompleted),
								"Status": Equal(corev1.ConditionTrue),
							}),
							MatchFields(IgnoreExtras, Fields{
								"Type":   Equal(v1alpha1.StatusCheckConditionSuccessThresholdExceed),
								"Status": Equal(corev1.ConditionTrue),
							}),
							MatchFields(IgnoreExtras, Fields{
								"Type":   Equal(v1alpha1.StatusCheckConditionFailureThresholdExceed),
								"Status": Equal(corev1.ConditionFalse),
							}),
							MatchFields(IgnoreExtras, Fields{
								"Type":   Equal(v1alpha1.StatusCheckConditionDurationExceed),
								"Status": Equal(corev1.ConditionTrue),
							}),
						))
				}

				By("deleting the created object")
				{
					Expect(k8sClient.Delete(context.TODO(), statusCheck)).To(Succeed())
				}
			})

			It("failure threshold exceed", func() {
				key := types.NamespacedName{
					Name:      "foo1",
					Namespace: "default",
				}
				duration := "10s"
				statusCheck := &v1alpha1.StatusCheck{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo1",
						Namespace: "default",
					},
					Spec: v1alpha1.StatusCheckSpec{
						Mode:                v1alpha1.StatusCheckContinuous,
						Type:                v1alpha1.TypeHTTP,
						Duration:            &duration,
						IntervalSeconds:     1,
						TimeoutSeconds:      1,
						FailureThreshold:    3,
						SuccessThreshold:    1,
						RecordsHistoryLimit: 10,
						EmbedStatusCheck: &v1alpha1.EmbedStatusCheck{
							HTTPStatusCheck: &v1alpha1.HTTPStatusCheck{
								RequestUrl:  "http://123.123.123.123",
								RequestBody: "failure",
								Criteria: v1alpha1.HTTPCriteria{
									StatusCode: "200",
								},
							},
						},
					},
				}

				By("creating a status check")
				{
					Expect(k8sClient.Create(context.TODO(), statusCheck)).To(Succeed())
				}

				By("reconciling status check, failure threshold exceed and completed (duration not exceed)")
				{
					Eventually(func() ([]v1alpha1.StatusCheckCondition, error) {
						err := k8sClient.Get(context.TODO(), key, statusCheck)
						if err != nil {
							return nil, err
						}
						return statusCheck.Status.Conditions, nil
					}, 5*time.Second, time.Second).Should(
						ConsistOf(
							MatchFields(IgnoreExtras, Fields{
								"Type":   Equal(v1alpha1.StatusCheckConditionCompleted),
								"Status": Equal(corev1.ConditionTrue),
							}),
							MatchFields(IgnoreExtras, Fields{
								"Type":   Equal(v1alpha1.StatusCheckConditionSuccessThresholdExceed),
								"Status": Equal(corev1.ConditionFalse),
							}),
							MatchFields(IgnoreExtras, Fields{
								"Type":   Equal(v1alpha1.StatusCheckConditionFailureThresholdExceed),
								"Status": Equal(corev1.ConditionTrue),
							}),
							MatchFields(IgnoreExtras, Fields{
								"Type":   Equal(v1alpha1.StatusCheckConditionDurationExceed),
								"Status": Equal(corev1.ConditionFalse),
							}),
						))
				}

				By("deleting the created object")
				{
					Expect(k8sClient.Delete(context.TODO(), statusCheck)).To(Succeed())
				}
			})
		})
	})
})
