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

	remoteObj := r.obj.DeepCopyObject().(v1alpha1.RemoteObject)
	err := r.remoteClient.Get(ctx, req.NamespacedName, remoteObj)
	remoteMissing := apierrors.IsNotFound(err)
	if err != nil && !remoteMissing {
		r.logger.Error(err, "unable to get remote chaos")
		return ctrl.Result{}, nil
	}
	if !remoteMissing && remoteObj.GetLabels()[v1alpha1.LabelControlledBy] != "remote-chaos" {
		return ctrl.Result{}, nil
	}

	localObj := r.obj.DeepCopyObject().(v1alpha1.RemoteObject)
	err = r.localClient.Get(ctx, req.NamespacedName, localObj)
	if apierrors.IsNotFound(err) {
		if remoteMissing {
			return ctrl.Result{}, nil
		}
		uid, resourceVersion := remoteObj.GetUID(), remoteObj.GetResourceVersion()
		err = r.remoteClient.Delete(ctx, remoteObj, client.Preconditions{UID: &uid, ResourceVersion: &resourceVersion})
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if err != nil {
		r.logger.Error(err, "unable to get local chaos")
		return ctrl.Result{}, nil
	}
	if r.clusterName == "" || localObj.GetRemoteCluster() != r.clusterName {
		return ctrl.Result{}, nil
	}

	localUID, remoteUID := localObj.GetUID(), remoteObj.GetUID()
	if remoteMissing {
		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			latest := r.obj.DeepCopyObject().(v1alpha1.RemoteObject)
			if err := r.localClient.Get(ctx, req.NamespacedName, latest); err != nil {
				return client.IgnoreNotFound(err)
			}
			if latest.GetUID() != localUID || latest.GetRemoteCluster() != r.clusterName {
				return nil
			}
			remote := r.obj.DeepCopyObject().(v1alpha1.RemoteObject)
			err := r.remoteClient.Get(ctx, req.NamespacedName, remote)
			if err == nil {
				return nil
			}
			if !apierrors.IsNotFound(err) {
				return err
			}
			latest.SetFinalizers(nil)
			if err := r.localClient.Update(ctx, latest); err != nil {
				return client.IgnoreNotFound(err)
			}
			uid, resourceVersion := latest.GetUID(), latest.GetResourceVersion()
			return client.IgnoreNotFound(r.localClient.Delete(ctx, latest, client.Preconditions{UID: &uid, ResourceVersion: &resourceVersion}))
		})
		return ctrl.Result{}, err
	}

	if reflect.DeepEqual(localObj.GetFinalizers(), remoteObj.GetFinalizers()) && reflect.DeepEqual(localObj.GetStatus(), remoteObj.GetStatus()) {
		return ctrl.Result{}, nil
	}
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest := r.obj.DeepCopyObject().(v1alpha1.RemoteObject)
		if err := r.localClient.Get(ctx, req.NamespacedName, latest); err != nil {
			return client.IgnoreNotFound(err)
		}
		if latest.GetUID() != localUID || latest.GetRemoteCluster() != r.clusterName {
			return nil
		}
		remote := r.obj.DeepCopyObject().(v1alpha1.RemoteObject)
		if err := r.remoteClient.Get(ctx, req.NamespacedName, remote); err != nil {
			return client.IgnoreNotFound(err)
		}
		if remote.GetUID() != remoteUID || remote.GetLabels()[v1alpha1.LabelControlledBy] != "remote-chaos" {
			return nil
		}
		latest.SetFinalizers(remote.GetFinalizers())
		remote.GetStatus().DeepCopyInto(latest.GetStatus())
		return r.localClient.Update(ctx, latest)
	})
	return ctrl.Result{}, err
}
