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
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
)

func taskWorkflow() (*v1alpha1.Workflow, *v1alpha1.WorkflowNode) {
	workflow := &v1alpha1.Workflow{
		ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "sequence"},
		Spec: v1alpha1.WorkflowSpec{Templates: []v1alpha1.Template{
			{Name: "passed", Type: v1alpha1.TypeSuspend},
			{Name: "failed", Type: v1alpha1.TypeSuspend},
		}},
	}
	node := &v1alpha1.WorkflowNode{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: workflow.Namespace, Name: "check", UID: "task-node",
			Labels: map[string]string{v1alpha1.LabelWorkflow: workflow.Name},
		},
		Spec: v1alpha1.WorkflowNodeSpec{
			Type: v1alpha1.TypeTask, WorkflowName: workflow.Name,
			Task: &v1alpha1.Task{
				Container: &corev1.Container{Name: "verify", Image: "test-image"},
				Labels: map[string]string{
					"team": "ops", v1alpha1.LabelControlledBy: "wrong-node", v1alpha1.LabelWorkflow: "wrong-workflow",
				},
			},
			ConditionalBranches: []v1alpha1.ConditionalBranch{
				{Target: "passed", Expression: `exitCode == 0 && stdout == "ready"`},
				{Target: "failed", Expression: "exitCode != 0"},
			},
		},
	}
	return workflow, node
}

func workflowUnitClient(t *testing.T, objects ...client.Object) client.WithWatch {
	t.Helper()
	g := NewWithT(t)
	scheme := runtime.NewScheme()
	g.Expect(corev1.AddToScheme(scheme)).To(Succeed())
	g.Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).
		WithStatusSubresource(&v1alpha1.WorkflowNode{}, &v1alpha1.StatusCheck{}).Build()
}

func TestTaskReconcileCreatesOnePodInNodeNamespace(t *testing.T) {
	g := NewWithT(t)
	ctx := t.Context()
	workflow, node := taskWorkflow()
	foreign := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{
		Namespace: "other", Name: "running", Labels: map[string]string{v1alpha1.LabelControlledBy: node.Name},
	}, Status: corev1.PodStatus{Phase: corev1.PodRunning}}
	kube := workflowUnitClient(t, workflow, node, foreign)
	events := recorder.NewDebugRecorder()
	controller := NewTaskReconciler(kube, nil, events, logr.Discard())
	key := client.ObjectKeyFromObject(node)
	for range 2 {
		_, err := controller.Reconcile(ctx, reconcile.Request{NamespacedName: key})
		g.Expect(err).NotTo(HaveOccurred())
	}

	pods, err := controller.FetchPodControlledByThisWorkflowNode(ctx, *node)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(pods).To(HaveLen(1))
	pod := pods[0]
	g.Expect(pod.Namespace).To(Equal(node.Namespace))
	g.Expect(pod.Labels).To(Equal(map[string]string{
		"team": "ops", v1alpha1.LabelControlledBy: node.Name, v1alpha1.LabelWorkflow: workflow.Name,
	}))
	g.Expect(pod.OwnerReferences).To(HaveLen(1))
	owner := pod.OwnerReferences[0]
	g.Expect(owner.Name).To(Equal(node.Name))
	g.Expect(owner.UID).To(Equal(node.UID))
	g.Expect(owner.Controller).NotTo(BeNil())
	g.Expect(*owner.Controller).To(BeTrue())
	g.Expect(pod.Spec.RestartPolicy).To(Equal(corev1.RestartPolicyNever))
	g.Expect(kube.Get(ctx, key, node)).To(Succeed())
	g.Expect(node.Status.ConditionalBranchesStatus.Branches).To(Equal([]v1alpha1.ConditionalBranchStatus{
		{Target: "passed", EvaluationResult: corev1.ConditionUnknown},
		{Target: "failed", EvaluationResult: corev1.ConditionUnknown},
	}))
	g.Expect(events.Events[key]).To(ConsistOf(recorder.TaskPodSpawned{PodName: pod.Name}))
}

