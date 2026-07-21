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

package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

var _ = Describe("EphemeralTaskReconciler", func() {
	var ns string

	BeforeEach(func() {
		ctx := context.TODO()
		namespace := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "chaos-mesh-ephemeral-task-",
			},
		}
		Expect(kubeClient.Create(ctx, &namespace)).To(Succeed())
		ns = namespace.Name
	})

	AfterEach(func() {
		ctx := context.TODO()
		namespace := corev1.Namespace{}
		Expect(kubeClient.Get(ctx, types.NamespacedName{Name: ns}, &namespace)).To(Succeed())
		Expect(kubeClient.Delete(ctx, &namespace)).To(Succeed())
	})

	It("spawns the ephemeral task pod only once", func() {
		ctx := context.TODO()
		workflow, node := createEphemeralTaskWorkflowAndEntryNode(ctx, ns, "ephemeral-spawn-once", nil)

		Eventually(func(g Gomega) {
			updatedNode := getWorkflowNode(ctx, ns, node.Name)
			g.Expect(updatedNode.Annotations[ephemeralTaskPodCreatedAnnotation]).To(Equal("true"))

			pods := listPodsControlledBy(ctx, ns, node.Name)
			g.Expect(pods).To(HaveLen(1))
		}, 10*time.Second, 500*time.Millisecond).Should(Succeed())

		Consistently(func(g Gomega) {
			pods := listPodsControlledBy(ctx, ns, node.Name)
			g.Expect(pods).To(HaveLen(1))
		}, 3*time.Second, 300*time.Millisecond).Should(Succeed())

		Expect(workflow.Name).To(Equal("ephemeral-spawn-once"))
	})

	It("records a synthetic failure and does not respawn when the pod is missing", func() {
		ctx := context.TODO()
		_, node := createEphemeralTaskWorkflowAndEntryNode(ctx, ns, "ephemeral-synthetic-failure", map[string]string{
			ephemeralTaskPodCreatedAnnotation: "true",
		})

		Eventually(func(g Gomega) {
			updatedNode := getWorkflowNode(ctx, ns, node.Name)
			g.Expect(updatedNode.Annotations[ephemeralTaskResultCollectedAnnotation]).To(Equal("true"))
			g.Expect(updatedNode.Status.ConditionalBranchesStatus).NotTo(BeNil())
			g.Expect(updatedNode.Status.ConditionalBranchesStatus.Branches).To(HaveLen(2))
			g.Expect(updatedNode.Status.ConditionalBranchesStatus.Context).To(ContainElement(`{"exitCode":-1,"stdout":""}`))
			g.Expect(updatedNode.Status.ConditionalBranchesStatus.Branches[0].EvaluationResult).To(Equal(corev1.ConditionFalse))
			g.Expect(updatedNode.Status.ConditionalBranchesStatus.Branches[1].EvaluationResult).To(Equal(corev1.ConditionTrue))
			g.Expect(listPodsControlledBy(ctx, ns, node.Name)).To(BeEmpty())

			children := listWorkflowNodesControlledBy(ctx, ns, node.Name)
			g.Expect(children).To(HaveLen(1))
			g.Expect(children[0].Spec.TemplateName).To(Equal("failure"))
		}, 10*time.Second, 500*time.Millisecond).Should(Succeed())
	})

	It("reuses the persisted result and does not respawn after completion", func() {
		ctx := context.TODO()
		_, node := createEphemeralTaskWorkflowAndEntryNode(ctx, ns, "ephemeral-persisted-result", map[string]string{
			ephemeralTaskPodCreatedAnnotation: "true",
		})
		node.Status.ConditionalBranchesStatus = &v1alpha1.ConditionalBranchesStatus{
			Branches: []v1alpha1.ConditionalBranchStatus{
				{
					Target:           "success",
					EvaluationResult: corev1.ConditionTrue,
				},
				{
					Target:           "failure",
					EvaluationResult: corev1.ConditionFalse,
				},
			},
			Context: []string{`{"exitCode":0,"stdout":"ok"}`},
		}
		Expect(kubeClient.Status().Update(ctx, node)).To(Succeed())

		Eventually(func(g Gomega) {
			updatedNode := getWorkflowNode(ctx, ns, node.Name)
			g.Expect(updatedNode.Status.ConditionalBranchesStatus).NotTo(BeNil())
			g.Expect(updatedNode.Status.ConditionalBranchesStatus.Context).To(ContainElement(`{"exitCode":0,"stdout":"ok"}`))
			g.Expect(updatedNode.Status.ConditionalBranchesStatus.Branches).To(HaveLen(2))
			g.Expect(updatedNode.Status.ConditionalBranchesStatus.Branches[0].EvaluationResult).To(Equal(corev1.ConditionTrue))
			g.Expect(updatedNode.Status.ConditionalBranchesStatus.Branches[1].EvaluationResult).To(Equal(corev1.ConditionFalse))
			g.Expect(listPodsControlledBy(ctx, ns, node.Name)).To(BeEmpty())

			children := listWorkflowNodesControlledBy(ctx, ns, node.Name)
			g.Expect(children).To(HaveLen(1))
			g.Expect(children[0].Spec.TemplateName).To(Equal("success"))
		}, 10*time.Second, 500*time.Millisecond).Should(Succeed())
	})
})

