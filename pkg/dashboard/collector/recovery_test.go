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

package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
	eventstore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/event"
	experimentstore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/experiment"
	schedulestore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/schedule"
	workflowstore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/workflow"
)

type recoveryFixture struct {
	db             *gorm.DB
	object         client.Object
	kube           client.Client
	reconciler     reconcile.Reconciler
	table          string
	getError       error
	cancelAfterGet context.CancelFunc
	getCalls       int
}

func newRecoveryFixture(t *testing.T, kind string) *recoveryFixture {
	t.Helper()
	f := &recoveryFixture{db: lifecycleDB(t)}
	meta := metav1.ObjectMeta{Name: "resource", Namespace: "default", UID: "resource-uid"}
	objects := []client.Object{}
	switch kind {
	case "chaos":
		f.object = &v1alpha1.PodChaos{ObjectMeta: meta, Spec: v1alpha1.PodChaosSpec{Action: v1alpha1.PodKillAction}}
		f.table = "experiments"
	case "schedule":
		f.object = &v1alpha1.Schedule{ObjectMeta: meta, Spec: v1alpha1.ScheduleSpec{Type: v1alpha1.ScheduleTypePodChaos, ScheduleItem: v1alpha1.ScheduleItem{EmbedChaos: v1alpha1.EmbedChaos{PodChaos: &v1alpha1.PodChaosSpec{Action: v1alpha1.PodKillAction}}}}}
		f.table = "schedules"
	case "workflow":
		f.object = &v1alpha1.Workflow{ObjectMeta: meta, Spec: v1alpha1.WorkflowSpec{Entry: "wait"}}
		f.table = "workflow_entities"
	case "event":
		f.object = &corev1.Event{ObjectMeta: meta, InvolvedObject: corev1.ObjectReference{Name: "target", Namespace: "default", UID: "target-uid", Kind: v1alpha1.KindPodChaos}, Reason: "Applied"}
		objects = append(objects, &v1alpha1.PodChaos{ObjectMeta: metav1.ObjectMeta{Name: "target", Namespace: "default", UID: "target-uid"}})
		f.table = "events"
	default:
		t.Fatalf("unknown fixture %s", kind)
	}
	objects = append(objects, f.object)
	f.kube = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).WithInterceptorFuncs(interceptor.Funcs{Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
		f.getCalls++
		if f.getError != nil {
			return f.getError
		}
		err := getWithTypeMeta(ctx, c, key, obj, opts...)
		if f.cancelAfterGet != nil {
			f.cancelAfterGet()
		}
		return err
	}}).Build()
	switch kind {
	case "chaos":
		store, err := experimentstore.NewStore(f.db)
		if err != nil {
			t.Fatal(err)
		}
		events, err := eventstore.NewStore(f.db)
		if err != nil {
			t.Fatal(err)
		}
		f.reconciler = &ChaosCollector{Client: f.kube, Log: logr.Discard(), apiType: &v1alpha1.PodChaos{}, store: store, eventStore: events}
	case "schedule":
		store, err := schedulestore.NewStore(f.db)
		if err != nil {
			t.Fatal(err)
		}
		f.reconciler = &ScheduleCollector{Client: f.kube, Log: logr.Discard(), apiType: &v1alpha1.Schedule{}, store: store}
	case "workflow":
		store, err := workflowstore.NewStore(f.db)
		if err != nil {
			t.Fatal(err)
		}
		f.reconciler = &WorkflowCollector{kubeClient: f.kube, Log: logr.Discard(), apiType: &v1alpha1.Workflow{}, store: store}
	case "event":
		store, err := eventstore.NewStore(f.db)
		if err != nil {
			t.Fatal(err)
		}
		f.reconciler = &EventCollector{Client: f.kube, Log: logr.Discard(), apiType: &corev1.Event{}, eventStore: store}
	}
	return f
}

func (f *recoveryFixture) run(ctx context.Context) (ctrl.Result, error) {
	return f.reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(f.object)})
}

func (f *recoveryFixture) rowCount(t *testing.T) int64 {
	t.Helper()
	var count int64
	if err := f.db.Table(f.table).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	return count
}

func requireCollectorRetry(t *testing.T, result ctrl.Result, err error) {
	t.Helper()
	if err != nil || !result.Requeue || result.RequeueAfter != 0 {
		t.Errorf("transient failure must request rate-limited retry, got %#v, %v", result, err)
	}
}

