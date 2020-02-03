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

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"k8s.io/client-go/util/retry"

	"github.com/go-logr/logr"
	"github.com/pingcap/chaos-mesh/pkg/apiinterface"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// InnerCommonObject used in common chaos reconcile
type InnerCommonObject interface {
	IsDeleted() bool
	apiinterface.StatefulObject
}

// InnerCommonReconcile used in common chaos reconcile
type InnerCommonReconcile interface {
	Apply(ctx context.Context, req ctrl.Request, chaos InnerCommonObject) error

	Recover(ctx context.Context, req ctrl.Request, chaos InnerCommonObject) error

	Object() InnerCommonObject
}

// Reconciler for common chaos
type Reconciler struct {
	InnerCommonReconcile
	client.Client
	Log logr.Logger
}

// NewReconciler would create Reconciler for common chaos
func NewReconciler(reconcile InnerCommonReconcile, c client.Client, log logr.Logger) *Reconciler {
	return &Reconciler{
		InnerCommonReconcile: reconcile,
		Client:               c,
		Log:                  log,
	}
}

// Reconcile the common chaos
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var err error

	r.Log.Info("reconciling a common chaos", "name", req.Name, "namespace", req.Namespace)
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
