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
	"fmt"

	"github.com/pingcap/chaos-mesh/controllers/common"

	"github.com/pingcap/chaos-mesh/controllers/twophase"

	"github.com/go-logr/logr"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	commonNetem "github.com/pingcap/chaos-mesh/controllers/common/networkchaos/netem"
	commonPartition "github.com/pingcap/chaos-mesh/controllers/common/networkchaos/partition"
	twophaseNetem "github.com/pingcap/chaos-mesh/controllers/twophase/networkchaos/netem"
	twophasePartition "github.com/pingcap/chaos-mesh/controllers/twophase/networkchaos/partition"
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

	scheduler := networkchaos.GetScheduler()
	duration, err := networkchaos.GetDuration()
	if err != nil {
		r.Log.Error(err, fmt.Sprintf("unable to get podchaos[%s/%s]'s duration", networkchaos.Namespace, networkchaos.Name))
		return ctrl.Result{}, nil
	}
	if scheduler == nil && duration == nil {
		return r.commonNetworkChaos(&networkchaos, req)
	} else if scheduler != nil && duration != nil {
		return r.scheduleNetworkChaos(&networkchaos, req)
	}

	// This should be ensured by admission webhook in the future
	r.Log.Error(fmt.Errorf("networkchaos[%s/%s] spec invalid", networkchaos.Namespace, networkchaos.Name), "scheduler and duration should be omitted or defined at the same time")
	return ctrl.Result{}, nil
}

func (r *Reconciler) commonNetworkChaos(networkchaos *v1alpha1.NetworkChaos, req ctrl.Request) (ctrl.Result, error) {
	switch networkchaos.Spec.Action {
	case v1alpha1.DelayAction, v1alpha1.DuplicateAction, v1alpha1.CorruptAction, v1alpha1.LossAction:
		r := commonNetem.NewCommonReconciler(r.Client, r.Log.WithValues("action", "netem"), req)
		reconciler := common.NewReconciler(r, r.Client, r.Log)
		return reconciler.Reconcile(req)
	case v1alpha1.PartitionAction:
		r := commonPartition.NewCommonReconciler(r.Client, r.Log.WithValues("action", "partition"), req)
		reconciler := common.NewReconciler(r, r.Client, r.Log)
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
		r := twophasePartition.NewTwoPhaseReconciler(r.Client, r.Log.WithValues("action", "partition"), req)
		reconciler := twophase.NewReconciler(r, r.Client, r.Log)
		return reconciler.Reconcile(req)
	default:
		return r.defaultResponse(networkchaos), nil
	}
}

func (r *Reconciler) defaultResponse(networkchaos *v1alpha1.NetworkChaos) ctrl.Result {
	r.Log.Error(nil, "networkchaos action is invalid", "action", networkchaos.Spec.Action)
	return ctrl.Result{}
}
