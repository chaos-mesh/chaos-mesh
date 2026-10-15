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
	"errors"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/multicluster/remotechaosmonitor"
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

func TestReconcileDeletionDetachesAfterOwnershipLoss(t *testing.T) {
	for _, owner := range []string{"", "another-controller"} {
		t.Run(owner, func(t *testing.T) {
			local, remote, localClient, remoteClient := forwarderObjects(t)
			ctx := context.Background()
			req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(local)}
			remote.Status.Experiment.DesiredPhase = v1alpha1.RunningPhase
			require.NoError(t, remoteClient.Update(ctx, remote))
			monitor := remotechaosmonitor.New(&v1alpha1.PodChaos{}, localClient, "cluster-a", remoteClient, logr.Discard())
			_, err := monitor.Reconcile(ctx, req)
			require.NoError(t, err)
			var mirrored v1alpha1.PodChaos
			require.NoError(t, localClient.Get(ctx, req.NamespacedName, &mirrored))
			require.Equal(t, []string{"remote-finalizer"}, mirrored.Finalizers)
			remote.Labels[v1alpha1.LabelControlledBy] = owner
			require.NoError(t, remoteClient.Update(ctx, remote))
			_, err = monitor.Reconcile(ctx, req)
			require.NoError(t, err)
			forwarder := &Reconciler{Client: localClient, Log: logr.Discard(), Object: &v1alpha1.PodChaos{}, registry: testRegistry{remoteClient}}
			_, err = forwarder.Reconcile(ctx, req)
			require.NoError(t, err)
			require.True(t, apierrors.IsNotFound(localClient.Get(ctx, req.NamespacedName, &mirrored)), "deleting management experiment must release mirrored finalizers after remote ownership is lost")
			var preserved v1alpha1.PodChaos
			require.NoError(t, remoteClient.Get(ctx, req.NamespacedName, &preserved))
			require.Equal(t, remote.UID, preserved.UID)
			require.Equal(t, remote.Spec, preserved.Spec)
			require.Equal(t, remote.Labels, preserved.Labels)
			require.Equal(t, remote.Finalizers, preserved.Finalizers)
			require.Nil(t, preserved.DeletionTimestamp)
			require.Equal(t, remote.Status, preserved.Status)
			_, err = forwarder.Reconcile(ctx, req)
			require.NoError(t, err)
			var repeated v1alpha1.PodChaos
			require.NoError(t, remoteClient.Get(ctx, req.NamespacedName, &repeated))
			require.Equal(t, preserved, repeated)
		})
	}
}

func TestReconcileDeletionMissingRemote(t *testing.T) {
	local, remote, localClient, remoteClient := forwarderObjects(t)
	ctx := context.Background()
	remote.Finalizers = nil
	require.NoError(t, remoteClient.Update(ctx, remote))
	require.NoError(t, remoteClient.Delete(ctx, remote))
	r := &Reconciler{Client: localClient, Log: logr.Discard(), Object: &v1alpha1.PodChaos{}, registry: testRegistry{remoteClient}}
	for range 2 {
		_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(local)})
		require.NoError(t, err)
		var got v1alpha1.PodChaos
		require.True(t, apierrors.IsNotFound(localClient.Get(ctx, client.ObjectKeyFromObject(local), &got)))
		require.True(t, apierrors.IsNotFound(remoteClient.Get(ctx, client.ObjectKeyFromObject(remote), &got)))
	}
}

