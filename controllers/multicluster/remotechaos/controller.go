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

package remotechaos

import (
	"context"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type Reconciler struct {
	client.Client
	Log logr.Logger

	Object v1alpha1.InnerObject

	registry interface {
		WithClient(string, func(client.Client) error) error
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	obj := r.Object.DeepCopyObject().(v1alpha1.RemoteObject)

	if err := r.Client.Get(ctx, req.NamespacedName, obj); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("chaos not found")
		} else {
			r.Log.Error(err, "unable to get chaos")
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, nil
	}

	if obj.GetRemoteCluster() == "" {
		return ctrl.Result{}, nil
	}

	result := ctrl.Result{}
	err := r.registry.WithClient(obj.GetRemoteCluster(), func(c client.Client) error {
		r.Log.Info("handling chaos with remote client", "cluster", obj.GetRemoteCluster())

		localObj := obj.DeepCopyObject().(v1alpha1.RemoteObject)
		if localObj.GetDeletionTimestamp() != nil {
			if err := r.reconcileDeletion(ctx, c, localObj); err != nil {
				return err
			}
			result.RequeueAfter = time.Second
			return nil
		}

		remoteObj := obj.DeepCopyObject().(v1alpha1.RemoteObject)
		err := c.Get(ctx, req.NamespacedName, remoteObj)
		if err != nil {
			if apierrors.IsNotFound(err) {
				// omit the remoteCluster
				localSpecValue := reflect.Indirect(reflect.ValueOf(localObj)).FieldByName("Spec")
				localSpecValue.FieldByName("RemoteCluster").Set(reflect.ValueOf(""))

				// only Spec, Name, Namespace and a label will be initialized
				newObj := r.Object.DeepCopyObject().(v1alpha1.RemoteObject)
				reflect.Indirect(reflect.ValueOf(newObj)).FieldByName("Spec").Set(localSpecValue)

				newObj.SetLabels(map[string]string{
					v1alpha1.LabelControlledBy: "remote-chaos",
				})
				newObj.SetName(obj.GetName())
				newObj.SetNamespace(obj.GetNamespace())

				return c.Create(ctx, newObj)
			}

			return err
		}

		return nil
	})
	if err != nil {
		r.Log.Error(err, "unable to handle chaos")
		return ctrl.Result{Requeue: true}, nil
	}

	return result, nil
}

func (r *Reconciler) reconcileDeletion(ctx context.Context, remoteClient client.Client, original v1alpha1.RemoteObject) error {
	key := client.ObjectKeyFromObject(original)
	var remoteToDelete v1alpha1.RemoteObject
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		remote := r.Object.DeepCopyObject().(v1alpha1.RemoteObject)
		remoteErr := remoteClient.Get(ctx, key, remote)
		if remoteErr != nil && !apierrors.IsNotFound(remoteErr) {
			return remoteErr
		}

		latest := r.Object.DeepCopyObject().(v1alpha1.RemoteObject)
		if err := r.Client.Get(ctx, key, latest); err != nil {
			return client.IgnoreNotFound(err)
		}
		if latest.GetUID() != original.GetUID() || latest.GetRemoteCluster() != original.GetRemoteCluster() || latest.GetDeletionTimestamp() == nil {
			return nil
		}
		if remoteErr == nil && remote.GetLabels()[v1alpha1.LabelControlledBy] == "remote-chaos" {
			remoteToDelete = remote
			return nil
		}
		if len(latest.GetFinalizers()) == 0 {
			return nil
		}
		latest.SetFinalizers(nil)
		return client.IgnoreNotFound(r.Client.Update(ctx, latest))
	})
	if err != nil || remoteToDelete == nil {
		return err
	}

	uid, resourceVersion := remoteToDelete.GetUID(), remoteToDelete.GetResourceVersion()
	return client.IgnoreNotFound(remoteClient.Delete(ctx, remoteToDelete, client.Preconditions{UID: &uid, ResourceVersion: &resourceVersion}))
}