func TestTaskReconcileReusesChildAfterStatusWriteFails(t *testing.T) {
	g := NewWithT(t)
	ctx := t.Context()
	workflow, node := taskWorkflow()
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: node.Namespace, Name: "check-pod", Labels: map[string]string{v1alpha1.LabelControlledBy: node.Name},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodSucceeded,
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "verify", State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: 0}}},
			},
		},
	}
	var reads atomic.Int32
	logs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/namespaces/test/pods/check-pod/log" || r.URL.Query().Get("container") != "verify" {
			http.NotFound(w, r)
			return
		}
		reads.Add(1)
		_, _ = io.WriteString(w, "ready\n")
	}))
	t.Cleanup(logs.Close)
	kube := workflowUnitClient(t, workflow, node, pod)
	statusFailure := errors.New("could not save active children")
	failNextChildStatusWrite := true
	withFailure := interceptor.NewClient(kube, interceptor.Funcs{
		SubResourceUpdate: func(ctx context.Context, c client.Client, name string, obj client.Object, opts ...client.SubResourceUpdateOption) error {
			if current, ok := obj.(*v1alpha1.WorkflowNode); ok && name == "status" &&
				len(current.Status.ActiveChildren) > 0 && failNextChildStatusWrite {
				failNextChildStatusWrite = false
				return statusFailure
			}
			return c.SubResource(name).Update(ctx, obj, opts...)
		},
	})
	controller := NewTaskReconciler(withFailure, &rest.Config{Host: logs.URL}, recorder.NewDebugRecorder(), logr.Discard())
	key := client.ObjectKeyFromObject(node)
	req := reconcile.Request{NamespacedName: key}
	_, err := controller.Reconcile(ctx, req)
	g.Expect(err).To(MatchError(statusFailure))
	g.Expect(kube.Get(ctx, key, node)).To(Succeed())
	g.Expect(node.Status.ConditionalBranchesStatus.Branches).To(Equal([]v1alpha1.ConditionalBranchStatus{
		{Target: "passed", EvaluationResult: corev1.ConditionTrue},
		{Target: "failed", EvaluationResult: corev1.ConditionFalse},
	}))
	g.Expect(node.Status.ConditionalBranchesStatus.Context).To(ConsistOf(MatchJSON(`{"exitCode":0,"stdout":"ready"}`)))
	g.Expect(WorkflowNodeFinished(node.Status)).To(BeFalse())
	children := &v1alpha1.WorkflowNodeList{}
	g.Expect(kube.List(ctx, children, client.InNamespace(node.Namespace), client.MatchingLabels{v1alpha1.LabelControlledBy: node.Name})).To(Succeed())
	g.Expect(children.Items).To(HaveLen(1))
	child := &children.Items[0]
	g.Expect(child.Spec.TemplateName).To(Equal("passed"))
	g.Expect(node.Status.ActiveChildren).To(BeEmpty())

	_, err = controller.Reconcile(ctx, req)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(kube.Get(ctx, key, node)).To(Succeed())
	g.Expect(node.Status.ActiveChildren).To(ConsistOf(corev1.LocalObjectReference{Name: child.Name}))
	g.Expect(kube.List(ctx, children, client.InNamespace(node.Namespace), client.MatchingLabels{v1alpha1.LabelControlledBy: node.Name})).To(Succeed())
	g.Expect(children.Items).To(HaveLen(1))

	SetCondition(&child.Status, v1alpha1.WorkflowNodeCondition{Type: v1alpha1.ConditionAccomplished, Status: corev1.ConditionTrue})
	g.Expect(kube.Status().Update(ctx, child)).To(Succeed())
	_, err = controller.Reconcile(ctx, req)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(kube.Get(ctx, key, node)).To(Succeed())
	g.Expect(WorkflowNodeFinished(node.Status)).To(BeTrue())
	g.Expect(node.Status.ActiveChildren).To(BeEmpty())
	g.Expect(node.Status.FinishedChildren).To(ConsistOf(corev1.LocalObjectReference{Name: child.Name}))
	g.Expect(reads.Load()).To(Equal(int32(1)))
	g.Expect(kube.List(ctx, children, client.InNamespace(node.Namespace), client.MatchingLabels{v1alpha1.LabelControlledBy: node.Name})).To(Succeed())
	g.Expect(children.Items).To(HaveLen(1))
}

