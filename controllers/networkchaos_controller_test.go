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

package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("NetworkChaos Controller", func() {
	supportingModes := []v1alpha1.NetworkChaosAction{
		v1alpha1.NetemAction,
		v1alpha1.DelayAction,
		v1alpha1.LossAction,
		v1alpha1.DuplicateAction,
		v1alpha1.CorruptAction,
		v1alpha1.PartitionAction,
	}

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	// Add Tests for OpenAPI validation (or additional CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("NetworkChaos Item", func() {
		It("should create successfully", func() {
			for _, action := range supportingModes {
				key := types.NamespacedName{
					Name:      "network-chaos" + "-" + randomStringWithCharset(10, charset),
					Namespace: "default",
				}

				created := &v1alpha1.NetworkChaos{
					ObjectMeta: metav1.ObjectMeta{
						Name:      key.Name,
						Namespace: key.Namespace,
					},
					Spec: v1alpha1.NetworkChaosSpec{
						Selector: v1alpha1.SelectorSpec{
							Namespaces: []string{"default"},
						},
						Scheduler: &v1alpha1.SchedulerSpec{
							Cron: "@every 2m",
						},
						Action: action,
						Mode:   v1alpha1.OnePodMode,
						TcParameter: v1alpha1.TcParameter{
							Delay: &v1alpha1.DelaySpec{
								Latency:     "10ms",
								Correlation: "0.15",
								Jitter:      "2ms",
							},
						},
					},
				}

				By("creating an API obj")
				Expect(k8sClient.Create(context.TODO(), created)).Should(Succeed())

				By("deleting the created object")
				Expect(k8sClient.Delete(context.TODO(), created)).To(Succeed())
				time.Sleep(1 * time.Second)
				Expect(k8sClient.Get(context.TODO(), key, created)).ToNot(Succeed())
			}
		})
	})
})
