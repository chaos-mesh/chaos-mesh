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

package networkchaos

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/pingcap/chaos-mesh/controllers/persist"
	persistNetem "github.com/pingcap/chaos-mesh/controllers/persist/netem"
	"github.com/pingcap/chaos-mesh/controllers/twophase"
	twophaseNetem "github.com/pingcap/chaos-mesh/controllers/twophase/netem"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/controllers/networkchaos/partition"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	client.Client
	Log logr.Logger
}

func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("reconciling networkchaos")
	ctx := context.Background()

	var networkchaos v1alpha1.NetworkChaos
	if err := r.Get(ctx, req.NamespacedName, &networkchaos); err != nil {
		r.Log.Error(err, "unable to get networkchaos")
		return ctrl.Result{}, nil
	}

	chaos := &v1alpha1.NetworkChaos{}
	if err := r.Get(ctx, req.NamespacedName, chaos); err != nil {
		r.Log.Error(err, "unable to get chaos")
		return ctrl.Result{}, nil
	}

	if chaos.GetScheduler() == nil {
		return r.persistNetworkChaos(&networkchaos, req)
	} else {
		return r.scheduleNetworkChaos(&networkchaos, req)
	}
}

func (r *Reconciler) persistNetworkChaos(networkchaos *v1alpha1.NetworkChaos, req ctrl.Request) (ctrl.Result, error) {
	switch networkchaos.Spec.Action {
	case v1alpha1.DelayAction, v1alpha1.DuplicateAction, v1alpha1.CorruptAction, v1alpha1.LossAction:
		r := persistNetem.NewPersistentReconciler(r.Client, r.Log.WithValues("action", "netem"), req)
		reconciler := persist.NewReconciler(r, r.Client, r.Log)
		return reconciler.Reconcile(req)
	case v1alpha1.PartitionAction:
		reconciler := partition.NewReconciler(r.Client, r.Log.WithValues("action", "partition"), req)
		return reconciler.Reconcile(req)
	default:
		return r.defaultResponse(networkchaos), nil
	}
}

func (r *Reconciler) scheduleNetworkChaos(networkchaos *v1alpha1.NetworkChaos, req ctrl.Request) (ctrl.Result, error) {
	switch networkchaos.Spec.Action {
	case v1alpha1.DelayAction, v1alpha1.DuplicateAction, v1alpha1.CorruptAction, v1alpha1.LossAction:
		r := twophaseNetem.NewTwoPhaseReconciler(r.Client, r.Log.WithValues("action", "netem"), req)
		reconciler := twophase.NewReconciler(r, r.Client, r.Log)
		return reconciler.Reconcile(req)
	case v1alpha1.PartitionAction:
		reconciler := partition.NewReconciler(r.Client, r.Log.WithValues("action", "partition"), req)
		return reconciler.Reconcile(req)
	default:
		return r.defaultResponse(networkchaos), nil
	}
}

func (r *Reconciler) defaultResponse(networkchaos *v1alpha1.NetworkChaos) ctrl.Result {
	r.Log.Error(nil, "networkchaos action is invalid", "action", networkchaos.Spec.Action)
	return ctrl.Result{}
}
