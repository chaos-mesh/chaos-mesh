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
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func TestReconcilePreservesUnownedChaos(t *testing.T) {
	for _, tc := range []struct {
		name          string
		localMissing  bool
		remoteMissing bool
		target        string
		controlledBy  string
	}{
		{name: "native remote without management object", localMissing: true},
		{name: "other controller remote without management object", localMissing: true, controlledBy: "another-controller"},
		{name: "native remote with matching management target", target: "cluster-a"},
		{name: "other controller remote with matching management target", target: "cluster-a", controlledBy: "another-controller"},
		{name: "owned remote with another management target", target: "cluster-b", controlledBy: "remote-chaos"},
		{name: "missing remote with another management target", target: "cluster-b", remoteMissing: true},
		{name: "owned remote with local management experiment", controlledBy: "remote-chaos"},
		{name: "missing remote with local management experiment", remoteMissing: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			local, remote := monitorChaosObjects()
			local.Spec.RemoteCluster = tc.target
			remote.Labels = nil
			if tc.controlledBy != "" {
				remote.Labels = map[string]string{"chaos-mesh.org/controlled-by": tc.controlledBy}
			}
			var localObjects, remoteObjects []client.Object
			if !tc.localMissing {
				localObjects = append(localObjects, local)
			}
			if !tc.remoteMissing {
				remoteObjects = append(remoteObjects, remote)
			}
			localClient := monitorClient(t, localObjects...)
			remoteClient := monitorClient(t, remoteObjects...)
			r := New(&v1alpha1.PodChaos{}, localClient, "cluster-a", remoteClient, logr.Discard())
			for range 2 {
				_, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(local)})
				require.NoError(t, err)
				if !tc.localMissing {
					assertChaosUnchanged(t, localClient, local)
				}
				if !tc.remoteMissing {
					assertChaosUnchanged(t, remoteClient, remote)
				}
			}
		})
	}
}

func TestReconcileOwnedChaos(t *testing.T) {
	t.Run("sync status and finalizers", func(t *testing.T) {
		local, remote := monitorChaosObjects()
		localClient, remoteClient := monitorClient(t, local), monitorClient(t, remote)
		r := New(&v1alpha1.PodChaos{}, localClient, "cluster-a", remoteClient, logr.Discard())
		for range 2 {
			_, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(local)})
			require.NoError(t, err)
			var got v1alpha1.PodChaos
			require.NoError(t, localClient.Get(context.Background(), client.ObjectKeyFromObject(local), &got))
			require.Equal(t, types.UID("management-uid"), got.UID)
			require.Equal(t, local.Spec, got.Spec)
			require.Equal(t, []string{"remote-finalizer"}, got.Finalizers)
			require.Equal(t, v1alpha1.PodChaosStatus{
				Experiment: v1alpha1.ExperimentStatus{
					DesiredPhase: "Run",
					Records:      []*v1alpha1.Record{{Id: "remote-pod", Phase: "Injected", InjectedCount: 1}},
				}}, got.Status)
			require.Nil(t, got.DeletionTimestamp)
			assertChaosUnchanged(t, remoteClient, remote)
		}
	})

	t.Run("remove management object after owned remote disappears", func(t *testing.T) {
		local, _ := monitorChaosObjects()
		localClient := monitorClient(t, local)
		r := New(&v1alpha1.PodChaos{}, localClient, "cluster-a", monitorClient(t), logr.Discard())
		for range 2 {
			_, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(local)})
			require.NoError(t, err)
			var got v1alpha1.PodChaos
			require.True(t, apierrors.IsNotFound(localClient.Get(context.Background(), client.ObjectKeyFromObject(local), &got)))
		}
	})

	t.Run("delete orphaned owned remote while preserving recovery finalizer", func(t *testing.T) {
		_, remote := monitorChaosObjects()
		remoteClient := monitorClient(t, remote)
		r := New(&v1alpha1.PodChaos{}, monitorClient(t), "cluster-a", remoteClient, logr.Discard())
		var deleting v1alpha1.PodChaos
		for i := range 2 {
			_, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(remote)})
			require.NoError(t, err)
			var got v1alpha1.PodChaos
			require.NoError(t, remoteClient.Get(context.Background(), client.ObjectKeyFromObject(remote), &got))
			require.Equal(t, types.UID("remote-uid"), got.UID)
			require.Equal(t, remote.Spec, got.Spec)
			require.Equal(t, remote.Status, got.Status)
			require.Equal(t, []string{"remote-finalizer"}, got.Finalizers)
			require.NotNil(t, got.DeletionTimestamp)
			if i == 1 {
				require.Equal(t, deleting.DeletionTimestamp, got.DeletionTimestamp)
			}
			deleting = got
		}
		require.NoError(t, remoteClient.Get(context.Background(), client.ObjectKeyFromObject(remote), &deleting))
		deleting.Finalizers = nil
		require.NoError(t, remoteClient.Update(context.Background(), &deleting))
		_, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(remote)})
		require.NoError(t, err)
		var got v1alpha1.PodChaos
		require.True(t, apierrors.IsNotFound(remoteClient.Get(context.Background(), client.ObjectKeyFromObject(remote), &got)))
	})
}

