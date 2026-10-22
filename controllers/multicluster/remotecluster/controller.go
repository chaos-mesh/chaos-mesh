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
	"encoding/json"
	"reflect"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	release "helm.sh/helm/v4/pkg/release/v1"
	"helm.sh/helm/v4/pkg/storage/driver"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/multicluster/clusterregistry"
	"github.com/chaos-mesh/chaos-mesh/pkg/helm"
)

const remoteClusterControllerFinalizer = "chaos-mesh/remotecluster-controllers"
const chaosMeshReleaseName = "chaos-mesh"

type clusterRegistry interface {
	Ensure(context.Context, string, *clientcmdapi.Config) error
	Stop(context.Context, string) error
}

type Reconciler struct {
	Log           logr.Logger
	registry      clusterRegistry
	ensureHelm    func(context.Context, *v1alpha1.RemoteCluster, clientcmd.ClientConfig) (string, error)
	uninstallHelm func(context.Context, *v1alpha1.RemoteCluster, clientcmd.ClientConfig) error

	client.Client
}

func (r *Reconciler) getKubeConfig(ctx context.Context, secretRef v1alpha1.RemoteClusterSecretRef) (*clientcmdapi.Config, error) {
	var secret corev1.Secret
	err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: secretRef.Namespace,
		Name:      secretRef.Name,
	}, &secret)
	if err != nil {
		return nil, errors.Wrapf(err, "get secret %s/%s", secretRef.Namespace, secretRef.Name)
	}

	kubeconfig, ok := secret.Data[secretRef.Key]
	if !ok {
		return nil, errors.Errorf("secret %s/%s has no key %s", secretRef.Namespace, secretRef.Name, secretRef.Key)
	}

	config, err := clientcmd.Load(kubeconfig)
	if err != nil {
		return nil, errors.Wrap(err, "load kubeconfig")
	}

	return config, nil
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var obj v1alpha1.RemoteCluster
	if err := r.Client.Get(ctx, req.NamespacedName, &obj); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		r.Log.Error(err, "unable to get remote cluster", "name", req.Name)
		return ctrl.Result{Requeue: true}, nil
	}
	snapshot := obj.DeepCopy()
	if !obj.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, snapshot)
	}

	kubeconfig, err := r.getKubeConfig(ctx, obj.Spec.KubeConfig.SecretRef)
	if err != nil {
		r.Log.Error(err, "fail to get kubeconfig from secret")
		return r.requeueNotReady(ctx, snapshot, "KubeConfigUnavailable")
	}

	if !controllerutil.ContainsFinalizer(&obj, remoteClusterControllerFinalizer) {
		if _, err := clientcmd.NewDefaultClientConfig(*kubeconfig.DeepCopy(), nil).ClientConfig(); err != nil {
			r.Log.Error(err, "invalid initial kubeconfig")
			return r.requeueNotReady(ctx, snapshot, "InvalidKubeConfig")
		}
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err := r.Client.Get(ctx, req.NamespacedName, &obj); err != nil {
				return err
			}
			if !obj.DeletionTimestamp.IsZero() || obj.Generation != snapshot.Generation || obj.UID != snapshot.UID {
				return nil
			}
			if !controllerutil.AddFinalizer(&obj, remoteClusterControllerFinalizer) {
				return nil
			}
			return r.Client.Update(ctx, &obj)
		})
		if err != nil {
			r.Log.Error(err, "fail to add finalizer", "name", obj.Name)
			return r.requeueNotReady(ctx, snapshot, "FinalizerUpdateFailed")
		}
		if !obj.DeletionTimestamp.IsZero() || obj.Generation != snapshot.Generation || obj.UID != snapshot.UID {
			return ctrl.Result{Requeue: true}, nil
		}
	}

	currentVersion := obj.Status.CurrentVersion
	if obj.Generation > obj.Status.ObservedGeneration {
		ensureHelm := r.ensureHelm
		if ensureHelm == nil {
			ensureHelm = r.ensureHelmRelease
		}
		currentVersion, err = ensureHelm(ctx, &obj, clientcmd.NewDefaultClientConfig(*kubeconfig.DeepCopy(), nil))
		if err != nil {
			r.Log.Error(err, "fail to list or install remote helm release")
			return r.requeueNotReady(ctx, snapshot, "HelmReconcileFailed")
		}
	}
	if err := r.registry.Ensure(ctx, obj.Name, kubeconfig); err != nil {
		r.Log.Error(err, "fail to ensure remote cluster controller manager")
		return r.requeueNotReady(ctx, snapshot, "ManagerReconcileFailed")
	}

	matched, err := r.updateStatus(ctx, snapshot, func(status *v1alpha1.RemoteClusterStatus) {
		status.CurrentVersion = currentVersion
		status.ObservedGeneration = snapshot.Generation
		setRemoteClusterCondition(status, v1alpha1.RemoteClusterConditionInstalled, corev1.ConditionTrue, "")
		setRemoteClusterCondition(status, v1alpha1.RemoteClusterConditionReady, corev1.ConditionTrue, "")
	})
	if err != nil {
		r.Log.Error(err, "fail to update remote cluster status", "name", snapshot.Name)
	}
	return ctrl.Result{Requeue: err != nil || !matched}, nil
}