func failCollectorDB(t *testing.T, db *gorm.DB, operation, table string) func() {
	t.Helper()
	enabled := true
	fail := func(tx *gorm.DB) {
		if enabled && tx.Statement.Table == table {
			tx.AddError(errors.New("injected database failure"))
		}
	}
	var err error
	switch operation {
	case "query":
		err = db.Callback().Query().Before("gorm:query").Register("collector_failure", fail)
	case "create":
		err = db.Callback().Create().Before("gorm:create").Register("collector_failure", fail)
	case "update":
		err = db.Callback().Update().Before("gorm:update").Register("collector_failure", fail)
	case "delete":
		err = db.Callback().Delete().Before("gorm:delete").Register("collector_failure", fail)
	default:
		t.Fatalf("unknown database operation %s", operation)
	}
	if err != nil {
		t.Fatal(err)
	}
	return func() { enabled = false }
}

func TestCollectorsRetryFailedGetWithoutDatabaseEffects(t *testing.T) {
	for _, kind := range []string{"chaos", "schedule", "workflow", "event"} {
		t.Run(kind, func(t *testing.T) {
			f := newRecoveryFixture(t, kind)
			dbCalls := 0
			observe := func(*gorm.DB) { dbCalls++ }
			if err := f.db.Callback().Query().Before("gorm:query").Register("observe", observe); err != nil {
				t.Fatal(err)
			}
			if err := f.db.Callback().Create().Before("gorm:create").Register("observe", observe); err != nil {
				t.Fatal(err)
			}
			if err := f.db.Callback().Update().Before("gorm:update").Register("observe", observe); err != nil {
				t.Fatal(err)
			}
			if err := f.db.Callback().Delete().Before("gorm:delete").Register("observe", observe); err != nil {
				t.Fatal(err)
			}
			f.getError = apierrors.NewServiceUnavailable("temporary outage")
			result, err := f.run(context.Background())
			requireCollectorRetry(t, result, err)
			if dbCalls != 0 {
				t.Errorf("failed Kubernetes Get made %d database calls", dbCalls)
			}
			f.getError = nil
			reconcileObject(t, f.reconciler, f.object)
			if got := f.rowCount(t); got != 1 {
				t.Fatalf("retry persisted %d rows, want 1", got)
			}
		})
	}
}

func TestCollectorsRetryDatabaseFailures(t *testing.T) {
	for _, kind := range []string{"chaos", "schedule", "workflow"} {
		for _, operation := range []string{"query", "create", "update", "archive"} {
			t.Run(kind+"/"+operation, func(t *testing.T) {
				f := newRecoveryFixture(t, kind)
				seeded := operation == "update" || operation == "archive"
				if seeded {
					reconcileObject(t, f.reconciler, f.object)
					if operation == "archive" {
						if err := f.kube.Delete(context.Background(), f.object); err != nil {
							t.Fatal(err)
						}
					} else {
						obj := f.object.DeepCopyObject().(client.Object)
						if err := f.kube.Get(context.Background(), client.ObjectKeyFromObject(obj), obj); err != nil {
							t.Fatal(err)
						}
						obj.SetFinalizers([]string{"test"})
						if err := f.kube.Update(context.Background(), obj); err != nil {
							t.Fatal(err)
						}
						if err := f.kube.Delete(context.Background(), obj); err != nil {
							t.Fatal(err)
						}
					}
				}
				dbOperation := operation
				if operation == "archive" {
					dbOperation = "update"
				}
				restore := failCollectorDB(t, f.db, dbOperation, f.table)
				result, err := f.run(context.Background())
				requireCollectorRetry(t, result, err)
				restore()
				wantCount := int64(0)
				if seeded {
					wantCount = 1
				}
				if got := f.rowCount(t); got != wantCount {
					t.Fatalf("failed %s left %d rows, want %d", operation, got, wantCount)
				}
				var before struct {
					ID         uint
					Archived   bool
					FinishTime *time.Time
				}
				if seeded {
					if err := f.db.Table(f.table).Take(&before).Error; err != nil {
						t.Fatal(err)
					}
					if before.Archived || before.FinishTime != nil {
						t.Fatalf("failed mutation changed row: %#v", before)
					}
				}
				reconcileObject(t, f.reconciler, f.object)
				if got := f.rowCount(t); got != 1 {
					t.Fatalf("retry persisted %d rows, want 1", got)
				}
				var after struct {
					ID         uint
					Archived   bool
					FinishTime *time.Time
				}
				if err := f.db.Table(f.table).Take(&after).Error; err != nil {
					t.Fatal(err)
				}
				if seeded && after.ID != before.ID {
					t.Errorf("retry replaced row %d with %d", before.ID, after.ID)
				}
				if operation == "archive" && (!after.Archived || after.FinishTime == nil) {
					t.Errorf("retry failed to archive row: %#v", after)
				}
				if operation == "update" && kind != "workflow" && after.FinishTime == nil {
					t.Errorf("retry failed to persist deletion time")
				}
				if operation == "update" && kind == "workflow" {
					collector := f.reconciler.(*WorkflowCollector)
					row, err := collector.store.FindByUID(context.Background(), "resource-uid")
					if err != nil {
						t.Fatal(err)
					}
					stored, err := core.WorkflowEntity2WorkflowCR(row)
					if err != nil {
						t.Fatal(err)
					}
					if stored.DeletionTimestamp == nil || len(stored.Finalizers) != 1 || stored.Finalizers[0] != "test" {
						t.Errorf("retry did not update stored workflow metadata: %#v", stored.ObjectMeta)
					}
				}
			})
		}
	}
}

