// Copyright 2019 Chaos Mesh Authors.
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

package twophase

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/client-go/util/retry"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/reconciler"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const emptyString = ""

// Reconciler for the twophase reconciler
type Reconciler struct {
	reconciler.InnerReconciler
	client.Client
	client.Reader
	Log logr.Logger
}

// NewReconciler would create reconciler for twophase controller
func NewReconciler(r reconciler.InnerReconciler, client client.Client, reader client.Reader, log logr.Logger) *Reconciler {
	return &Reconciler{
		InnerReconciler: r,
		Client:          client,
		Reader:          reader,
		Log:             log,
	}
}

// Reconcile is twophase reconcile implement
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var err error
	now := time.Now()

	r.Log.Info("Reconciling a two phase chaos", "name", req.Name, "namespace", req.Namespace)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_chaos := r.Object()
	if err = r.Client.Get(ctx, req.NamespacedName, _chaos); err != nil {
		r.Log.Error(err, "unable to get chaos")
		return ctrl.Result{}, err
	}
	chaos := _chaos.(v1alpha1.InnerSchedulerObject)

	duration, err := chaos.GetDuration()
	if err != nil {
		r.Log.Error(err, "failed to get chaos duration")
		return ctrl.Result{}, err
	}

	scheduler := chaos.GetScheduler()
	if scheduler == nil {
		r.Log.Info("Scheduler should be defined currently")
		return ctrl.Result{}, fmt.Errorf("misdefined scheduler")
	}

	if duration == nil {
		zero := 0 * time.Second
		duration = &zero
	}

	status := chaos.GetStatus()
	result := ctrl.Result{}

	if chaos.IsDeleted() {
		// This chaos was deleted
		r.Log.Info("Removing self")
		err = r.Recover(ctx, req, chaos)
		if err != nil {
			r.Log.Error(err, "failed to recover chaos")
			updateFailedMessage(ctx, r, chaos, err.Error())
			return ctrl.Result{Requeue: true}, err
		}

		status.Experiment.Phase = v1alpha1.ExperimentPhaseFinished
		status.FailedMessage = emptyString
	} else if chaos.IsPaused() {
		if status.Experiment.Phase == v1alpha1.ExperimentPhaseRunning {
			r.Log.Info("Pausing")

			err = r.Recover(ctx, req, chaos)
			if err != nil {
				r.Log.Error(err, "failed to pause chaos")
				updateFailedMessage(ctx, r, chaos, err.Error())
				return ctrl.Result{Requeue: true}, err
			}

			now := time.Now()
			status.Experiment.EndTime = &metav1.Time{
				Time: now,
			}
			if status.Experiment.StartTime != nil {
				status.Experiment.Duration = now.Sub(status.Experiment.StartTime.Time).String()
			}
		}
		status.Experiment.Phase = v1alpha1.ExperimentPhasePaused
		status.FailedMessage = emptyString
	} else if !chaos.GetNextRecover().IsZero() && chaos.GetNextRecover().Before(now) {
		// Start recover
		r.Log.Info("Recovering")

		// Don't need to recover again if chaos was paused before
		if status.Experiment.Phase != v1alpha1.ExperimentPhasePaused {
			if err = r.Recover(ctx, req, chaos); err != nil {
				r.Log.Error(err, "failed to recover chaos")
				updateFailedMessage(ctx, r, chaos, err.Error())
				return ctrl.Result{Requeue: true}, err
			}
		}

		chaos.SetNextRecover(time.Time{})

		status.Experiment.EndTime = &metav1.Time{
			Time: time.Now(),
		}
		status.Experiment.Phase = v1alpha1.ExperimentPhaseWaiting
		status.FailedMessage = emptyString
	} else if (status.Experiment.Phase == v1alpha1.ExperimentPhaseFailed ||
		status.Experiment.Phase == v1alpha1.ExperimentPhasePaused) &&
		!chaos.GetNextRecover().IsZero() && chaos.GetNextRecover().After(now) {
		// Only resume/retry chaos in the case when current round is not finished,
		// which means the current time is before recover time. Otherwise we
		// don't resume the chaos and just wait for the start of next round.

		r.Log.Info("Resuming/Retrying")

		dur := chaos.GetNextRecover().Sub(now)
		if err = applyAction(ctx, r, req, dur, chaos); err != nil {
			updateFailedMessage(ctx, r, chaos, err.Error())
			return ctrl.Result{Requeue: true}, err
		}

		status.FailedMessage = emptyString
	} else {
		if chaos.GetNextStart() == (time.Time{}) {
			// this is the first time start, calculate the next start time
			nextStart, err := utils.NextTime(*chaos.GetScheduler(), now, true)
			if err != nil {
				r.Log.Error(err, "failed to calculate the start time")
				updateFailedMessage(ctx, r, chaos, err.Error())
				return ctrl.Result{}, err
			}
			chaos.SetNextStart(*nextStart)
			r.Log.Info("First time start, set the next start time", "next start time", *nextStart)
			status.Experiment.StartTime = &metav1.Time{Time: now}
		}

		// start the chaos before or equal now
		if !chaos.GetNextStart().After(now) {
			r.Log.Info("Starting")
			if err = applyAction(ctx, r, req, *duration, chaos); err != nil {
				updateFailedMessage(ctx, r, chaos, err.Error())
				return ctrl.Result{Requeue: true}, err
			}

			nextStart, err := utils.NextTime(*chaos.GetScheduler(), status.Experiment.StartTime.Time, false)
			if err != nil {
				r.Log.Error(err, "failed to get the next start time")
				return ctrl.Result{}, err
			}
			nextRecover := status.Experiment.StartTime.Time.Add(*duration)
			if nextStart.Before(nextRecover) {
				r.Log.Info("nextRecover shouldn't be later than nextStart, update the nextStart time")
				*nextStart = nextRecover.Add(time.Second)
			}

			chaos.SetNextStart(*nextStart)
			chaos.SetNextRecover(nextRecover)
			status.FailedMessage = emptyString
		} else {
			r.Log.Info("Waiting")
			nextStart, err := utils.NextTime(*chaos.GetScheduler(), status.Experiment.StartTime.Time, false)
			if err != nil {
				r.Log.Error(err, "failed to get next start time")
				return ctrl.Result{}, err
			}

			nextTime := chaos.GetNextStart()
			var duration time.Duration
			// if nextStart is not equal to nextTime, the scheduler may have been modified.
			// So set nextStart to new next start time.
			if nextStart.Equal(nextTime) {
				if !chaos.GetNextRecover().IsZero() && chaos.GetNextRecover().Before(nextTime) {
					nextTime = chaos.GetNextRecover()
				}
				duration = nextTime.Sub(now)
			} else {
				chaos.SetNextStart(*nextStart)
				duration = nextStart.Sub(now)
			}

			r.Log.Info("Requeue request", "after", duration)
			result.RequeueAfter = duration
		}
	}

	if err := r.Update(ctx, chaos); err != nil {
		r.Log.Error(err, "unable to update chaos status")
		return ctrl.Result{}, err
	}

	return result, nil
}

