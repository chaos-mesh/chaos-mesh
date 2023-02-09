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

package common

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common/finalizers"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var _ = Describe("Finalizer", func() {

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context("Adding finalizer", func() {
		It("should add record finalizer", func() {
			key := types.NamespacedName{
				Name:      "final1",
				Namespace: "default",
			}
			duration := "1000s"
			chaos := &v1alpha1.TimeChaos{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "final1",
					Namespace: "default",
				},
				Spec: v1alpha1.TimeChaosSpec{
					TimeOffset: "100ms",
					ClockIds:   []string{"CLOCK_REALTIME"},
					Duration:   &duration,
					ContainerSelector: v1alpha1.ContainerSelector{
						PodSelector: v1alpha1.PodSelector{
							Mode: v1alpha1.OneMode,
						},
					},
				},
			}

			By("creating a chaos")
			{
				Expect(k8sClient.Create(context.TODO(), chaos)).To(Succeed())
			}

			By("Adding finalizers")
			{
				err := wait.Poll(time.Second*1, time.Second*10, func() (ok bool, err error) {
					err = k8sClient.Get(context.TODO(), key, chaos)
					if err != nil {
						return false, err
					}
					return len(chaos.GetObjectMeta().GetFinalizers()) > 0 && chaos.GetObjectMeta().GetFinalizers()[0] == finalizers.RecordFinalizer, nil
				})
				Expect(err).ToNot(HaveOccurred())
			}

			By("deleting the created object")
			{
				Expect(k8sClient.Delete(context.TODO(), chaos)).To(Succeed())
			}
		})
	})
})
