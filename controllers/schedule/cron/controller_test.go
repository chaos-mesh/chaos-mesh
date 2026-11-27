// Copyright 2026 Chaos Mesh Authors.
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

package cron

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/schedule/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
)

func TestCron(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Schedule cron suite")
}

var _ = Describe("Reconcile", func() {
	DescribeTable("records an already created workflow after an API failure",
		func(ctx SpecContext, policy v1alpha1.ConcurrencyPolicy, lostCreateResponse bool) {
			scheme := runtime.NewScheme()
			Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
			started := time.Now().UTC().Truncate(time.Second)
			schedule := &v1alpha1.Schedule{
				Namespace: "test", Name: "retry", UID: "schedule-uid",
				CreationTimestamp: metav1.NewTime(started.Add(-15 * time.Minute)),
				Spec: v1alpha1.ScheduleSpec{
					Schedule: "@every 10m", ConcurrencyPolicy: policy, Type: v1alpha1.ScheduleTypeWorkflow,
					ScheduleItem: v1alpha1.ScheduleItem{Workflow: &v1alpha1.WorkflowSpec{
						Entry: "wait", Templates: []v1alpha1.Template{{Name: "wait", Type: v1alpha1.TypeSuspend}},
					}},
				},
			}
			failure := k8sError.NewServiceUnavailable("schedule write unavailable")
			creates := 0
			updates := 0
			kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(schedule).
				WithStatusSubresource(&v1alpha1.Workflow{}).
				WithInterceptorFuncs(interceptor.Funcs{
					Create: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
						creates++
						if err := c.Create(ctx, obj, opts...); err != nil {
							return err
						}
						if lostCreateResponse && creates == 1 {
							return failure
						}
						return nil
					},
					Update: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
						updates++
						if !lostCreateResponse && updates == 1 {
							return failure
						}
						return c.Update(ctx, obj, opts...)
					},
				}).Build()
			events := recorder.NewDebugRecorder()
			r := &Reconciler{Client: kube, Log: logr.Discard(), ActiveLister: utils.NewActiveLister(kube, logr.Discard()), Recorder: events}
			req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(schedule)}
			_, firstErr := r.reconcile(ctx, req, started)
			Expect(kube.Get(ctx, req.NamespacedName, schedule)).To(Succeed())
			Expect(schedule.Status.LastScheduleTime.IsZero()).To(BeTrue())
			Expect(creates).To(Equal(1))

			// Recover across a later tick and a controller restart.
			retryAt := started.Add(25 * time.Minute)
			r = &Reconciler{Client: kube, Log: logr.Discard(), ActiveLister: utils.NewActiveLister(kube, logr.Discard()), Recorder: events}
			_, err := r.reconcile(ctx, req, retryAt)
			Expect(err).NotTo(HaveOccurred())
			Expect(creates).To(Equal(1))
			Expect(kube.Get(ctx, req.NamespacedName, schedule)).To(Succeed())
			Expect(schedule.Status.LastScheduleTime.Time).To(BeTemporally("==", started))
			Expect(firstErr).To(MatchError(failure))

			workflows := &v1alpha1.WorkflowList{}
			Expect(kube.List(ctx, workflows, client.InNamespace(schedule.Namespace))).To(Succeed())
			Expect(workflows.Items).To(HaveLen(1))
			original := &workflows.Items[0]
			if policy.IsForbid() {
				_, err = r.reconcile(ctx, req, retryAt)
				Expect(err).NotTo(HaveOccurred())
				Expect(creates).To(Equal(1))
				original.Status.Conditions = []v1alpha1.WorkflowCondition{{Type: v1alpha1.WorkflowConditionAccomplished, Status: corev1.ConditionTrue}}
				Expect(kube.Status().Update(ctx, original)).To(Succeed())
			}
			_, err = r.reconcile(ctx, req, retryAt)
			Expect(err).NotTo(HaveOccurred())
			Expect(creates).To(Equal(2))
			Expect(kube.Get(ctx, client.ObjectKeyFromObject(original), original)).To(Succeed())
			Expect(kube.Get(ctx, req.NamespacedName, schedule)).To(Succeed())
			Expect(schedule.Status.LastScheduleTime.Time).To(BeTemporally("==", retryAt))
		},
		Entry("Allow policy with a failed schedule update", v1alpha1.AllowConcurrent, false),
		Entry("Forbid policy with a failed schedule update", v1alpha1.ForbidConcurrent, false),
		Entry("a lost create response", v1alpha1.AllowConcurrent, true),
	)

	It("does not adopt a same-named workflow from another Schedule", func(ctx SpecContext) {
		scheme := runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		now := time.Now().UTC().Truncate(time.Second)
		schedule := &v1alpha1.Schedule{
			Namespace: "test", Name: "retry", UID: "new-schedule",
			CreationTimestamp: metav1.NewTime(now.Add(-time.Hour)),
			Spec: v1alpha1.ScheduleSpec{
				Schedule: "@every 10m", ConcurrencyPolicy: v1alpha1.AllowConcurrent, Type: v1alpha1.ScheduleTypeWorkflow,
				ScheduleItem: v1alpha1.ScheduleItem{Workflow: &v1alpha1.WorkflowSpec{Entry: "wait"}},
			},
		}
		old := schedule.DeepCopy()
		old.UID = "old-schedule"
		workflow := &v1alpha1.Workflow{ObjectMeta: metav1.ObjectMeta{
			Namespace: schedule.Namespace, Name: nextRunName(schedule),
			Annotations:     map[string]string{runStartedAtAnnotation: now.Format(time.RFC3339)},
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(old, v1alpha1.SchemeGroupVersion.WithKind(v1alpha1.KindSchedule))},
		}}
		kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(schedule, workflow).Build()
		r := &Reconciler{Client: kube, Log: logr.Discard(), Recorder: recorder.NewDebugRecorder()}
		req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(schedule)}
		_, err := r.reconcile(ctx, req, now)
		Expect(err).To(MatchError(ContainSubstring("belongs to a different controller")))
		Expect(kube.Get(ctx, req.NamespacedName, schedule)).To(Succeed())
		Expect(schedule.Status.LastScheduleTime.IsZero()).To(BeTrue())
	})

	It("reports missing experiment specs in persisted Schedules", func(ctx SpecContext) {
		scheme := runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		for _, kind := range []v1alpha1.ScheduleTemplateType{v1alpha1.ScheduleTypeAzureChaos, v1alpha1.ScheduleTypeWorkflow} {
			schedule := &v1alpha1.Schedule{
				Namespace: "test", Name: "invalid", CreationTimestamp: metav1.NewTime(time.Now().Add(-15 * time.Minute)),
				Spec: v1alpha1.ScheduleSpec{Schedule: "@every 10m", ConcurrencyPolicy: v1alpha1.AllowConcurrent, Type: kind},
			}
			creates := 0
			kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(schedule).
				WithInterceptorFuncs(interceptor.Funcs{
					Create: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
						creates++
						return c.Create(ctx, obj, opts...)
					},
				}).Build()
			events := recorder.NewDebugRecorder()
			r := &Reconciler{Client: kube, Log: logr.Discard(), Recorder: events}
			key := client.ObjectKeyFromObject(schedule)

			Expect(r.Reconcile(ctx, ctrl.Request{NamespacedName: key})).To(Equal(ctrl.Result{}))
			Expect(creates).To(BeZero())
			Expect(events.Events[key]).To(HaveLen(1))
			Expect(events.Events[key][0]).To(BeAssignableToTypeOf(recorder.Failed{}))
			failure := events.Events[key][0].(recorder.Failed)
			Expect(failure.Activity).To(Equal("generate new object"))
			Expect(failure.Err).To(ContainSubstring("missing " + string(kind) + " spec"))
			Expect(kube.Get(ctx, key, schedule)).To(Succeed())
			Expect(schedule.Status.LastScheduleTime.IsZero()).To(BeTrue())
		}
	}, SpecTimeout(5*time.Second))

	It("logs no error when the Schedule has been deleted", func() {
		scheme := runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())

		var errorLogs []string
		inner := funcr.New(func(_, _ string) {}, funcr.Options{})
		r := &Reconciler{
			Client: fake.NewClientBuilder().WithScheme(scheme).Build(),
			Log:    logr.New(&errorCountingSink{LogSink: inner.GetSink(), count: &errorLogs}),
		}

		result, err := r.Reconcile(context.Background(), ctrl.Request{
			NamespacedName: types.NamespacedName{Namespace: "default", Name: "deleted-schedule"},
		})

		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(ctrl.Result{}))
		Expect(errorLogs).To(BeEmpty())
	})

	It("returns other Schedule read errors for retry", func() {
		readErr := k8sError.NewServiceUnavailable("failed to read Schedule")
		r := &Reconciler{
			Client: fake.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{
				Get: func(context.Context, client.WithWatch, client.ObjectKey, client.Object, ...client.GetOption) error {
					return readErr
				},
			}).Build(),
		}

		result, err := r.Reconcile(context.Background(), ctrl.Request{
			NamespacedName: types.NamespacedName{Namespace: "default", Name: "schedule"},
		})

		Expect(err).To(MatchError(readErr))
		Expect(result).To(Equal(ctrl.Result{}))
	})
})

