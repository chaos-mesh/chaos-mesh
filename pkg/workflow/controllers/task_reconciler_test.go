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

package controllers

import (
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
)

func TestTaskReconcileCreatesOnePodInNodeNamespace(t *testing.T) {
	g := NewWithT(t)
	ctx := t.Context()
	workflow := &v1alpha1.Workflow{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "sequence"}}
	node := &v1alpha1.WorkflowNode{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: workflow.Namespace, Name: "check", UID: "task-node",
			Labels: map[string]string{v1alpha1.LabelWorkflow: workflow.Name},
		},
		Spec: v1alpha1.WorkflowNodeSpec{
			Type: v1alpha1.TypeTask, WorkflowName: workflow.Name,
			Task: &v1alpha1.Task{Container: &corev1.Container{Name: "verify", Image: "test-image"}},
		},
	}
	foreign := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "other", Name: "running",
			Labels: map[string]string{v1alpha1.LabelControlledBy: node.Name},
		},
		Status: corev1.PodStatus{Phase: corev1.PodRunning},
	}
	scheme := runtime.NewScheme()
	g.Expect(corev1.AddToScheme(scheme)).To(Succeed())
	g.Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
	kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(workflow, node, foreign).
		WithStatusSubresource(&v1alpha1.WorkflowNode{}).Build()
	controller := NewTaskReconciler(kube, nil, recorder.NewDebugRecorder(), logr.Discard())
	req := reconcile.Request{NamespacedName: client.ObjectKeyFromObject(node)}
	for range 2 {
		_, err := controller.Reconcile(ctx, req)
		g.Expect(err).NotTo(HaveOccurred())
	}

	var pods corev1.PodList
	g.Expect(kube.List(ctx, &pods, client.InNamespace(node.Namespace))).To(Succeed())
	g.Expect(pods.Items).To(HaveLen(1))
	g.Expect(pods.Items[0].OwnerReferences).To(HaveLen(1))
	g.Expect(pods.Items[0].OwnerReferences[0].UID).To(Equal(node.UID))
}
