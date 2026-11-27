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
	"fmt"
	"testing"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
)

func TestParallelWorkflowPreservesChildrenAcrossPartialCreation(t *testing.T) {
	g := NewWithT(t)
	ctx := t.Context()
	workflow := &v1alpha1.Workflow{
		ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "parallel-workflow"},
		Spec: v1alpha1.WorkflowSpec{Templates: []v1alpha1.Template{
			{Name: "verify", Type: v1alpha1.TypeSuspend},
			{Name: "cleanup", Type: v1alpha1.TypeSuspend},
		}},
	}
	node := &v1alpha1.WorkflowNode{
		ObjectMeta: metav1.ObjectMeta{Namespace: workflow.Namespace, Name: "parallel", UID: "parent-uid"},
		Spec: v1alpha1.WorkflowNodeSpec{
			Type: v1alpha1.TypeParallel, WorkflowName: workflow.Name,
			Children: []string{"verify", "verify", "cleanup"},
		},
	}
	kube := workflowUnitClient(t, workflow, node)
	failure := errors.New("second creation interrupted")
	var attempts []string
	wrapped := interceptor.NewClient(kube, interceptor.Funcs{
		Create: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
			child := obj.(*v1alpha1.WorkflowNode)
			attempts = append(attempts, child.Spec.TemplateName)
			if len(attempts) == 2 {
				return failure
			}
			return c.Create(ctx, obj, opts...)
		},
	})
	r := NewParallelNodeReconciler(wrapped, recorder.NewDebugRecorder(), logr.Discard())
	req := reconcile.Request{NamespacedName: client.ObjectKeyFromObject(node)}
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(MatchError(failure))

	children := &v1alpha1.WorkflowNodeList{}
	opts := []client.ListOption{client.InNamespace(node.Namespace), client.MatchingLabels{v1alpha1.LabelControlledBy: node.Name}}
	g.Expect(kube.List(ctx, children, opts...)).To(Succeed())
	g.Expect(children.Items).To(HaveLen(1))
	first := children.Items[0]
	SetCondition(&first.Status, v1alpha1.WorkflowNodeCondition{Type: v1alpha1.ConditionAccomplished, Status: corev1.ConditionTrue})
	g.Expect(kube.Status().Update(ctx, &first)).To(Succeed())

	for range 2 {
		_, err = r.Reconcile(ctx, req)
		g.Expect(err).NotTo(HaveOccurred())
	}
	g.Expect(attempts).To(Equal([]string{"verify", "verify", "verify", "cleanup"}))
	g.Expect(kube.List(ctx, children, opts...)).To(Succeed())
	var templates []string
	for _, child := range children.Items {
		g.Expect(child.DeletionTimestamp).To(BeNil(), fmt.Sprintf("child %s", child.Name))
		templates = append(templates, child.Spec.TemplateName)
	}
	g.Expect(templates).To(ConsistOf("verify", "verify", "cleanup"))
	g.Expect(kube.Get(ctx, req.NamespacedName, node)).To(Succeed())
	g.Expect(node.Status.FinishedChildren).To(ConsistOf(corev1.LocalObjectReference{Name: first.Name}))
	g.Expect(node.Status.ActiveChildren).To(HaveLen(2))
}

