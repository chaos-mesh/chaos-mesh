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
	"reflect"

	"k8s.io/client-go/tools/record"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector"
)

type InnerObjectWithCustomStatus interface {
	v1alpha1.InnerObject

	GetCustomStatus() interface{}
}

type InnerObjectWithSelector interface {
	v1alpha1.InnerObject

	GetSelectorSpecs() map[string]interface{}
}

type ChaosImpl interface {
	Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error)
	Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error)
}

// Reconciler for common chaos
type Reconciler struct {
	Impl ChaosImpl

	// Object is used to mark the target type of this Reconciler
	Object InnerObjectWithSelector

	// Client is used to operate on the Kubernetes cluster
	client.Client
	client.Reader

	Recorder record.EventRecorder

	Selector *selector.Selector

	Log logr.Logger
}

// Reconcile the common chaos
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	obj := r.Object.DeepCopyObject().(InnerObjectWithSelector)

	if err := r.Client.Get(context.TODO(), req.NamespacedName, obj); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("chaos not found")
		} else {
			// TODO: handle this error
			r.Log.Error(err, "unable to get chaos")
		}
		return ctrl.Result{}, nil
	}

	shouldUpdate := false

	desiredPhase := obj.GetStatus().Experiment.DesiredPhase
	records := obj.GetStatus().Experiment.Records
	selectors := obj.GetSelectorSpecs()

	if records == nil {
		for name, sel := range selectors {
			targets, err := r.Selector.Select(context.TODO(), sel)
			if err != nil {
				r.Log.Error(err, "fail to select")
				r.Recorder.Eventf(obj, "Warning", "Failed", "Failed to select targets: %s", err.Error())
				return ctrl.Result{}, nil
			}

			for _, target := range targets {
				records = append(records, &v1alpha1.Record{
					Id:          target.Id(),
					SelectorKey: name,
					Phase:       v1alpha1.NotInjected,
				})
				shouldUpdate = true
			}
		}
		// TODO: dynamic upgrade the records when some of these pods/containers stopped
	}

	for index, record := range records {
		var err error
		r.Log.Info("iterating record", "record", record, "desiredPhase", desiredPhase)

		// TODO: fix the cache invalidate problem
		// If the common controller is running, applying and modifying records,
		// during the same time, the desirephase controller updates the resource several
		// times. For example, the user update and remove the pause mark.
		// Then in the next reconcilation, the controller will found the records are not
		// updated, and apply again.
		// In this controller, it should at least get the latest records to keep safety.
		if desiredPhase == v1alpha1.RunningPhase && record.Phase != v1alpha1.Injected {
			originalPhase := record.Phase

			r.Log.Info("apply chaos", "id", records[index].Id)
			record.Phase, err = r.Impl.Apply(context.TODO(), index, records, obj)
			if record.Phase != originalPhase {
				shouldUpdate = true
			}
			if err != nil {
				// TODO: add backoff and retry mechanism
				// but the retry shouldn't block other resource process
				r.Log.Error(err, "fail to apply chaos")
				r.Recorder.Eventf(obj, "Warning", "Failed", "Failed to apply chaos: %s", err.Error())
				continue
			}

			r.Recorder.Eventf(obj, "Normal", "Applied", "Successfully apply chaos for %s", records[index].Id)
		}
		if desiredPhase == v1alpha1.StoppedPhase && record.Phase != v1alpha1.NotInjected {
			originalPhase := record.Phase

			r.Log.Info("recover chaos", "id", records[index].Id)
			record.Phase, err = r.Impl.Recover(context.TODO(), index, records, obj)
			if record.Phase != originalPhase {
				shouldUpdate = true
			}
			if err != nil {
				// TODO: add backoff and retry mechanism
				// but the retry shouldn't block other resource process
				r.Log.Error(err, "fail to recover chaos")
				r.Recorder.Eventf(obj, "Warning", "Failed", "Failed to recover chaos: %s", err.Error())
				continue
			}

			r.Recorder.Eventf(obj, "Normal", "Recovered", "Successfully recover chaos for %s", records[index].Id)
		}
	}

	// TODO: auto generate SetCustomStatus rather than reflect
	var customStatus reflect.Value
	if objWithStatus, ok := obj.(InnerObjectWithCustomStatus); ok {
		customStatus = reflect.Indirect(reflect.ValueOf(objWithStatus.GetCustomStatus()))
	}
	if shouldUpdate {
		updateError := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			r.Log.Info("updating records", "records", records)
			obj := r.Object.DeepCopyObject().(InnerObjectWithSelector)

			if err := r.Client.Get(context.TODO(), req.NamespacedName, obj); err != nil {
				r.Log.Error(err, "unable to get chaos")
				return err
			}

			obj.GetStatus().Experiment.Records = records
			if objWithStatus, ok := obj.(InnerObjectWithCustomStatus); ok {
				ptrToCustomStatus := objWithStatus.GetCustomStatus()
				// TODO: auto generate SetCustomStatus rather than reflect
				reflect.Indirect(reflect.ValueOf(ptrToCustomStatus)).Set(reflect.Indirect(customStatus))
			}
			return r.Client.Update(context.TODO(), obj)
		})
		if updateError != nil {
			r.Log.Error(updateError, "fail to update")
			r.Recorder.Eventf(obj, "Normal", "Failed", "Failed to update records: %s", updateError.Error())
			return ctrl.Result{}, nil
		}

		r.Recorder.Event(obj, "Normal", "Updated", "Successfully update records of resource")
	}
	return ctrl.Result{}, nil
}
