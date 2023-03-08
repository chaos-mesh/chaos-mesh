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

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/multicluster/clusterregistry"
)

type Reconciler struct {
	client.Client
	Log logr.Logger

	Object v1alpha1.InnerObject

	registry *clusterregistry.RemoteClusterRegistry
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	obj := r.Object.DeepCopyObject().(v1alpha1.RemoteObject)

	if err := r.Client.Get(context.TODO(), req.NamespacedName, obj); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("chaos not found")
		} else {
			// TODO: handle this error
			r.Log.Error(err, "unable to get chaos")
		}
		return ctrl.Result{}, nil
	}

	err := r.registry.WithClient(obj.GetRemoteCluster(), func(c client.Client) error {
		r.Log.Info("handling chaos with remote client", "cluster", obj.GetRemoteCluster())

		localObj := obj.DeepCopyObject().(v1alpha1.RemoteObject)

		remoteObj := obj.DeepCopyObject().(v1alpha1.RemoteObject)
		err := c.Get(ctx, req.NamespacedName, remoteObj)
		if err != nil {
			if apierrors.IsNotFound(err) {
				// remote chaos doesn't exist, while the local one is being deleted
				if localObj.GetDeletionTimestamp() != nil {
					return retry.RetryOnConflict(retry.DefaultRetry, func() error {
						var obj v1alpha1.RemoteObject
						r.Log.Info("resetting finalizers of local objects")
						if err := r.Client.Get(context.TODO(), req.NamespacedName, obj); err != nil {
							if apierrors.IsNotFound(err) {
								r.Log.Info("chaos has been removed")
								return nil
							}
							// TODO: handle this error
							r.Log.Error(err, "unable to get chaos")

							return err
						}

						obj.SetFinalizers([]string{})
						return r.Client.Update(ctx, obj)
					})
				}

				// omit the remoteCluster
				localSpecValue := reflect.Indirect(reflect.ValueOf(localObj)).FieldByName("Spec")
				localSpecValue.FieldByName("RemoteCluster").Set(reflect.ValueOf(""))

				// only Spec, Name, Namespace and a label will be initialized
				newObj := r.Object.DeepCopyObject().(v1alpha1.RemoteObject)
				reflect.Indirect(reflect.ValueOf(newObj)).FieldByName("Spec").Set(localSpecValue)

				newObj.SetLabels(map[string]string{
					"chaos-mesh.org/controlled-by": "remote-chaos",
				})
				newObj.SetName(obj.GetName())
				newObj.SetNamespace(obj.GetNamespace())

				return c.Create(ctx, newObj)
			}

			// TODO: handle this error
			r.Log.Error(err, "unable to get chaos")
			return nil
		}

		// remote chaos exists
		if localObj.GetDeletionTimestamp() != nil {
			r.Log.Info("deleting remote obj", "namespace", remoteObj.GetNamespace(), "name", remoteObj.GetName())
			err := c.Delete(ctx, remoteObj)
			if err != nil {
				if apierrors.IsNotFound(err) {
					return nil
				}
				return err
			}
		}
		return nil
	})
	if err != nil {
		r.Log.Error(err, "unable to handle chaos")
		// TODO: handle the error
		// TODO: retry
	}

	return ctrl.Result{}, nil
}
