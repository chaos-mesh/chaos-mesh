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

var _ = Describe("IoChaos Controller", func() {
	supportingModes := []v1alpha1.IOChaosAction{
		v1alpha1.IODelayAction,
		v1alpha1.IOErrnoAction,
		v1alpha1.IOMixedAction,
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
	Context("IoChaos Item", func() {
		It("should create successfully", func() {
			for _, actionTp := range supportingModes {
				key := types.NamespacedName{
					Name:      "t-iochaos" + "-" + randomStringWithCharset(10, charset),
					Namespace: "default",
				}

				duration := "60s"
				created := &v1alpha1.IoChaos{
					ObjectMeta: metav1.ObjectMeta{
						Name:      key.Name,
						Namespace: key.Namespace,
					},
					Spec: v1alpha1.IoChaosSpec{
						Selector: v1alpha1.SelectorSpec{
							Namespaces: []string{"default"},
						},
						Scheduler: &v1alpha1.SchedulerSpec{
							Cron: "@every 2m",
						},
						Action:   actionTp,
						Mode:     v1alpha1.OnePodMode,
						Duration: &duration,
						Layer:    v1alpha1.FileSystemLayer,
						Delay:    "10ms",
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