func TestSerialWorkflowRetriesCleanupBeforeStartingReplacement(t *testing.T) {
	g := NewWithT(t)
	ctx := t.Context()
	workflow := &v1alpha1.Workflow{
		ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "serial-workflow"},
		Spec: v1alpha1.WorkflowSpec{Templates: []v1alpha1.Template{
			{Name: "first", Type: v1alpha1.TypeSuspend},
			{Name: "second", Type: v1alpha1.TypeSuspend},
		}},
	}
	node := &v1alpha1.WorkflowNode{
		ObjectMeta: metav1.ObjectMeta{Namespace: workflow.Namespace, Name: "serial", UID: "parent-uid"},
		Spec: v1alpha1.WorkflowNodeSpec{
			Type: v1alpha1.TypeSerial, WorkflowName: workflow.Name, Children: []string{"first", "second"},
		},
	}
	finished := func(name, template, uid string, second int64) *v1alpha1.WorkflowNode {
		child := &v1alpha1.WorkflowNode{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: node.Namespace, Name: name, UID: types.UID(uid), CreationTimestamp: metav1.NewTime(time.Unix(second, 0)),
				Labels:          map[string]string{v1alpha1.LabelControlledBy: node.Name},
				OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(node, v1alpha1.SchemeGroupVersion.WithKind("WorkflowNode"))},
			},
			Spec: v1alpha1.WorkflowNodeSpec{TemplateName: template},
		}
		SetCondition(&child.Status, v1alpha1.WorkflowNodeCondition{Type: v1alpha1.ConditionAccomplished, Status: corev1.ConditionTrue})
		return child
	}
	first := finished("first-old", "", "first-uid", 1)
	stale := finished("second-before", "removed-template", "stale-uid", 2)
	stale.Finalizers = []string{"fixture.invalid/cleanup"}
	tail := finished("old-tail", "second", "tail-uid", 3)
	unrelated := finished("other-owner", "unrelated", "other-uid", 4)
	unrelated.OwnerReferences[0].UID = "old-parent-uid"
	kube := workflowUnitClient(t, workflow, node, first, stale, tail, unrelated)
	failure := errors.New("temporary cleanup failure")
	var deletes []string
	creates := 0
	wrapped := interceptor.NewClient(kube, interceptor.Funcs{
		Delete: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
			deletes = append(deletes, obj.GetName())
			g.Expect(obj.GetName()).To(BeElementOf(tail.Name, stale.Name))
			options := (&client.DeleteOptions{}).ApplyOptions(opts)
			g.Expect(options.Preconditions).NotTo(BeNil())
			g.Expect(options.Preconditions.UID).To(HaveValue(Equal(obj.GetUID())))
			g.Expect(options.Preconditions.ResourceVersion).To(HaveValue(Equal(obj.GetResourceVersion())))
			if len(deletes) == 1 {
				return failure
			}
			return c.Delete(ctx, obj, opts...)
		},
		Create: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
			creates++
			g.Expect(obj.(*v1alpha1.WorkflowNode).Spec.TemplateName).To(Equal("second"))
			return c.Create(ctx, obj, opts...)
		},
	})
	r := NewSerialNodeReconciler(wrapped, recorder.NewDebugRecorder(), logr.Discard())
	req := reconcile.Request{NamespacedName: client.ObjectKeyFromObject(node)}
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(MatchError(failure))
	for range 2 {
		result, err := r.Reconcile(ctx, req)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(result.RequeueAfter).To(BeNumerically(">", 0))
		g.Expect(creates).To(BeZero())
		g.Expect(kube.Get(ctx, req.NamespacedName, node)).To(Succeed())
		g.Expect(WorkflowNodeFinished(node.Status)).To(BeFalse())
	}
	g.Expect(deletes).To(Equal([]string{tail.Name, tail.Name, stale.Name}))
	g.Expect(kube.Get(ctx, client.ObjectKeyFromObject(stale), stale)).To(Succeed())
	g.Expect(stale.DeletionTimestamp).NotTo(BeNil())
	stale.Finalizers = nil
	g.Expect(kube.Update(ctx, stale)).To(Succeed())

	for range 2 {
		_, err := r.Reconcile(ctx, req)
		g.Expect(err).NotTo(HaveOccurred())
	}
	g.Expect(creates).To(Equal(1))
	g.Expect(kube.Get(ctx, client.ObjectKeyFromObject(first), first)).To(Succeed())
	g.Expect(kube.Get(ctx, client.ObjectKeyFromObject(unrelated), unrelated)).To(Succeed())
	g.Expect(kube.Get(ctx, req.NamespacedName, node)).To(Succeed())
	g.Expect(node.Status.FinishedChildren).To(Equal([]corev1.LocalObjectReference{{Name: first.Name}}))
	g.Expect(node.Status.ActiveChildren).To(HaveLen(1))
}