func (r *Reconciler) updateStatus(ctx context.Context, snapshot *v1alpha1.RemoteCluster, mutate func(*v1alpha1.RemoteClusterStatus)) (bool, error) {
	matched := false
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		matched = false
		var latest v1alpha1.RemoteCluster
		if err := r.Client.Get(ctx, client.ObjectKeyFromObject(snapshot), &latest); err != nil {
			return client.IgnoreNotFound(err)
		}
		if latest.UID != snapshot.UID || latest.Generation != snapshot.Generation ||
			latest.DeletionTimestamp.IsZero() != snapshot.DeletionTimestamp.IsZero() ||
			latest.Status.ObservedGeneration > snapshot.Generation {
			return nil
		}
		matched = true
		before := latest.DeepCopy()
		mutate(&latest.Status)
		if reflect.DeepEqual(before.Status, latest.Status) {
			return nil
		}
		return r.Client.Status().Update(ctx, &latest)
	})
	return matched, err
}

func (r *Reconciler) requeueNotReady(ctx context.Context, snapshot *v1alpha1.RemoteCluster, reason string) (ctrl.Result, error) {
	_, err := r.updateStatus(ctx, snapshot, func(status *v1alpha1.RemoteClusterStatus) {
		setRemoteClusterCondition(status, v1alpha1.RemoteClusterConditionReady, corev1.ConditionFalse, reason)
	})
	if err != nil {
		r.Log.Error(err, "fail to update remote cluster status", "name", snapshot.Name)
	}
	return ctrl.Result{Requeue: true}, nil
}

func (r *Reconciler) reconcileDelete(ctx context.Context, obj *v1alpha1.RemoteCluster) (ctrl.Result, error) {
	matched, statusErr := r.updateStatus(ctx, obj, func(status *v1alpha1.RemoteClusterStatus) {
		setRemoteClusterCondition(status, v1alpha1.RemoteClusterConditionReady, corev1.ConditionFalse, "Deleting")
	})
	if statusErr != nil {
		r.Log.Error(statusErr, "fail to update deleting remote cluster status", "name", obj.Name)
	} else if !matched {
		return ctrl.Result{Requeue: true}, nil
	}
	if err := r.registry.Stop(ctx, obj.Name); err != nil && !errors.Is(err, clusterregistry.ErrNotExist) {
		r.Log.Error(err, "fail to stop cluster")
		return ctrl.Result{Requeue: true}, nil
	}
	if !controllerutil.ContainsFinalizer(obj, remoteClusterControllerFinalizer) {
		return ctrl.Result{Requeue: statusErr != nil}, nil
	}
	kubeconfig, err := r.getKubeConfig(ctx, obj.Spec.KubeConfig.SecretRef)
	if err != nil {
		r.Log.Error(err, "fail to get kubeconfig from secret")
		return ctrl.Result{Requeue: true}, nil
	}
	uninstallHelm := r.uninstallHelm
	if uninstallHelm == nil {
		uninstallHelm = r.uninstallHelmRelease
	}
	if err := uninstallHelm(ctx, obj, clientcmd.NewDefaultClientConfig(*kubeconfig.DeepCopy(), nil)); err != nil {
		r.Log.Error(err, "fail to uninstall helm release")
		return ctrl.Result{Requeue: true}, nil
	}
	matched, err = r.updateStatus(ctx, obj, func(status *v1alpha1.RemoteClusterStatus) {
		setRemoteClusterCondition(status, v1alpha1.RemoteClusterConditionInstalled, corev1.ConditionFalse, "")
		setRemoteClusterCondition(status, v1alpha1.RemoteClusterConditionReady, corev1.ConditionFalse, "Deleting")
	})
	if err != nil {
		r.Log.Error(err, "fail to update deleted remote cluster status", "name", obj.Name)
	}
	if err != nil || !matched {
		return ctrl.Result{Requeue: true}, nil
	}
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		matched = false
		var latest v1alpha1.RemoteCluster
		if err := r.Client.Get(ctx, client.ObjectKeyFromObject(obj), &latest); err != nil {
			return client.IgnoreNotFound(err)
		}
		if latest.UID != obj.UID || latest.Generation != obj.Generation || latest.DeletionTimestamp.IsZero() {
			return nil
		}
		matched = true
		if !controllerutil.RemoveFinalizer(&latest, remoteClusterControllerFinalizer) {
			return nil
		}
		return r.Client.Update(ctx, &latest)
	})
	if err != nil {
		r.Log.Error(err, "fail to remove finalizer", "name", obj.Name)
	}
	return ctrl.Result{Requeue: err != nil || !matched}, nil
}