func TestReconcilePreservesRetargetedChaosAfterConflict(t *testing.T) {
	for _, remoteMissing := range []bool{false, true} {
		t.Run(fmt.Sprintf("remote missing %t", remoteMissing), func(t *testing.T) {
			local, remote := monitorChaosObjects()
			localClient := monitorClient(t, local)
			retargeted := false
			conflictingClient := interceptor.NewClient(localClient, interceptor.Funcs{
				Update: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
					if !retargeted {
						var latest v1alpha1.PodChaos
						require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(obj), &latest))
						latest.Spec.RemoteCluster = "cluster-b"
						require.NoError(t, c.Update(ctx, &latest))
						retargeted = true
						return apierrors.NewConflict(schema.GroupResource{Group: "chaos-mesh.org", Resource: "podchaos"}, obj.GetName(), fmt.Errorf("experiment retargeted"))
					}
					return c.Update(ctx, obj, opts...)
				},
			})
			var remoteObjects []client.Object
			if !remoteMissing {
				remoteObjects = append(remoteObjects, remote)
			}
			remoteClient := monitorClient(t, remoteObjects...)
			r := New(&v1alpha1.PodChaos{}, conflictingClient, "cluster-a", remoteClient, logr.Discard())
			for range 2 {
				_, _ = r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(local)})
				require.True(t, retargeted, "the first management write must exercise a concurrent retarget")
				want := local.DeepCopy()
				want.Spec.RemoteCluster = "cluster-b"
				assertChaosUnchanged(t, localClient, want)
				if !remoteMissing {
					assertChaosUnchanged(t, remoteClient, remote)
				}
			}
		})
	}
}

func monitorClient(t *testing.T, objects ...client.Object) client.WithWatch {
	t.Helper()
	scheme := runtime.NewScheme()
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
}

func monitorChaosObjects() (*v1alpha1.PodChaos, *v1alpha1.PodChaos) {
	local := &v1alpha1.PodChaos{
		Namespace: "experiments", Name: "same-name", UID: "management-uid",
		Finalizers: []string{"management-finalizer"},
		Spec: v1alpha1.PodChaosSpec{
			Action: "pod-kill", RemoteCluster: "cluster-a", GracePeriod: 7,
			ContainerSelector: v1alpha1.ContainerSelector{PodSelector: v1alpha1.PodSelector{
				Mode: "one", Selector: v1alpha1.PodSelectorSpec{Pods: map[string][]string{"workloads": {"management-pod"}}},
			}},
		},
		Status: v1alpha1.PodChaosStatus{ChaosStatus: v1alpha1.ChaosStatus{
			Experiment: v1alpha1.ExperimentStatus{
				DesiredPhase: "Stop", Records: []*v1alpha1.Record{{Id: "management-pod", Phase: "Not Injected"}},
			},
		}},
	}
	remote := local.DeepCopy()
	remote.UID = "remote-uid"
	remote.Labels = map[string]string{"chaos-mesh.org/controlled-by": "remote-chaos"}
	remote.Finalizers = []string{"remote-finalizer"}
	remote.Spec.RemoteCluster = ""
	remote.Spec.GracePeriod = 13
	remote.Spec.Selector.Pods = map[string][]string{"workloads": {"remote-pod"}}
	remote.Status = v1alpha1.PodChaosStatus{
		Experiment: v1alpha1.ExperimentStatus{
			DesiredPhase: "Run", Records: []*v1alpha1.Record{{Id: "remote-pod", Phase: "Injected", InjectedCount: 1}},
		}}
	return local, remote
}

func assertChaosUnchanged(t *testing.T, c client.Client, want *v1alpha1.PodChaos) {
	t.Helper()
	var got v1alpha1.PodChaos
	require.NoError(t, c.Get(context.Background(), client.ObjectKeyFromObject(want), &got))
	require.Equal(t, want.UID, got.UID)
	require.Equal(t, want.Spec, got.Spec)
	require.Equal(t, want.Status, got.Status)
	require.Equal(t, want.Labels, got.Labels)
	require.Equal(t, want.Finalizers, got.Finalizers)
	require.Equal(t, want.DeletionTimestamp, got.DeletionTimestamp)
}