func TestSerialWorkflowRecognizesCompletedChildrenCreatedInTheSameSecond(t *testing.T) {
	g := NewWithT(t)
	ctx := t.Context()
	node := &v1alpha1.WorkflowNode{
		ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "serial"},
		Spec:       v1alpha1.WorkflowNodeSpec{Type: v1alpha1.TypeSerial, Children: []string{"verify", "prepare", "verify"}},
	}
	makeChild := func(name, template string) *v1alpha1.WorkflowNode {
		child := &v1alpha1.WorkflowNode{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: node.Namespace, Name: name, CreationTimestamp: metav1.NewTime(time.Unix(10, 0)),
				Labels: map[string]string{v1alpha1.LabelControlledBy: node.Name},
			},
			Spec: v1alpha1.WorkflowNodeSpec{TemplateName: template},
		}
		SetCondition(&child.Status, v1alpha1.WorkflowNodeCondition{Type: v1alpha1.ConditionAccomplished, Status: corev1.ConditionTrue})
		return child
	}
	prepare := makeChild("a-prepare", "prepare")
	first := makeChild("b-verify", "verify")
	last := makeChild("c-verify", "verify")
	kube := workflowUnitClient(t, node, prepare, first, last)
	r := NewSerialNodeReconciler(kube, recorder.NewDebugRecorder(), logr.Discard())
	key := client.ObjectKeyFromObject(node)
	_, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: key})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(kube.Get(ctx, key, node)).To(Succeed())
	g.Expect(WorkflowNodeFinished(node.Status)).To(BeTrue())
	g.Expect(node.Status.FinishedChildren).To(Equal([]corev1.LocalObjectReference{
		{Name: first.Name}, {Name: prepare.Name}, {Name: last.Name},
	}))
}

func TestSerialWorkflowCleansStaleChildrenWhenChildrenCleared(t *testing.T) {
	g := NewWithT(t)
	ctx := t.Context()
	node := &v1alpha1.WorkflowNode{
		ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "serial", UID: "parent-uid"},
		Spec:       v1alpha1.WorkflowNodeSpec{Type: v1alpha1.TypeSerial},
	}
	stale := &v1alpha1.WorkflowNode{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: node.Namespace, Name: "first-old", UID: "stale-uid",
			Labels:          map[string]string{v1alpha1.LabelControlledBy: node.Name},
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(node, v1alpha1.SchemeGroupVersion.WithKind("WorkflowNode"))},
			Finalizers:      []string{"fixture.invalid/cleanup"},
		},
		Spec: v1alpha1.WorkflowNodeSpec{TemplateName: "first"},
	}
	SetCondition(&stale.Status, v1alpha1.WorkflowNodeCondition{Type: v1alpha1.ConditionAccomplished, Status: corev1.ConditionTrue})
	kube := workflowUnitClient(t, node, stale)
	r := NewSerialNodeReconciler(kube, recorder.NewDebugRecorder(), logr.Discard())
	key := client.ObjectKeyFromObject(node)
	req := reconcile.Request{NamespacedName: key}

	for range 2 {
		result, err := r.Reconcile(ctx, req)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(result.RequeueAfter).To(BeNumerically(">", 0))
		g.Expect(kube.Get(ctx, key, node)).To(Succeed())
		g.Expect(WorkflowNodeFinished(node.Status)).To(BeFalse())
	}
	g.Expect(kube.Get(ctx, client.ObjectKeyFromObject(stale), stale)).To(Succeed())
	g.Expect(stale.DeletionTimestamp).NotTo(BeNil())
	stale.Finalizers = nil
	g.Expect(kube.Update(ctx, stale)).To(Succeed())

	result, err := r.Reconcile(ctx, req)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(result.RequeueAfter).To(BeZero())
	g.Expect(kube.Get(ctx, key, node)).To(Succeed())
	g.Expect(WorkflowNodeFinished(node.Status)).To(BeTrue())
	g.Expect(node.Status.FinishedChildren).To(BeEmpty())
}