func applyAction(
	ctx context.Context,
	r *Reconciler,
	req ctrl.Request,
	duration time.Duration,
	chaos v1alpha1.InnerSchedulerObject,
) error {
	status := chaos.GetStatus()
	r.Log.Info("Chaos action:", "chaos", chaos)

	// Start to apply action
	r.Log.Info("Performing Action")

	if err := r.Apply(ctx, req, chaos); err != nil {
		r.Log.Error(err, "failed to apply chaos action")

		status.Experiment.Phase = v1alpha1.ExperimentPhaseFailed

		updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return r.Update(ctx, chaos)
		})
		if updateError != nil {
			r.Log.Error(updateError, "unable to update chaos finalizers")
		}

		return err
	}

	status.Experiment.StartTime = &metav1.Time{Time: time.Now()}
	status.Experiment.Phase = v1alpha1.ExperimentPhaseRunning
	status.Experiment.Duration = duration.String()
	return nil
}

func updateFailedMessage(
	ctx context.Context,
	r *Reconciler,
	chaos v1alpha1.InnerSchedulerObject,
	err string,
) {
	status := chaos.GetStatus()
	status.FailedMessage = err
	if err := r.Update(ctx, chaos); err != nil {
		r.Log.Error(err, "unable to update chaos status")
	}
}
