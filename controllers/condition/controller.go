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

package condition

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
)

// Reconciler for common chaos
type Reconciler struct {
	// Object is used to mark the target type of this Reconciler
	Object runtime.Object

	// Client is used to operate on the Kubernetes cluster
	client.Client

	Recorder record.EventRecorder

	Log logr.Logger
}

type StatusAndReason struct {
	Status corev1.ConditionStatus
	Reason string
}

func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	obj := r.Object.DeepCopyObject().(common.InnerObjectWithSelector)
	if err := r.Client.Get(context.TODO(), req.NamespacedName, obj); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("chaos not found")
		} else {
			// TODO: handle this error
			r.Log.Error(err, "unable to get chaos")
		}
		return ctrl.Result{}, nil
	}

	updateError := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		conditionMap := make(map[v1alpha1.ChaosConditionType]StatusAndReason)
		for _, c := range obj.GetStatus().Conditions {
			conditionMap[c.Type] = StatusAndReason{
				Status: c.Status,
				Reason: c.Reason,
			}
		}

		newConditionMap := make(map[v1alpha1.ChaosConditionType]StatusAndReason)
		if obj.GetStatus().Experiment.Records != nil {
			newConditionMap[v1alpha1.ConditionSelected] = StatusAndReason{
				Status: corev1.ConditionTrue,
			}
		} else {
			newConditionMap[v1alpha1.ConditionSelected] = StatusAndReason{
				Status: corev1.ConditionFalse,
			}
		}

		allInjected := corev1.ConditionTrue
		allRecovered := corev1.ConditionTrue
		for _, record := range obj.GetStatus().Experiment.Records {
			if record.Phase != v1alpha1.NotInjected {
				allRecovered = corev1.ConditionFalse
			}

			if record.Phase != v1alpha1.Injected {
				allInjected = corev1.ConditionFalse
			}
		}
		newConditionMap[v1alpha1.ConditionAllInjected] = StatusAndReason{
			Status: allInjected,
		}
		newConditionMap[v1alpha1.ConditionAllRecovered] = StatusAndReason{
			Status: allRecovered,
		}

		if obj.IsPaused() {
			newConditionMap[v1alpha1.ConditionPaused] = StatusAndReason{
				Status: corev1.ConditionTrue,
			}
		} else {
			newConditionMap[v1alpha1.ConditionPaused] = StatusAndReason{
				Status: corev1.ConditionFalse,
			}
		}

		if !reflect.DeepEqual(newConditionMap, conditionMap) {
			conditions := make([]v1alpha1.ChaosCondition, 0, 5)
			for k, v := range newConditionMap {
				conditions = append(conditions, v1alpha1.ChaosCondition{
					Type:   k,
					Status: v.Status,
					Reason: v.Reason,
				})
			}

			r.Log.Info("updating conditions", "conditions", conditions)
			obj := r.Object.DeepCopyObject().(common.InnerObjectWithSelector)

			if err := r.Client.Get(context.TODO(), req.NamespacedName, obj); err != nil {
				r.Log.Error(err, "unable to get chaos")
				return err
			}

			obj.GetStatus().Conditions = conditions
			return r.Client.Update(context.TODO(), obj)
		}

		return nil
	})

	if updateError != nil {
		r.Log.Error(updateError, "fail to update")
		r.Recorder.Eventf(obj, "Normal", "Failed", "Failed to update conditions: %s", updateError.Error())
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}