func (r *Reconciler) requestsForSecret(ctx context.Context, secret client.Object) []ctrl.Request {
	var clusters v1alpha1.RemoteClusterList
	if err := r.Client.List(ctx, &clusters); err != nil {
		r.Log.Error(err, "fail to list remote clusters for secret", "namespace", secret.GetNamespace(), "name", secret.GetName())
		return nil
	}
	var requests []ctrl.Request
	for _, cluster := range clusters.Items {
		ref := cluster.Spec.KubeConfig.SecretRef
		if ref.Namespace == secret.GetNamespace() && ref.Name == secret.GetName() {
			requests = append(requests, ctrl.Request{Name: cluster.Name})
		}
	}
	return requests
}

func (r *Reconciler) getHelmClient(_ context.Context, clientConfig clientcmd.ClientConfig) (*helm.Client, error) {
	restClientGetter := helm.NewRESTClientGetter(clientConfig)

	helmClient, err := helm.NewHelmClient(restClientGetter, r.Log)
	if err != nil {
		return nil, err
	}

	return helmClient, nil
}

func (r *Reconciler) ensureHelmRelease(ctx context.Context, obj *v1alpha1.RemoteCluster, clientConfig clientcmd.ClientConfig) (string, error) {
	helmClient, err := r.getHelmClient(ctx, clientConfig)
	if err != nil {
		return "", err
	}
	_, releaseErr := helmClient.GetRelease(obj.Spec.Namespace, chaosMeshReleaseName)
	if releaseErr != nil && !errors.Is(releaseErr, driver.ErrReleaseNotFound) {
		return "", releaseErr
	}
	chart, err := helm.FetchChaosMeshChart(ctx, obj.Spec.Version, config.ControllerCfg.LocalHelmChartPath)
	if err != nil {
		return "", err
	}

	values := make(map[string]any)
	if obj.Spec.ConfigOverride != nil {
		err = json.Unmarshal(obj.Spec.ConfigOverride, &values)
		if err != nil {
			return "", err
		}
	}
	var release *release.Release
	if errors.Is(releaseErr, driver.ErrReleaseNotFound) {
		release, err = helmClient.InstallRelease(obj.Spec.Namespace, chaosMeshReleaseName, chart, values)
		if err != nil {
			return "", err
		}
	} else {
		release, err = helmClient.UpgradeRelease(obj.Spec.Namespace, chaosMeshReleaseName, chart, values)
		if err != nil {
			return "", err
		}
	}

	return release.Chart.AppVersion(), nil
}

func (r *Reconciler) uninstallHelmRelease(ctx context.Context, obj *v1alpha1.RemoteCluster, clientConfig clientcmd.ClientConfig) error {
	helmClient, err := r.getHelmClient(ctx, clientConfig)
	if err != nil {
		return err
	}

	_, err = helmClient.GetRelease(obj.Spec.Namespace, chaosMeshReleaseName)
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			return nil
		}

		return err
	}

	// the release still exist
	_, err = helmClient.UninstallRelease(obj.Spec.Namespace, chaosMeshReleaseName)
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			return nil
		}

		return err
	}

	return nil
}
