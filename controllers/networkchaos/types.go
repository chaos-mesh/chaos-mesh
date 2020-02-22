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
	"fmt"

	"k8s.io/client-go/tools/record"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/controllers/common"
	"github.com/pingcap/chaos-mesh/controllers/networkchaos/netem"
	"github.com/pingcap/chaos-mesh/controllers/networkchaos/partition"
	"github.com/pingcap/chaos-mesh/controllers/twophase"
)

type Reconciler struct {
	client.Client
	record.EventRecorder
	Log logr.Logger
}

// Reconcile reconciles a NetworkChaos resource
func (r *Reconciler) Reconcile(req ctrl.Request, chaos *v1alpha1.NetworkChaos) (ctrl.Result, error) {
	r.Log.Info("reconciling networkchaos")

	scheduler := chaos.GetScheduler()
	duration, err := chaos.GetDuration()
	if err != nil {
		r.Log.Error(err, fmt.Sprintf("unable to get networkchaos[%s/%s]'s duration", chaos.Namespace, chaos.Name))
		return ctrl.Result{}, err
	}
	if scheduler == nil && duration == nil {
		return r.commonNetworkChaos(chaos, req)
	} else if scheduler != nil && duration != nil {
		return r.scheduleNetworkChaos(chaos, req)
	}

	err = fmt.Errorf("networkchaos[%s/%s] spec invalid", chaos.Namespace, chaos.Name)
	// This should be ensured by admission webhook in the future
	r.Log.Error(err, "scheduler and duration should be omitted or defined at the same time")
	return ctrl.Result{}, err
}

func (r *Reconciler) commonNetworkChaos(networkchaos *v1alpha1.NetworkChaos, req ctrl.Request) (ctrl.Result, error) {
	var cr *common.Reconciler
	switch networkchaos.Spec.Action {
	case v1alpha1.DelayAction, v1alpha1.DuplicateAction, v1alpha1.CorruptAction, v1alpha1.LossAction:
		cr = netem.NewCommonReconciler(r.Client, r.Log.WithValues("action", "netem"),
			req, r.EventRecorder)
	case v1alpha1.PartitionAction:
		cr = partition.NewCommonReconciler(r.Client, r.Log.WithValues("action", "partition"),
			req, r.EventRecorder)
	default:
		return r.invalidActionResponse(networkchaos)
	}
	return cr.Reconcile(req)
}

func (r *Reconciler) scheduleNetworkChaos(networkchaos *v1alpha1.NetworkChaos, req ctrl.Request) (ctrl.Result, error) {
	var sr *twophase.Reconciler
	switch networkchaos.Spec.Action {
	case v1alpha1.DelayAction, v1alpha1.DuplicateAction, v1alpha1.CorruptAction, v1alpha1.LossAction:
		sr = netem.NewTwoPhaseReconciler(r.Client, r.Log.WithValues("action", "netem"),
			req, r.EventRecorder)
	case v1alpha1.PartitionAction:
		sr = partition.NewTwoPhaseReconciler(r.Client, r.Log.WithValues("action", "partition"),
			req, r.EventRecorder)
	default:
		return r.invalidActionResponse(networkchaos)
	}
	return sr.Reconcile(req)
}

func (r *Reconciler) invalidActionResponse(networkchaos *v1alpha1.NetworkChaos) (ctrl.Result, error) {
	r.Log.Error(nil, "networkchaos action is invalid", "action", networkchaos.Spec.Action)
	return ctrl.Result{}, fmt.Errorf("invalid networkchaos action")
}
