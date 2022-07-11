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

package remotecluster

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/multicluster/clusterregistry"
)

const remoteClusterControllerFinalizer = "chaos-mesh/remotecluster-controllers"

type Reconciler struct {
	Log      logr.Logger
	registry *clusterregistry.RemoteClusterRegistry

	client.Client
}

func (r *Reconciler) getRestConfig(ctx context.Context, secretRef v1alpha1.RemoteClusterSecretRef) (*rest.Config, error) {
	var secret corev1.Secret
	err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: secretRef.Namespace,
		Name:      secretRef.Name,
	}, &secret)
	if err != nil {
		return nil, errors.Wrapf(err, "get secret %s/%s", secretRef.Namespace, secretRef.Name)
	}

	kubeconfig := secret.Data[secretRef.Key]

	config, err := clientcmd.Load(kubeconfig)
	if err != nil {
		return nil, errors.Wrap(err, "load kubeconfig")
	}

	return clientcmd.NewDefaultClientConfig(*config, nil).ClientConfig()
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var obj v1alpha1.RemoteCluster
	err := r.Client.Get(ctx, req.NamespacedName, &obj)
	if err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("remote cluster not found", "namespace", req.Namespace, "name", req.Name)
		} else {
			// TODO: handle this error
			r.Log.Error(err, "unable to get remote cluster", "namespace", req.Namespace, "name", req.Name)
		}
		return ctrl.Result{}, nil
	}

	if !obj.DeletionTimestamp.IsZero() {
		err := r.registry.Stop(ctx, obj.Name)
		if err != nil {
			if !errors.Is(err, clusterregistry.ErrNotExist) {
				r.Log.Error(err, "fail to stop cluster")
				return ctrl.Result{Requeue: true}, nil
			}
		}

		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			obj.Finalizers = []string{}
			return r.Client.Update(ctx, &obj)
		})
		if err != nil {
			r.Log.Error(err, "fail to update finalizer", "name", obj.Name)
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, nil
	}

	restConfig, err := r.getRestConfig(ctx, obj.Spec.KubeConfig.SecretRef)
	if err != nil {
		r.Log.Error(err, "fail to get rest config")
		return ctrl.Result{Requeue: true}, nil
	}

	err = r.registry.Spawn(obj.Name, restConfig)
	if err != nil {
		if !errors.Is(err, clusterregistry.ErrAlreadyExist) {
			r.Log.Error(err, "fail to spawn controllers", "name", obj.Name)
			return ctrl.Result{Requeue: true}, nil
		}
	}

	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		obj.Finalizers = []string{remoteClusterControllerFinalizer}
		return r.Client.Update(ctx, &obj)
	})
	if err != nil {
		r.Log.Error(err, "fail to update finalizer", "name", obj.Name)
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}
