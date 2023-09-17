// Copyright 2023 Chaos Mesh Authors.
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

package condition

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

var _ = Describe("Test condition controller", func() {
	reconciler := Reconciler{
		Object: &v1alpha1.PodChaos{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					v1alpha1.PauseAnnotationKey: "true",
				},
			},
		},
	}

	Context("Test diffConditions", func() {
		It("selected/allInjected/allRecovered state should be false when records is empty", func() {
			newConditionMap := diffConditions(reconciler.Object.DeepCopyObject().(v1alpha1.InnerObject))

			Expect(newConditionMap[v1alpha1.ConditionSelected].Status).To(Equal(corev1.ConditionFalse))
			Expect(newConditionMap[v1alpha1.ConditionAllInjected].Status).To(Equal(corev1.ConditionFalse))
			Expect(newConditionMap[v1alpha1.ConditionAllRecovered].Status).To(Equal(corev1.ConditionFalse))
		})

		It("Paused state should be true when pause annotation is true", func() {
			obj := reconciler.Object.DeepCopyObject().(v1alpha1.InnerObject)
			obj.SetAnnotations(map[string]string{
				v1alpha1.PauseAnnotationKey: "true",
			})
			newConditionMap := diffConditions(obj)

			Expect(newConditionMap[v1alpha1.ConditionPaused].Status).To(Equal(corev1.ConditionTrue))
		})

		It("AllInjected state should be true when all records are injected", func() {
			obj := reconciler.Object.DeepCopyObject().(v1alpha1.InnerObject)
			obj.GetStatus().Experiment.Records = append(obj.GetStatus().Experiment.Records, &v1alpha1.Record{
				Phase: v1alpha1.Injected,
			})
			newConditionMap := diffConditions(obj)

			Expect(newConditionMap[v1alpha1.ConditionAllInjected].Status).To(Equal(corev1.ConditionTrue))
		})

		It("AllRecovered state should be true when all records are recovered", func() {
			obj := reconciler.Object.DeepCopyObject().(v1alpha1.InnerObject)
			obj.GetStatus().Experiment.Records = append(obj.GetStatus().Experiment.Records, &v1alpha1.Record{
				Phase: v1alpha1.NotInjected,
			})
			newConditionMap := diffConditions(obj)

			Expect(newConditionMap[v1alpha1.ConditionAllRecovered].Status).To(Equal(corev1.ConditionTrue))
		})
	})
})
