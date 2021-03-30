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

package common

import (
	"context"
	"time"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	endpoint "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	// AnnotationCleanFinalizer key
	AnnotationCleanFinalizer = `chaos-mesh.chaos-mesh.org/cleanFinalizer`
	// AnnotationCleanFinalizerForced value
	AnnotationCleanFinalizerForced = `forced`
)

const emptyString = ""

// Reconciler for common chaos
type Reconciler struct {
	endpoint.Endpoint
	ctx.Context
}

// NewReconciler would create Reconciler for common chaos
func NewReconciler(req ctrl.Request, e endpoint.Endpoint, ctx ctx.Context) *Reconciler {
	ctx.Log = ctx.Log.WithName(req.NamespacedName.String())

	return &Reconciler{
		Endpoint: e,
		Context:  ctx,
	}
}

// Reconcile the common chaos
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var err error

	r.Log.Info("Reconciling a common chaos", "name", req.Name, "namespace", req.Namespace)
	ctx := context.Background()

	chaos := r.Object()
	if err = r.Client.Get(ctx, req.NamespacedName, chaos); err != nil {
		r.Log.Error(err, "unable to get chaos")
		return ctrl.Result{}, err
	}

	status := chaos.GetStatus()

	phase := status.Experiment.Phase
	failedMessage := status.FailedMessage
	startTime := status.Experiment.StartTime
	endTime := status.Experiment.EndTime
	duration := status.Experiment.Duration

	if chaos.IsDeleted() {
		// This chaos was deleted
		r.Log.Info("Removing self")
		if err = r.Recover(ctx, req, chaos); err != nil {
			r.Log.Error(err, "failed to recover chaos")
			updateFailedMessage(ctx, r, chaos, err.Error())
			return ctrl.Result{Requeue: true}, err
		}
		phase = v1alpha1.ExperimentPhaseFinished
		failedMessage = emptyString
	} else if chaos.IsPaused() {
		if status.Experiment.Phase == v1alpha1.ExperimentPhaseRunning {
			r.Log.Info("Pausing")

			if err = r.Recover(ctx, req, chaos); err != nil {
				r.Log.Error(err, "failed to pause chaos")
				updateFailedMessage(ctx, r, chaos, err.Error())
				return ctrl.Result{Requeue: true}, err
			}
			now := time.Now()
			endTime = &metav1.Time{
				Time: now,
			}
			if status.Experiment.StartTime != nil {
				duration = now.Sub(status.Experiment.StartTime.Time).String()
			}
		}
		phase = v1alpha1.ExperimentPhasePaused
		failedMessage = emptyString
	} else if status.Experiment.Phase == v1alpha1.ExperimentPhaseRunning {
		r.Log.Info("The common chaos is already running", "name", req.Name, "namespace", req.Namespace)
		return ctrl.Result{}, nil
	} else {
		// Start chaos action
		r.Log.Info("Performing Action")

		if err = r.Apply(ctx, req, chaos); err != nil {
			r.Log.Error(err, "failed to apply chaos action")
			updateFailedMessage(ctx, r, chaos, err.Error())

			phase = v1alpha1.ExperimentPhaseFailed

			return ctrl.Result{Requeue: true}, err
		}
		startTime = &metav1.Time{
			Time: time.Now(),
		}
		phase = v1alpha1.ExperimentPhaseRunning
		failedMessage = emptyString
	}
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err = r.Client.Get(ctx, req.NamespacedName, chaos); err != nil {
			r.Log.Error(err, "unable to get chaos")
			return err
		}

		status := chaos.GetStatus()
		status.Experiment.Phase = phase
		status.FailedMessage = failedMessage
		status.Experiment.StartTime = startTime
		status.Experiment.EndTime = endTime
		status.Experiment.Duration = duration
		return r.Update(ctx, chaos)
	})
	if err != nil {
		r.Log.Error(err, "unable to update chaos status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// Since this will make the reconciler requeued, we do not retry it.
func updateFailedMessage(
	ctx context.Context,
	r *Reconciler,
	chaos v1alpha1.InnerObject,
	err string,
) {
	status := chaos.GetStatus()
	status.FailedMessage = err
	if err := r.Update(ctx, chaos); err != nil {
		r.Log.Error(err, "unable to update chaos status")
	}
}
