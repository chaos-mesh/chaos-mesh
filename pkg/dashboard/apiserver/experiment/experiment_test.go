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

package experiment

import (
	"context"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

var _ = Describe("Experiment.APIServer", func() {
	BeforeEach(func() {
		Expect(k8sClient.Create(context.TODO(), &v1alpha1.PodChaos{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod-chaos",
				Namespace: "default",
			},
			Spec: v1alpha1.PodChaosSpec{
				Action: v1alpha1.PodKillAction,
				ContainerSelector: v1alpha1.ContainerSelector{
					PodSelector: v1alpha1.PodSelector{
						Mode: v1alpha1.OneMode,
					},
				},
			},
		})).To(Succeed())
	})

	Context("Delete Experiment", func() {
		It("should delete experiment by UID", func() {
			chaos := &v1alpha1.PodChaos{}
			Expect(k8sClient.Get(context.TODO(), types.NamespacedName{
				Name:      "test-pod-chaos",
				Namespace: "default",
			}, chaos)).To(Succeed())

			Expect(deleteChaosByUID(&gin.Context{}, k8sClient, string(chaos.UID))).To(BeTrue())

			err := k8sClient.Get(context.TODO(), types.NamespacedName{
				Name:      "test-pod-chaos",
				Namespace: "default",
			}, chaos)
			Expect(err).To(HaveOccurred())
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
		})
		It("should fail if experiment UID does not found", func() {
			chaos := &v1alpha1.PodChaos{}
			Expect(k8sClient.Get(context.TODO(), types.NamespacedName{
				Name:      "test-pod-chaos-not-found",
				Namespace: "default",
			}, chaos)).To(Succeed())

			Expect(deleteChaosByUID(&gin.Context{}, k8sClient, string(chaos.UID))).To(BeFalse())
		})
	})
})
