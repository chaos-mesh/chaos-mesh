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
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
)

func statusCheckWorkflow() (*v1alpha1.Workflow, *v1alpha1.WorkflowNode) {
	workflow := &v1alpha1.Workflow{
		ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "workflow"},
	}
	node := &v1alpha1.WorkflowNode{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: workflow.Namespace, Name: "probe", UID: "probe-node",
			Labels: map[string]string{v1alpha1.LabelWorkflow: workflow.Name},
		},
		Spec: v1alpha1.WorkflowNodeSpec{
			Type: v1alpha1.TypeStatusCheck, WorkflowName: workflow.Name, AbortWithStatusCheck: true,
			StatusCheck: &v1alpha1.StatusCheckSpec{Type: v1alpha1.TypeHTTP, Mode: v1alpha1.StatusCheckSynchronous},
		},
	}
	return workflow, node
}

func TestStatusCheckReconcileAbortsWorkflowAndCleansUpLocalCheck(t *testing.T) {
	g := NewWithT(t)
	ctx := t.Context()
	workflow, node := statusCheckWorkflow()
	foreign := &v1alpha1.StatusCheck{ObjectMeta: metav1.ObjectMeta{
		Namespace: "other", Name: "foreign-probe",
		Labels: map[string]string{v1alpha1.LabelControlledBy: node.Name},
	}}
	kube := workflowUnitClient(t, workflow, node, foreign)
	events := recorder.NewDebugRecorder()
	controller := NewStatusCheckReconciler(kube, events, logr.Discard())
	key := client.ObjectKeyFromObject(node)
	req := reconcile.Request{NamespacedName: key}

	_, err := controller.Reconcile(ctx, req)
	g.Expect(err).NotTo(HaveOccurred())
	var local v1alpha1.StatusCheckList
	g.Expect(kube.List(ctx, &local, client.InNamespace(node.Namespace))).To(Succeed())
	g.Expect(local.Items).To(HaveLen(1))
	check := &local.Items[0]
	g.Expect(check.Spec).To(Equal(*node.Spec.StatusCheck))
	g.Expect(check.Labels).To(HaveKeyWithValue(v1alpha1.LabelWorkflow, workflow.Name))
	g.Expect(check.Labels).To(HaveKeyWithValue(v1alpha1.LabelControlledBy, node.Name))
	owner := metav1.GetControllerOf(check)
	g.Expect(owner).NotTo(BeNil())
	g.Expect(owner.UID).To(Equal(node.UID))
	g.Expect(events.Events[key]).To(ConsistOf(recorder.StatusCheckCreated{Name: check.Name}))

	g.Expect(kube.Get(ctx, key, node)).To(Succeed())
	g.Expect(ConditionEqualsTo(node.Status, v1alpha1.ConditionAccomplished, corev1.ConditionFalse)).To(BeTrue())
	g.Expect(ConditionEqualsTo(node.Status, v1alpha1.ConditionAborted, corev1.ConditionFalse)).To(BeTrue())

	check.Status.Conditions = []v1alpha1.StatusCheckCondition{
		{Type: v1alpha1.StatusCheckConditionCompleted, Status: corev1.ConditionTrue},
		{Type: v1alpha1.StatusCheckConditionSuccessThresholdExceed, Status: corev1.ConditionFalse},
	}
	g.Expect(kube.Status().Update(ctx, check)).To(Succeed())
	_, err = controller.Reconcile(ctx, req)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(kube.Get(ctx, key, node)).To(Succeed())
	g.Expect(GetCondition(node.Status, v1alpha1.ConditionAccomplished)).To(Equal(&v1alpha1.WorkflowNodeCondition{
		Type: v1alpha1.ConditionAccomplished, Status: corev1.ConditionTrue, Reason: v1alpha1.StatusCheckCompleted,
	}))
	g.Expect(GetCondition(node.Status, v1alpha1.ConditionAborted)).To(Equal(&v1alpha1.WorkflowNodeCondition{
		Type: v1alpha1.ConditionAborted, Status: corev1.ConditionTrue, Reason: v1alpha1.StatusCheckNotExceedSuccessThreshold,
	}))

	abort := NewAbortNodeReconciler(kube, events, logr.Discard())
	for range 2 {
		_, err = abort.Reconcile(ctx, req)
		g.Expect(err).NotTo(HaveOccurred())
	}
	g.Expect(kube.Get(ctx, client.ObjectKeyFromObject(workflow), workflow)).To(Succeed())
	g.Expect(workflow.Annotations).To(HaveKeyWithValue(v1alpha1.WorkflowAnnotationAbort, "true"))

	_, err = controller.Reconcile(ctx, req)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(kube.Get(ctx, client.ObjectKeyFromObject(check), check)).To(Succeed())
	g.Expect(check.DeletionTimestamp).NotTo(BeNil())
	g.Expect(kube.Get(ctx, client.ObjectKeyFromObject(foreign), foreign)).To(Succeed())
	g.Expect(foreign.DeletionTimestamp).To(BeNil())
	g.Expect(events.Events[key]).To(ConsistOf(
		recorder.StatusCheckCreated{Name: check.Name}, recorder.StatusCheckDeleted{Name: check.Name},
	))
}

