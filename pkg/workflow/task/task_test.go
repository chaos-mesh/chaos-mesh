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

package task

import (
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func TestTask(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Workflow Task")
}

var _ = Describe("Workflow tasks", func() {
	It("builds a one-shot pod using the task's configured identity and placement", func() {
		limits := corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("250m")}
		template := v1alpha1.Task{
			Container: &corev1.Container{
				Name:      "task",
				Image:     "example:v1",
				Args:      []string{"verify"},
				Resources: corev1.ResourceRequirements{Limits: limits},
			},
			ServiceAccountName: "workflow-runner",
			Volumes:            []corev1.Volume{{Name: "scratch"}},
			NodeSelector:       map[string]string{"pool": "workflow"},
			Tolerations:        []corev1.Toleration{{Key: "dedicated", Value: "workflow"}},
		}

		pod, err := SpawnPodForTask(template)
		Expect(err).NotTo(HaveOccurred())
		Expect(pod.RestartPolicy).To(Equal(corev1.RestartPolicyNever))
		Expect(pod.ServiceAccountName).To(Equal(template.ServiceAccountName))
		Expect(pod.Volumes).To(Equal(template.Volumes))
		Expect(pod.NodeSelector).To(Equal(template.NodeSelector))
		Expect(pod.Tolerations).To(Equal(template.Tolerations))
		Expect(pod.Containers).To(Equal([]corev1.Container{*template.Container}))
		pod.Containers[0].Args[0] = "changed"
		pod.Containers[0].Resources.Limits[corev1.ResourceCPU] = resource.MustParse("2")
		pod.NodeSelector["pool"] = "other"
		pod.Tolerations[0].Value = "other"
		Expect(template.Container.Args).To(Equal([]string{"verify"}))
		Expect(template.Container.Resources.Limits.Cpu().MilliValue()).To(Equal(int64(250)))
		Expect(template.NodeSelector).To(HaveKeyWithValue("pool", "workflow"))
		Expect(template.Tolerations[0].Value).To(Equal("workflow"))
	})

	It("distinguishes false branch expressions from unevaluable ones", func() {
		evaluator := NewEvaluator(logr.Discard(), nil)
		branches := []v1alpha1.ConditionalBranch{
			{Target: "success", Expression: "exitCode == 0"},
			{Target: "failure", Expression: "exitCode != 0"},
			{Target: "invalid-condition", Expression: "exitCode =="},
			{Target: "not-a-condition", Expression: "exitCode + 1"},
		}

		Expect(evaluator.EvaluateConditionBranches(branches, map[string]any{"exitCode": int32(0)})).To(Equal([]v1alpha1.ConditionalBranchStatus{
			{Target: "success", EvaluationResult: corev1.ConditionTrue},
			{Target: "failure", EvaluationResult: corev1.ConditionFalse},
			{Target: "invalid-condition", EvaluationResult: corev1.ConditionUnknown},
			{Target: "not-a-condition", EvaluationResult: corev1.ConditionUnknown},
		}))
	})
})
