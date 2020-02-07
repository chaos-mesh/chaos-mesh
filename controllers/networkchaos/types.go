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

	"github.com/pingcap/chaos-mesh/controllers/networkchaos/netem"
	"github.com/pingcap/chaos-mesh/controllers/networkchaos/partition"

	"github.com/pingcap/chaos-mesh/controllers/common"

	"github.com/pingcap/chaos-mesh/controllers/twophase"

	"github.com/go-logr/logr"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
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
	var cr *common.Reconciler
	switch networkchaos.Spec.Action {
	case v1alpha1.DelayAction, v1alpha1.DuplicateAction, v1alpha1.CorruptAction, v1alpha1.LossAction:
		cr = netem.NewCommonReconciler(r.Client, r.Log.WithValues("action", "netem"), req)
	case v1alpha1.PartitionAction:
		cr = partition.NewCommonReconciler(r.Client, r.Log.WithValues("action", "partition"), req)
	default:
		return r.invalidActionResponse(networkchaos), nil
	}
	return cr.Reconcile(req)
}

func (r *Reconciler) scheduleNetworkChaos(networkchaos *v1alpha1.NetworkChaos, req ctrl.Request) (ctrl.Result, error) {
	var sr *twophase.Reconciler
	switch networkchaos.Spec.Action {
	case v1alpha1.DelayAction, v1alpha1.DuplicateAction, v1alpha1.CorruptAction, v1alpha1.LossAction:
		sr = netem.NewTwoPhaseReconciler(r.Client, r.Log.WithValues("action", "netem"), req)
	case v1alpha1.PartitionAction:
		sr = partition.NewTwoPhaseReconciler(r.Client, r.Log.WithValues("action", "partition"), req)
	default:
		return r.invalidActionResponse(networkchaos), nil
	}
	return sr.Reconcile(req)
}

func (r *Reconciler) invalidActionResponse(networkchaos *v1alpha1.NetworkChaos) ctrl.Result {
	r.Log.Error(nil, "networkchaos action is invalid", "action", networkchaos.Spec.Action)
	return ctrl.Result{}
}
