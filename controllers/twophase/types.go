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

// Reconcile is two phase reconcile
func (r *Reconciler) Reconcile(chaos v1alpha1.InnerObject, req ctrl.Request) (ctrl.Result, error) {
	var err error
	now := time.Now()

	r.Log.Info("Reconciling a two phase chaos", "name", req.Name, "namespace", req.Namespace)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	schedChaos := chaos.(v1alpha1.InnerSchedulerObject)
	duration, err := schedChaos.GetDuration()
	if err != nil {
		r.Log.Error(err, "failed to get chaos duration")
		return ctrl.Result{}, err
	}

	scheduler := schedChaos.GetScheduler()
	if scheduler == nil {
		r.Log.Info("Scheduler should be defined currently")
		return ctrl.Result{}, fmt.Errorf("misdefined scheduler")
	}

	if duration == nil {
		zero := 0 * time.Second
		duration = &zero
	}

	status := schedChaos.GetStatus()

	if schedChaos.IsDeleted() {
		// This chaos was deleted
		r.Log.Info("Removing self")
		err = r.Recover(ctx, req, schedChaos)
		if err != nil {
			r.Log.Error(err, "failed to recover chaos")
			return ctrl.Result{Requeue: true}, err
		}

		status.Experiment.Phase = v1alpha1.ExperimentPhaseFinished

		if err := r.Update(ctx, schedChaos); err != nil {
			r.Log.Error(err, "unable to update chaos status")
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// if it is renewed, do recover first and promotes the select items.
	isRenewed := schedChaos.IsRenewed()
	if isRenewed {
		r.Log.Info("Renewing self")

		if err = r.Recover(ctx, req, chaos); err != nil {
			r.Log.Error(err, "failed to pause chaos while renewing")
			return ctrl.Result{Requeue: true}, err
		}

		// promote staging items to production.
		if err = r.Promotes(ctx, req, chaos); err != nil {
			r.Log.Error(err, "failed to promote chaos select items from staging to prod(renew)")
			return ctrl.Result{Requeue: true}, err
		}

		// if the experiment is still running, update the endtime(renew is more likely to did pause and resume).
		if status.Experiment.Phase == v1alpha1.ExperimentPhaseRunning {
			now := time.Now()
			status.Experiment.EndTime = &metav1.Time{
				Time: now,
			}
			if status.Experiment.StartTime != nil {
				status.Experiment.Duration = now.Sub(status.Experiment.StartTime.Time).String()
			}
		}

		// mark phase at ExperimentPhaseWaiting to leave it enter the following steps
		status.Experiment.Phase = v1alpha1.ExperimentPhaseWaiting
	}

	if schedChaos.IsPaused() {
		if status.Experiment.Phase == v1alpha1.ExperimentPhaseRunning {
			r.Log.Info("Pausing")

			err = r.Recover(ctx, req, schedChaos)
			if err != nil {
				r.Log.Error(err, "failed to pause chaos")
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
	} else if !schedChaos.GetNextRecover().IsZero() && schedChaos.GetNextRecover().Before(now) {
		// Start recover
		r.Log.Info("Recovering")

		// Don't need to recover again if chaos was paused or renewed before
		if !isRenewed || status.Experiment.Phase != v1alpha1.ExperimentPhasePaused {
			if err = r.Recover(ctx, req, schedChaos); err != nil {
				r.Log.Error(err, "failed to recover chaos")
				return ctrl.Result{Requeue: true}, err
			}
		}

		schedChaos.SetNextRecover(time.Time{})

		status.Experiment.EndTime = &metav1.Time{
			Time: time.Now(),
		}
		status.Experiment.Phase = v1alpha1.ExperimentPhaseWaiting
	} else if status.Experiment.Phase == v1alpha1.ExperimentPhasePaused &&
		!schedChaos.GetNextRecover().IsZero() && schedChaos.GetNextRecover().After(now) {
		// Only resume chaos in the case when current round is not finished,
		// which means the current time is before recover time. Otherwise we
		// don't resume the chaos and just wait for the start of next round.

		r.Log.Info("Resuming")

		dur := schedChaos.GetNextRecover().Sub(now)
		if err := applyAction(ctx, r, req, dur, schedChaos); err != nil {
			return ctrl.Result{Requeue: true}, err
		}

	} else if schedChaos.GetNextStart().Before(now) {
		nextStart, err := utils.NextTime(*schedChaos.GetScheduler(), now)
		if err != nil {
			r.Log.Error(err, "failed to get next start time")
			return ctrl.Result{}, err
		}

		nextRecover := now.Add(*duration)
		if nextStart.Before(nextRecover) {
			err := fmt.Errorf("nextRecover shouldn't be later than nextStart")
			r.Log.Error(err, "nextRecover is later than nextStart. Then recover can never be reached",
				"nextRecover", nextRecover, "nextStart", nextStart)
			return ctrl.Result{}, err
		}

		if err := applyAction(ctx, r, req, *duration, schedChaos); err != nil {
			return ctrl.Result{Requeue: true}, err
		}

		schedChaos.SetNextStart(*nextStart)
		schedChaos.SetNextRecover(nextRecover)
	} else {
		nextTime := schedChaos.GetNextStart()

		if !schedChaos.GetNextRecover().IsZero() && schedChaos.GetNextRecover().Before(nextTime) {
			nextTime = schedChaos.GetNextRecover()
		}
		duration := nextTime.Sub(now)
		r.Log.Info("Requeue request", "after", duration)

		return ctrl.Result{RequeueAfter: duration}, nil
	}

	if err := r.Update(ctx, schedChaos); err != nil {
		r.Log.Error(err, "unable to update chaos status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
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
