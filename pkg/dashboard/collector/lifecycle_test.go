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

package collector

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/go-logr/logr"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
	eventstore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/event"
	experimentstore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/experiment"
	schedulestore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/schedule"
	workflowstore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/workflow"
)

func lifecycleDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "dashboard.db")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Error(err)
		}
	})
	return db
}

func reconcileObject(t *testing.T, reconciler interface {
	Reconcile(context.Context, ctrl.Request) (ctrl.Result, error)
}, obj client.Object) {
	t.Helper()
	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(obj)})
	if err != nil || result != (ctrl.Result{}) {
		t.Fatalf("reconcile = %#v, %v", result, err)
	}
}

func TestChaosCollectorLifecycle(t *testing.T) {
	for _, managed := range []bool{false, true} {
		t.Run(map[bool]string{false: "standalone archive", true: "managed cleanup"}[managed], func(t *testing.T) {
			ctx := context.Background()
			db := lifecycleDB(t)
			experiments, err := experimentstore.NewStore(db)
			if err != nil {
				t.Fatal(err)
			}
			events, err := eventstore.NewStore(db)
			if err != nil {
				t.Fatal(err)
			}
			chaos := &v1alpha1.PodChaos{ObjectMeta: metav1.ObjectMeta{Name: "experiment", Namespace: "default", UID: "experiment-uid", Finalizers: []string{"test"}}, Spec: v1alpha1.PodChaosSpec{Action: v1alpha1.PodKillAction}}
			if managed {
				chaos.Labels = map[string]string{v1alpha1.LabelManagedBy: "workflow"}
			}
			kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(chaos).WithInterceptorFuncs(interceptor.Funcs{Get: getWithTypeMeta}).Build()
			collector := &ChaosCollector{Client: kube, Log: logr.Discard(), apiType: &v1alpha1.PodChaos{}, store: experiments, eventStore: events}
			reconcileObject(t, collector, chaos)
			first, err := experiments.FindByUID(ctx, string(chaos.UID))
			if err != nil {
				t.Fatal(err)
			}
			if first.ID == 0 || first.Archived || first.FinishTime != nil || first.Action != string(v1alpha1.PodKillAction) {
				t.Fatalf("created = %#v", first)
			}
			reconcileObject(t, collector, chaos)
			if err := events.Create(ctx, &core.Event{ObjectID: string(chaos.UID), Reason: "Applied"}); err != nil {
				t.Fatal(err)
			}
			if err := events.Create(ctx, &core.Event{ObjectID: "other-uid", Reason: "Unrelated"}); err != nil {
				t.Fatal(err)
			}
			if err := kube.Delete(ctx, chaos); err != nil {
				t.Fatal(err)
			}
			reconcileObject(t, collector, chaos)
			deleting, err := experiments.FindByUID(ctx, string(chaos.UID))
			if err != nil {
				t.Fatal(err)
			}
			if deleting.ID != first.ID || deleting.FinishTime == nil || deleting.Archived || deleting.Experiment != first.Experiment {
				t.Fatalf("deleting = %#v", deleting)
			}
			current := &v1alpha1.PodChaos{}
			if err := kube.Get(ctx, client.ObjectKeyFromObject(chaos), current); err != nil {
				t.Fatal(err)
			}
			current.Finalizers = nil
			if err := kube.Update(ctx, current); err != nil {
				t.Fatal(err)
			}
			reconcileObject(t, collector, chaos)
			reconcileObject(t, collector, chaos)
			archived, err := experiments.FindByUID(ctx, string(chaos.UID))
			related, eventErr := events.ListByUID(ctx, string(chaos.UID))
			if eventErr != nil {
				t.Fatal(eventErr)
			}
			if managed {
				if !errors.Is(err, gorm.ErrRecordNotFound) || len(related) != 0 {
					t.Fatalf("managed row = %#v, error = %v, events = %#v", archived, err, related)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if !archived.Archived || archived.ID != first.ID || archived.FinishTime == nil || !archived.FinishTime.Equal(*deleting.FinishTime) || len(related) != 1 {
					t.Fatalf("archived = %#v, events = %#v", archived, related)
				}
			}
			other, err := events.ListByUID(ctx, "other-uid")
			if err != nil || len(other) != 1 {
				t.Fatalf("unrelated events = %#v, error = %v", other, err)
			}
			all, err := experiments.ListMeta(ctx, "", "", "", false)
			if err != nil || len(all) != 0 {
				t.Fatalf("active rows = %#v, error = %v", all, err)
			}
		})
	}
}

func TestChaosCollectorPersistsAzure(t *testing.T) {
	store, err := experimentstore.NewStore(lifecycleDB(t))
	if err != nil {
		t.Fatal(err)
	}
	chaos := &v1alpha1.AzureChaos{ObjectMeta: metav1.ObjectMeta{Name: "azure", Namespace: "default", UID: "azure-uid"}, Spec: v1alpha1.AzureChaosSpec{Action: v1alpha1.AzureVmStop}}
	kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(chaos).WithInterceptorFuncs(interceptor.Funcs{Get: getWithTypeMeta}).Build()
	collector := &ChaosCollector{Client: kube, Log: logr.Discard(), apiType: &v1alpha1.AzureChaos{}, store: store}
	reconcileObject(t, collector, chaos)
	got, err := store.FindByUID(context.Background(), "azure-uid")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != v1alpha1.KindAzureChaos || got.Action != string(v1alpha1.AzureVmStop) {
		t.Fatalf("Azure archive = %#v", got)
	}
}

func TestScheduleCollectorLifecycle(t *testing.T) {
	ctx := context.Background()
	store, err := schedulestore.NewStore(lifecycleDB(t))
	if err != nil {
		t.Fatal(err)
	}
	schedule := &v1alpha1.Schedule{ObjectMeta: metav1.ObjectMeta{Name: "schedule", Namespace: "default", UID: "schedule-uid", Finalizers: []string{"test"}}, Spec: v1alpha1.ScheduleSpec{Type: v1alpha1.ScheduleTypePodChaos, ScheduleItem: v1alpha1.ScheduleItem{EmbedChaos: v1alpha1.EmbedChaos{PodChaos: &v1alpha1.PodChaosSpec{Action: v1alpha1.PodKillAction}}}}}
	kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(schedule).Build()
	collector := &ScheduleCollector{Client: kube, Log: logr.Discard(), apiType: &v1alpha1.Schedule{}, store: store}
	reconcileObject(t, collector, schedule)
	first, err := store.FindByUID(ctx, "schedule-uid")
	if err != nil {
		t.Fatal(err)
	}
	if first.ID == 0 || first.Action != string(v1alpha1.PodKillAction) || first.Archived || first.FinishTime != nil {
		t.Fatalf("created schedule = %#v", first)
	}
	reconcileObject(t, collector, schedule)
	if err := kube.Delete(ctx, schedule); err != nil {
		t.Fatal(err)
	}
	reconcileObject(t, collector, schedule)
	deleting, err := store.FindByUID(ctx, "schedule-uid")
	if err != nil {
		t.Fatal(err)
	}
	if deleting.FinishTime == nil || deleting.ID != first.ID || deleting.Schedule != first.Schedule {
		t.Fatalf("deleting schedule = %#v", deleting)
	}
	current := &v1alpha1.Schedule{}
	if err := kube.Get(ctx, client.ObjectKeyFromObject(schedule), current); err != nil {
		t.Fatal(err)
	}
	current.Finalizers = nil
	if err := kube.Update(ctx, current); err != nil {
		t.Fatal(err)
	}
	reconcileObject(t, collector, schedule)
	reconcileObject(t, collector, schedule)
	archived, err := store.FindByUID(ctx, "schedule-uid")
	if err != nil {
		t.Fatal(err)
	}
	if !archived.Archived || archived.ID != first.ID || !archived.FinishTime.Equal(*deleting.FinishTime) {
		t.Fatalf("archived schedule = %#v", archived)
	}
	rows, err := store.ListMeta(ctx, "", "", true)
	if err != nil || len(rows) != 1 {
		t.Fatalf("archived schedules = %#v, error = %v", rows, err)
	}
}

func TestWorkflowCollectorLifecycle(t *testing.T) {
	ctx := context.Background()
	store, err := workflowstore.NewStore(lifecycleDB(t))
	if err != nil {
		t.Fatal(err)
	}
	wf := &v1alpha1.Workflow{ObjectMeta: metav1.ObjectMeta{Name: "workflow", Namespace: "default", UID: "workflow-uid"}, Spec: v1alpha1.WorkflowSpec{Entry: "wait"}}
	kube := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(&v1alpha1.Workflow{}).WithObjects(wf).Build()
	collector := &WorkflowCollector{kubeClient: kube, Log: logr.Discard(), apiType: &v1alpha1.Workflow{}, store: store}
	reconcileObject(t, collector, wf)
	first, err := store.FindByUID(ctx, "workflow-uid")
	if err != nil {
		t.Fatal(err)
	}
	if first.ID == 0 || first.Status != core.WorkflowUnknown || first.Archived {
		t.Fatalf("initial workflow = %#v", first)
	}
	current := &v1alpha1.Workflow{}
	if err := kube.Get(ctx, client.ObjectKeyFromObject(wf), current); err != nil {
		t.Fatal(err)
	}
	end := metav1.NewTime(time.Now().Add(-time.Hour).UTC().Truncate(time.Second))
	current.Status.EndTime = &end
	current.Status.Conditions = []v1alpha1.WorkflowCondition{{Type: v1alpha1.WorkflowConditionAccomplished, Status: corev1.ConditionTrue}}
	if err := kube.Status().Update(ctx, current); err != nil {
		t.Fatal(err)
	}
	reconcileObject(t, collector, wf)
	completed, err := store.FindByUID(ctx, "workflow-uid")
	if err != nil {
		t.Fatal(err)
	}
	if completed.ID != first.ID || completed.Status != core.WorkflowSucceed || parseTime(t, completed.EndTime) != end.Time || completed.FinishTime != nil {
		t.Fatalf("completed workflow = %#v", completed)
	}
	restored, err := core.WorkflowEntity2WorkflowCR(completed)
	if err != nil || restored.Status.EndTime == nil || !restored.Status.EndTime.Equal(&end) {
		t.Fatalf("persisted workflow = %#v, error = %v", restored, err)
	}
	if err := kube.Delete(ctx, current); err != nil {
		t.Fatal(err)
	}
	reconcileObject(t, collector, wf)
	archived, err := store.FindByUID(ctx, "workflow-uid")
	if err != nil {
		t.Fatal(err)
	}
	if !archived.Archived || archived.FinishTime == nil || archived.EndTime != completed.EndTime {
		t.Fatalf("archived workflow = %#v", archived)
	}
	reconcileObject(t, collector, wf)
	again, err := store.FindByUID(ctx, "workflow-uid")
	if err != nil || again.ID != archived.ID || !again.FinishTime.Equal(*archived.FinishTime) {
		t.Fatalf("repeated archive = %#v, error = %v", again, err)
	}
}

func TestEventCollectorFiltersMissingAndUnrelatedObjects(t *testing.T) {
	for _, tc := range []struct {
		name, kind string
		exists     bool
		want       int
	}{
		{"chaos", v1alpha1.KindPodChaos, true, 1}, {"schedule", v1alpha1.KindSchedule, true, 1}, {"workflow", v1alpha1.KindWorkflow, true, 1}, {"node", v1alpha1.KindWorkflowNode, true, 1}, {"missing", v1alpha1.KindPodChaos, false, 0}, {"unrelated", "Pod", true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store, err := eventstore.NewStore(lifecycleDB(t))
			if err != nil {
				t.Fatal(err)
			}
			event := &corev1.Event{ObjectMeta: metav1.ObjectMeta{Name: "event", Namespace: "default", CreationTimestamp: metav1.NewTime(time.Now().UTC().Truncate(time.Second))}, InvolvedObject: corev1.ObjectReference{Name: "object", Namespace: "default", UID: types.UID("object-uid"), Kind: tc.kind}, Type: "Warning", Reason: "Failed", Message: "injection failed"}
			objects := []client.Object{event}
			if tc.exists {
				meta := metav1.ObjectMeta{Name: "object", Namespace: "default", UID: "object-uid"}
				switch tc.kind {
				case v1alpha1.KindPodChaos:
					objects = append(objects, &v1alpha1.PodChaos{ObjectMeta: meta})
				case v1alpha1.KindSchedule:
					objects = append(objects, &v1alpha1.Schedule{ObjectMeta: meta})
				case v1alpha1.KindWorkflow:
					objects = append(objects, &v1alpha1.Workflow{ObjectMeta: meta})
				case v1alpha1.KindWorkflowNode:
					objects = append(objects, &v1alpha1.WorkflowNode{ObjectMeta: meta})
				case "Pod":
					objects = append(objects, &corev1.Pod{ObjectMeta: meta})
				}
			}
			kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
			collector := &EventCollector{Client: kube, Log: logr.Discard(), apiType: &corev1.Event{}, eventStore: store}
			reconcileObject(t, collector, event)
			rows, err := store.List(context.Background())
			if err != nil || len(rows) != tc.want {
				t.Fatalf("events = %#v, error = %v", rows, err)
			}
			if tc.want == 1 && (rows[0].ObjectID != "object-uid" || rows[0].Kind != tc.kind || rows[0].Name != "object" || rows[0].Namespace != "default" || rows[0].Reason != event.Reason || rows[0].Message != event.Message || rows[0].Type != event.Type || !rows[0].CreatedAt.Equal(event.CreationTimestamp.Time)) {
				t.Fatalf("persisted event = %#v", rows[0])
			}
		})
	}
}

func getWithTypeMeta(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if err := c.Get(ctx, key, obj, opts...); err != nil {
		return err
	}
	gvk, err := apiutil.GVKForObject(obj, scheme)
	if err != nil {
		return err
	}
	obj.GetObjectKind().SetGroupVersionKind(gvk)
	return nil
}

func parseTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatal(err)
	}
	return parsed.UTC()
}
