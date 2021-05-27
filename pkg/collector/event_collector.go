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
	"fmt"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/event"

	"github.com/go-logr/logr"

	"github.com/chaos-mesh/chaos-mesh/pkg/core"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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
	_ = r.Get(ctx, req.NamespacedName, event)
	fmt.Println(event.Message + "!!!!! " + event.InvolvedObject.Kind)

	return ctrl.Result{}, nil
}

// Setup setups collectors by Manager.
func (r *EventCollector) Setup(mgr ctrl.Manager, apiType runtime.Object) error {
	r.apiType = apiType

	return ctrl.NewControllerManagedBy(mgr).
		For(apiType).
		WithEventFilter(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldObj := e.ObjectOld.(*v1alpha1.PodIoChaos)
				newObj := e.ObjectNew.(*v1alpha1.PodIoChaos)

				return !reflect.DeepEqual(oldObj.Spec, newObj.Spec)
			},
		}).
		Complete(r)
}
