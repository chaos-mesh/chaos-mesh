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
	durationPodfailure "github.com/pingcap/chaos-mesh/controllers/common/podfailure"
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
		r.Log.Error(err, "unable to get duration")
		return ctrl.Result{}, nil
	}
	if scheduler == nil && duration == nil {
		return r.durationPodChaos(&podchaos, req)
	} else if scheduler != nil && duration != nil {
		return r.schedulePodChaos(&podchaos, req)
	}

	// This should be ensured by admission webhook in the future
	r.Log.Error(fmt.Errorf("podchaos[%s/%s] spec invaild", podchaos.Namespace, podchaos.Name), "scheduler and duration should be omiited or defined at the same time")
	return ctrl.Result{}, nil
}

func (r *Reconciler) durationPodChaos(podchaos *v1alpha1.PodChaos, req ctrl.Request) (ctrl.Result, error) {
	switch podchaos.Spec.Action {
	case v1alpha1.PodKillAction:
		return r.notSupportedResponse(podchaos), nil
	case v1alpha1.PodFailureAction:
		r := durationPodfailure.NewDurationReconciler(r.Client, r.Log.WithValues("action", "pod-failure"), req)
		reconciler := common.Reconciler{
			InnerCommonReconcile: r,
			Client:               r.Client,
			Log:                  r.Log,
		}
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
		r := twophasePodfailure.NewTwoPhaseReconciler(r.Client, r.Log.WithValues("action", "pod-failure"), req)
		reconciler := twophase.Reconciler{
			InnerReconciler: r,
			Client:          r.Client,
			Log:             r.Log,
		}
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
