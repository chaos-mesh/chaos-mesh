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

package reconciler

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// InnerReconciler is interface for reconciler
type InnerReconciler interface {

	// Apply means the reconciler perform the chaos action
	Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error

	// Recover means the reconciler recovers the chaos action
	Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error

	// Object would return the instance of chaos
	Object() v1alpha1.InnerObject
}
