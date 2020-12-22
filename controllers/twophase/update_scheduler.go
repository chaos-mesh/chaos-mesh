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
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
)

// SchedulerUpdater updates nextStart and nextRecover for resources
type SchedulerUpdater struct {
	Object runtime.Object
	ctx.Context
}

// Reconcile is twophase reconcile implement
func (r *SchedulerUpdater) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var err error

	r.Log.Info("Modifying scheduler for a resource", "name", req.Name, "namespace", req.Namespace, "time", time.Now())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_chaos := r.Object.DeepCopyObject()
	if err = r.Client.Get(ctx, req.NamespacedName, _chaos); err != nil {
		r.Log.Error(err, "unable to get chaos")
		return ctrl.Result{}, nil
	}
	chaos := _chaos.(v1alpha1.InnerSchedulerObject)

	// update scheduler will start a waiting experiment
	if chaos.GetStatus().Experiment.Phase == v1alpha1.ExperimentPhaseWaiting {
		chaos.SetNextStart(time.Now())
	}

	if err := r.Update(ctx, chaos); err != nil {
		r.Log.Error(err, "unable to update chaos")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}