func TestTaskReconcileReportsCreationFailure(t *testing.T) {
	g := NewWithT(t)
	workflow, node := taskWorkflow()
	kube := workflowUnitClient(t, workflow, node)
	failure := errors.New("pod creation rejected")
	withFailure := interceptor.NewClient(kube, interceptor.Funcs{
		Create: func(context.Context, client.WithWatch, client.Object, ...client.CreateOption) error {
			return failure
		},
	})
	events := recorder.NewDebugRecorder()
	controller := NewTaskReconciler(withFailure, nil, events, logr.Discard())
	key := client.ObjectKeyFromObject(node)
	_, err := controller.Reconcile(t.Context(), reconcile.Request{NamespacedName: key})
	g.Expect(err).To(MatchError(failure))
	g.Expect(events.Events[key]).To(ConsistOf(recorder.TaskPodSpawnFailed{}))
}

func TestTaskReconcileKeepsChildrenFromPartialCreation(t *testing.T) {
	g := NewWithT(t)
	ctx := t.Context()
	workflow, node := taskWorkflow()
	node.Status.ConditionalBranchesStatus = &v1alpha1.ConditionalBranchesStatus{
		Branches: []v1alpha1.ConditionalBranchStatus{
			{Target: "passed", EvaluationResult: corev1.ConditionTrue},
			{Target: "failed", EvaluationResult: corev1.ConditionTrue},
		},
	}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: node.Namespace, Name: "completed", Labels: map[string]string{v1alpha1.LabelControlledBy: node.Name}},
		Status:     corev1.PodStatus{Phase: corev1.PodSucceeded},
	}
	kube := workflowUnitClient(t, workflow, node, pod)
	failure := errors.New("second child creation temporarily unavailable")
	failSecond := true
	var attempts []string
	wrapped := interceptor.NewClient(kube, interceptor.Funcs{
		Create: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
			child, ok := obj.(*v1alpha1.WorkflowNode)
			if ok {
				attempts = append(attempts, child.Spec.TemplateName)
				if failSecond && child.Spec.TemplateName == "failed" {
					return failure
				}
			}
			return c.Create(ctx, obj, opts...)
		},
	})
	r := NewTaskReconciler(wrapped, nil, recorder.NewDebugRecorder(), logr.Discard())
	req := reconcile.Request{NamespacedName: client.ObjectKeyFromObject(node)}
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(MatchError(failure))
	children := &v1alpha1.WorkflowNodeList{}
	opts := []client.ListOption{client.InNamespace(node.Namespace), client.MatchingLabels{v1alpha1.LabelControlledBy: node.Name}}
	g.Expect(kube.List(ctx, children, opts...)).To(Succeed())
	g.Expect(children.Items).To(HaveLen(1))
	first := children.Items[0]
	g.Expect(first.Spec.TemplateName).To(Equal("passed"))
	SetCondition(&first.Status, v1alpha1.WorkflowNodeCondition{Type: v1alpha1.ConditionAccomplished, Status: corev1.ConditionTrue})
	g.Expect(kube.Status().Update(ctx, &first)).To(Succeed())

	failSecond = false
	for range 2 {
		_, err = r.Reconcile(ctx, req)
		g.Expect(err).NotTo(HaveOccurred())
	}
	g.Expect(attempts).To(Equal([]string{"passed", "failed", "failed"}))
	g.Expect(kube.List(ctx, children, opts...)).To(Succeed())
	g.Expect(children.Items).To(HaveLen(2))
	g.Expect(kube.Get(ctx, req.NamespacedName, node)).To(Succeed())
	g.Expect(node.Status.FinishedChildren).To(Equal([]corev1.LocalObjectReference{{Name: first.Name}}))
	g.Expect(node.Status.ActiveChildren).To(HaveLen(1))
}
