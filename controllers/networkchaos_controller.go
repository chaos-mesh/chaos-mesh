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

package controllers

import (
	"context"

	"github.com/chaos-mesh/chaos-mesh/controllers/common"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/networkchaos"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"
)

// NetworkChaosReconciler reconciles a NetworkChaos object
type NetworkChaosReconciler struct {
	client.Client
	client.Reader
	record.EventRecorder
	Log logr.Logger
}

// +kubebuilder:rbac:groups=chaos-mesh.org,resources=networkchaos,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=chaos-mesh.org,resources=networkchaos/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;watch;list

// Reconcile reconciles a NetworkChaos resource
func (r *NetworkChaosReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	logger := r.Log.WithValues("reconciler", "networkchaos")

	if !common.ControllerCfg.ClusterScoped && req.Namespace != common.ControllerCfg.TargetNamespace {
		// NOOP
		logger.Info("ignore chaos which belongs to an unexpected namespace within namespace scoped mode",
			"chaosName", req.Name, "expectedNamespace", common.ControllerCfg.TargetNamespace, "actualNamespace", req.Namespace)
		return ctrl.Result{}, nil
	}

	reconciler := networkchaos.Reconciler{
		Client:        r.Client,
		Reader:        r.Reader,
		EventRecorder: r.EventRecorder,
		Log:           logger,
	}

	chaos := &v1alpha1.NetworkChaos{}
	if err := r.Client.Get(context.Background(), req.NamespacedName, chaos); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("network chaos not found")
		} else {
			r.Log.Error(err, "unable to get network chaos")
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

// SetupWithManager setup networkchaos reconciler which called by controller-manager
func (r *NetworkChaosReconciler) SetupWithManager(mgr ctrl.Manager) error {
	podToChaosMapFn := handler.ToRequestsFunc(
		func(a handler.MapObject) []reconcile.Request {
			reqs := []reconcile.Request{}

			_, ok := a.Object.(*v1.Pod)
			if !ok {
				return reqs
			}

			associateChaos, err := utils.SelectAndFilterChaosByPod(context.Background(), r.Client, r.Reader, nil, &v1alpha1.NetworkChaosList{})
			if err != nil {
				r.Log.Error(err, "error filter chaos by pod")
				return reqs
			}

			for _, chaos := range associateChaos {
				reqs = append(reqs, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      chaos.GetChaos().Name,
						Namespace: chaos.GetChaos().Namespace,
					},
				})
				r.Log.Info("issued chaos reconcile request", "chaos", chaos.GetChaos().Name)
			}

			return reqs
		})

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.NetworkChaos{}).
		Watches(&source.Kind{Type: &v1.Pod{}}, &handler.EnqueueRequestsFromMapFunc{
			ToRequests: podToChaosMapFn,
		}). // NOTE: we need to subscribe pod events to sync networkchaos
		Complete(r)
}
