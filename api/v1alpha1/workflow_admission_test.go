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

package v1alpha1

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Workflow admission", func() {
	ctx := context.Background()
	admission := &Workflow{}

	It("accepts a serial workflow containing tasks, a schedule and a status check", func() {
		deadline := "5s"
		workflow := &Workflow{Spec: WorkflowSpec{
			Entry: "entry",
			Templates: []Template{
				{Name: "entry", Type: TypeSerial, Children: []string{"task", "schedule", "check", "wait"}},
				{Name: "task", Type: TypeTask, Task: &Task{Container: &corev1.Container{Name: "job", Image: "test-image"}}},
				{Name: "schedule", Type: TypeSchedule, Schedule: &ChaosOnlyScheduleSpec{Schedule: "*/5 * * * *"}},
				{Name: "check", Type: TypeStatusCheck, StatusCheck: &StatusCheckSpec{}},
				{Name: "wait", Type: TypeSuspend, Deadline: &deadline},
			},
		}}
		Expect(admission.ValidateCreate(ctx, workflow)).Error().NotTo(HaveOccurred())
		Expect(admission.ValidateUpdate(ctx, workflow, workflow.DeepCopy())).Error().NotTo(HaveOccurred())
		Expect(admission.ValidateDelete(ctx, workflow)).Error().NotTo(HaveOccurred())

		invalid := workflow.DeepCopy()
		invalid.Spec.Templates[0].Children = []string{"removed"}
		Expect(admission.ValidateUpdate(ctx, workflow, invalid)).Error().To(MatchError(ContainSubstring("spec.templates[0].children[0]")))
	})

	DescribeTable("rejects malformed templates during creation",
		func(template Template, message string) {
			workflow := &Workflow{Spec: WorkflowSpec{Entry: template.Name, Templates: []Template{template}}}
			Expect(admission.ValidateCreate(ctx, workflow)).Error().To(MatchError(ContainSubstring(message)))
		},
		Entry("invalid name", Template{Name: "Invalid_Name", Type: TypeTask}, "DNS-1123 subdomain"),
		Entry("suspend without deadline", Template{Name: "wait", Type: TypeSuspend}, "deadline in template with type Suspend"),
		Entry("unrecognized type", Template{Name: "step", Type: "Unsupported"}, "unrecognized template type"),
		Entry("missing embedded chaos", Template{Name: "experiment", Type: TypePodChaos}, "the value of chaos PodChaos is required"),
		Entry("missing spec for the selected chaos", Template{Name: "experiment", Type: TypePodChaos, EmbedChaos: &EmbedChaos{}}, "missing chaos spec PodChaos"),
		Entry("container kill without a container name", Template{
			Name: "experiment", Type: TypePodChaos,
			EmbedChaos: &EmbedChaos{PodChaos: &PodChaosSpec{Action: ContainerKillAction}},
		}, "containerNames"),
	)

	It("rejects objects of the wrong kind for every admission operation", func() {
		pod := &corev1.Pod{}
		Expect(admission.ValidateCreate(ctx, pod)).Error().To(MatchError(ContainSubstring("expected type *Workflow")))
		Expect(admission.ValidateUpdate(ctx, pod, admission)).Error().To(MatchError(ContainSubstring("expected type *Workflow")))
		Expect(admission.ValidateUpdate(ctx, admission, pod)).Error().To(MatchError(ContainSubstring("expected type *Workflow")))
		Expect(admission.ValidateDelete(ctx, pod)).Error().To(MatchError(ContainSubstring("expected type *Workflow")))
	})
})
