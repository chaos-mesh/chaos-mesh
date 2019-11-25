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

package iochaos

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"

	"github.com/pingcap/chaos-operator/api/v1alpha1"
	"github.com/pingcap/chaos-operator/controllers/iochaos/delay"
	"github.com/pingcap/chaos-operator/pkg/utils"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	client.Client
	Log logr.Logger
}

func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("reconciling iochaos")
	ctx := context.Background()

	var iochaos v1alpha1.IoChaos
	if err := r.Get(ctx, req.NamespacedName, &iochaos); err != nil {
		r.Log.Error(err, "unable to get iochaos")
		return utils.HandleError(false, err)
	}

	switch iochaos.Spec.Action {
	case v1alpha1.IODelayAction:
		reconciler := delay.NewConciler(r.Client, r.Log.WithValues("reconciler", "delay"), req)
		return reconciler.Reconcile(req)
	default:
		err := fmt.Errorf("unknown action %s", string(iochaos.Spec.Action))
		r.Log.Error(err, "unknown action %s", string(iochaos.Spec.Action))

		return utils.HandleError(false, err)
	}
}