// errorCountingSink records the message of every Error call.
type errorCountingSink struct {
	logr.LogSink
	count *[]string
}

func (s *errorCountingSink) Error(err error, msg string, keysAndValues ...interface{}) {
	*s.count = append(*s.count, msg)
	s.LogSink.Error(err, msg, keysAndValues...)
}

var _ = Describe("Scheduling workflows with Forbid concurrency", func() {
	It("waits for the current workflow, preserves the schedule time on create failure, and retries a status conflict", func(ctx SpecContext) {
		scheme := runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		schedule := &v1alpha1.Schedule{
			TypeMeta: metav1.TypeMeta{Kind: "Schedule", APIVersion: v1alpha1.SchemeGroupVersion.String()},
			Name:     "fixture", Namespace: "test", UID: "schedule-uid", CreationTimestamp: metav1.NewTime(time.Now().Add(-35 * time.Minute)),
			Spec: v1alpha1.ScheduleSpec{
				Type: v1alpha1.ScheduleTypeWorkflow, Schedule: "@every 10m", ConcurrencyPolicy: v1alpha1.ForbidConcurrent,
				ScheduleItem: v1alpha1.ScheduleItem{Workflow: &v1alpha1.WorkflowSpec{Entry: "suspend", Templates: []v1alpha1.Template{{Name: "suspend", Type: v1alpha1.TypeSuspend}}}},
			},
		}
		active := &v1alpha1.Workflow{ObjectMeta: metav1.ObjectMeta{Namespace: schedule.Namespace, Name: "active", Labels: map[string]string{v1alpha1.LabelManagedBy: schedule.Name}}}
		foreign := active.DeepCopy()
		foreign.Namespace = "other"
		failCreate := false
		updates := 0
		kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(schedule, active, foreign).
			WithStatusSubresource(active).
			WithInterceptorFuncs(interceptor.Funcs{
				Create: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
					if failCreate {
						return errors.New("workflow creation failed")
					}
					return c.Create(ctx, obj, opts...)
				},
				Update: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
					updates++
					if updates == 1 {
						return k8sError.NewConflict(schema.GroupResource{Group: "chaos-mesh.org", Resource: "schedules"}, obj.GetName(), errors.New("concurrent update"))
					}
					return c.Update(ctx, obj, opts...)
				},
			}).Build()
		events := recorder.NewDebugRecorder()
		r := &Reconciler{Client: kube, Log: logr.Discard(), ActiveLister: utils.NewActiveLister(kube, logr.Discard()), Recorder: events}
		req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(schedule)}
		Expect(r.Reconcile(ctx, req)).To(Equal(ctrl.Result{}))
		Expect(events.Events[req.NamespacedName]).To(ConsistOf(recorder.ScheduleForbid{RunningName: active.Name}))
		Expect(updates).To(BeZero())

		active.Status.Conditions = []v1alpha1.WorkflowCondition{{Type: v1alpha1.WorkflowConditionAccomplished, Status: corev1.ConditionTrue}}
		Expect(kube.Status().Update(ctx, active)).To(Succeed())
		failCreate = true
		_, err := r.Reconcile(ctx, req)
		Expect(err).To(MatchError("workflow creation failed"))
		Expect(events.Events[req.NamespacedName]).To(ContainElement(recorder.Failed{Activity: "create new object", Err: "workflow creation failed"}))
		Expect(kube.Get(ctx, req.NamespacedName, schedule)).To(Succeed())
		Expect(schedule.Status.LastScheduleTime.IsZero()).To(BeTrue())

		failCreate = false
		Expect(r.Reconcile(ctx, req)).To(Equal(ctrl.Result{}))
		Expect(updates).To(Equal(2))
		Expect(kube.Get(ctx, req.NamespacedName, schedule)).To(Succeed())
		Expect(schedule.Status.LastScheduleTime.IsZero()).To(BeFalse())
		var workflows v1alpha1.WorkflowList
		Expect(kube.List(ctx, &workflows, client.InNamespace(schedule.Namespace))).To(Succeed())
		Expect(workflows.Items).To(HaveLen(2))
		for _, item := range workflows.Items {
			if item.Name == active.Name {
				continue
			}
			Expect(item.GetLabels()).To(HaveKeyWithValue(v1alpha1.LabelManagedBy, schedule.Name))
			Expect(metav1.GetControllerOf(&item).UID).To(Equal(schedule.UID))
			Expect(item.Spec).To(Equal(*schedule.Spec.Workflow))
		}
		result, err := r.Reconcile(ctx, req)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.RequeueAfter).To(BeNumerically(">", 0))
		Expect(updates).To(Equal(2))
	}, SpecTimeout(5*time.Second))
})
