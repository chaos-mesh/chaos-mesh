// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package cron

import (
	"context"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/schedule/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/types"
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

var t = true

func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	schedule := &v1alpha1.Schedule{}
	err := r.Get(ctx, req.NamespacedName, schedule)
	if err != nil {
		r.Log.Error(err, "unable to get chaos")
		return ctrl.Result{}, nil
	}

	if schedule.IsPaused() {
		r.Log.Info("not starting chaos as it is paused")
		return ctrl.Result{}, nil
	}

	now := time.Now()
	shouldSpawn := false
	r.Log.Info("calculate schedule time", "schedule", schedule.Spec.Schedule, "lastScheduleTime", schedule.Status.LastScheduleTime, "now", now)
	missedRun, nextRun, err := getRecentUnmetScheduleTime(schedule, now)
	if err != nil {
		r.Recorder.Event(schedule, recorder.Failed{
			Activity: "get run time",
			Err:      err.Error(),
		})
		return ctrl.Result{}, nil
	}
	if missedRun == nil {
		r.Log.Info("requeue later to reconcile the schedule", "requeue-after", nextRun.Sub(now))
		return ctrl.Result{RequeueAfter: nextRun.Sub(now)}, nil
	}

	if schedule.Spec.StartingDeadlineSeconds != nil {
		if missedRun.Add(time.Second * time.Duration(*schedule.Spec.StartingDeadlineSeconds)).Before(now) {
			r.Recorder.Event(schedule, recorder.MissedSchedule{
				MissedRun: *missedRun,
			})
			return ctrl.Result{}, nil
		}
	}

	r.Log.Info("schedule to spawn new chaos", "missedRun", missedRun, "nextRun", nextRun)
	shouldSpawn = true

	if shouldSpawn && schedule.Spec.ConcurrencyPolicy.IsForbid() {
		list, err := r.ActiveLister.ListActiveJobs(ctx, schedule)
		if err != nil {
			r.Recorder.Event(schedule, recorder.Failed{
				Activity: "list active jobs",
				Err:      err.Error(),
			})
			return ctrl.Result{}, nil
		}

		items := reflect.ValueOf(list).Elem().FieldByName("Items")
		for i := 0; i < items.Len(); i++ {
			if schedule.Spec.Type != v1alpha1.ScheduleTypeWorkflow {
				item := items.Index(i).Addr().Interface().(v1alpha1.InnerObject)
				if !controller.IsChaosFinished(item, now) {
					shouldSpawn = false
					r.Recorder.Event(schedule, recorder.ScheduleForbid{
						RunningName: item.GetObjectMeta().Name,
					})
					r.Log.Info("forbid to spawn new chaos", "running", item.GetChaos().Name)
					break
				}
			} else {
				workflow := items.Index(i).Addr().Interface().(*v1alpha1.Workflow)
				if !controllers.WorkflowConditionEqualsTo(workflow.Status, v1alpha1.WorkflowConditionAccomplished, corev1.ConditionTrue) {
					shouldSpawn = false
					r.Recorder.Event(schedule, recorder.ScheduleForbid{
						RunningName: workflow.GetObjectMeta().Name,
					})
					r.Log.Info("forbid to spawn new workflow", "running", workflow.GetChaos().Name)
					break
				}
			}
		}
	}

	if shouldSpawn {
		newObj, meta, err := schedule.Spec.ScheduleItem.SpawnNewObject(schedule.Spec.Type)
		if err != nil {
			r.Recorder.Event(schedule, recorder.Failed{
				Activity: "generate new object",
				Err:      err.Error(),
			})
			return ctrl.Result{}, nil
		}

		meta.SetOwnerReferences([]metav1.OwnerReference{
			{
				APIVersion:         schedule.APIVersion,
				Kind:               schedule.Kind,
				Name:               schedule.Name,
				UID:                schedule.UID,
				Controller:         &t,
				BlockOwnerDeletion: &t,
			},
		})
		meta.SetLabels(map[string]string{
			"managed-by": schedule.Name,
		})
		meta.SetNamespace(schedule.Namespace)
		meta.SetName(names.SimpleNameGenerator.GenerateName(schedule.Name + "-"))

		err = r.Create(ctx, newObj)
		if err != nil {
			r.Recorder.Event(schedule, recorder.Failed{
				Activity: "create new object",
				Err:      err.Error(),
			})
			r.Log.Error(err, "fail to create new object", "obj", newObj)
			return ctrl.Result{}, nil
		}
		r.Recorder.Event(schedule, recorder.ScheduleSpawn{
			Name: meta.GetName(),
		})
		r.Log.Info("create new object", "namespace", meta.GetNamespace(), "name", meta.GetName())

		lastScheduleTime := now
		updateError := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			r.Log.Info("updating lastScheduleTime", "time", lastScheduleTime)
			schedule = schedule.DeepCopyObject().(*v1alpha1.Schedule)

			if err := r.Client.Get(ctx, req.NamespacedName, schedule); err != nil {
				r.Log.Error(err, "unable to get schedule")
				return err
			}

			schedule.Status.LastScheduleTime.Time = lastScheduleTime
			return r.Client.Update(ctx, schedule)
		})
		if updateError != nil {
			r.Log.Error(updateError, "fail to update")
			r.Recorder.Event(schedule, recorder.Failed{
				Activity: "update lastScheduleTime",
				Err:      updateError.Error(),
			})
			return ctrl.Result{}, nil
		}

		r.Recorder.Event(schedule, recorder.Updated{
			Field: "lastScheduleTime",
		})
	}

	return ctrl.Result{}, nil
}

func NewController(mgr ctrl.Manager, client client.Client, log logr.Logger, lister *utils.ActiveLister, recorderBuilder *recorder.RecorderBuilder) (types.Controller, error) {
	builder.Default(mgr).
		For(&v1alpha1.Schedule{}).
		Named("schedule-cron").
		Complete(&Reconciler{
			client,
			log.WithName("schedule-cron"),
			lister,
			recorderBuilder.Build("schedule-cron"),
		})
	return "schedule-cron", nil
}
