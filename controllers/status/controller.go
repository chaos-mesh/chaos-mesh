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

package status

import (
	"context"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// This `Reconciler` set the `.Status.Experiment.Phase` field of an object
// this field is only used in dashboard to display the status of chaos.
// Never depend on the `.Status.Experiment.Phase` value, it's not stable.

type Reconciler struct {
	// Object is used to mark the target type of this Reconciler
	Object v1alpha1.InnerObject

	// Client is used to operate on the Kubernetes cluster
	client.Client
	client.Reader

	Log logr.Logger
}

func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	obj := r.Object.DeepCopyObject().(v1alpha1.InnerObject)

	if err := r.Client.Get(context.TODO(), req.NamespacedName, obj); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("chaos not found")
		} else {
			// TODO: handle this error
			r.Log.Error(err, "unable to get chaos")
		}
		return ctrl.Result{}, nil
	}

	status := obj.GetStatus()
	phase := status.Experiment.Phase
	if obj.IsPaused() {
		phase = v1alpha1.ExperimentPhasePaused
	}

	for _, record := range status.Experiment.Records {
		if record.Phase == v1alpha1.Injected {
			phase = v1alpha1.ExperimentPhaseRunning
		}
	}

	if phase != status.Experiment.Phase {
		updateError := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			obj := r.Object.DeepCopyObject().(v1alpha1.InnerObject)

			if err := r.Client.Get(context.TODO(), req.NamespacedName, obj); err != nil {
				r.Log.Error(err, "unable to get chaos")
				return err
			}

			obj.GetStatus().Experiment.Phase = phase
			return r.Client.Update(context.TODO(), obj)
		})
		if updateError != nil {
			// TODO: handle this error
			r.Log.Error(updateError, "fail to update")
		}
	}

	return ctrl.Result{}, nil
}
