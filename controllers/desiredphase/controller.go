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

	ctx := &reconcileContext{
		obj:          obj,
		Reconciler:   r,
		shouldUpdate: false,
	}
	return ctx.Reconcile(req)
}

type reconcileContext struct {
	obj v1alpha1.InnerObject

	*Reconciler
	shouldUpdate bool
	requeueAfter time.Duration
}

func (ctx *reconcileContext) GetCreationTimestamp() metav1.Time {
	return ctx.obj.GetObjectMeta().CreationTimestamp
}

func (ctx *reconcileContext) CalcDesiredPhase() (v1alpha1.DesiredPhase, []recorder.ChaosEvent) {
	events := []recorder.ChaosEvent{}

	// Consider the finalizers
	if ctx.obj.IsDeleted() {
		if ctx.obj.GetStatus().Experiment.DesiredPhase != v1alpha1.StoppedPhase {
			events = append(events, recorder.Deleted{})
		}
		return v1alpha1.StoppedPhase, events
	}

	if ctx.obj.IsOneShot() {
		// An oneshot chaos should always be in running phase, so that it cannot
		// be applied multiple times or cause other bugs :(
		return v1alpha1.RunningPhase, events
	}

	// Consider the duration
	now := time.Now()

	durationExceeded, untilStop, err := ctx.obj.DurationExceeded(now)
	if err != nil {
		ctx.Log.Error(err, "failed to parse duration")
	}
	if durationExceeded {
		if ctx.obj.GetStatus().Experiment.DesiredPhase != v1alpha1.StoppedPhase {
			events = append(events, recorder.TimeUp{})
		}
		return v1alpha1.StoppedPhase, events
	}

	ctx.requeueAfter = untilStop

	// Then decide the pause logic
	if ctx.obj.IsPaused() {
		if ctx.obj.GetStatus().Experiment.DesiredPhase != v1alpha1.StoppedPhase {
			events = append(events, recorder.Paused{})
		}
		return v1alpha1.StoppedPhase, events
	}

	if ctx.obj.GetStatus().Experiment.DesiredPhase != v1alpha1.RunningPhase {
		events = append(events, recorder.Started{})
	}
	return v1alpha1.RunningPhase, events
}

func (ctx *reconcileContext) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	desiredPhase, events := ctx.CalcDesiredPhase()

	ctx.Log.Info("modify desiredPhase", "desiredPhase", desiredPhase)
	if ctx.obj.GetStatus().Experiment.DesiredPhase != desiredPhase {
		for _, ev := range events {
			ctx.Recorder.Event(ctx.obj, ev)
		}

		updateError := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			obj := ctx.Object.DeepCopyObject().(v1alpha1.InnerObject)

			if err := ctx.Client.Get(context.TODO(), req.NamespacedName, obj); err != nil {
				ctx.Log.Error(err, "unable to get chaos")
				return err
			}

			if obj.GetStatus().Experiment.DesiredPhase != desiredPhase {
				obj.GetStatus().Experiment.DesiredPhase = desiredPhase
				ctx.Log.Info("update object", "namespace", obj.GetObjectMeta().GetNamespace(), "name", obj.GetObjectMeta().GetName())
				return ctx.Client.Update(context.TODO(), obj)
			}

			return nil
		})
		if updateError != nil {
			ctx.Log.Error(updateError, "fail to update")
			ctx.Recorder.Event(ctx.obj, recorder.Failed{
				Activity: "update desiredphase",
				Err:      updateError.Error(),
			})
			return ctrl.Result{}, nil
		}

		ctx.Recorder.Event(ctx.obj, recorder.Updated{
			Field: "desiredPhase",
		})
	}
	return ctrl.Result{RequeueAfter: ctx.requeueAfter}, nil
}
