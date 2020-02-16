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

package controllers

import (
	"github.com/go-logr/logr"

	chaosmeshv1alpha1 "github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/controllers/iochaos"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// IoChaosReconciler reconciles a IoChaos object
type IoChaosReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=chaosmesh.pingcap.com,resources=iochaos,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=chaosmesh.pingcap.com,resources=iochaos/status,verbs=get;update;patch

func (r *IoChaosReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("iochaos", req.NamespacedName)

	reconciler := iochaos.Reconciler{
		Client: r.Client,
		Log:    logger,
	}

	return reconciler.Reconcile(req)
}

func (r *IoChaosReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&chaosmeshv1alpha1.IoChaos{}).
		Complete(r)
}
