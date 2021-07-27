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

package desiredphase

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"

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

	Context("Setting phase", func() {
		It("should set phase to running", func() {
			key := types.NamespacedName{
				Name:      "foo1",
				Namespace: "default",
			}
			duration := "10s"
			chaos := &v1alpha1.TimeChaos{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo1",
					Namespace: "default",
				},
				Spec: v1alpha1.TimeChaosSpec{
					TimeOffset: "100ms",
					ClockIds:   []string{"CLOCK_REALTIME"},
					Duration:   &duration,
					ContainerSelector: v1alpha1.ContainerSelector{
						PodSelector: v1alpha1.PodSelector{
							Mode: v1alpha1.OnePodMode,
						},
					},
				},
			}

			By("creating a chaos")
			{
				Expect(k8sClient.Create(context.TODO(), chaos)).To(Succeed())
			}

			By("Reconciling desired phase")
			{
				err := wait.Poll(time.Second*1, time.Second*10, func() (ok bool, err error) {
					err = k8sClient.Get(context.TODO(), key, chaos)
					if err != nil {
						return false, err
					}
					return chaos.GetStatus().Experiment.DesiredPhase == v1alpha1.RunningPhase, nil
				})
				Expect(err).ToNot(HaveOccurred())
				err = wait.Poll(time.Second*1, time.Second*10, func() (ok bool, err error) {
					err = k8sClient.Get(context.TODO(), key, chaos)
					if err != nil {
						return false, err
					}
					return chaos.GetStatus().Experiment.DesiredPhase == v1alpha1.StoppedPhase, nil
				})
				Expect(err).ToNot(HaveOccurred())
			}

			By("deleting the created object")
			{
				Expect(k8sClient.Delete(context.TODO(), chaos)).To(Succeed())
				Expect(k8sClient.Get(context.TODO(), key, chaos)).ToNot(Succeed())
			}
		})
		It("should stop paused chaos", func() {
			key := types.NamespacedName{
				Name:      "foo2",
				Namespace: "default",
			}
			duration := "1000s"
			chaos := &v1alpha1.TimeChaos{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo2",
					Namespace: "default",
				},
				Spec: v1alpha1.TimeChaosSpec{
					TimeOffset: "100ms",
					ClockIds:   []string{"CLOCK_REALTIME"},
					Duration:   &duration,
					ContainerSelector: v1alpha1.ContainerSelector{
						PodSelector: v1alpha1.PodSelector{
							Mode: v1alpha1.OnePodMode,
						},
					},
				},
			}

			By("creating a chaos")
			{
				Expect(k8sClient.Create(context.TODO(), chaos)).To(Succeed())
			}

			By("Reconciling desired phase")
			{
				err := wait.Poll(time.Second*1, time.Second*10, func() (ok bool, err error) {
					err = k8sClient.Get(context.TODO(), key, chaos)
					if err != nil {
						return false, err
					}
					return chaos.GetStatus().Experiment.DesiredPhase == v1alpha1.RunningPhase, nil
				})
				Expect(err).ToNot(HaveOccurred())
			}
			By("Pause chaos")
			{
				err := retry.RetryOnConflict(retry.DefaultRetry, func() (err error) {
					err = k8sClient.Get(context.TODO(), key, chaos)
					if err != nil {
						return err
					}
					chaos.SetAnnotations(map[string]string{v1alpha1.PauseAnnotationKey: "true"})
					return k8sClient.Update(context.TODO(), chaos)
				})
				Expect(err).ToNot(HaveOccurred())
				err = wait.Poll(time.Second*5, time.Second*60, func() (ok bool, err error) {
					err = k8sClient.Get(context.TODO(), key, chaos)
					if err != nil {
						return false, err
					}
					return chaos.GetStatus().Experiment.DesiredPhase == v1alpha1.StoppedPhase, nil
				})
				Expect(err).ToNot(HaveOccurred())
			}

			By("Resume chaos")
			{
				err := retry.RetryOnConflict(retry.DefaultRetry, func() (err error) {
					err = k8sClient.Get(context.TODO(), key, chaos)
					if err != nil {
						return err
					}
					chaos.SetAnnotations(map[string]string{v1alpha1.PauseAnnotationKey: "false"})
					return k8sClient.Update(context.TODO(), chaos)
				})
				Expect(err).ToNot(HaveOccurred())
				err = wait.Poll(time.Second*5, time.Second*60, func() (ok bool, err error) {
					err = k8sClient.Get(context.TODO(), key, chaos)
					if err != nil {
						return false, err
					}
					return chaos.GetStatus().Experiment.DesiredPhase == v1alpha1.RunningPhase, nil
				})
				Expect(err).ToNot(HaveOccurred())
			}

			By("deleting the created object")
			{
				Expect(k8sClient.Delete(context.TODO(), chaos)).To(Succeed())
				Expect(k8sClient.Get(context.TODO(), key, chaos)).ToNot(Succeed())
			}
		})
	})
})