func TestWorkflowEntryReconcilesDuplicatesAndCompletion(t *testing.T) {
	g := NewWithT(t)
	ctx := t.Context()
	workflow := &v1alpha1.Workflow{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "flow"},
		Spec:       v1alpha1.WorkflowSpec{Entry: "entry", Templates: []v1alpha1.Template{{Name: "entry", Type: v1alpha1.TypeSerial}}},
	}
	kube := newCreateTrackingClient(t, workflow)
	blocked := true
	reconciler := NewWorkflowEntryReconciler(interceptor.NewClient(kube.Client.(client.WithWatch), interceptor.Funcs{
		Create: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
			if blocked {
				return errors.New("entry creation temporarily unavailable")
			}
			return c.Create(ctx, obj, opts...)
		},
	}), recorder.NewDebugRecorder(), logr.Discard())
	req := reconcile.Request{NamespacedName: client.ObjectKeyFromObject(workflow)}
	run := func() {
		t.Helper()
		g.Expect(reconciler.Reconcile(ctx, req)).To(Equal(reconcile.Result{}))
		g.Expect(kube.Get(ctx, req.NamespacedName, workflow)).To(Succeed())
	}
	run()
	g.Expect(GetWorkflowCondition(workflow.Status, v1alpha1.WorkflowConditionScheduled).Status).To(Equal(corev1.ConditionFalse))
	g.Expect(workflow.Status.EntryNode).To(BeNil())
	g.Expect(workflow.Status.StartTime).NotTo(BeNil())
	started := workflow.Status.StartTime.DeepCopy()

	blocked = false
	run()
	g.Expect(workflow.Status.EntryNode).NotTo(BeNil())
	first := *workflow.Status.EntryNode
	g.Expect(GetWorkflowCondition(workflow.Status, v1alpha1.WorkflowConditionScheduled).Status).To(Equal(corev1.ConditionTrue))
	g.Expect(GetWorkflowCondition(workflow.Status, v1alpha1.WorkflowConditionAccomplished).Status).To(Equal(corev1.ConditionFalse))
	g.Expect(workflow.Status.StartTime).To(Equal(started))

	entry := &v1alpha1.WorkflowNode{}
	g.Expect(kube.Get(ctx, client.ObjectKey{Namespace: workflow.Namespace, Name: first}, entry)).To(Succeed())
	duplicate := &v1alpha1.WorkflowNode{
		ObjectMeta: metav1.ObjectMeta{Namespace: workflow.Namespace, Name: "redundant", Labels: entry.Labels, CreationTimestamp: metav1.NewTime(time.Now().Add(time.Hour))},
		Spec:       entry.Spec,
	}
	g.Expect(kube.Create(ctx, duplicate)).To(Succeed())
	foreign := duplicate.DeepCopy()
	foreign.Namespace, foreign.ResourceVersion = "other", ""
	g.Expect(kube.Create(ctx, foreign)).To(Succeed())

	SetCondition(&entry.Status, v1alpha1.WorkflowNodeCondition{Type: v1alpha1.ConditionAccomplished, Status: corev1.ConditionTrue})
	g.Expect(kube.Status().Update(ctx, entry)).To(Succeed())
	run()
	g.Expect(*workflow.Status.EntryNode).To(Equal(first))
	g.Expect(GetWorkflowCondition(workflow.Status, v1alpha1.WorkflowConditionAccomplished).Status).To(Equal(corev1.ConditionTrue))
	g.Expect(workflow.Status.EndTime).NotTo(BeNil())
	finished := workflow.Status.EndTime.DeepCopy()
	g.Expect(kube.Get(ctx, client.ObjectKeyFromObject(duplicate), &v1alpha1.WorkflowNode{})).To(Satisfy(apierrors.IsNotFound))
	g.Expect(kube.Get(ctx, client.ObjectKeyFromObject(foreign), &v1alpha1.WorkflowNode{})).To(Succeed())
	run()
	g.Expect(workflow.Status.EndTime).To(Equal(finished))
}

