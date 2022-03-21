// Copyright Chaos Mesh Authors.
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

package statuscheck

import (
	"context"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconciler applys statuscheck
type Reconciler struct {
	kubeClient    client.Client
	eventRecorder recorder.ChaosRecorder

	manager Manager

	logger logr.Logger
}

func NewReconciler(kubeClient client.Client, logger logr.Logger, eventRecorder recorder.ChaosRecorder) *Reconciler {
	return &Reconciler{
		kubeClient:    kubeClient,
		eventRecorder: eventRecorder,
		logger:        logger,
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	obj := &v1alpha1.StatusCheck{}
	if err := r.kubeClient.Get(ctx, req.NamespacedName, obj); err != nil {
		if apierrors.IsNotFound(err) {
			// TODO remove from manager
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// TODO get status from manager
	// TODO if nil, add status check to manager

	return ctrl.Result{}, nil
}