func TestStatusCheckReconcileReturnsCreationError(t *testing.T) {
	rejected := errors.New("status check creation rejected")
	for _, scenario := range []struct {
		name    string
		hasSpec bool
		wantErr string
	}{
		{name: "missing specification", wantErr: "does not contains spec"},
		{name: "API rejects creation", hasSpec: true, wantErr: rejected.Error()},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			g := NewWithT(t)
			workflow, node := statusCheckWorkflow()
			if !scenario.hasSpec {
				node.Spec.StatusCheck = nil
			}
			kube := workflowUnitClient(t, workflow, node)
			failing := interceptor.NewClient(kube, interceptor.Funcs{
				Create: func(context.Context, client.WithWatch, client.Object, ...client.CreateOption) error {
					return rejected
				},
			})
			events := recorder.NewDebugRecorder()
			controller := NewStatusCheckReconciler(failing, events, logr.Discard())
			key := client.ObjectKeyFromObject(node)
			_, err := controller.Reconcile(t.Context(), reconcile.Request{NamespacedName: key})
			g.Expect(err).To(MatchError(ContainSubstring(scenario.wantErr)))
			g.Expect(events.Events[key]).To(ConsistOf(recorder.StatusCheckCreatedFailed{}))
		})
	}
}

func TestStatusCheckReconcileRemovesDuplicatesWithoutAborting(t *testing.T) {
	g := NewWithT(t)
	ctx := t.Context()
	workflow, node := statusCheckWorkflow()
	node.Spec.AbortWithStatusCheck = false
	check := &v1alpha1.StatusCheck{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: node.Namespace, Name: "first", Labels: map[string]string{v1alpha1.LabelControlledBy: node.Name},
		},
		Status: v1alpha1.StatusCheckStatus{Conditions: []v1alpha1.StatusCheckCondition{
			{Type: v1alpha1.StatusCheckConditionCompleted, Status: corev1.ConditionTrue},
			{Type: v1alpha1.StatusCheckConditionSuccessThresholdExceed, Status: corev1.ConditionFalse},
		}},
	}
	duplicate := check.DeepCopy()
	duplicate.Name = "second"
	kube := workflowUnitClient(t, workflow, node, check, duplicate)
	controller := NewStatusCheckReconciler(kube, recorder.NewDebugRecorder(), logr.Discard())
	key := client.ObjectKeyFromObject(node)
	_, err := controller.Reconcile(ctx, reconcile.Request{NamespacedName: key})
	g.Expect(err).NotTo(HaveOccurred())

	var checks v1alpha1.StatusCheckList
	g.Expect(kube.List(ctx, &checks, client.InNamespace(node.Namespace))).To(Succeed())
	g.Expect(checks.Items).To(HaveLen(1))
	g.Expect(kube.Get(ctx, key, node)).To(Succeed())
	g.Expect(ConditionEqualsTo(node.Status, v1alpha1.ConditionAccomplished, corev1.ConditionTrue)).To(BeTrue())
	g.Expect(ConditionEqualsTo(node.Status, v1alpha1.ConditionAborted, corev1.ConditionFalse)).To(BeTrue())
}

func TestAbortNodeReconcilePropagatesOnlyToActiveChildren(t *testing.T) {
	g := NewWithT(t)
	ctx := t.Context()
	parent := &v1alpha1.WorkflowNode{
		ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "parent"},
		Spec:       v1alpha1.WorkflowNodeSpec{Type: v1alpha1.TypeParallel},
	}
	SetCondition(&parent.Status, v1alpha1.WorkflowNodeCondition{Type: v1alpha1.ConditionAborted, Status: corev1.ConditionTrue})
	active := &v1alpha1.WorkflowNode{ObjectMeta: metav1.ObjectMeta{
		Namespace: parent.Namespace, Name: "active", Labels: map[string]string{v1alpha1.LabelControlledBy: parent.Name},
	}}
	finished := active.DeepCopy()
	finished.Name = "finished"
	SetCondition(&finished.Status, v1alpha1.WorkflowNodeCondition{Type: v1alpha1.ConditionAccomplished, Status: corev1.ConditionTrue})
	foreign := active.DeepCopy()
	foreign.Namespace = "other"
	kube := workflowUnitClient(t, parent, active, finished, foreign)
	events := recorder.NewDebugRecorder()
	controller := NewAbortNodeReconciler(kube, events, logr.Discard())
	req := reconcile.Request{NamespacedName: client.ObjectKeyFromObject(parent)}

	for range 2 {
		_, err := controller.Reconcile(ctx, req)
		g.Expect(err).NotTo(HaveOccurred())
	}
	activeKey := client.ObjectKeyFromObject(active)
	g.Expect(kube.Get(ctx, activeKey, active)).To(Succeed())
	g.Expect(GetCondition(active.Status, v1alpha1.ConditionAborted)).To(Equal(&v1alpha1.WorkflowNodeCondition{
		Type: v1alpha1.ConditionAborted, Status: corev1.ConditionTrue, Reason: v1alpha1.ParentNodeAborted,
	}))
	for _, child := range []*v1alpha1.WorkflowNode{finished, foreign} {
		g.Expect(kube.Get(ctx, client.ObjectKeyFromObject(child), child)).To(Succeed())
		g.Expect(GetCondition(child.Status, v1alpha1.ConditionAborted)).To(BeNil())
	}
	g.Expect(events.Events[activeKey]).To(ConsistOf(recorder.ParentNodeAborted{ParentNodeName: parent.Name}))

	req.Name = "gone"
	_, err := controller.Reconcile(ctx, req)
	g.Expect(err).NotTo(HaveOccurred())
}