func TestWorkflowScheduleRetainsExperimentAndRetriesCreation(t *testing.T) {
	for _, tc := range []struct {
		kind       v1alpha1.ScheduleTemplateType
		experiment v1alpha1.EmbedChaos
	}{
		{v1alpha1.ScheduleTypeAzureChaos, v1alpha1.EmbedChaos{
			AzureChaos: &v1alpha1.AzureChaosSpec{Action: v1alpha1.AzureVmStop},
		}},
		{v1alpha1.ScheduleTypeBlockChaos, v1alpha1.EmbedChaos{
			BlockChaos: &v1alpha1.BlockChaosSpec{Action: v1alpha1.BlockDelay, VolumeName: "fixture-volume"},
		}},
	} {
		t.Run(string(tc.kind), func(t *testing.T) {
			g := NewWithT(t)
			ctx := t.Context()
			workflow := &v1alpha1.Workflow{
				ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "scheduled-chaos"},
				Spec: v1alpha1.WorkflowSpec{
					Entry: "schedule",
					Templates: []v1alpha1.Template{{
						Name: "schedule", Type: v1alpha1.TypeSchedule,
						Schedule: &v1alpha1.ChaosOnlyScheduleSpec{
							Type: tc.kind, Schedule: "@every 10m", ConcurrencyPolicy: v1alpha1.ForbidConcurrent,
							HistoryLimit: 3, StartingDeadlineSeconds: new(int64(120)), EmbedChaos: tc.experiment,
						},
					}},
				},
			}
			kube := newCreateTrackingClient(t, workflow)
			log := logr.Discard()
			events := recorder.NewDebugRecorder()
			entry := NewWorkflowEntryReconciler(kube, events, log)
			key := client.ObjectKeyFromObject(workflow)
			_, err := entry.Reconcile(ctx, reconcile.Request{NamespacedName: key})
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(kube.Get(ctx, key, workflow)).To(Succeed())
			g.Expect(workflow.Status.EntryNode).NotTo(BeNil())
			nodeKey := client.ObjectKey{Namespace: workflow.Namespace, Name: *workflow.Status.EntryNode}

			failure := errors.New("schedule admission temporarily unavailable")
			attempts := 0
			wrapped := interceptor.NewClient(kube.Client.(client.WithWatch), interceptor.Funcs{
				Create: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
					attempts++
					if attempts == 1 {
						return failure
					}
					return c.Create(ctx, obj, opts...)
				},
			})
			r := NewChaosNodeReconciler(wrapped, events, log)
			req := reconcile.Request{NamespacedName: nodeKey}
			_, err = r.Reconcile(ctx, req)
			g.Expect(err).To(MatchError(failure))
			var schedules v1alpha1.ScheduleList
			g.Expect(kube.List(ctx, &schedules, client.InNamespace(workflow.Namespace))).To(Succeed())
			g.Expect(schedules.Items).To(BeEmpty())

			for range 2 {
				_, err = r.Reconcile(ctx, req)
				g.Expect(err).NotTo(HaveOccurred())
			}
			g.Expect(attempts).To(Equal(2))
			g.Expect(kube.List(ctx, &schedules, client.InNamespace(workflow.Namespace))).To(Succeed())
			g.Expect(schedules.Items).To(HaveLen(1))
			spec := schedules.Items[0].Spec
			original := workflow.Spec.Templates[0].Schedule
			g.Expect(spec.Type).To(Equal(original.Type))
			g.Expect(spec.Schedule).To(Equal(original.Schedule))
			g.Expect(spec.ConcurrencyPolicy).To(Equal(original.ConcurrencyPolicy))
			g.Expect(spec.HistoryLimit).To(Equal(original.HistoryLimit))
			g.Expect(spec.StartingDeadlineSeconds).To(Equal(original.StartingDeadlineSeconds))
			spawned, err := spec.ScheduleItem.SpawnNewObject(tc.kind)
			g.Expect(err).NotTo(HaveOccurred())
			expected, err := tc.experiment.SpawnNewObject(v1alpha1.TemplateType(tc.kind))
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(spawned).To(Equal(expected))
		})
	}
}

