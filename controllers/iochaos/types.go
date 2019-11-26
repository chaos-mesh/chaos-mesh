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

	"github.com/go-logr/logr"

	"github.com/pingcap/chaos-operator/api/v1alpha1"
	"github.com/pingcap/chaos-operator/controllers/iochaos/fs"

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
		return ctrl.Result{}, nil
	}

	switch iochaos.Spec.Action {
	case v1alpha1.FileSystemLayer:
		reconciler := fs.NewConciler(r.Client, r.Log.WithValues("reconciler", "chaosfs"), req)
		return reconciler.Reconcile(req)
	default:
		r.Log.Error(nil, "unknown file system I/O layer %s", string(iochaos.Spec.Layer))

		return ctrl.Result{}, nil
	}
}