func TestCollectorsHonorCancellationAfterGet(t *testing.T) {
	for _, kind := range []string{"chaos", "schedule", "workflow", "event"} {
		t.Run(kind, func(t *testing.T) {
			f := newRecoveryFixture(t, kind)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			f.cancelAfterGet = cancel
			_, err := f.run(ctx)
			if !errors.Is(err, context.Canceled) {
				t.Errorf("canceled reconciliation returned %v", err)
			}
			if got := f.rowCount(t); got != 0 {
				t.Errorf("canceled reconciliation persisted %d rows", got)
			}
			f.cancelAfterGet = nil
			reconcileObject(t, f.reconciler, f.object)
			if got := f.rowCount(t); got != 1 {
				t.Errorf("fresh request persisted %d rows, want 1", got)
			}
		})
	}
}

func TestManagedCleanupRetriesPartialFailure(t *testing.T) {
	for _, failure := range []string{"list", "events", "experiment"} {
		t.Run(failure, func(t *testing.T) {
			f := newRecoveryFixture(t, "chaos")
			collector := f.reconciler.(*ChaosCollector)
			for _, uid := range []string{"first", "second"} {
				exp := &core.Experiment{ExperimentMeta: core.ExperimentMeta{Name: "resource", Namespace: "default", UID: uid, Archived: true}, Experiment: `{"apiVersion":"chaos-mesh.org/v1alpha1","kind":"PodChaos","metadata":{"labels":{"managed-by":"workflow"}}}`}
				if err := collector.store.Save(context.Background(), exp); err != nil {
					t.Fatal(err)
				}
				if err := collector.eventStore.Create(context.Background(), &core.Event{ObjectID: uid}); err != nil {
					t.Fatal(err)
				}
			}
			if err := collector.eventStore.Create(context.Background(), &core.Event{ObjectID: "unrelated"}); err != nil {
				t.Fatal(err)
			}
			if err := f.kube.Delete(context.Background(), f.object); err != nil {
				t.Fatal(err)
			}
			restore := func() {}
			if failure == "list" {
				restore = failCollectorDB(t, f.db, "query", "experiments")
			} else {
				table := "events"
				if failure == "experiment" {
					table = "experiments"
				}
				enabled := true
				deletes := 0
				if err := f.db.Callback().Delete().Before("gorm:delete").Register("partial_failure", func(tx *gorm.DB) {
					if enabled && tx.Statement.Table == table {
						deletes++
						if deletes == 2 {
							tx.AddError(errors.New("second delete failed"))
						}
					}
				}); err != nil {
					t.Fatal(err)
				}
				restore = func() { enabled = false }
			}
			result, err := f.run(context.Background())
			requireCollectorRetry(t, result, err)
			restore()
			if failure != "list" {
				if got := f.rowCount(t); got != 1 {
					t.Errorf("partial cleanup retained %d retry anchors, want 1", got)
				}
				remaining, err := collector.store.ListArchivedManagedByNamespaceName(context.Background(), "default", "resource")
				if err != nil || len(remaining) != 1 {
					t.Errorf("failed cleanup lost retry anchor, rows=%#v err=%v", remaining, err)
				} else {
					events, err := collector.eventStore.ListByUID(context.Background(), remaining[0].UID)
					if err != nil {
						t.Fatal(err)
					}
					want := 0
					if failure == "events" {
						want = 1
					}
					if len(events) != want {
						t.Errorf("remaining events %d, want %d", len(events), want)
					}
				}
			}
			reconcileObject(t, f.reconciler, f.object)
			reconcileObject(t, f.reconciler, f.object)
			if got := f.rowCount(t); got != 0 {
				t.Errorf("retry left %d managed rows", got)
			}
			events, err := collector.eventStore.List(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if len(events) != 1 || events[0].ObjectID != "unrelated" {
				t.Errorf("retry left events %#v", events)
			}
		})
	}
}

func TestEventCollectorRetriesTargetAndCreateFailures(t *testing.T) {
	for _, kind := range []string{v1alpha1.KindPodChaos, v1alpha1.KindSchedule, v1alpha1.KindWorkflow, v1alpha1.KindWorkflowNode} {
		for _, failure := range []string{"target", "create", "missing"} {
			t.Run(kind+"/"+failure, func(t *testing.T) {
				f := newRecoveryFixture(t, "event")
				ev := f.object.(*corev1.Event)
				ev.InvolvedObject.Kind = kind
				if err := f.kube.Update(context.Background(), ev); err != nil {
					t.Fatal(err)
				}
				collector := f.reconciler.(*EventCollector)
				unavailable := failure == "target" || failure == "missing"
				collector.Client = interceptor.NewClient(f.kube.(client.WithWatch), interceptor.Funcs{Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
					if key.Name == "target" {
						if unavailable {
							if failure == "missing" {
								return apierrors.NewNotFound(v1alpha1.Resource("target"), key.Name)
							}
							return apierrors.NewServiceUnavailable("target unavailable")
						}
						obj.SetName(key.Name)
						obj.SetNamespace(key.Namespace)
						obj.SetUID("target-uid")
						return nil
					}
					return c.Get(ctx, key, obj, opts...)
				}})
				restore := func() {}
				if failure == "create" {
					restore = failCollectorDB(t, f.db, "create", "events")
				}
				result, err := f.run(context.Background())
				if failure == "missing" {
					if err != nil || result != (ctrl.Result{}) {
						t.Errorf("missing target result %#v, %v", result, err)
					}
				} else {
					requireCollectorRetry(t, result, err)
				}
				restore()
				if got := f.rowCount(t); got != 0 {
					t.Errorf("failed event collection persisted %d rows", got)
				}
				if failure == "missing" {
					return
				}
				unavailable = false
				reconcileObject(t, collector, ev)
				rows, err := collector.eventStore.List(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				if len(rows) != 1 || rows[0].ObjectID != "target-uid" || rows[0].Kind != kind || rows[0].Reason != "Applied" {
					t.Errorf("retry events %#v", rows)
				}
			})
		}
	}
}

func TestCollectorsPropagateCancellationDuringPersistence(t *testing.T) {
	for _, kind := range []string{"chaos", "schedule", "workflow", "event"} {
		t.Run(kind, func(t *testing.T) {
			f := newRecoveryFixture(t, kind)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if kind == "event" {
				if err := f.db.Callback().Create().Before("gorm:create").Register("cancel_request", func(*gorm.DB) { cancel() }); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := f.db.Callback().Query().After("gorm:query").Register("cancel_request", func(*gorm.DB) { cancel() }); err != nil {
					t.Fatal(err)
				}
			}
			_, err := f.run(ctx)
			if !errors.Is(err, context.Canceled) {
				t.Errorf("canceled persistence returned %v", err)
			}
			if got := f.rowCount(t); got != 0 {
				t.Errorf("canceled persistence saved %d rows", got)
			}
			reconcileObject(t, f.reconciler, f.object)
			if got := f.rowCount(t); got != 1 {
				t.Errorf("fresh request persisted %d rows, want 1", got)
			}
		})
	}
}

func TestCollectorsHonorCancellationBeforeGet(t *testing.T) {
	for _, kind := range []string{"chaos", "schedule", "workflow", "event"} {
		t.Run(kind, func(t *testing.T) {
			f := newRecoveryFixture(t, kind)
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			result, err := f.run(ctx)
			if !errors.Is(err, context.Canceled) || result != (ctrl.Result{}) {
				t.Errorf("canceled reconciliation = %#v, %v", result, err)
			}
			if f.getCalls != 0 {
				t.Errorf("canceled reconciliation made %d Kubernetes calls", f.getCalls)
			}
			if got := f.rowCount(t); got != 0 {
				t.Errorf("canceled reconciliation saved %d rows", got)
			}
		})
	}
}