func TestChaosNodeChildReconciliation(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typ      v1alpha1.TemplateType
		newChild func() client.Object
	}{
		{"schedule", v1alpha1.TypeSchedule, func() client.Object { return &v1alpha1.Schedule{} }},
		{"pod chaos", v1alpha1.TypePodChaos, func() client.Object { return &v1alpha1.PodChaos{} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			ctx := t.Context()
			node := &v1alpha1.WorkflowNode{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "node"},
				Spec:       v1alpha1.WorkflowNodeSpec{Type: tc.typ},
			}
			child := func(namespace, name string, created time.Time) client.Object {
				obj := tc.newChild()
				obj.SetNamespace(namespace)
				obj.SetName(name)
				obj.SetLabels(map[string]string{v1alpha1.LabelControlledBy: node.Name})
				obj.SetCreationTimestamp(metav1.NewTime(created))
				return obj
			}
			then := time.Now().Add(-time.Hour)
			first := child(node.Namespace, "first", then)
			duplicate := child(node.Namespace, "duplicate", then.Add(time.Minute))
			foreign := child("other", "unrelated", then)
			kube := newCreateTrackingClient(t, node, first, duplicate, foreign)

			statusWrites := 0
			blockDelete := true
			c := interceptor.NewClient(kube.Client.(client.WithWatch), interceptor.Funcs{
				Delete: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
					if blockDelete && obj.GetName() == duplicate.GetName() {
						return errors.New("temporary deletion failure")
					}
					return c.Delete(ctx, obj, opts...)
				},
				SubResourceUpdate: func(ctx context.Context, c client.Client, sub string, obj client.Object, opts ...client.SubResourceUpdateOption) error {
					statusWrites++
					if statusWrites == 1 {
						return apierrors.NewConflict(schema.GroupResource{Group: "chaos-mesh.org", Resource: "workflownodes"}, obj.GetName(), errors.New("concurrent status change"))
					}
					return c.SubResource(sub).Update(ctx, obj, opts...)
				},
			})
			r := NewChaosNodeReconciler(c, recorder.NewDebugRecorder(), logr.Discard())
			req := reconcile.Request{NamespacedName: client.ObjectKeyFromObject(node)}
			run := func() {
				t.Helper()
				g.Expect(r.Reconcile(ctx, req)).To(Equal(reconcile.Result{}))
				g.Expect(kube.Get(ctx, req.NamespacedName, node)).To(Succeed())
			}
			run()
			g.Expect(statusWrites).To(Equal(2))
			g.Expect(node.Status.ChaosResource).NotTo(BeNil())
			g.Expect(node.Status.ChaosResource.Name).To(Equal(first.GetName()))
			g.Expect(ConditionEqualsTo(node.Status, v1alpha1.ConditionChaosInjected, corev1.ConditionTrue)).To(BeTrue())
			g.Expect(kube.Get(ctx, client.ObjectKeyFromObject(duplicate), tc.newChild())).To(Succeed())
			blockDelete = false
			run()
			g.Expect(kube.Get(ctx, client.ObjectKeyFromObject(duplicate), tc.newChild())).To(Satisfy(apierrors.IsNotFound))
			run()
			g.Expect(node.Status.ChaosResource.Name).To(Equal(first.GetName()))
			SetCondition(&node.Status, v1alpha1.WorkflowNodeCondition{Type: v1alpha1.ConditionAccomplished, Status: corev1.ConditionTrue})
			g.Expect(kube.Status().Update(ctx, node)).To(Succeed())
			run()
			g.Expect(node.Status.ChaosResource).To(BeNil())
			g.Expect(ConditionEqualsTo(node.Status, v1alpha1.ConditionChaosInjected, corev1.ConditionFalse)).To(BeTrue())
			g.Expect(kube.Get(ctx, client.ObjectKeyFromObject(first), tc.newChild())).To(Satisfy(apierrors.IsNotFound))
			g.Expect(kube.Get(ctx, client.ObjectKeyFromObject(foreign), tc.newChild())).To(Succeed())
			g.Expect(kube.created).To(BeEmpty())
		})
	}
}

