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
	"reflect"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"
)

type routeEntry struct {
	Name      string
	Object    runtime.Object
	Endpoints []routeEndpoint
}

func newEntry(name string, obj runtime.Object) *routeEntry {
	return &routeEntry{
		Name:      name,
		Object:    obj,
		Endpoints: []routeEndpoint{},
	}
}

type routeEndpoint struct {
	RouteFunc   func(runtime.Object) bool
	NewEndpoint endpoint.NewEndpoint
}

var routeTable map[reflect.Type]*routeEntry

var log logr.Logger

// Register registers an endpoint
func Register(name string, obj runtime.Object, routeFunc func(runtime.Object) bool, newEndpoint endpoint.NewEndpoint) {
	typ := reflect.TypeOf(obj)
	_, ok := routeTable[typ]
	if !ok {
		routeTable[typ] = newEntry(name, obj)
	}

	entry := routeTable[typ]
	if entry.Name != name {
		err := errors.Errorf("different names for one type of resource")
		log.Error(err, "different names for one type of resource", "name", name, "existingName", entry.Name)
	}

	entry.Endpoints = append(entry.Endpoints, routeEndpoint{
		RouteFunc:   routeFunc,
		NewEndpoint: newEndpoint,
	})
}

// SetupWithManagerAndConfigs setups reconciler with manager and controller configs
func SetupWithManagerAndConfigs(mgr ctrl.Manager, cfg *config.ChaosControllerConfig) error {
	for typ, end := range routeTable {
		log.Info("setup reconciler with manager", "type", typ, "endpoint", end.Name)
		reconciler := NewReconciler(end.Name, end.Object, mgr, end.Endpoints, cfg.ClusterScoped, cfg.TargetNamespace)
		err := reconciler.SetupWithManager(mgr)
		if err != nil {
			log.Error(err, "fail to setup reconciler with manager")

			return err
		}

		if err := ctrl.NewWebhookManagedBy(mgr).
			For(end.Object).
			Complete(); err != nil {
			log.Error(err, "fail to setup webhook")
			return err
		}
	}

	return nil
}

func init() {
	routeTable = make(map[reflect.Type]*routeEntry)
	log = ctrl.Log.WithName("router")
}
