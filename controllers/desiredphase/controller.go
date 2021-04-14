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
	"k8s.io/client-go/tools/record"
	"time"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// Reconciler for common chaos
type Reconciler struct {
	// Object is used to mark the target type of this Reconciler
	Object v1alpha1.InnerObject

	// Client is used to operate on the Kubernetes cluster
	client.Client
	client.Reader

	Recorder record.EventRecorder
	Log logr.Logger
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

func (ctx *reconcileContext) CalcDesiredPhase() v1alpha1.DesiredPhase {
	// Consider the finalizers
	if ctx.obj.IsDeleted() {
		ctx.Recorder.Eventf(ctx.obj,"Normal", "Deleted", "Turn into StoppedPhase")
		return v1alpha1.StoppedPhase
	}

	// Consider the duration
	now := time.Now()

	duration, err := ctx.obj.GetDuration()
	if err != nil {
		ctx.Log.Error(err, "failed to parse duration")
	}
	if duration != nil {
		stopTime := ctx.GetCreationTimestamp().Add(*duration)
		if stopTime.Before(now) {
			ctx.Recorder.Eventf(ctx.obj,"Normal", "TimeUp", "Turn into StoppedPhase")
			return v1alpha1.StoppedPhase
		} else {
			ctx.requeueAfter = stopTime.Sub(now)
		}
	}

	// Then decide the pause logic
	if ctx.obj.IsPaused() {
		ctx.Recorder.Eventf(ctx.obj,"Normal", "Paused", "Turn into StoppedPhase")
		return v1alpha1.StoppedPhase
	} else {
		return v1alpha1.RunningPhase
	}
}

func (ctx *reconcileContext) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	desiredPhase := ctx.CalcDesiredPhase()

	ctx.Log.Info("modify desiredPhase", "desiredPhase", desiredPhase)
	if ctx.obj.GetStatus().Experiment.DesiredPhase != desiredPhase {
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
			} else {
				return nil
			}
		})
		if updateError != nil {
			ctx.Log.Error(updateError, "fail to update")
			ctx.Recorder.Eventf(ctx.obj, "Normal", "Failed", "Failed to update desiredphase: %s", updateError.Error())
			return ctrl.Result{}, nil
		}

		ctx.Recorder.Event(ctx.obj, "Normal", "Updated", "Successfully update desiredPhase of resource")
	}
	return ctrl.Result{RequeueAfter: ctx.requeueAfter}, nil
}
