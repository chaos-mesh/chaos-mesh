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

package duration

import (
	"context"
	"time"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"k8s.io/client-go/util/retry"

	"github.com/go-logr/logr"
	"github.com/pingcap/chaos-mesh/pkg/apiinterface"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// InnerDurationObject used in duration chaos reconcile
type InnerDurationObject interface {
	IsDeleted() bool
	apiinterface.StatefulObject
}

// InnerDurationReconcile used in duration chaos reconcile
type InnerDurationReconcile interface {
	Apply(ctx context.Context, req ctrl.Request, chaos InnerDurationObject) error

	Recover(ctx context.Context, req ctrl.Request, chaos InnerDurationObject) error

	Object() InnerDurationObject
}

// Reconciler for persist chaos
type Reconciler struct {
	InnerDurationReconcile
	client.Client
	Log logr.Logger
}

// NewReconciler would create Reconciler for persist chaos
func NewReconciler(reconcile InnerDurationReconcile, c client.Client, log logr.Logger) *Reconciler {
	return &Reconciler{
		InnerDurationReconcile: reconcile,
		Client:                 c,
		Log:                    log,
	}
}

// Reconcile the persist chaos
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var err error

	r.Log.Info("reconciling a persistent chaos", "name", req.Name, "namespace", req.Namespace)
	ctx := context.Background()

	chaos := r.Object()
	if err = r.Get(ctx, req.NamespacedName, chaos); err != nil {
		r.Log.Error(err, "unable to get chaos")
		return ctrl.Result{}, nil
	}
	if chaos.IsDeleted() {
		// This chaos was deleted
		r.Log.Info("Removing self")
		err = r.Recover(ctx, req, chaos)
		if err != nil {
			r.Log.Error(err, "failed to recover chaos")
			return ctrl.Result{Requeue: true}, nil
		}
	}

	// Start failure action
	r.Log.Info("Performing Action")

	status := chaos.GetStatus()

	err = r.Apply(ctx, req, chaos)
	if err != nil {
		r.Log.Error(err, "failed to apply chaos action")

		updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return r.Update(ctx, chaos)
		})
		if updateError != nil {
			r.Log.Error(updateError, "unable to update chaos finalizers")
		}

		return ctrl.Result{Requeue: true}, nil
	}
	status.Experiment.StartTime = &metav1.Time{
		Time: time.Now(),
	}
	status.Experiment.Phase = v1alpha1.ExperimentPhaseRunning

	if err := r.Update(ctx, chaos); err != nil {
		r.Log.Error(err, "unable to update chaosctl status")
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}
