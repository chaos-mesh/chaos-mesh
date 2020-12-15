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

package trafficcontrol

import (
	"context"
	"errors"

	"github.com/hashicorp/go-multierror"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/controllers/networkchaos/podnetworkchaosmanager"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"
)

// Recover means the reconciler recovers the chaos action
func (r *endpoint) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	somechaos, ok := chaos.(*v1alpha1.NetworkChaos)
	if !ok {
		err := errors.New("chaos is not NetworkChaos")
		r.Log.Error(err, "chaos is not NetworkChaos", "chaos", chaos)
		return err
	}

	if err := r.cleanFinalizersAndRecover(ctx, somechaos); err != nil {
		return err
	}
	r.Event(somechaos, v1.EventTypeNormal, utils.EventChaosRecovered, "")

	return nil
}

func (r *endpoint) cleanFinalizersAndRecover(ctx context.Context, chaos *v1alpha1.NetworkChaos) error {
	var result error

	source := chaos.Namespace + "/" + chaos.Name
	m := podnetworkchaosmanager.New(source, r.Log, r.Client)

	for _, key := range chaos.Finalizers {
		ns, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		_ = m.WithInit(types.NamespacedName{
			Namespace: ns,
			Name:      name,
		})
	}
	responses := m.Commit(ctx)
	for _, response := range responses {
		key := response.Key
		err := response.Err
		// if pod not found or not running, directly return and giveup recover.
		if err != nil {
			if err != podnetworkchaosmanager.ErrPodNotFound && err != podnetworkchaosmanager.ErrPodNotRunning {
				r.Log.Error(err, "fail to commit", "key", key)

				result = multierror.Append(result, err)
				continue
			}

			r.Log.Info("pod is not found or not running", "key", key)
		}

		chaos.Finalizers = utils.RemoveFromFinalizer(chaos.Finalizers, response.Key.String())
	}
	r.Log.Info("After recovering", "finalizers", chaos.Finalizers)

	if chaos.Annotations[common.AnnotationCleanFinalizer] == common.AnnotationCleanFinalizerForced {
		r.Log.Info("Force cleanup all finalizers", "chaos", chaos)
		chaos.Finalizers = make([]string, 0)
		return nil
	}

	return result
}
