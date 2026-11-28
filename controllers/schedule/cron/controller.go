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

package cron

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/schedule/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/builder"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/controller"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/controllers"
)

type Reconciler struct {
	client.Client
	Log          logr.Logger
	ActiveLister *utils.ActiveLister

	Recorder recorder.ChaosRecorder
}

const runStartedAtAnnotation = "chaos-mesh.org/schedule-run-started-at"

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.reconcile(ctx, req, time.Now())
}

func (r *Reconciler) reconcile(ctx context.Context, req ctrl.Request, now time.Time) (ctrl.Result, error) {
	schedule := &v1alpha1.Schedule{}
	if err := r.Get(ctx, req.NamespacedName, schedule); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if schedule.IsPaused() || schedule.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	runKey := client.ObjectKey{Namespace: schedule.Namespace, Name: nextRunName(schedule)}
	started, err := r.startedRun(ctx, schedule, runKey)
	if err != nil {
		return ctrl.Result{}, err
	}
	if !started.IsZero() {
		return ctrl.Result{}, r.recordStart(ctx, schedule, started)
	}

	r.Log.Info("calculate schedule time", "schedule", schedule.Spec.Schedule, "lastScheduleTime", schedule.Status.LastScheduleTime, "now", now)
	missedRun, nextRun, err := getRecentUnmetScheduleTime(schedule, now)
	if err != nil {
		r.Recorder.Event(schedule, recorder.Failed{Activity: "get run time", Err: err.Error()})
		return ctrl.Result{}, nil
	}
	if missedRun == nil {
		return ctrl.Result{RequeueAfter: nextRun.Sub(now)}, nil
	}

	if deadline := schedule.Spec.StartingDeadlineSeconds; deadline != nil &&
		missedRun.Add(time.Second*time.Duration(*deadline)).Before(now) {
		r.Recorder.Event(schedule, recorder.MissedSchedule{MissedRun: *missedRun})
		return ctrl.Result{}, nil
	}

	if schedule.Spec.ConcurrencyPolicy.IsForbid() {
		list, err := r.ActiveLister.ListActiveJobs(ctx, schedule)
		if err != nil {
			r.Recorder.Event(schedule, recorder.Failed{Activity: "list active jobs", Err: err.Error()})
			return ctrl.Result{}, err
		}

		for _, item := range list.GetItems() {
			var finished bool
			if workflow, ok := item.(*v1alpha1.Workflow); ok {
				finished = controllers.WorkflowConditionEqualsTo(workflow.Status, v1alpha1.WorkflowConditionAccomplished, corev1.ConditionTrue)
			} else {
				finished = controller.IsChaosFinished(item.(v1alpha1.InnerObject), now)
			}
			if !finished {
				r.Recorder.Event(schedule, recorder.ScheduleForbid{RunningName: item.GetName()})
				return ctrl.Result{}, nil
			}
		}
	}

	newObj, err := schedule.Spec.ScheduleItem.SpawnNewObject(schedule.Spec.Type)
	if err != nil {
		r.Recorder.Event(schedule, recorder.Failed{Activity: "generate new object", Err: err.Error()})
		return ctrl.Result{}, nil
	}
	owner := metav1.NewControllerRef(schedule, v1alpha1.SchemeGroupVersion.WithKind(v1alpha1.KindSchedule))
	started = now.UTC().Truncate(time.Second)
	newObj.SetOwnerReferences([]metav1.OwnerReference{*owner})
	newObj.SetLabels(map[string]string{v1alpha1.LabelManagedBy: schedule.Name})
	newObj.SetAnnotations(map[string]string{runStartedAtAnnotation: started.Format(time.RFC3339)})
	newObj.SetNamespace(runKey.Namespace)
	newObj.SetName(runKey.Name)

	if err = r.Create(ctx, newObj); err != nil {
		r.Recorder.Event(schedule, recorder.Failed{Activity: "create new object", Err: err.Error()})
		return ctrl.Result{}, err
	}
	r.Recorder.Event(schedule, recorder.ScheduleSpawn{Name: newObj.GetName()})
	return ctrl.Result{}, r.recordStart(ctx, schedule, started)
}

// nextRunName identifies the successor to the last recorded run. It remains
// stable if a retry crosses a cron tick or the starting deadline moves.
func nextRunName(schedule *v1alpha1.Schedule) string {
	key := fmt.Sprintf("%s/%d", schedule.UID, schedule.Status.LastScheduleTime.Unix())
	hash := sha256.Sum256([]byte(key))
	suffix := fmt.Sprintf("-%x", hash[:8])
	base := schedule.Name[:min(len(schedule.Name), validation.DNS1123LabelMaxLength-len(suffix))]
	return strings.TrimRight(base, "-.") + suffix
}

func (r *Reconciler) startedRun(ctx context.Context, schedule *v1alpha1.Schedule, key client.ObjectKey) (time.Time, error) {
	kind, ok := v1alpha1.AllScheduleItemKinds()[string(schedule.Spec.Type)]
	if !ok {
		return time.Time{}, nil
	}
	run := kind.SpawnObject()
	if err := r.Get(ctx, key, run); err != nil {
		return time.Time{}, client.IgnoreNotFound(err)
	}
	owner := metav1.GetControllerOf(run)
	if owner == nil || owner.UID != schedule.UID || owner.Name != schedule.Name || owner.Kind != v1alpha1.KindSchedule {
		return time.Time{}, fmt.Errorf("scheduled run %s belongs to a different controller", key)
	}
	started, err := time.Parse(time.RFC3339, run.GetAnnotations()[runStartedAtAnnotation])
	if err != nil {
		return time.Time{}, fmt.Errorf("read start time of scheduled run %s: %w", key, err)
	}
	if !started.After(schedule.Status.LastScheduleTime.Time) {
		return time.Time{}, fmt.Errorf("start time of scheduled run %s does not advance the schedule", key)
	}
	return started, nil
}

func (r *Reconciler) recordStart(ctx context.Context, schedule *v1alpha1.Schedule, started time.Time) error {
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		latest := &v1alpha1.Schedule{}
		if err := r.Get(ctx, client.ObjectKeyFromObject(schedule), latest); err != nil {
			return client.IgnoreNotFound(err)
		}
		if latest.UID != schedule.UID || !latest.Status.LastScheduleTime.Equal(&schedule.Status.LastScheduleTime) {
			return nil
		}

		latest.Status.LastScheduleTime = metav1.NewTime(started)
		if err := r.Update(ctx, latest); err != nil {
			return err
		}
		r.Recorder.Event(latest, recorder.Updated{Field: "lastScheduleTime"})
		return nil
	})
	if err != nil {
		r.Recorder.Event(schedule, recorder.Failed{Activity: "update lastScheduleTime", Err: err.Error()})
	}
	return err
}

const controllerName = "schedule-cron"

func Bootstrap(mgr ctrl.Manager, client client.Client, log logr.Logger, lister *utils.ActiveLister, recorderBuilder *recorder.Builder) error {
	if !config.ShouldSpawnController(controllerName) {
		return nil
	}

	return builder.Default(mgr).
		For(&v1alpha1.Schedule{}).
		Named(controllerName).
		Complete(&Reconciler{
			client,
			log.WithName(controllerName),
			lister,
			recorderBuilder.Build(controllerName),
		})
}
