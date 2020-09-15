// Copyright 2019 Chaos Mesh Authors.
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
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/controllers/iochaos/fs"
	"github.com/chaos-mesh/chaos-mesh/controllers/twophase"
)

type Reconciler struct {
	client.Client
	client.Reader
	record.EventRecorder
	Log logr.Logger
}

// Reconcile reconciles an IOChaos resource
func (r *Reconciler) Reconcile(req ctrl.Request, chaos *v1alpha1.IoChaos) (ctrl.Result, error) {
	r.Log.Info("Reconciling iochaos")
	scheduler := chaos.GetScheduler()
	duration, err := chaos.GetDuration()
	if err != nil {
		msg := fmt.Sprintf("unable to get iochaos[%s/%s]'s duration",
			req.Namespace, req.Name)
		r.Log.Error(err, msg)
		return ctrl.Result{}, err
	}
	if scheduler == nil && duration == nil {
		return r.commonIOChaos(chaos, req)
	} else if scheduler != nil && duration != nil {
		return r.scheduleIOChaos(chaos, req)
	}

	// This should be ensured by admission webhook in the future
	err = fmt.Errorf("iochaos[%s/%s] spec invalid", req.Namespace, req.Name)
	r.Log.Error(err, "scheduler and duration should be omitted or defined at the same time")
	return ctrl.Result{}, err
}

func (r *Reconciler) commonIOChaos(iochaos *v1alpha1.IoChaos, req ctrl.Request) (ctrl.Result, error) {
	var cr *common.Reconciler
	switch iochaos.Spec.Layer {
	case v1alpha1.FileSystemLayer:
		cr = fs.NewCommonReconciler(r.Client, r.Reader, r.Log.WithValues("reconciler", "chaosfs"),
			req, r.EventRecorder)
	default:
		return r.invalidActionResponse(iochaos)
	}
	return cr.Reconcile(iochaos, req)
}

func (r *Reconciler) scheduleIOChaos(iochaos *v1alpha1.IoChaos, req ctrl.Request) (ctrl.Result, error) {
	var sr *twophase.Reconciler
	switch iochaos.Spec.Layer {
	case v1alpha1.FileSystemLayer:
		sr = fs.NewTwoPhaseReconciler(r.Client, r.Reader, r.Log.WithValues("reconciler", "chaosfs"),
			req, r.EventRecorder)
	default:
		return r.invalidActionResponse(iochaos)
	}
	return sr.Reconcile(iochaos, req)
}

func (r *Reconciler) invalidActionResponse(iochaos *v1alpha1.IoChaos) (ctrl.Result, error) {
	r.Log.Error(nil, "unknown file system I/O layer", "action", iochaos.Spec.Action)
	return ctrl.Result{}, fmt.Errorf("unknown file system I/O layer")
}
