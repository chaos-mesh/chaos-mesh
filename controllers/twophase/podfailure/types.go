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
	"github.com/pingcap/chaos-mesh/controllers/podchaos/podfailure"
	"github.com/pingcap/chaos-mesh/controllers/twophase"

	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewTwoPhaseReconciler would Create TwoPhaseReconciler
func NewTwoPhaseReconciler(c client.Client, log logr.Logger, req ctrl.Request) *TwoPhaseReconciler {
	r := &podfailure.Reconciler{
		Client: c,
		Log:    log,
	}
	return &TwoPhaseReconciler{
		Reconciler: r,
	}
}

// TwoPhaseReconciler reconcile the networkchaos
type TwoPhaseReconciler struct {
	*podfailure.Reconciler
}

// Apply implement InnerReconciler.Apply
func (r *TwoPhaseReconciler) Apply(ctx context.Context, req ctrl.Request, chaos twophase.InnerObject) error {
	podchaos, ok := chaos.(*v1alpha1.PodChaos)
	if !ok {
		err := errors.New("chaos is not PodChaos")
		r.Log.Error(err, "chaos is not PodChaos", "chaos", chaos)
		return err
	}
	return r.Perform(ctx, req, podchaos)
}

// Recover implement InnerReconciler.Recover
func (r *TwoPhaseReconciler) Recover(ctx context.Context, req ctrl.Request, chaos twophase.InnerObject) error {
	podchaos, ok := chaos.(*v1alpha1.PodChaos)
	if !ok {
		err := errors.New("chaos is not PodChaos")
		r.Log.Error(err, "chaos is not PodChaos", "chaos", chaos)
		return err
	}
	return r.Clean(ctx, req, podchaos)
}

// Object implement InnerReconciler.Object
func (r *TwoPhaseReconciler) Object() twophase.InnerObject {
	return r.Instance()
}
