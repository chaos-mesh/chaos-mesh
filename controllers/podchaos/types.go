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
	"context"
	"fmt"

	"github.com/pingcap/chaos-mesh/controllers/common"

	"github.com/pingcap/chaos-mesh/controllers/twophase"

	"github.com/go-logr/logr"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	commonPodfailure "github.com/pingcap/chaos-mesh/controllers/common/podfailure"
	"github.com/pingcap/chaos-mesh/controllers/podchaos/podkill"
	twophasePodfailure "github.com/pingcap/chaos-mesh/controllers/twophase/podfailure"
)

type Reconciler struct {
	client.Client
	Log logr.Logger
}

func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("reconciling podchaos")
	ctx := context.Background()

	var podchaos v1alpha1.PodChaos
	if err := r.Get(ctx, req.NamespacedName, &podchaos); err != nil {
		r.Log.Error(err, "unable to get podchaos")
		return ctrl.Result{}, nil
	}
	scheduler := podchaos.GetScheduler()
	duration, err := podchaos.GetDuration()
	if err != nil {
		r.Log.Error(err, fmt.Sprintf("unable to get podchaos[%s/%s]'s duration", podchaos.Namespace, podchaos.Name))
		return ctrl.Result{}, nil
	}
	if scheduler == nil && duration == nil {
		return r.commonPodChaos(&podchaos, req)
	} else if scheduler != nil {
		return r.schedulePodChaos(&podchaos, req)
	}

	// This should be ensured by admission webhook in the future
	r.Log.Error(fmt.Errorf("podchaos[%s/%s] spec invalid", podchaos.Namespace, podchaos.Name), "scheduler and duration should be omitted or defined at the same time")
	return ctrl.Result{}, nil
}

func (r *Reconciler) commonPodChaos(podchaos *v1alpha1.PodChaos, req ctrl.Request) (ctrl.Result, error) {
	switch podchaos.Spec.Action {
	case v1alpha1.PodKillAction:
		return r.notSupportedResponse(podchaos), nil
	case v1alpha1.PodFailureAction:
		r := commonPodfailure.NewCommonReconciler(r.Client, r.Log.WithValues("action", "pod-failure"), req)
		reconciler := common.NewReconciler(r, r.Client, r.Log)
		return reconciler.Reconcile(req)
	default:
		return r.defaultResponse(podchaos), nil
	}
}

func (r *Reconciler) schedulePodChaos(podchaos *v1alpha1.PodChaos, req ctrl.Request) (ctrl.Result, error) {
	switch podchaos.Spec.Action {
	case v1alpha1.PodKillAction:
		reconciler := podkill.Reconciler{
			Client: r.Client,
			Log:    r.Log.WithValues("action", "pod-kill"),
		}
		return reconciler.Reconcile(req)
	case v1alpha1.PodFailureAction:
		if podchaos.Spec.Duration == nil {
			r.Log.Error(nil, "schedule podchaos should define duration", "action", podchaos.Spec.Action)
			return ctrl.Result{}, nil
		}
		r := twophasePodfailure.NewTwoPhaseReconciler(r.Client, r.Log.WithValues("action", "pod-failure"), req)
		reconciler := twophase.NewReconciler(r, r.Client, r.Log)
		return reconciler.Reconcile(req)
	default:
		return r.defaultResponse(podchaos), nil
	}
}

func (r *Reconciler) defaultResponse(podchaos *v1alpha1.PodChaos) ctrl.Result {
	r.Log.Error(nil, "podchaos action is invalid", "action", podchaos.Spec.Action)
	return ctrl.Result{}
}

func (r *Reconciler) notSupportedResponse(podchaos *v1alpha1.PodChaos) ctrl.Result {
	r.Log.Error(nil, "podchaos action hasn't support duration chaos yet", "action", podchaos.Spec.Action)
	return ctrl.Result{}
}
