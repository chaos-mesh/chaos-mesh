// Copyright 2020 Chaos Mesh Authors.
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
	"context"

	"github.com/chaos-mesh/chaos-mesh/controllers/common"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/timechaos"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"
)

// TimeChaosReconciler reconciles a TimeChaos object
type TimeChaosReconciler struct {
	client.Client
	client.Reader
	record.EventRecorder
	Log logr.Logger
}

// +kubebuilder:rbac:groups=chaos-mesh.org,resources=timechaos,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=chaos-mesh.org,resources=timechaos/status,verbs=get;update;patch

// Reconcile reconciles a TimeChaos resource
func (r *TimeChaosReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	logger := r.Log.WithValues("reconciler", "timechaos")

	if !common.ControllerCfg.ClusterScoped && req.Namespace != common.ControllerCfg.TargetNamespace {
		// NOOP
		logger.Info("ignore chaos which belongs to an unexpected namespace within namespace scoped mode",
			"chaosName", req.Name, "expectedNamespace", common.ControllerCfg.TargetNamespace, "actualNamespace", req.Namespace)
		return ctrl.Result{}, nil
	}

	reconciler := timechaos.Reconciler{
		Client:        r.Client,
		Reader:        r.Reader,
		EventRecorder: r.EventRecorder,
		Log:           logger,
	}

	chaos := &v1alpha1.TimeChaos{}
	if err := r.Client.Get(context.Background(), req.NamespacedName, chaos); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("time chaos not found")
		} else {
			r.Log.Error(err, "unable to get time chaos")
		}
		return ctrl.Result{}, nil
	}

	result, err = reconciler.Reconcile(req, chaos)
	if err != nil {
		if chaos.IsDeleted() || chaos.IsPaused() {
			r.Event(chaos, v1.EventTypeWarning, utils.EventChaosRecoverFailed, err.Error())
		} else {
			r.Event(chaos, v1.EventTypeWarning, utils.EventChaosInjectFailed, err.Error())
		}
	}

	return result, nil
}

// SetupWithManager setups a time chaos reconciler on controller-manager
func (r *TimeChaosReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.TimeChaos{}).
		Complete(r)
}
