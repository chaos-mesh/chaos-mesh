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

	"github.com/go-logr/logr"

	"github.com/pingcap/chaos-operator/api/v1alpha1"
	"github.com/pingcap/chaos-operator/controllers/podchaos/podfailure"
	"github.com/pingcap/chaos-operator/controllers/podchaos/podkill"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	switch podchaos.Spec.Action {
	case v1alpha1.PodKillAction:
		reconciler := podkill.Reconciler{
			Client: r.Client,
			Log:    r.Log.WithValues("action", "pod-kill"),
		}
		return reconciler.Reconcile(req)
	case v1alpha1.PodFailureAction:
		reconciler := podfailure.NewReconciler(r.Client, r.Log.WithValues("action", "pod-failure"), req)
		return reconciler.Reconcile(req)
	default:
		err := fmt.Errorf("unknown action %s", string(podchaos.Spec.Action))
		r.Log.Error(err, "unknown action %s", string(podchaos.Spec.Action))

		return ctrl.Result{}, nil
	}
}
