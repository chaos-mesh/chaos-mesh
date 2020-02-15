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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/controllers/common"
	"github.com/pingcap/chaos-mesh/controllers/iochaos/fs"
	"github.com/pingcap/chaos-mesh/controllers/twophase"
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

	scheduler := iochaos.GetScheduler()
	duration, err := iochaos.GetDuration()
	if err != nil {
		r.Log.Error(err, fmt.Sprintf("unable to get podchaos[%s/%s]'s duration", iochaos.Namespace, iochaos.Name))
		return ctrl.Result{}, nil
	}
	if scheduler == nil && duration == nil {
		return r.commonIOChaos(&iochaos, req)
	} else if scheduler != nil && duration != nil {
		return r.scheduleIOChaos(&iochaos, req)
	}

	// This should be ensured by admission webhook in the future
	r.Log.Error(fmt.Errorf("iochaos[%s/%s] spec invalid", iochaos.Namespace, iochaos.Name), "scheduler and duration should be omitted or defined at the same time")
	return ctrl.Result{}, nil
}

func (r *Reconciler) commonIOChaos(iochaos *v1alpha1.IoChaos, req ctrl.Request) (ctrl.Result, error) {
	var cr *common.Reconciler
	switch iochaos.Spec.Layer {
	case v1alpha1.FileSystemLayer:
		cr = fs.NewCommonReconciler(r.Client, r.Log.WithValues("reconciler", "chaosfs"), req)
	default:
		return r.invalidActionResponse(iochaos), nil
	}
	return cr.Reconcile(req)
}

func (r *Reconciler) scheduleIOChaos(iochaos *v1alpha1.IoChaos, req ctrl.Request) (ctrl.Result, error) {
	var sr *twophase.Reconciler
	switch iochaos.Spec.Layer {
	case v1alpha1.FileSystemLayer:
		sr = fs.NewTwoPhaseReconciler(r.Client, r.Log.WithValues("reconciler", "chaosfs"), req)
	default:
		return r.invalidActionResponse(iochaos), nil
	}
	return sr.Reconcile(req)
}

func (r *Reconciler) invalidActionResponse(iochaos *v1alpha1.IoChaos) ctrl.Result {
	r.Log.Error(nil, "unknown file system I/O layer", "action", iochaos.Spec.Action)
	return ctrl.Result{}
}
