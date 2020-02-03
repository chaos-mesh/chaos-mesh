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

package podfailure

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/controllers/duration"
	"github.com/pingcap/chaos-mesh/controllers/podchaos/podfailure"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DurationReconciler is reconciler for podfailure
type DurationReconciler struct {
	*podfailure.Reconciler
}

// NewDurationReconciler would create reconciler for duration chaos
func NewDurationReconciler(c client.Client, log logr.Logger, req ctrl.Request) *DurationReconciler {
	r := &podfailure.Reconciler{
		Client: c,
		Log:    log,
	}
	return &DurationReconciler{
		Reconciler: r,
	}
}

// Apply would perform duration chaos for podchaos
func (r *DurationReconciler) Apply(ctx context.Context, req ctrl.Request, chaos duration.InnerDurationObject) error {
	podChaos, ok := chaos.(*v1alpha1.PodChaos)
	if !ok {
		err := errors.New("chaos is not PodChaos")
		r.Log.Error(err, "chaos is not PodChaos", "chaos", chaos)
		return err
	}
	return r.Perform(ctx, req, podChaos)
}

// Recover would recover the duration chaos for podchaos
func (r *DurationReconciler) Recover(ctx context.Context, req ctrl.Request, chaos duration.InnerDurationObject) error {
	podChaos, ok := chaos.(*v1alpha1.PodChaos)
	if !ok {
		err := errors.New("chaos is not PodChaos")
		r.Log.Error(err, "chaos is not PodChaos", "chaos", chaos)
		return err
	}
	return r.Clean(ctx, req, podChaos)
}

// Object implement duration.Object
func (r *DurationReconciler) Object() duration.InnerDurationObject {
	return r.Instance()
}
