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

package router

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/controllers/twophase"
	"github.com/chaos-mesh/chaos-mesh/pkg/events"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	end "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"
)

// Reconciler reconciles a chaos resource
type Reconciler struct {
	Name            string
	Object          runtime.Object
	Endpoints       []routeEndpoint
	ClusterScoped   bool
	TargetNamespace string

	ctx.Context
}

// Reconcile reconciles a chaos resource
func (r *Reconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	if !r.ClusterScoped && req.Namespace != r.TargetNamespace {
		// NOOP
		r.Log.Info("ignore chaos which belongs to an unexpected namespace within namespace scoped mode",
			"chaosName", req.Name, "expectedNamespace", r.TargetNamespace, "actualNamespace", req.Namespace)
		return ctrl.Result{}, nil
	}

	ctx := r.Context.LogWithValues("reconciler", r.Name, "resource name", req.NamespacedName)

	chaos, ok := r.Object.DeepCopyObject().(v1alpha1.InnerSchedulerObject)
	if !ok {
		err := errors.New("object is not InnerSchedulerObject")
		r.Log.Error(err, "object is not InnerSchedulerObject", "object", r.Object.DeepCopyObject())
		return ctrl.Result{}, err
	}

	if err := r.Client.Get(context.Background(), req.NamespacedName, chaos); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("chaos not found")
		} else {
			r.Log.Error(err, "unable to get chaos")
		}
		return ctrl.Result{}, nil
	}

	scheduler := chaos.GetScheduler()
	duration, err := chaos.GetDuration()
	if err != nil {
		r.Log.Error(err, fmt.Sprintf("unable to get chaos[%s/%s]'s duration", chaos.GetChaos().Namespace, chaos.GetChaos().Name))
		return ctrl.Result{}, err
	}

	var controller end.Endpoint
	for _, end := range r.Endpoints {
		if end.RouteFunc(chaos.(runtime.Object)) {
			controller = end.NewEndpoint(ctx)
		}
	}
	if controller == nil {
		err := errors.Errorf("cannot route object to one of the endpoint")
		r.Log.Error(err, "fail to route to endpoint", "object", chaos, "endpoints", r.Endpoints)
		return ctrl.Result{}, err
	}

	var reconciler reconcile.Reconciler
	if scheduler == nil && duration == nil {
		reconciler = common.NewReconciler(req, controller, ctx)
	} else if scheduler != nil {
		// scheduler != nil && duration != nil
		// but PodKill is an expection
		reconciler = twophase.NewReconciler(req, controller, ctx)
	} else {
		err := errors.Errorf("both scheduler and duration should be nil or not nil")
		r.Log.Error(err, "fail to construct reconciler", "scheduler", scheduler, "duration", duration)
		return ctrl.Result{}, err
	}

	result, err = reconciler.Reconcile(req)
	if err != nil {
		if chaos.IsDeleted() || chaos.IsPaused() {
			r.Event(chaos, v1.EventTypeWarning, events.ChaosRecoverFailed, err.Error())
		} else {
			r.Event(chaos, v1.EventTypeWarning, events.ChaosInjectFailed, err.Error())
		}
	}
	return result, nil
}

// NewReconciler creates a new reconciler
func NewReconciler(name string, object runtime.Object, mgr ctrl.Manager, endpoints []routeEndpoint, clusterScoped bool, targetNamespace string) *Reconciler {
	return &Reconciler{
		Name:            name,
		Object:          object,
		Endpoints:       endpoints,
		ClusterScoped:   clusterScoped,
		TargetNamespace: targetNamespace,

		Context: ctx.Context{
			Client:        mgr.GetClient(),
			Reader:        mgr.GetAPIReader(),
			EventRecorder: mgr.GetEventRecorderFor(name + "-controller"),
			Log:           ctrl.Log.WithName("controllers").WithName(name),
		},
	}
}

// SetupWithManager registers controller to manager
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := ctrl.NewControllerManagedBy(mgr).
		For(r.Object.DeepCopyObject()).
		Complete(r)

	if err != nil {
		return err
	}

	kind, err := apiutil.GVKForObject(r.Object.DeepCopyObject(), mgr.GetScheme())
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(r.Object.DeepCopyObject()).
		Named(strings.ToLower(kind.Kind) + "-scheduler-updater").
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(_ event.CreateEvent) bool {
				return false
			},
			DeleteFunc: func(_ event.DeleteEvent) bool {
				return false
			},
			GenericFunc: func(_ event.GenericEvent) bool {
				return false
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				old := e.ObjectOld.(v1alpha1.InnerSchedulerObject).GetScheduler()
				new := e.ObjectNew.(v1alpha1.InnerSchedulerObject).GetScheduler()

				if (old == nil) || (new == nil) {
					return false
				}

				return old.Cron != new.Cron
			},
		}).
		Complete(&twophase.SchedulerUpdater{
			Context: r.Context,
			Object:  r.Object,
		})
}
