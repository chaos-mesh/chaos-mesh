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
	"github.com/chaos-mesh/chaos-mesh/pkg/finalizer"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	endpoint "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	// AnnotationCleanFinalizer key
	AnnotationCleanFinalizer = `chaos-mesh.chaos-mesh.org/cleanFinalizer`
	// AnnotationCleanFinalizerForced value
	AnnotationCleanFinalizerForced = `forced`

	// Prefinalizer protect chaos from being deleted straightly
	Prefinalizer = `pre-finalizer.common.chaos-mesh.org`
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

	if chaos.IsDeleted() {
		// This chaos was deleted
		r.Log.Info("Removing self")
		if err = r.Recover(ctx, req, chaos); err != nil {
			r.Log.Error(err, "failed to recover chaos")
			updateFailedMessage(ctx, r, chaos, err.Error())
			return ctrl.Result{Requeue: true}, err
		}
		chaos.GetMeta().SetFinalizers(finalizer.RemoveFromFinalizer(chaos.GetMeta().GetFinalizers(), Prefinalizer))
		status.Experiment.Phase = v1alpha1.ExperimentPhaseFinished
		status.FailedMessage = emptyString
	} else if !finalizer.ContainsFinalizer(chaos.GetMeta().GetFinalizers(), Prefinalizer) {
		chaos.GetMeta().SetFinalizers(finalizer.InsertFinalizer(chaos.GetMeta().GetFinalizers(), Prefinalizer))
	} else if chaos.IsPaused() {
		if status.Experiment.Phase == v1alpha1.ExperimentPhaseRunning {
			r.Log.Info("Pausing")

			if err = r.Recover(ctx, req, chaos); err != nil {
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
	} else if status.Experiment.Phase == v1alpha1.ExperimentPhaseRunning {
		r.Log.Info("The common chaos is already running", "name", req.Name, "namespace", req.Namespace)
		return ctrl.Result{}, nil
	} else {
		// Start chaos action
		r.Log.Info("Performing Action")

		if err = r.Apply(ctx, req, chaos); err != nil {
			r.Log.Error(err, "failed to apply chaos action")
			updateFailedMessage(ctx, r, chaos, err.Error())

			status.Experiment.Phase = v1alpha1.ExperimentPhaseFailed

			updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				return r.Update(ctx, chaos)
			})
			if updateError != nil {
				r.Log.Error(updateError, "unable to update chaos finalizers")
				updateFailedMessage(ctx, r, chaos, updateError.Error())
			}

			return ctrl.Result{Requeue: true}, err
		}
		status.Experiment.StartTime = &metav1.Time{
			Time: time.Now(),
		}
		status.Experiment.Phase = v1alpha1.ExperimentPhaseRunning
		status.FailedMessage = emptyString
	}

	if err := r.Update(ctx, chaos); err != nil {
		if !errors.IsConflict(err) {
			r.Log.Error(err, "fail to update chaos")
			return ctrl.Result{}, err
		}

		r.Log.Error(err, "conflict on updating chaos")

		updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			obj := r.Object().DeepCopyObject().(v1alpha1.InnerObject)
			if err = r.Client.Get(ctx, req.NamespacedName, obj); err != nil {
				r.Log.Error(err, "unable to get chaos")
				return err
			}

			*obj.GetStatus() = *status
			obj.GetMeta().SetFinalizers(chaos.GetMeta().GetFinalizers())
			return r.Update(ctx, obj)
		})

		if updateError != nil {
			r.Log.Error(updateError, "unable to update chaos")
			return ctrl.Result{}, updateError
		}
	}

	return ctrl.Result{}, nil
}

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
