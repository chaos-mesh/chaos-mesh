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

package finalizers

import (
	"context"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
)

const (
	// AnnotationCleanFinalizer key
	AnnotationCleanFinalizer = `chaos-mesh.chaos-mesh.org/cleanFinalizer`
	// AnnotationCleanFinalizerForced value
	AnnotationCleanFinalizerForced = `forced`

	RecordFinalizer = "chaos-mesh/records"
)

// ReconcilerMeta defines the meta of InitReconciler and CleanReconciler struct.
type ReconcilerMeta struct {
	// Object is used to mark the target type of this Reconciler
	Object v1alpha1.InnerObject

	// Client is used to operate on the Kubernetes cluster
	client.Client

	Recorder recorder.ChaosRecorder

	Log logr.Logger
}

// InitReconciler for common chaos to init the finalizer
type InitReconciler struct {
	ReconcilerMeta
}

// Reconcile the common chaos to init the finalizer
func (r *InitReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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

	if !obj.IsDeleted() {
		if !ContainsFinalizer(obj.(metav1.Object), RecordFinalizer) {
			r.Recorder.Event(obj, recorder.FinalizerInited{})
			finalizers := append(obj.GetFinalizers(), RecordFinalizer)
			return updateFinalizer(r.ReconcilerMeta, obj, req, finalizers)
		}
	}

	return ctrl.Result{}, nil
}

// CleanReconciler for common chaos to clean the finalizer
type CleanReconciler struct {
	ReconcilerMeta
}

// Reconcile the common chaos to clean the finalizer
func (r *CleanReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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

	finalizers := obj.GetFinalizers()
	records := obj.GetStatus().Experiment.Records
	if obj.IsDeleted() {
		resumed := true
		for _, record := range records {
			if record.Phase != v1alpha1.NotInjected {
				resumed = false
			}
		}

		if obj.GetAnnotations()[AnnotationCleanFinalizer] == AnnotationCleanFinalizerForced || (resumed && len(finalizers) != 0) {
			r.Recorder.Event(obj, recorder.FinalizerRemoved{})
			finalizers = []string{}
			return updateFinalizer(r.ReconcilerMeta, obj, req, finalizers)
		}
	}

	return ctrl.Result{}, nil
}

func updateFinalizer(r ReconcilerMeta, obj v1alpha1.InnerObject, req ctrl.Request, finalizers []string) (ctrl.Result, error) {
	updateError := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		obj := r.Object.DeepCopyObject().(v1alpha1.InnerObject)

		if err := r.Client.Get(context.TODO(), req.NamespacedName, obj); err != nil {
			r.Log.Error(err, "unable to get chaos")
			return err
		}

		obj.SetFinalizers(finalizers)
		return r.Client.Update(context.TODO(), obj)
	})
	if updateError != nil {
		// TODO: handle this error
		r.Log.Error(updateError, "fail to update")
		r.Recorder.Event(obj, recorder.Failed{
			Activity: "update finalizer",
			Err:      "updateError.Error()",
		})
		return ctrl.Result{}, nil
	}

	r.Recorder.Event(obj, recorder.Updated{
		Field: "finalizer",
	})
	return ctrl.Result{}, nil
}

// ContainsFinalizer checks an Object that the provided finalizer is present.
func ContainsFinalizer(o metav1.Object, finalizer string) bool {
	f := o.GetFinalizers()
	for _, e := range f {
		if e == finalizer {
			return true
		}
	}
	return false
}
