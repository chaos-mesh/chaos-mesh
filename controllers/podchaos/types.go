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

package podchaos

import (
	"fmt"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/controllers/common"
	"github.com/pingcap/chaos-mesh/controllers/podchaos/containerkill"
	"github.com/pingcap/chaos-mesh/controllers/podchaos/podfailure"
	"github.com/pingcap/chaos-mesh/controllers/podchaos/podkill"
	"github.com/pingcap/chaos-mesh/controllers/twophase"
)

type Reconciler struct {
	client.Client
	Log logr.Logger
}

// Reconcile reconciles a PodChaos resource
func (r *Reconciler) Reconcile(req ctrl.Request, chaos *v1alpha1.PodChaos) (ctrl.Result, error) {
	r.Log.Info("reconciling podchaos")
	scheduler := chaos.GetScheduler()
	duration, err := chaos.GetDuration()
	if err != nil {
		r.Log.Error(err, fmt.Sprintf("unable to get podchaos[%s/%s]'s duration", chaos.Namespace, chaos.Name))
		return ctrl.Result{}, err
	}
	if scheduler == nil && duration == nil {
		return r.commonPodChaos(chaos, req)
	} else if scheduler != nil {
		return r.schedulePodChaos(chaos, req)
	}

	// This should be ensured by admission webhook in the future
	r.Log.Error(fmt.Errorf("podchaos[%s/%s] spec invalid", chaos.Namespace, chaos.Name), "scheduler and duration should be omitted or defined at the same time")
	return ctrl.Result{}, fmt.Errorf("invalid scheduler and duration")
}

func (r *Reconciler) commonPodChaos(podchaos *v1alpha1.PodChaos, req ctrl.Request) (ctrl.Result, error) {
	var pr *common.Reconciler
	switch podchaos.Spec.Action {
	case v1alpha1.PodKillAction:
		return r.notSupportedResponse(podchaos)
	case v1alpha1.ContainerKillAction:
		return r.notSupportedResponse(podchaos)
	case v1alpha1.PodFailureAction:
		pr = podfailure.NewCommonReconciler(r.Client, r.Log.WithValues("action", "pod-failure"), req)
	default:
		return r.invalidActionResponse(podchaos)
	}
	return pr.Reconcile(req)
}

func (r *Reconciler) schedulePodChaos(podchaos *v1alpha1.PodChaos, req ctrl.Request) (ctrl.Result, error) {
	var tr *twophase.Reconciler
	switch podchaos.Spec.Action {
	case v1alpha1.PodKillAction:
		tr = podkill.NewTwoPhaseReconciler(r.Client, r.Log.WithValues("action", "pod-kill"), req)
	case v1alpha1.PodFailureAction:
		tr = podfailure.NewTwoPhaseReconciler(r.Client, r.Log.WithValues("action", "pod-failure"), req)
	case v1alpha1.ContainerKillAction:
		tr = containerkill.NewTwoPhaseReconciler(r.Client, r.Log.WithValues("action", "container-kill"), req)
	default:
		return r.invalidActionResponse(podchaos)
	}
	return tr.Reconcile(req)
}

func (r *Reconciler) invalidActionResponse(podchaos *v1alpha1.PodChaos) (ctrl.Result, error) {
	r.Log.Error(nil, "podchaos action is invalid", "action", podchaos.Spec.Action)
	return ctrl.Result{}, fmt.Errorf("invalid chaos action")
}

func (r *Reconciler) notSupportedResponse(podchaos *v1alpha1.PodChaos) (ctrl.Result, error) {
	r.Log.Error(nil, "podchaos action hasn't support duration chaos yet", "action", podchaos.Spec.Action)
	return ctrl.Result{}, fmt.Errorf("unsupported chaos action")
}
