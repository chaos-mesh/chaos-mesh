// Copyright 2021 Chaos Mesh Authors.
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

package collector

import (
	"context"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/core"
)

// EventCollector represents a collector for Event Object.
type EventCollector struct {
	client.Client
	Log     logr.Logger
	apiType runtime.Object
	event   core.EventStore
}

// Reconcile reconciles a Event collector.
func (r *EventCollector) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	if r.apiType == nil {
		r.Log.Error(nil, "apiType has not been initialized")
		return ctrl.Result{}, nil
	}
	ctx := context.Background()

	event := &v1.Event{}
	err := r.Get(ctx, req.NamespacedName, event)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			r.Log.Error(err, "unable to get event")
		}
		return ctrl.Result{}, nil
	}

	chaosKind, ok := v1alpha1.AllKinds()[event.InvolvedObject.Kind]
	if ok {
		if err = r.Get(ctx, types.NamespacedName{
			Namespace: event.InvolvedObject.Namespace,
			Name:      event.InvolvedObject.Name,
		}, chaosKind.Chaos); err != nil {
			return ctrl.Result{}, nil
		}
	} else {
		if err = r.Get(ctx, types.NamespacedName{
			Namespace: event.InvolvedObject.Namespace,
			Name:      event.InvolvedObject.Name,
		}, &v1alpha1.Schedule{}); err != nil {
			return ctrl.Result{}, nil
		}
	}

	et := core.Event{
		CreatedAt: event.CreationTimestamp.Time,
		Kind:      event.InvolvedObject.Kind,
		Type:      event.Type,
		Reason:    event.Reason,
		Message:   event.Message,
		Name:      event.InvolvedObject.Name,
		Namespace: event.InvolvedObject.Namespace,
		ObjectID:  string(event.InvolvedObject.UID),
	}
	if err := r.event.Create(context.Background(), &et); err != nil {
		r.Log.Error(err, "failed to save event", "event", et)
	}

	return ctrl.Result{}, nil
}

// Setup setups collectors by Manager.
func (r *EventCollector) Setup(mgr ctrl.Manager, apiType runtime.Object) error {
	r.apiType = apiType

	return ctrl.NewControllerManagedBy(mgr).
		For(apiType).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				event, ok := e.Object.(*v1.Event)
				if !ok {
					return false
				}
				flag := false
				_, ok = v1alpha1.AllKinds()[event.InvolvedObject.Kind]
				if ok {
					flag = true
				}
				if event.InvolvedObject.Kind == v1alpha1.KindSchedule {
					flag = true
				}
				return flag

			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return false
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				return false
			},
			GenericFunc: func(e event.GenericEvent) bool {
				return false
			},
		}).
		Complete(r)
}