func TestReconcileDeletionDetachRetries(t *testing.T) {
	for _, change := range []string{"unchanged", "retarget", "replacement", "removed", "remote reowned"} {
		t.Run(change, func(t *testing.T) {
			local, remote, localClient, remoteClient := forwarderObjects(t)
			ctx := context.Background()
			key := client.ObjectKeyFromObject(local)
			remote.Labels = nil
			require.NoError(t, remoteClient.Update(ctx, remote))
			var before v1alpha1.PodChaos
			require.NoError(t, remoteClient.Get(ctx, key, &before))
			updates := 0
			conflictingLocal := interceptor.NewClient(localClient, interceptor.Funcs{
				Update: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
					updates++
					if updates != 1 {
						return c.Update(ctx, obj, opts...)
					}
					var latest v1alpha1.PodChaos
					require.NoError(t, c.Get(ctx, key, &latest))
					switch change {
					case "retarget":
						latest.Spec.RemoteCluster = "cluster-b"
						require.NoError(t, c.Update(ctx, &latest))
					case "replacement", "removed":
						latest.Finalizers = nil
						require.NoError(t, c.Update(ctx, &latest))
						if change == "replacement" {
							replacement := local.DeepCopy()
							replacement.UID = "replacement-management"
							replacement.ResourceVersion = ""
							replacement.DeletionTimestamp = nil
							require.NoError(t, c.Create(ctx, replacement))
							require.NoError(t, c.Delete(ctx, replacement))
						}
					case "remote reowned":
						remote.Labels = map[string]string{v1alpha1.LabelControlledBy: "remote-chaos"}
						require.NoError(t, remoteClient.Update(ctx, remote))
					}
					return apierrors.NewConflict(schema.GroupResource{Group: "chaos-mesh.org", Resource: "podchaos"}, obj.GetName(), errors.New("concurrent update"))
				},
			})
			r := &Reconciler{Client: conflictingLocal, Log: logr.Discard(), Object: &v1alpha1.PodChaos{}, registry: testRegistry{remoteClient}}
			_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: key})
			require.NoError(t, err)
			var gotLocal, gotRemote v1alpha1.PodChaos
			localErr := localClient.Get(ctx, key, &gotLocal)
			if change == "unchanged" || change == "removed" {
				require.True(t, apierrors.IsNotFound(localErr))
			} else {
				require.NoError(t, localErr)
				require.Equal(t, []string{"management-finalizer"}, gotLocal.Finalizers)
				require.NotNil(t, gotLocal.DeletionTimestamp)
				if change == "retarget" {
					require.Equal(t, "cluster-b", gotLocal.Spec.RemoteCluster)
				} else if change == "replacement" {
					require.Equal(t, "replacement-management", string(gotLocal.UID))
				}
			}
			require.NoError(t, remoteClient.Get(ctx, key, &gotRemote))
			if change == "remote reowned" {
				require.Equal(t, "remote-chaos", gotRemote.Labels[v1alpha1.LabelControlledBy])
				require.Equal(t, before.UID, gotRemote.UID)
				require.Equal(t, before.Spec, gotRemote.Spec)
				require.Equal(t, before.Finalizers, gotRemote.Finalizers)
				require.NotNil(t, gotRemote.DeletionTimestamp)
			} else {
				require.Equal(t, before, gotRemote)
			}
			if change == "unchanged" {
				require.Equal(t, 2, updates)
			} else {
				require.Equal(t, 1, updates)
			}
		})
	}
}

func TestReconcileDeletionReturnsAPIErrors(t *testing.T) {
	for _, operation := range []string{"remote get", "local get", "local update"} {
		t.Run(operation, func(t *testing.T) {
			local, remote, localClient, remoteClient := forwarderObjects(t)
			ctx := context.Background()
			key := client.ObjectKeyFromObject(local)
			remote.Labels = nil
			require.NoError(t, remoteClient.Update(ctx, remote))
			failure := errors.New("API unavailable")
			gets := 0
			failingLocal := interceptor.NewClient(localClient, interceptor.Funcs{
				Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
					gets++
					if operation == "local get" && gets > 1 {
						return failure
					}
					return c.Get(ctx, key, obj, opts...)
				},
				Update: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
					if operation == "local update" {
						return failure
					}
					return c.Update(ctx, obj, opts...)
				},
			})
			failingRemote := interceptor.NewClient(remoteClient, interceptor.Funcs{
				Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
					if operation == "remote get" {
						return failure
					}
					return c.Get(ctx, key, obj, opts...)
				},
			})
			r := &Reconciler{Client: failingLocal, Log: logr.Discard(), Object: &v1alpha1.PodChaos{}, registry: testRegistry{failingRemote}}
			_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: key})
			require.ErrorIs(t, err, failure)
			var gotLocal, gotRemote v1alpha1.PodChaos
			require.NoError(t, localClient.Get(ctx, key, &gotLocal))
			require.Equal(t, []string{"management-finalizer"}, gotLocal.Finalizers)
			require.NoError(t, remoteClient.Get(ctx, key, &gotRemote))
			require.Nil(t, gotRemote.DeletionTimestamp)
		})
	}
}
