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

	"github.com/go-logr/logr"
	"github.com/pingcap/chaos-operator/controllers/networkchaos/partition"

	"github.com/pingcap/chaos-operator/api/v1alpha1"
	"github.com/pingcap/chaos-operator/controllers/networkchaos/delay"
	"github.com/pingcap/chaos-operator/pkg/utils"

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
		return utils.HandleError(false, err)
	}

	switch networkchaos.Spec.Action {
	case v1alpha1.DelayAction:
		reconciler := delay.NewReconciler(r.Client, r.Log.WithValues("action", "delay"), req)
		return reconciler.Reconcile(req)
	case v1alpha1.PartitionAction:
		reconciler := partition.NewReconciler(r.Client, r.Log.WithValues("action", "partition"), req)
		return reconciler.Reconcile(req)
	default:
		err := fmt.Errorf("unknown action %s", string(networkchaos.Spec.Action))
		r.Log.Error(err, "unknown action %s", string(networkchaos.Spec.Action))

		return utils.HandleError(false, err)
	}
}
