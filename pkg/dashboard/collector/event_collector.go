// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

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
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

// EventCollector represents a collector for Event Object.
type EventCollector struct {
	client.Client
	Log        logr.Logger
	apiType    runtime.Object
	eventStore core.EventStore
}

// Reconcile reconciles a Event collector.
func (r *EventCollector) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if err := ctx.Err(); err != nil {
		return ctrl.Result{}, err
	}
	if r.apiType == nil {
		r.Log.Error(nil, "apiType has not been initialized")
		return ctrl.Result{}, nil
	}

	event := &v1.Event{}
	err := r.Get(ctx, req.NamespacedName, event)
	if err := ctx.Err(); err != nil {
		return ctrl.Result{}, err
	}
	if err != nil {
		if !apierrors.IsNotFound(err) {
			r.Log.Error(err, "unable to get event")
			return retryCollector(ctx)
		}
		return ctrl.Result{}, nil
	}
	var target client.Object
	if chaosKind, ok := v1alpha1.AllKinds()[event.InvolvedObject.Kind]; ok {
		target = chaosKind.SpawnObject()
	} else {
		switch event.InvolvedObject.Kind {
		case v1alpha1.KindSchedule:
			target = &v1alpha1.Schedule{}
		case v1alpha1.KindWorkflow:
			target = &v1alpha1.Workflow{}
		case v1alpha1.KindWorkflowNode:
			target = &v1alpha1.WorkflowNode{}
		default:
			r.Log.Info("event collector: omitted event", "involved object name", event.InvolvedObject.Name, "involved object namespace", event.InvolvedObject.Namespace, "involved object kind", event.InvolvedObject.Kind)
			return ctrl.Result{}, nil
		}
	}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: event.InvolvedObject.Namespace,
		Name:      event.InvolvedObject.Name,
	}, target); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		r.Log.Error(err, "unable to get event target")
		return retryCollector(ctx)
	}

	et := core.Event{
		CreatedAt: event.CreationTimestamp.Time.UTC(),
		Kind:      event.InvolvedObject.Kind,
		Type:      event.Type,
		Reason:    event.Reason,
		Message:   event.Message,
		Name:      event.InvolvedObject.Name,
		Namespace: event.InvolvedObject.Namespace,
		ObjectID:  string(event.InvolvedObject.UID),
	}
	if uid := string(event.UID); uid != "" {
		et.SourceUID = &uid
	}
	if err := r.eventStore.Create(ctx, &et); err != nil {
		r.Log.Error(err, "failed to save event", "event", et)
		return retryCollector(ctx)
	}

	return ctrl.Result{}, nil
}

// Setup setups collectors by Manager.
func (r *EventCollector) Setup(mgr ctrl.Manager, apiType client.Object) error {
	r.apiType = apiType

	return ctrl.NewControllerManagedBy(mgr).
		For(apiType).
		WithEventFilter(eventCollectorFilter()).
		Complete(r)
}

func eventCollectorFilter() predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			event, ok := e.Object.(*v1.Event)
			if !ok {
				return false
			}

			_, ok = v1alpha1.AllKindsIncludeScheduleAndWorkflow()[event.InvolvedObject.Kind]
			return ok || event.InvolvedObject.Kind == v1alpha1.KindWorkflowNode
		},
		DeleteFunc: func(_ event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(_ event.UpdateEvent) bool {
			return false
		},
		GenericFunc: func(_ event.GenericEvent) bool {
			return false
		},
	}
}
