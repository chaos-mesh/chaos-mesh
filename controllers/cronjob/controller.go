// Copyright 2020 Chaos Mesh Authors.
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

package cronjob

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/pingcap/errors"
	"github.com/robfig/cron/v3"
	apiError "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/storage/names"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
)

// Reconciler for the twophase reconciler
type Reconciler struct {
	ctx.Context
	shouldUpdate bool

	v1alpha1.CronJob
}

// Reconcile is twophase reconcile implement
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	now := time.Now()
	r.Log.Info("Reconciling a cron CronJob", "name", req.Name, "namespace", req.Namespace, "time", now)

	job, err := r.syncActiveJob(ctx)
	if err != nil {
		r.Log.Error(err, "fail to sync active job")

		return r.ret(ctx, err)
	}

	nextRecover := r.CronJob.GetNextRecover()
	nextStart := r.CronJob.GetNextStart()
	r.Log.Info("reconcile", "nextRecover", nextRecover, "nextStart", nextStart)

	if !nextRecover.IsZero() && nextRecover.Before(now) && job != nil {
		r.Log.Info("delete job", "name", job.GetName())
		// Need to remove the job
		err := r.Delete(ctx, job)
		if err != nil {
			r.Log.Error(err, "fail to remove object", "namespace", req.Namespace, "name", job.GetName())

			return r.ret(ctx, err)
		}
		r.shouldUpdate = true
		r.CronJob.SetActiveJob(nil)

		return r.ret(ctx, nil)
	}

	if nextStart.Before(now) {
		if job == nil && !r.CronJob.IsPaused() {
			// Need to create the job
			newJob := r.CronJob.IntoJobWithoutName()
			name := names.SimpleNameGenerator.GenerateName(req.Name + "-")
			newJob.SetName(name)
			r.Log.Info("create new job", "name", name)

			err := r.Create(ctx, newJob)
			if err != nil {
				r.Log.Error(err, "fail to create new job")

				return r.ret(ctx, err)
			}
			r.shouldUpdate = true
			r.CronJob.SetActiveJob(&types.NamespacedName{
				Namespace: req.Namespace,
				Name:      name,
			})
		} else {
			r.Log.Info("skip this iteration as the job is still running or the cronjob is paused")
		}

		scheduler := r.CronJob.GetScheduler()
		if scheduler == nil {
			return r.ret(ctx, errors.Errorf("misdefined scheduler"))
		}
		nextStart, err := nextTime(*scheduler, now)
		if err != nil {
			return r.ret(ctx, err)
		}
		r.CronJob.SetNextStart(*nextStart)

		duration, err := r.CronJob.GetDuration()
		if err != nil {
			return r.ret(ctx, err)
		}
		r.CronJob.SetNextRecover(now.Add(*duration))
	}

	if job != nil {
		r.Log.Info("update job")
		// Need to synchronize the configuration
		shouldUpdate := r.CronJob.UpdateJob(job)

		if shouldUpdate {
			err := r.Update(ctx, job)
			if err != nil {
				r.Log.Error(err, "fail to update new job")

				return r.ret(ctx, err)
			}
		}
	}

	return r.ret(ctx, nil)
}

func nextTime(spec v1alpha1.SchedulerSpec, now time.Time) (*time.Time, error) {
	scheduler, err := cron.ParseStandard(spec.Cron)
	if err != nil {
		return nil, fmt.Errorf("fail to parse runner rule %s, %v", spec.Cron, err)
	}

	next := scheduler.Next(now)
	return &next, nil
}

func (r *Reconciler) ret(ctx context.Context, executionError error) (ctrl.Result, error) {
	var err error

	r.Log.Info("return a reconcile request")

	r.CronJob.GetStatus().FailedMessage = ""
	if executionError != nil {
		r.CronJob.GetStatus().FailedMessage = executionError.Error()
	}

	if r.shouldUpdate {
		err = r.Context.Client.Update(ctx, r.CronJob)
		if err != nil {
			r.Log.Error(err, "fail to update cronjob status")
		}
	}

	requeueAfter := time.Duration(0)
	if executionError == nil && err == nil {
		requeueAfter, err = calcRequeueAfterTime(r.CronJob, time.Now())
		if err != nil {
			r.Log.Error(err, "fail to calc requeue time")
		}
	}

	return ctrl.Result{RequeueAfter: requeueAfter}, nil
}

func (r *Reconciler) syncActiveJob(ctx context.Context) (v1alpha1.Job, error) {
	activeJob := r.CronJob.GetActiveJob()

	if activeJob == nil {
		return nil, nil
	}

	r.Log.Info("try to lookup active job", "job", activeJob)
	obj := r.CronJob.GetJobObject()
	err := r.Context.Client.Get(ctx, *activeJob, obj)
	if err != nil {
		if apiError.IsNotFound(err) {
			r.shouldUpdate = true
			r.CronJob.SetActiveJob(nil)
			return nil, nil
		}

		return nil, err
	}

	// synchronise status from job to the cronjob
	newStatus := obj.GetStatus()
	cronStatus := r.GetStatus()
	cronStatus.FailedMessage = newStatus.FailedMessage
	cronStatus.Experiment = *newStatus.Experiment.DeepCopy()
	return obj, nil
}

func calcRequeueAfterTime(chaos v1alpha1.InnerSchedulerObject, now time.Time) (time.Duration, error) {
	requeueAfter := time.Duration(math.MaxInt64)
	// requeueAfter = min(filter([nextRecoverAfter, nextStartAfter], >0))
	nextRecoverAfter := chaos.GetNextRecover().Sub(now)
	nextStartAfter := chaos.GetNextStart().Sub(now)
	if nextRecoverAfter > 0 && requeueAfter > nextRecoverAfter {
		requeueAfter = nextRecoverAfter
	}
	if nextStartAfter > 0 && requeueAfter > nextStartAfter {
		requeueAfter = nextStartAfter
	}

	var err error
	if requeueAfter == math.MaxInt64 {
		err = errors.Errorf("unexpected behavior, now is greater than nextRecover and nextStart")
	}

	return requeueAfter, err
}
