// Copyright 2019 PingCAP, Inc.
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

	"github.com/go-logr/logr"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/controllers/reconciler"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// AnnotationCleanFinalizer key
	AnnotationCleanFinalizer = `chaos-mesh.pingcap.com/cleanFinalizer`
	// AnnotationCleanFinalizerForced value
	AnnotationCleanFinalizerForced = `forced`
)

// Reconciler for common chaos
type Reconciler struct {
	reconciler.InnerReconciler
	client.Client
	Log logr.Logger
}

// NewReconciler would create Reconciler for common chaos
func NewReconciler(reconcile reconciler.InnerReconciler, c client.Client, log logr.Logger) *Reconciler {
	return &Reconciler{
		InnerReconciler: reconcile,
		Client:          c,
		Log:             log,
	}
}

// Reconcile the common chaos
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var err error

	r.Log.Info("Reconciling a common chaos", "name", req.Name, "namespace", req.Namespace)
	ctx := context.Background()

	chaos := r.Object()
	if err = r.Get(ctx, req.NamespacedName, chaos); err != nil {
		r.Log.Error(err, "unable to get chaos")
		return ctrl.Result{}, err
	}

	status := chaos.GetStatus()

	if chaos.IsDeleted() {
		// This chaos was deleted
		r.Log.Info("Removing self")
		if err = r.Recover(ctx, req, chaos); err != nil {
			r.Log.Error(err, "failed to recover chaos")
			return ctrl.Result{Requeue: true}, err
		}
		status.Experiment.Phase = v1alpha1.ExperimentPhaseFinished
	} else if chaos.IsPaused() {
		if status.Experiment.Phase == v1alpha1.ExperimentPhaseRunning {
			r.Log.Info("Pausing")

			if err = r.Recover(ctx, req, chaos); err != nil {
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
	} else if status.Experiment.Phase == v1alpha1.ExperimentPhaseRunning {
		r.Log.Info("The common chaos is already running", "name", req.Name, "namespace", req.Namespace)
		return ctrl.Result{}, nil
	} else {
		// Start chaos action
		r.Log.Info("Performing Action")

		if err = r.Apply(ctx, req, chaos); err != nil {
			r.Log.Error(err, "failed to apply chaos action")

			status.Experiment.Phase = v1alpha1.ExperimentPhaseFailed

			updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				return r.Update(ctx, chaos)
			})
			if updateError != nil {
				r.Log.Error(updateError, "unable to update chaos finalizers")
			}

			return ctrl.Result{Requeue: true}, err
		}
		status.Experiment.StartTime = &metav1.Time{
			Time: time.Now(),
		}
		status.Experiment.Phase = v1alpha1.ExperimentPhaseRunning
	}

	if err := r.Update(ctx, chaos); err != nil {
		r.Log.Error(err, "unable to update chaos status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
