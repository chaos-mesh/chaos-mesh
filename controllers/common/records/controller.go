// Copyright 2021 Chaos Mesh Authors.
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

package records

import (
	"context"
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector"
)

// Reconciler for chaos records
type Reconciler struct {
	Impl types.ChaosImpl

	// Object is used to mark the target type of this Reconciler
	Object v1alpha1.InnerObject

	// Client is used to operate on the Kubernetes cluster
	client.Client
	client.Reader

	Recorder recorder.ChaosRecorder

	Selector *selector.Selector

	Log logr.Logger
}

type Operation string

const (
	Apply   Operation = "apply"
	Recover Operation = "recover"
	Nothing Operation = ""
)

// Reconcile the chaos records
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	obj := r.Object.DeepCopyObject().(v1alpha1.InnerObjectWithSelector)
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

	logger := r.Log.WithValues("name", obj.GetName(), "namespace", obj.GetNamespace(), "kind", obj.GetObjectKind().GroupVersionKind().Kind)

	if records == nil {
		for name, sel := range selectors {
			targets, err := r.Selector.Select(context.TODO(), sel)
			if err != nil {
				logger.Error(err, "fail to select")
				r.Recorder.Event(obj, recorder.Failed{
					Activity: "select targets",
					Err:      err.Error(),
				})
				return ctrl.Result{}, nil
			}

			if len(targets) == 0 {
				logger.Info("no target has been selected")
				r.Recorder.Event(obj, recorder.Failed{
					Activity: "select targets",
					Err:      "no target has been selected",
				})
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

	needRetry := false
	for index, record := range records {
		var err error
		idLogger := logger.WithValues("id", records[index].Id)
		idLogger.V(1).Info("iterating record", "record", record, "desiredPhase", desiredPhase)

		// The whole running logic is a cycle:
		// Not Injected -> Not Injected/* -> Injected -> Injected/* -> Not Injected
		// Every step should follow the cycle. For example, if it's in "Not Injected/*" status, and it wants to recover
		// then it has to apply and then recover, but not recover directly.

		originalPhase := record.Phase
		operation := Nothing
		if desiredPhase == v1alpha1.RunningPhase && originalPhase != v1alpha1.Injected {
			// The originalPhase has three possible situations: Not Injected, Not Injected/* or Injected/*
			// In the first two situations, it should apply, in the last situation, it should recover

			if strings.HasPrefix(string(originalPhase), string(v1alpha1.NotInjected)) {
				operation = Apply
			} else {
				operation = Recover
			}
		}
		if desiredPhase == v1alpha1.StoppedPhase && originalPhase != v1alpha1.NotInjected {
			// The originalPhase has three possible situations: Not Injected/*, Injected, or Injected/*
			// In the first one situation, it should apply, in the last two situations, it should recover

			if strings.HasPrefix(string(originalPhase), string(v1alpha1.NotInjected)) {
				operation = Apply
			} else {
				operation = Recover
			}
		}

		if operation == Apply {
			idLogger.Info("apply chaos")
			record.Phase, err = r.Impl.Apply(context.TODO(), index, records, obj)
			if record.Phase != originalPhase {
				shouldUpdate = true
			}
			if err != nil {
				// TODO: add backoff and retry mechanism
				// but the retry shouldn't block other resource process
				idLogger.Error(err, "fail to apply chaos")
				applyFailedEvent := newRecordEvent(v1alpha1.TypeFailed, v1alpha1.Apply, err.Error())
				records[index].Events = append(records[index].Events, *applyFailedEvent)
				r.Recorder.Event(obj, recorder.Failed{
					Activity: "apply chaos",
					Err:      err.Error(),
				})
				needRetry = true
				// if the impl.Apply() failed, we need to update the status to update the records[index].Events
				shouldUpdate = true
				continue
			}

			if record.Phase == v1alpha1.Injected {
				records[index].InjectedCount++
				applySucceedEvent := newRecordEvent(v1alpha1.TypeSucceeded, v1alpha1.Apply, "")
				records[index].Events = append(records[index].Events, *applySucceedEvent)
				r.Recorder.Event(obj, recorder.Applied{
					Id: records[index].Id,
				})
			}
		} else if operation == Recover {
			idLogger.Info("recover chaos")
			record.Phase, err = r.Impl.Recover(context.TODO(), index, records, obj)
			if record.Phase != originalPhase {
				shouldUpdate = true
			}
			if err != nil {
				// TODO: add backoff and retry mechanism
				// but the retry shouldn't block other resource process
				idLogger.Error(err, "fail to recover chaos")
				recoverFailedEvent := newRecordEvent(v1alpha1.TypeFailed, v1alpha1.Recover, err.Error())
				records[index].Events = append(records[index].Events, *recoverFailedEvent)
				r.Recorder.Event(obj, recorder.Failed{
					Activity: "recover chaos",
					Err:      err.Error(),
				})
				needRetry = true
				// if the impl.Recover() failed, we need to update the status to update the records[index].Events
				shouldUpdate = true
				continue
			}

			if record.Phase == v1alpha1.NotInjected {
				records[index].RecoveredCount++
				recoverSucceedEvent := newRecordEvent(v1alpha1.TypeSucceeded, v1alpha1.Recover, "")
				records[index].Events = append(records[index].Events, *recoverSucceedEvent)
				r.Recorder.Event(obj, recorder.Recovered{
					Id: records[index].Id,
				})
			}
		}
	}

	// TODO: auto generate SetCustomStatus rather than reflect
	var customStatus reflect.Value
	if objWithStatus, ok := obj.(v1alpha1.InnerObjectWithCustomStatus); ok {
		customStatus = reflect.Indirect(reflect.ValueOf(objWithStatus.GetCustomStatus()))
	}
	if shouldUpdate {
		updateError := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			logger.V(1).Info("updating records", "records", records)
			obj := r.Object.DeepCopyObject().(v1alpha1.InnerObjectWithSelector)

			if err := r.Client.Get(context.TODO(), req.NamespacedName, obj); err != nil {
				logger.Error(err, "unable to get chaos")
				return err
			}

			obj.GetStatus().Experiment.Records = records
			if objWithStatus, ok := obj.(v1alpha1.InnerObjectWithCustomStatus); ok {
				ptrToCustomStatus := objWithStatus.GetCustomStatus()
				// TODO: auto generate SetCustomStatus rather than reflect
				reflect.Indirect(reflect.ValueOf(ptrToCustomStatus)).Set(reflect.Indirect(customStatus))
			}
			return r.Client.Update(context.TODO(), obj)
		})
		if updateError != nil {
			logger.Error(updateError, "fail to update")
			r.Recorder.Event(obj, recorder.Failed{
				Activity: "update records",
				Err:      updateError.Error(),
			})
			return ctrl.Result{Requeue: true}, nil
		}

		r.Recorder.Event(obj, recorder.Updated{
			Field: "records",
		})
	}
	return ctrl.Result{Requeue: needRetry}, nil
}

func newRecordEvent(eventType v1alpha1.RecordEventType, eventStage v1alpha1.RecordEventOperation, msg string) *v1alpha1.RecordEvent {
	return v1alpha1.NewRecordEvent(eventType, eventStage, msg, metav1.Now())
}
