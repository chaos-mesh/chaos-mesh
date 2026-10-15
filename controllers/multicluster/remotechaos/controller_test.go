// Copyright 2026 Chaos Mesh Authors.
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
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type testRegistry struct{ remote client.Client }

func (r testRegistry) WithClient(_ string, f func(client.Client) error) error { return f(r.remote) }

func TestReconcileDeletionOwnership(t *testing.T) {
	for _, controlledBy := range []string{"", "another-controller", "remote-chaos"} {
		t.Run("controlled by "+controlledBy, func(t *testing.T) {
			local, remote, localClient, remoteClient := forwarderObjects(t)
			remote.Labels[v1alpha1.LabelControlledBy] = controlledBy
			require.NoError(t, remoteClient.Update(context.Background(), remote))
			r := &Reconciler{Client: localClient, Log: logr.Discard(), Object: &v1alpha1.PodChaos{}, registry: testRegistry{remoteClient}}
			for range 2 {
				_, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(local)})
				require.NoError(t, err)
				var got v1alpha1.PodChaos
				require.NoError(t, remoteClient.Get(context.Background(), client.ObjectKeyFromObject(remote), &got))
				require.Equal(t, remote.UID, got.UID)
				require.Equal(t, remote.Spec, got.Spec)
				require.Equal(t, remote.Labels, got.Labels)
				require.Equal(t, remote.Finalizers, got.Finalizers)
				if controlledBy == "remote-chaos" {
					require.NotNil(t, got.DeletionTimestamp)
				} else {
					require.Nil(t, got.DeletionTimestamp)
				}
			}
		})
	}
}

func TestReconcileDeletionRechecksManagement(t *testing.T) {
	for _, change := range []string{"retarget", "replacement"} {
		t.Run(change, func(t *testing.T) {
			local, remote, localClient, remoteClient := forwarderObjects(t)
			changed := false
			observingRemote := interceptor.NewClient(remoteClient, interceptor.Funcs{
				Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
					if !changed {
						var latest v1alpha1.PodChaos
						require.NoError(t, localClient.Get(ctx, key, &latest))
						if change == "retarget" {
							latest.Spec.RemoteCluster = "cluster-b"
							require.NoError(t, localClient.Update(ctx, &latest))
						} else {
							latest.Finalizers = nil
							require.NoError(t, localClient.Update(ctx, &latest))
							replacement := local.DeepCopy()
							replacement.UID = "replacement-management"
							replacement.ResourceVersion = ""
							replacement.DeletionTimestamp = nil
							require.NoError(t, localClient.Create(ctx, replacement))
						}
						changed = true
					}
					return c.Get(ctx, key, obj, opts...)
				},
			})
			r := &Reconciler{Client: localClient, Log: logr.Discard(), Object: &v1alpha1.PodChaos{}, registry: testRegistry{observingRemote}}
			_, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(local)})
			require.NoError(t, err)
			require.True(t, changed)
			var got v1alpha1.PodChaos
			require.NoError(t, remoteClient.Get(context.Background(), client.ObjectKeyFromObject(remote), &got))
			require.Equal(t, remote.UID, got.UID)
			require.Equal(t, remote.Spec, got.Spec)
			require.Equal(t, remote.Finalizers, got.Finalizers)
			require.Nil(t, got.DeletionTimestamp)
		})
	}
}

func TestReconcileDeletionPreservesRelabelledRemote(t *testing.T) {
	local, remote, localClient, remoteClient := forwarderObjects(t)
	deletingRemote := interceptor.NewClient(remoteClient, interceptor.Funcs{
		Delete: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
			var options client.DeleteOptions
			options.ApplyOptions(opts)
			require.NotNil(t, options.Preconditions)
			require.Equal(t, remote.UID, *options.Preconditions.UID)
			require.Equal(t, obj.GetResourceVersion(), *options.Preconditions.ResourceVersion)
			var latest v1alpha1.PodChaos
			require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(remote), &latest))
			latest.Labels = nil
			require.NoError(t, c.Update(ctx, &latest))
			return c.Delete(ctx, obj, opts...)
		},
	})
	r := &Reconciler{Client: localClient, Log: logr.Discard(), Object: &v1alpha1.PodChaos{}, registry: testRegistry{deletingRemote}}
	_, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(local)})
	require.True(t, apierrors.IsConflict(err))
	_, err = r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(local)})
	require.NoError(t, err)
	var got v1alpha1.PodChaos
	require.NoError(t, remoteClient.Get(context.Background(), client.ObjectKeyFromObject(remote), &got))
	require.Equal(t, remote.UID, got.UID)
	require.Equal(t, remote.Spec, got.Spec)
	require.Equal(t, remote.Finalizers, got.Finalizers)
	require.Empty(t, got.Labels)
	require.Nil(t, got.DeletionTimestamp)
}

func forwarderObjects(t *testing.T) (*v1alpha1.PodChaos, *v1alpha1.PodChaos, client.WithWatch, client.WithWatch) {
	t.Helper()
	scheme := runtime.NewScheme()
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	now := metav1.Now()
	local := &v1alpha1.PodChaos{ObjectMeta: metav1.ObjectMeta{
		Namespace: "experiments", Name: "same-name", UID: "management-uid",
		Finalizers: []string{"management-finalizer"}, DeletionTimestamp: &now,
	}, Spec: v1alpha1.PodChaosSpec{Action: "pod-kill", RemoteCluster: "cluster-a"}}
	remote := &v1alpha1.PodChaos{ObjectMeta: metav1.ObjectMeta{
		Namespace: "experiments", Name: "same-name", UID: "remote-uid",
		Labels:     map[string]string{v1alpha1.LabelControlledBy: "remote-chaos"},
		Finalizers: []string{"remote-finalizer"},
	}, Spec: v1alpha1.PodChaosSpec{Action: "pod-kill", GracePeriod: 13}}
	localClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(local).Build()
	remoteClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(remote).Build()
	return local, remote, localClient, remoteClient
}