func TestWorkflowWaitsForStaleParallelAndConditionalChildren(t *testing.T) {
	for _, nodeType := range []v1alpha1.TemplateType{v1alpha1.TypeParallel, v1alpha1.TypeTask} {
		t.Run(string(nodeType), func(t *testing.T) {
			g := NewWithT(t)
			ctx := t.Context()
			workflow := &v1alpha1.Workflow{Namespace: "test", Name: "parent", Spec: v1alpha1.WorkflowSpec{
				Templates: []v1alpha1.Template{{Name: "keep", Type: v1alpha1.TypeSuspend}, {Name: "next", Type: v1alpha1.TypeSuspend}},
			}}
			parent := &v1alpha1.WorkflowNode{
				Namespace: workflow.Namespace, Name: "parent-node", UID: "parent-node-uid",
				Spec: v1alpha1.WorkflowNodeSpec{Type: nodeType, WorkflowName: workflow.Name},
			}
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: parent.Namespace, Name: "task-pod",
					Labels: map[string]string{v1alpha1.LabelControlledBy: parent.Name},
				},
				Status: corev1.PodStatus{Phase: corev1.PodSucceeded},
			}
			if nodeType == v1alpha1.TypeParallel {
				parent.Spec.Children = []string{"keep", "next"}
			} else {
				parent.Spec.ConditionalBranches = []v1alpha1.ConditionalBranch{{Target: "keep"}, {Target: "next"}}
				parent.Status.ConditionalBranchesStatus = &v1alpha1.ConditionalBranchesStatus{
					Branches: []v1alpha1.ConditionalBranchStatus{
						{Target: "keep", EvaluationResult: corev1.ConditionTrue},
						{Target: "next", EvaluationResult: corev1.ConditionTrue},
					},
				}
			}
			child := func(template string) *v1alpha1.WorkflowNode {
				result := &v1alpha1.WorkflowNode{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: parent.Namespace, Name: template + "-old", UID: types.UID(template + "-uid"),
						Labels:          map[string]string{v1alpha1.LabelControlledBy: parent.Name},
						OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(parent, v1alpha1.SchemeGroupVersion.WithKind("WorkflowNode"))},
					},
					Spec: v1alpha1.WorkflowNodeSpec{TemplateName: template},
				}
				SetCondition(&result.Status, v1alpha1.WorkflowNodeCondition{Type: v1alpha1.ConditionAccomplished, Status: corev1.ConditionTrue})
				return result
			}
			keep, obsolete := child("keep"), child("obsolete")
			obsolete.Finalizers = []string{"fixture.invalid/cleanup"}
			kube := workflowUnitClient(t, workflow, parent, pod, keep, obsolete)
			failure := errors.New("delete interrupted")
			deletes, creates := 0, 0
			wrapped := interceptor.NewClient(kube, interceptor.Funcs{
				Delete: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
					deletes++
					g.Expect(obj.GetUID()).To(Equal(obsolete.UID))
					options := (&client.DeleteOptions{}).ApplyOptions(opts)
					g.Expect(options.Preconditions).NotTo(BeNil())
					g.Expect(options.Preconditions.UID).To(HaveValue(Equal(obsolete.UID)))
					if deletes == 1 {
						return failure
					}
					return c.Delete(ctx, obj, opts...)
				},
				Create: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
					creates++
					g.Expect(obj.(*v1alpha1.WorkflowNode).Spec.TemplateName).To(Equal("next"))
					return c.Create(ctx, obj, opts...)
				},
			})
			var r reconcile.Reconciler = NewParallelNodeReconciler(wrapped, recorder.NewDebugRecorder(), logr.Discard())
			if nodeType == v1alpha1.TypeTask {
				r = NewTaskReconciler(wrapped, nil, recorder.NewDebugRecorder(), logr.Discard())
			}
			key := client.ObjectKeyFromObject(parent)
			req := reconcile.Request{NamespacedName: key}
			_, err := r.Reconcile(ctx, req)
			g.Expect(err).To(MatchError(failure))
			for range 2 {
				result, err := r.Reconcile(ctx, req)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(result.RequeueAfter).To(BeNumerically(">", 0))
				g.Expect(creates).To(BeZero())
				g.Expect(kube.Get(ctx, key, parent)).To(Succeed())
				g.Expect(WorkflowNodeFinished(parent.Status)).To(BeFalse())
			}
			g.Expect(deletes).To(Equal(2))
			g.Expect(kube.Get(ctx, client.ObjectKeyFromObject(obsolete), obsolete)).To(Succeed())
			g.Expect(obsolete.DeletionTimestamp).NotTo(BeNil())
			obsolete.Finalizers = nil
			g.Expect(kube.Update(ctx, obsolete)).To(Succeed())

			for range 2 {
				_, err := r.Reconcile(ctx, req)
				g.Expect(err).NotTo(HaveOccurred())
			}
			g.Expect(creates).To(Equal(1))
			g.Expect(kube.Get(ctx, client.ObjectKeyFromObject(keep), keep)).To(Succeed())
			g.Expect(kube.Get(ctx, key, parent)).To(Succeed())
			g.Expect(parent.Status.FinishedChildren).To(Equal([]corev1.LocalObjectReference{{Name: keep.Name}}))
			g.Expect(parent.Status.ActiveChildren).To(HaveLen(1))
		})
	}
}
