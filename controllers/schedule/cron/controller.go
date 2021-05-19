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
	"go.uber.org/fx"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/schedule/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/controller"
)

type Reconciler struct {
	client.Client
	Log          logr.Logger
	ActiveLister *utils.ActiveLister

	Recorder record.EventRecorder
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

	now := time.Now()
	shouldSpawn := false
	r.Log.Info("calculate schedule time", "schedule", schedule.Spec.Schedule, "lastScheduleTime", schedule.Status.LastScheduleTime, "now", now)
	missedRun, nextRun, err := getRecentUnmetScheduleTime(schedule, now)
	if err != nil {
		r.Recorder.Eventf(schedule, "Warning", "Failed", "Failed to get run time: %s", err.Error())
		return ctrl.Result{}, nil
	}
	if missedRun == nil {
		r.Log.Info("requeue later to reconcile the schedule", "requeue-after", nextRun.Sub(now))
		return ctrl.Result{RequeueAfter: nextRun.Sub(now)}, nil
	}

	if schedule.Spec.StartingDeadlineSeconds != nil {
		if missedRun.Add(time.Second * time.Duration(*schedule.Spec.StartingDeadlineSeconds)).Before(now) {
			r.Recorder.Eventf(schedule, "Warning", "MissSchedule", "Missed scheduled time to start a job: %s", missedRun.Format(time.RFC1123Z))
			return ctrl.Result{}, nil
		}
	}

	r.Log.Info("schedule to spawn new chaos", "missedRun", missedRun, "nextRun", nextRun)
	shouldSpawn = true

	if shouldSpawn && schedule.Spec.ConcurrencyPolicy.IsForbid() {
		list, err := r.ActiveLister.ListActiveJobs(ctx, schedule)
		if err != nil {
			r.Recorder.Eventf(schedule, "Warning", "Failed", "Failed to list active jobs: %s", err.Error())
			return ctrl.Result{}, nil
		}

		items := reflect.ValueOf(list).Elem().FieldByName("Items")
		for i := 0; i < items.Len(); i++ {
			item := items.Index(i).Addr().Interface().(v1alpha1.InnerObject)
			if !controller.IsChaosFinished(item, now) {
				shouldSpawn = false
				r.Recorder.Eventf(schedule, "Warning", "Forbid", "Forbid spawning new job because: %s is still running", item.GetObjectMeta().Name)
				r.Log.Info("forbid to spawn new chaos", "running", item.GetChaos().Name)
				break
			}
		}
	}

	if shouldSpawn {
		r.Recorder.Event(schedule, "Info", "Spawn", "Spawn new chaos")

		newObj, meta, err := schedule.Spec.EmbedChaos.SpawnNewObject(schedule.Spec.Type)
		if err != nil {
			r.Recorder.Eventf(schedule, "Warning", "Failed", "Failed to generate new object: %s", err.Error())
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
			r.Recorder.Eventf(schedule, "Warning", "Failed", "Failed to create new object: %s", err.Error())
			r.Log.Error(err, "fail to create new object", "obj", newObj)
			return ctrl.Result{}, nil
		}
		r.Recorder.Eventf(schedule, "Normal", "Created", "Create new object: %s", meta.GetName())
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
			r.Recorder.Eventf(schedule, "Warning", "Failed", "Failed to update lastScheduleTime: %s", updateError.Error())
			return ctrl.Result{}, nil
		}

		r.Recorder.Event(schedule, "Normal", "Updated", "Successfully update lastScheduleTime of resource")

		// TODO: make the interval and total time configurable
		// The following codes ensure the Schedule in cache has the latest lastScheduleTime
		ensureLatestError := wait.Poll(100*time.Millisecond, 2*time.Second, func() (bool, error) {
			schedule = schedule.DeepCopyObject().(*v1alpha1.Schedule)

			if err := r.Client.Get(ctx, req.NamespacedName, schedule); err != nil {
				r.Log.Error(err, "unable to get schedule")
				return false, err
			}

			return schedule.Status.LastScheduleTime.Time == lastScheduleTime, nil
		})
		if ensureLatestError != nil {
			r.Log.Error(ensureLatestError, "Fail to ensure that the resource in cache has the latest lastScheduleTime")
			return ctrl.Result{}, nil
		}
	}

	return ctrl.Result{}, nil
}

type Objs struct {
	fx.In

	Objs []types.Object `group:"objs"`
}

func NewController(mgr ctrl.Manager, client client.Client, log logr.Logger, objs Objs, lister *utils.ActiveLister) (types.Controller, error) {
	ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Schedule{}).
		Named("schedule-cron").
		Complete(&Reconciler{
			client,
			log.WithName("schedule-cron"),
			lister,
			mgr.GetEventRecorderFor("schedule-cron"),
		})
	return "schedule-cron", nil
}
