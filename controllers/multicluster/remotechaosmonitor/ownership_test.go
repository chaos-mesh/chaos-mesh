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

package remotechaosmonitor

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/event"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func TestMonitorEventOwnership(t *testing.T) {
	_, owned := monitorChaosObjects()
	native := owned.DeepCopy()
	native.Labels = nil
	require.False(t, remoteChaosPredicate.Create(event.CreateEvent{Object: native}))
	require.False(t, remoteChaosPredicate.Update(event.UpdateEvent{ObjectOld: owned, ObjectNew: native}))
	require.False(t, remoteChaosPredicate.Delete(event.DeleteEvent{Object: native}))
	require.True(t, remoteChaosPredicate.Create(event.CreateEvent{Object: owned}))
	require.True(t, remoteChaosPredicate.Update(event.UpdateEvent{ObjectOld: native, ObjectNew: owned}))
	require.True(t, remoteChaosPredicate.Delete(event.DeleteEvent{Object: owned}))
}

func TestReconcileRechecksPairAfterConflict(t *testing.T) {
	for _, change := range []string{"management replacement", "missing remote management replacement", "remote replacement", "remote label removed", "remote status changed"} {
		t.Run(change, func(t *testing.T) {
			local, remote := monitorChaosObjects()
			localClient, remoteClient := monitorClient(t, local), monitorClient(t, remote)
			if change == "missing remote management replacement" {
				remoteClient = monitorClient(t)
			}
			conflicted := false
			localWant, remoteWant := local.DeepCopy(), remote.DeepCopy()
			conflictingClient := interceptor.NewClient(localClient, interceptor.Funcs{
				Update: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
					if conflicted {
						return c.Update(ctx, obj, opts...)
					}
					conflicted = true
					var latest v1alpha1.PodChaos
					switch change {
					case "management replacement", "missing remote management replacement":
						require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(local), &latest))
						latest.Finalizers = nil
						require.NoError(t, c.Update(ctx, &latest))
						require.NoError(t, c.Delete(ctx, &latest))
						localWant.UID = "replacement-management"
						localWant.ResourceVersion = ""
						require.NoError(t, c.Create(ctx, localWant))
					case "remote replacement":
						require.NoError(t, remoteClient.Get(ctx, client.ObjectKeyFromObject(remote), &latest))
						latest.Finalizers = nil
						require.NoError(t, remoteClient.Update(ctx, &latest))
						require.NoError(t, remoteClient.Delete(ctx, &latest))
						remoteWant.UID = "replacement-remote"
						remoteWant.ResourceVersion = ""
						require.NoError(t, remoteClient.Create(ctx, remoteWant))
					default:
						require.NoError(t, remoteClient.Get(ctx, client.ObjectKeyFromObject(remote), &latest))
						if change == "remote label removed" {
							latest.Labels = nil
						} else {
							latest.Status.Experiment.DesiredPhase = "Stop"
						}
						require.NoError(t, remoteClient.Update(ctx, &latest))
						remoteWant = latest.DeepCopy()
					}
					return apierrors.NewConflict(schema.GroupResource{Group: "chaos-mesh.org", Resource: "podchaos"}, obj.GetName(), errors.New("concurrent change"))
				},
			})
			r := New(&v1alpha1.PodChaos{}, conflictingClient, "cluster-a", remoteClient, logr.Discard())
			_, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(local)})
			require.NoError(t, err)
			require.True(t, conflicted)
			if change == "remote status changed" {
				localWant.Status = remoteWant.Status
				localWant.Finalizers = remoteWant.Finalizers
			}
			assertChaosUnchanged(t, localClient, localWant)
			if change != "missing remote management replacement" {
				assertChaosUnchanged(t, remoteClient, remoteWant)
			}
		})
	}
}

func TestReconcileCleanupStopsAfterFailedFinalizerUpdate(t *testing.T) {
	local, _ := monitorChaosObjects()
	localClient := monitorClient(t, local)
	failure := errors.New("write failed")
	deletes := 0
	failingClient := interceptor.NewClient(localClient, interceptor.Funcs{
		Update: func(context.Context, client.WithWatch, client.Object, ...client.UpdateOption) error { return failure },
		Delete: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
			deletes++
			return c.Delete(ctx, obj, opts...)
		},
	})
	r := New(&v1alpha1.PodChaos{}, failingClient, "cluster-a", monitorClient(t), logr.Discard())
	_, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(local)})
	require.ErrorIs(t, err, failure)
	require.Zero(t, deletes)
	assertChaosUnchanged(t, localClient, local)
}

func TestReconcileCleanupRechecksRemoteAbsence(t *testing.T) {
	local, remote := monitorChaosObjects()
	localClient, remoteClient := monitorClient(t, local), monitorClient(t)
	gets := 0
	appearingClient := interceptor.NewClient(remoteClient, interceptor.Funcs{
		Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			gets++
			if gets == 2 {
				require.NoError(t, c.Create(ctx, remote))
			}
			return c.Get(ctx, key, obj, opts...)
		},
	})
	r := New(&v1alpha1.PodChaos{}, localClient, "cluster-a", appearingClient, logr.Discard())
	_, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(local)})
	require.NoError(t, err)
	assertChaosUnchanged(t, localClient, local)
	assertChaosUnchanged(t, remoteClient, remote)
}

func TestReconcileOrphanDeletePreservesRelabelledRemote(t *testing.T) {
	_, remote := monitorChaosObjects()
	remoteClient := monitorClient(t, remote)
	deletingClient := interceptor.NewClient(remoteClient, interceptor.Funcs{
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
	r := New(&v1alpha1.PodChaos{}, monitorClient(t), "cluster-a", deletingClient, logr.Discard())
	_, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(remote)})
	require.True(t, apierrors.IsConflict(err))
	remote.Labels = nil
	assertChaosUnchanged(t, remoteClient, remote)
	_, err = r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(remote)})
	require.NoError(t, err)
	assertChaosUnchanged(t, remoteClient, remote)
}

func TestReconcileCleanupPreservesRetargetAtDelete(t *testing.T) {
	local, _ := monitorChaosObjects()
	localClient := monitorClient(t, local)
	deletingClient := interceptor.NewClient(localClient, interceptor.Funcs{
		Delete: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
			var options client.DeleteOptions
			options.ApplyOptions(opts)
			require.NotNil(t, options.Preconditions)
			require.Equal(t, local.UID, *options.Preconditions.UID)
			require.Equal(t, obj.GetResourceVersion(), *options.Preconditions.ResourceVersion)
			var latest v1alpha1.PodChaos
			require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(local), &latest))
			latest.Spec.RemoteCluster = "cluster-b"
			latest.Finalizers = []string{"cluster-b-finalizer"}
			require.NoError(t, c.Update(ctx, &latest))
			return c.Delete(ctx, obj, opts...)
		},
	})
	r := New(&v1alpha1.PodChaos{}, deletingClient, "cluster-a", monitorClient(t), logr.Discard())
	_, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(local)})
	require.NoError(t, err)
	local.Spec.RemoteCluster = "cluster-b"
	local.Finalizers = []string{"cluster-b-finalizer"}
	assertChaosUnchanged(t, localClient, local)
}
