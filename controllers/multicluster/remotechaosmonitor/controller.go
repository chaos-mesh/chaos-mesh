// Copyright 2022 Chaos Mesh Authors.
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

package remotechaosmonitor

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type Reconciler struct {
	clusterName string

	obj v1alpha1.InnerObject

	logger       logr.Logger
	localClient  client.Client
	remoteClient client.Client
}

func New(obj v1alpha1.InnerObject, manageClient client.Client, clusterName string, localClient client.Client, logger logr.Logger) *Reconciler {
	return &Reconciler{
		clusterName:  clusterName,
		obj:          obj,
		logger:       logger.WithName("remotechaos-monitor"),
		localClient:  manageClient,
		remoteClient: localClient,
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger.Info("reconcile chaos", "clusterName", r.clusterName, "namespace", req.Namespace, "name", req.Name)

	shouldDelete := false
	remoteObj := r.obj.DeepCopyObject().(v1alpha1.RemoteObject)
	err := r.remoteClient.Get(ctx, req.NamespacedName, remoteObj)
	if err != nil {
		if apierrors.IsNotFound(err) {
			r.logger.Info("chaos not found")

			shouldDelete = true
		} else {
			// TODO: handle this error
			r.logger.Error(err, "unable to get remote chaos")
			return ctrl.Result{}, nil
		}
	}

	localObj := r.obj.DeepCopyObject().(v1alpha1.RemoteObject)
	err = r.localClient.Get(ctx, req.NamespacedName, localObj)
	if err != nil {
		if apierrors.IsNotFound(err) {
			r.logger.Info("local chaos not found")

			err = r.remoteClient.Delete(ctx, remoteObj)
			if err != nil {
				if apierrors.IsNotFound(err) {
					r.logger.Info("remote chaos deleted")
				} else {
					// TODO: retry for some error
					r.logger.Error(err, "unable to get remote chaos")
				}
			}

			return ctrl.Result{}, nil
		}

		// TODO: handle this error
		r.logger.Error(err, "unable to get local chaos")
		return ctrl.Result{}, nil
	}

	if shouldDelete {
		r.logger.Info("deleting local obj", "namespace", localObj.GetNamespace(), "name", localObj.GetName())
		localObj.SetFinalizers([]string{})
		err := r.localClient.Update(ctx, localObj)
		if err != nil {
			// TODO: retry
			r.logger.Error(err, "fail to update localObj")
			// it's expected to not return, as we could still try
			// to remove this object
		}

		err = r.localClient.Delete(ctx, localObj)
		if err != nil {
			// TODO: retry
			r.logger.Error(err, "fail to delete local object")
		}

		return ctrl.Result{}, nil
	}

	shouldUpdate := false
	if !reflect.DeepEqual(localObj.GetFinalizers(), remoteObj.GetFinalizers()) {
		r.logger.Info("setting new finalizers")
		localObj.SetFinalizers(remoteObj.GetFinalizers())

		shouldUpdate = true
	}
	// TODO: set the status

	if shouldUpdate {
		retry.RetryOnConflict(retry.DefaultRetry, func() error {
			localObj := r.obj.DeepCopyObject().(v1alpha1.RemoteObject)
			err = r.localClient.Get(ctx, req.NamespacedName, localObj)
			if err != nil {
				return err
			}

			// TODO: also refresh the remote chaos object
			localObj.SetFinalizers(remoteObj.GetFinalizers())

			return r.localClient.Update(ctx, localObj)
		})
	}
	return ctrl.Result{}, nil
}
