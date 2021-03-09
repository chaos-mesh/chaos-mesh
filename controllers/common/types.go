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

package common

import (
	"context"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	endpoint "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"

	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	// AnnotationCleanFinalizer key
	AnnotationCleanFinalizer = `chaos-mesh.chaos-mesh.org/cleanFinalizer`
	// AnnotationCleanFinalizerForced value
	AnnotationCleanFinalizerForced = `forced`
)

const emptyString = ""

// Reconciler for common chaos
type Reconciler struct {
	endpoint.Endpoint
	ctx.Context
}

// Reconcile the common chaos
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var err error

	r.Log.Info("Reconciling a common chaos", "name", req.Name, "namespace", req.Namespace)
	ctx := context.Background()

	chaos := r.Object()
	if err = r.Client.Get(ctx, req.NamespacedName, chaos); err != nil {
		r.Log.Error(err, "unable to get chaos")
		return ctrl.Result{}, err
	}

	if err := r.Update(ctx, chaos); err != nil {
		r.Log.Error(err, "unable to update chaos status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
