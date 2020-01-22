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

package netem

import (
	"context"
	"errors"
	"github.com/go-logr/logr"
	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	networkchaosNetem "github.com/pingcap/chaos-mesh/controllers/networkchaos/netem"
	"github.com/pingcap/chaos-mesh/controllers/persist"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PersistentReconciler is reconciler for netem
type PersistentReconciler struct {
	*networkchaosNetem.Reconciler
}

// NewPersistentReconciler would create PersistentReconciler
func NewPersistentReconciler(c client.Client, log logr.Logger, req ctrl.Request) *PersistentReconciler {
	r := &networkchaosNetem.Reconciler{
		Client: c,
		Log:    log,
	}
	return &PersistentReconciler{
		Reconciler: r,
	}
}

// implement persist.Apply
func (r *PersistentReconciler) Apply(ctx context.Context, req ctrl.Request, chaos persist.InnerPersistObject) error {
	networkchaos, ok := chaos.(*v1alpha1.NetworkChaos)
	if !ok {
		err := errors.New("chaos is not NetworkChaos")
		r.Log.Error(err, "chaos is not NetworkChaos", "chaos", chaos)
		return err
	}
	return r.Perform(ctx, req, networkchaos)
}

// implement persist.Recover
func (r *PersistentReconciler) Recover(ctx context.Context, req ctrl.Request, chaos persist.InnerPersistObject) error {
	networkchaos, ok := chaos.(*v1alpha1.NetworkChaos)
	if !ok {
		err := errors.New("chaos is not NetworkChaos")
		r.Log.Error(err, "chaos is not NetworkChaos", "chaos", chaos)
		return err
	}
	return r.Clean(ctx, req, networkchaos)
}

// implement persist.Object
func (r *PersistentReconciler) Object() persist.InnerPersistObject {
	return r.Instance()
}
