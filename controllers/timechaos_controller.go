// Copyright 2020 PingCAP, Inc.
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

package controllers

import (
	"context"

	"k8s.io/client-go/tools/record"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/controllers/timechaos"
	"github.com/pingcap/chaos-mesh/pkg/utils"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TimeChaosReconciler reconciles a TimeChaos object
type TimeChaosReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=pingcap.com,resources=timechaos,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pingcap.com,resources=timechaos/status,verbs=get;update;patch

// Reconcile reconciles a TimeChaos resource
func (r *TimeChaosReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	logger := r.Log.WithValues("reconciler", "timechaos")

	reconciler := timechaos.Reconciler{
		Client: r.Client,
		Log:    logger,
	}

	chaos := &v1alpha1.TimeChaos{}
	if err := r.Get(context.Background(), req.NamespacedName, chaos); err != nil {
		r.Log.Error(err, "unable to get time chaos")
		return ctrl.Result{}, nil
	}

	if !chaos.IsDeleted() {
		r.Recorder.Event(chaos, v1.EventTypeNormal, utils.EventChaosStarted, "")
		result, err = reconciler.Reconcile(req, chaos)
	} else {
		result, err = reconciler.Reconcile(req, chaos)
		r.Recorder.Event(chaos, v1.EventTypeNormal, utils.EventChaosCompleted, "")
	}

	return result, err
}

// SetupWithManager setups a time chaos reconciler on controller-manager
func (r *TimeChaosReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.TimeChaos{}).
		Complete(r)
}
