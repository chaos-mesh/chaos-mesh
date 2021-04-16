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
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/types"
	"github.com/go-logr/logr"
	"github.com/robfig/cron"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type Reconciler struct {
	client.Client
	Log logr.Logger

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

	cron, err := cron.Parse(schedule.Spec.Schedule)
	if err != nil {
		r.Recorder.Eventf(schedule, "Warning", "Failed", "Failed to parse schedule: %s", err.Error())
		return ctrl.Result{}, nil
	}

	now := time.Now()
	shouldSpawn := false
	lastScheduleTime := schedule.Status.LastScheduleTime.Time
	expectScheduleTime := cron.Next(lastScheduleTime)
	if lastScheduleTime.IsZero() {
		shouldSpawn = true
	} else if expectScheduleTime.Before(now) {
		shouldSpawn = true
	} else if expectScheduleTime.After(now) {
		shouldSpawn = false
		r.Log.Info("requeue later to reconcile the schedule", "requeue-after", expectScheduleTime.Sub(now))
		return ctrl.Result{RequeueAfter: expectScheduleTime.Sub(now)}, nil
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
		meta.SetLabels(map[string]string {
			"managed-by": schedule.Name,
		})
		meta.SetNamespace(schedule.Namespace)
		meta.SetName(names.SimpleNameGenerator.GenerateName(schedule.Name + "-"))

		err = r.Create(ctx, newObj)
		if err != nil {
			r.Recorder.Eventf(schedule, "Warning", "Failed", "Failed to create new object: %s", err.Error())
			return ctrl.Result{}, nil
		}

		lastScheduleTime = now
		updateError := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			r.Log.Info("updating lastScheduleTime", "time", lastScheduleTime)
			schedule = schedule.DeepCopyObject().(*v1alpha1.Schedule)

			if err := r.Client.Get(context.TODO(), req.NamespacedName, schedule); err != nil {
				r.Log.Error(err, "unable to get schedule")
				return err
			}

			schedule.Status.LastScheduleTime.Time = lastScheduleTime
			return r.Client.Update(context.TODO(), schedule)
		})
		if updateError != nil {
			r.Log.Error(updateError, "fail to update")
			r.Recorder.Eventf(schedule, "Normal", "Failed", "Failed to update lastScheduleTime: %s", updateError.Error())
			return ctrl.Result{}, nil
		}

		r.Recorder.Event(schedule, "Normal", "Updated", "Successfully update lastScheduleTime of resource")
	}

	return ctrl.Result{}, nil
}

func NewController(mgr ctrl.Manager, client client.Client, log logr.Logger ) (types.Controller, error) {
	ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Schedule{}).
		Named("schedule-cron").
		Complete(&Reconciler{
			client,
			log.WithName("schedule-cron"),
			mgr.GetEventRecorderFor("schedule-cron"),
	})
	return "schedule", nil
}