func createEphemeralTaskWorkflowAndEntryNode(ctx context.Context, namespace, name string, annotations map[string]string) (*v1alpha1.Workflow, *v1alpha1.WorkflowNode) {
	nodeAnnotations := map[string]string{}
	for key, value := range annotations {
		nodeAnnotations[key] = value
	}

	now := metav1.Now()
	node := &v1alpha1.WorkflowNode{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "main-entry",
			Namespace:   namespace,
			Labels:      map[string]string{v1alpha1.LabelControlledBy: name, v1alpha1.LabelWorkflow: name},
			Annotations: nodeAnnotations,
		},
		Spec: v1alpha1.WorkflowNodeSpec{
			TemplateName: "main",
			WorkflowName: name,
			Type:         v1alpha1.TypeEphemeralTask,
			StartTime:    &now,
			Task: &v1alpha1.Task{
				Container: &corev1.Container{
					Name:    "main",
					Image:   "busybox:1.36",
					Command: []string{"sh", "-c", "echo ok"},
				},
			},
			ConditionalBranches: []v1alpha1.ConditionalBranch{
				{
					Target:     "success",
					Expression: `exitCode == 0 && stdout == "ok"`,
				},
				{
					Target:     "failure",
					Expression: `exitCode != 0`,
				},
			},
		},
	}
	Expect(kubeClient.Create(ctx, node)).To(Succeed())

	workflow := &v1alpha1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.WorkflowSpec{
			Entry: "main",
			Templates: []v1alpha1.Template{
				{
					Name: "main",
					Type: v1alpha1.TypeEphemeralTask,
					Task: &v1alpha1.Task{
						Container: &corev1.Container{
							Name:    "main",
							Image:   "busybox:1.36",
							Command: []string{"sh", "-c", "echo ok"},
						},
					},
					ConditionalBranches: []v1alpha1.ConditionalBranch{
						{
							Target:     "success",
							Expression: `exitCode == 0 && stdout == "ok"`,
						},
						{
							Target:     "failure",
							Expression: `exitCode != 0`,
						},
					},
				},
				{
					Name: "success",
					Type: v1alpha1.TypeTask,
					Task: &v1alpha1.Task{
						Container: &corev1.Container{
							Name:    "success",
							Image:   "busybox:1.36",
							Command: []string{"sh", "-c", "echo success"},
						},
					},
				},
				{
					Name: "failure",
					Type: v1alpha1.TypeTask,
					Task: &v1alpha1.Task{
						Container: &corev1.Container{
							Name:    "failure",
							Image:   "busybox:1.36",
							Command: []string{"sh", "-c", "echo failure"},
						},
					},
				},
			},
		},
	}
	Expect(kubeClient.Create(ctx, workflow)).To(Succeed())

	return workflow, node
}

func getWorkflowNode(ctx context.Context, namespace, name string) v1alpha1.WorkflowNode {
	node := v1alpha1.WorkflowNode{}
	Expect(kubeClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &node)).To(Succeed())
	return node
}

func listPodsControlledBy(ctx context.Context, namespace, controller string) []corev1.Pod {
	pods := corev1.PodList{}
	Expect(kubeClient.List(ctx, &pods,
		client.InNamespace(namespace),
		client.MatchingLabels{v1alpha1.LabelControlledBy: controller},
	)).To(Succeed())
	return pods.Items
}

func listWorkflowNodesControlledBy(ctx context.Context, namespace, controller string) []v1alpha1.WorkflowNode {
	nodes := v1alpha1.WorkflowNodeList{}
	Expect(kubeClient.List(ctx, &nodes,
		client.InNamespace(namespace),
		client.MatchingLabels{v1alpha1.LabelControlledBy: controller},
	)).To(Succeed())
	return nodes.Items
}
