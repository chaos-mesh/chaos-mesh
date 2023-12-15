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

package desiredphase

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
)

// Reconciler for common chaos
type Reconciler struct {
	// Object is used to mark the target type of this Reconciler
	Object v1alpha1.InnerObject

	// Client is used to operate on the Kubernetes cluster
	client.Client

	Recorder recorder.ChaosRecorder
	Log      logr.Logger
}

// Reconcile the common chaos
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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

	reconcileInfo := &reconcileInfo{
		obj:          obj,
		Reconciler:   r,
		shouldUpdate: false,
	}
	return reconcileInfo.Reconcile(req)
}

type reconcileInfo struct {
	obj v1alpha1.InnerObject

	*Reconciler
	shouldUpdate bool
	requeueAfter time.Duration
}

func (info *reconcileInfo) GetCreationTimestamp() metav1.Time {
	return info.obj.GetCreationTimestamp()
}

func (info *reconcileInfo) CalcDesiredPhase() (v1alpha1.DesiredPhase, []recorder.ChaosEvent) {
	events := []recorder.ChaosEvent{}

	// Consider the finalizers
	if info.obj.IsDeleted() {
		if info.obj.GetStatus().Experiment.DesiredPhase != v1alpha1.StoppedPhase {
			events = append(events, recorder.Deleted{})
		}
		return v1alpha1.StoppedPhase, events
	}

	if info.obj.IsOneShot() {
		// An oneshot chaos should always be in running phase, so that it cannot
		// be applied multiple times or cause other bugs :(
		return v1alpha1.RunningPhase, events
	}

	// Consider the duration
	now := time.Now()

	durationExceeded, untilStop, err := info.obj.DurationExceeded(now)
	if err != nil {
		info.Log.Error(err, "failed to parse duration")
	}
	if durationExceeded {
		if info.obj.GetStatus().Experiment.DesiredPhase != v1alpha1.StoppedPhase {
			events = append(events, recorder.TimeUp{})
		}
		return v1alpha1.StoppedPhase, events
	}

	info.requeueAfter = untilStop

	// Then decide the pause logic
	if info.obj.IsPaused() {
		if info.obj.GetStatus().Experiment.DesiredPhase != v1alpha1.StoppedPhase {
			events = append(events, recorder.Paused{})
		}
		return v1alpha1.StoppedPhase, events
	}

	if info.obj.GetStatus().Experiment.DesiredPhase != v1alpha1.RunningPhase {
		events = append(events, recorder.Started{})
	}
	return v1alpha1.RunningPhase, events
}

func (info *reconcileInfo) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	desiredPhase, events := info.CalcDesiredPhase()

	info.Log.V(1).Info("modify desiredPhase", "desiredPhase", desiredPhase)
	if info.obj.GetStatus().Experiment.DesiredPhase != desiredPhase {
		for _, ev := range events {
			info.Recorder.Event(info.obj, ev)
		}

		updateError := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			obj := info.Object.DeepCopyObject().(v1alpha1.InnerObject)

			if err := info.Client.Get(context.TODO(), req.NamespacedName, obj); err != nil {
				info.Log.Error(err, "unable to get chaos")
				return err
			}

			if obj.GetStatus().Experiment.DesiredPhase != desiredPhase {
				obj.GetStatus().Experiment.DesiredPhase = desiredPhase
				info.Log.V(1).Info("update object", "namespace", obj.GetNamespace(), "name", obj.GetName())
				return info.Client.Update(context.TODO(), obj)
			}

			return nil
		})
		if updateError != nil {
			info.Log.Error(updateError, "fail to update")
			info.Recorder.Event(info.obj, recorder.Failed{
				Activity: "update desiredphase",
				Err:      updateError.Error(),
			})
			return ctrl.Result{}, nil
		}

		info.Recorder.Event(info.obj, recorder.Updated{
			Field: "desiredPhase",
		})
	}
	return ctrl.Result{RequeueAfter: info.requeueAfter}, nil
}
