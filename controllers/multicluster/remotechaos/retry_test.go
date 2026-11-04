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
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/multicluster/clusterregistry"
)

func TestReconcileRetriesWhenRemoteManagerMissing(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	experiment := &v1alpha1.PodChaos{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "experiment"}, Spec: v1alpha1.PodChaosSpec{RemoteCluster: "not-yet-registered"}}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(experiment).Build()
	r := Reconciler{Client: c, Log: logr.Discard(), Object: &v1alpha1.PodChaos{}, registry: clusterregistry.New(logr.Discard(), c)}
	result, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: types.NamespacedName{Namespace: experiment.Namespace, Name: experiment.Name}})
	require.True(t, err != nil || result.Requeue || result.RequeueAfter > 0, "missing manager must retry without requiring another experiment edit")
}

func TestReconcileRetriesManagementReadFailure(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	c := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{Get: func(context.Context, client.WithWatch, client.ObjectKey, client.Object, ...client.GetOption) error {
		return errors.New("temporary management API failure")
	}}).Build()
	r := Reconciler{Client: c, Log: logr.Discard(), Object: &v1alpha1.PodChaos{}}
	result, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "test", Name: "experiment"}})
	require.True(t, err != nil || result.Requeue || result.RequeueAfter > 0, "management read failure must retry")
}

func TestRemoteErrorsRetryAndRecover(t *testing.T) {
	for _, operation := range []string{"get", "create", "delete"} {
		t.Run(operation, func(t *testing.T) {
			scheme := runtime.NewScheme()
			require.NoError(t, v1alpha1.AddToScheme(scheme))
			obj := &v1alpha1.PodChaos{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "experiment", UID: "local"}, Spec: v1alpha1.PodChaosSpec{RemoteCluster: "a"}}
			objects := []client.Object{}
			if operation == "delete" {
				obj.DeletionTimestamp = &metav1.Time{Time: time.Now()}
				obj.Finalizers = []string{"remote-finalizer"}
				objects = append(objects, &v1alpha1.PodChaos{ObjectMeta: metav1.ObjectMeta{Namespace: obj.Namespace, Name: obj.Name, Labels: map[string]string{v1alpha1.LabelControlledBy: "remote-chaos"}}})
			}
			local := fake.NewClientBuilder().WithScheme(scheme).WithObjects(obj).Build()
			fail := true
			check := func(ctx context.Context) {
				_, ok := ctx.Deadline()
				require.True(t, ok, "every remote call needs a deadline")
			}
			remote := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).WithInterceptorFuncs(interceptor.Funcs{
				Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, out client.Object, opts ...client.GetOption) error {
					check(ctx)
					if fail && operation == "get" {
						return errors.New("unavailable")
					}
					return c.Get(ctx, key, out, opts...)
				},
				Create: func(ctx context.Context, c client.WithWatch, out client.Object, opts ...client.CreateOption) error {
					check(ctx)
					if fail && operation == "create" {
						return errors.New("unavailable")
					}
					return c.Create(ctx, out, opts...)
				},
				Delete: func(ctx context.Context, c client.WithWatch, out client.Object, opts ...client.DeleteOption) error {
					check(ctx)
					if fail && operation == "delete" {
						return errors.New("unavailable")
					}
					return c.Delete(ctx, out, opts...)
				},
			}).Build()
			r := Reconciler{Client: local, Log: logr.Discard(), Object: &v1alpha1.PodChaos{}, registry: testRegistry{remote}}
			req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(obj)}
			result, err := r.Reconcile(context.Background(), req)
			require.NoError(t, err)
			require.True(t, result.Requeue)
			fail = false
			result, err = r.Reconcile(context.Background(), req)
			require.NoError(t, err)
			require.False(t, result.Requeue)
			if operation == "delete" {
				require.Positive(t, result.RequeueAfter)
				result, err = r.Reconcile(context.Background(), req)
				require.NoError(t, err)
				require.Equal(t, ctrl.Result{RequeueAfter: time.Second}, result)
				require.True(t, apierrors.IsNotFound(local.Get(context.Background(), req.NamespacedName, &v1alpha1.PodChaos{})))
			} else {
				require.NoError(t, func() error {
					ctx, cancel := context.WithTimeout(context.Background(), time.Second)
					defer cancel()
					return remote.Get(ctx, req.NamespacedName, &v1alpha1.PodChaos{})
				}())
			}
		})
	}
}

func TestDeletionCleanupRefetchesOnConflictAndRetriesFailure(t *testing.T) {
	for _, failure := range []error{apierrors.NewConflict(schema.GroupResource{Resource: "podchaos"}, "experiment", errors.New("conflict")), errors.New("unavailable")} {
		t.Run(failure.Error(), func(t *testing.T) {
			scheme := runtime.NewScheme()
			require.NoError(t, v1alpha1.AddToScheme(scheme))
			obj := &v1alpha1.PodChaos{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "experiment", UID: "local", DeletionTimestamp: &metav1.Time{Time: time.Now()}, Finalizers: []string{"remote-finalizer"}}, Spec: v1alpha1.PodChaosSpec{RemoteCluster: "a"}}
			updates, gets := 0, 0
			local := fake.NewClientBuilder().WithScheme(scheme).WithObjects(obj).WithInterceptorFuncs(interceptor.Funcs{
				Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, out client.Object, opts ...client.GetOption) error {
					require.NotNil(t, out)
					_, ok := ctx.Deadline()
					require.True(t, ok)
					gets++
					return c.Get(ctx, key, out, opts...)
				},
				Update: func(ctx context.Context, c client.WithWatch, out client.Object, opts ...client.UpdateOption) error {
					updates++
					if updates == 1 {
						return failure
					}
					return c.Update(ctx, out, opts...)
				},
			}).Build()
			remote := fake.NewClientBuilder().WithScheme(scheme).Build()
			r := Reconciler{Client: local, Log: logr.Discard(), Object: &v1alpha1.PodChaos{}, registry: testRegistry{remote}}
			req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(obj)}
			result, err := r.Reconcile(context.Background(), req)
			require.NoError(t, err)
			if apierrors.IsConflict(failure) {
				require.False(t, result.Requeue)
				require.GreaterOrEqual(t, gets, 3)
			} else {
				require.True(t, result.Requeue)
				result, err = r.Reconcile(context.Background(), req)
				require.NoError(t, err)
				require.False(t, result.Requeue)
			}
			require.Equal(t, 2, updates)
		})
	}
}

func TestReconcilePropagatesEarlierCancellation(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	obj := &v1alpha1.PodChaos{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "experiment"}, Spec: v1alpha1.PodChaosSpec{RemoteCluster: "a"}}
	local := fake.NewClientBuilder().WithScheme(scheme).WithObjects(obj).Build()
	remote := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{Get: func(ctx context.Context, _ client.WithWatch, _ client.ObjectKey, _ client.Object, _ ...client.GetOption) error {
		<-ctx.Done()
		return ctx.Err()
	}}).Build()
	r := Reconciler{Client: local, Log: logr.Discard(), Object: &v1alpha1.PodChaos{}, registry: testRegistry{remote}}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	result, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(obj)})
	require.NoError(t, err)
	require.True(t, result.Requeue)
	require.ErrorIs(t, ctx.Err(), context.DeadlineExceeded)
}

func TestDeletionWaitsForLiveAbsence(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	obj := &v1alpha1.PodChaos{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "experiment", UID: "local", Finalizers: []string{"cleanup"}, DeletionTimestamp: &metav1.Time{Time: time.Now()}}, Spec: v1alpha1.PodChaosSpec{RemoteCluster: "a"}}
	local := fake.NewClientBuilder().WithScheme(scheme).WithObjects(obj).Build()
	remoteObj := &v1alpha1.PodChaos{ObjectMeta: metav1.ObjectMeta{Namespace: obj.Namespace, Name: obj.Name, UID: "remote", Labels: map[string]string{v1alpha1.LabelControlledBy: "remote-chaos"}, Finalizers: []string{"remote-cleanup"}}}
	unavailable := true
	remote := fake.NewClientBuilder().WithScheme(scheme).WithObjects(remoteObj).WithInterceptorFuncs(interceptor.Funcs{Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, out client.Object, opts ...client.GetOption) error {
		if unavailable {
			return errors.New("unavailable")
		}
		return c.Get(ctx, key, out, opts...)
	}}).Build()
	r := Reconciler{Client: local, Log: logr.Discard(), Object: &v1alpha1.PodChaos{}, registry: testRegistry{remote}}
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(obj)}
	result, err := r.Reconcile(context.Background(), req)
	require.NoError(t, err)
	require.True(t, result.Requeue)
	current := &v1alpha1.PodChaos{}
	require.NoError(t, local.Get(context.Background(), req.NamespacedName, current))
	require.Equal(t, obj.Finalizers, current.Finalizers)
	unavailable = false
	for i := 0; i < 2; i++ {
		result, err = r.Reconcile(context.Background(), req)
		require.NoError(t, err)
		require.Positive(t, result.RequeueAfter)
		require.NoError(t, local.Get(context.Background(), req.NamespacedName, current))
		require.Equal(t, obj.Finalizers, current.Finalizers)
	}
	require.NoError(t, remote.Get(context.Background(), req.NamespacedName, remoteObj))
	remoteObj.Finalizers = nil
	require.NoError(t, remote.Update(context.Background(), remoteObj))
	result, err = r.Reconcile(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, ctrl.Result{RequeueAfter: time.Second}, result)
	require.True(t, apierrors.IsNotFound(local.Get(context.Background(), req.NamespacedName, current)))
}

func TestDeletionCleanupRejectsChangedLocalIdentity(t *testing.T) {
	for _, change := range []string{"uid", "target", "deletion"} {
		t.Run(change, func(t *testing.T) {
			scheme := runtime.NewScheme()
			require.NoError(t, v1alpha1.AddToScheme(scheme))
			obj := &v1alpha1.PodChaos{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "experiment", UID: "local", Finalizers: []string{"cleanup"}, DeletionTimestamp: &metav1.Time{Time: time.Now()}}, Spec: v1alpha1.PodChaosSpec{RemoteCluster: "a"}}
			gets, updates := 0, 0
			local := fake.NewClientBuilder().WithScheme(scheme).WithObjects(obj).WithInterceptorFuncs(interceptor.Funcs{
				Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, out client.Object, opts ...client.GetOption) error {
					if err := c.Get(ctx, key, out, opts...); err != nil {
						return err
					}
					gets++
					if gets > 1 {
						current := out.(*v1alpha1.PodChaos)
						switch change {
						case "uid":
							current.UID = "replacement"
						case "target":
							current.Spec.RemoteCluster = "b"
						case "deletion":
							current.DeletionTimestamp = nil
						}
					}
					return nil
				},
				Update: func(context.Context, client.WithWatch, client.Object, ...client.UpdateOption) error {
					updates++
					return nil
				},
			}).Build()
			remote := fake.NewClientBuilder().WithScheme(scheme).Build()
			r := Reconciler{Client: local, Log: logr.Discard(), Object: &v1alpha1.PodChaos{}, registry: testRegistry{remote}}
			result, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(obj)})
			require.NoError(t, err)
			require.Equal(t, ctrl.Result{RequeueAfter: time.Second}, result)
			require.Zero(t, updates)
		})
	}
}

type admissionRegistry struct {
	remote client.Client
	err    error
}

func (r *admissionRegistry) WithClient(_ string, f func(client.Client) error) error {
	if r.err != nil {
		return r.err
	}
	return f(r.remote)
}

func TestOwnedDeletionConvergesAfterBusyAndAPIFailure(t *testing.T) {
	ctx := context.Background()
	local, remote, localClient, remoteClient := forwarderObjects(t)
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(local)}
	unavailable := false
	forwarded := interceptor.NewClient(remoteClient, interceptor.Funcs{Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
		if unavailable {
			return errors.New("remote API unavailable")
		}
		return c.Get(ctx, key, obj, opts...)
	}})
	registry := &admissionRegistry{remote: forwarded, err: clusterregistry.ErrBusy}
	r := &Reconciler{Client: localClient, Log: logr.Discard(), Object: &v1alpha1.PodChaos{}, registry: registry}
	for _, blocked := range []string{"busy", "API"} {
		if blocked == "API" {
			registry.err = nil
			unavailable = true
		}
		result, err := r.Reconcile(ctx, req)
		require.NoError(t, err)
		require.Equal(t, ctrl.Result{Requeue: true}, result)
		var gotLocal, gotRemote v1alpha1.PodChaos
		require.NoError(t, localClient.Get(ctx, req.NamespacedName, &gotLocal))
		require.Equal(t, []string{"management-finalizer"}, gotLocal.Finalizers)
		require.NoError(t, remoteClient.Get(ctx, req.NamespacedName, &gotRemote))
		require.Equal(t, remote.UID, gotRemote.UID)
		require.Nil(t, gotRemote.DeletionTimestamp)
	}
	unavailable = false
	for range 2 {
		result, err := r.Reconcile(ctx, req)
		require.NoError(t, err)
		require.Equal(t, ctrl.Result{RequeueAfter: time.Second}, result)
		var gotLocal, gotRemote v1alpha1.PodChaos
		require.NoError(t, localClient.Get(ctx, req.NamespacedName, &gotLocal))
		require.Equal(t, []string{"management-finalizer"}, gotLocal.Finalizers)
		require.NoError(t, remoteClient.Get(ctx, req.NamespacedName, &gotRemote))
		require.Equal(t, remote.UID, gotRemote.UID)
		require.NotNil(t, gotRemote.DeletionTimestamp)
	}
	require.NoError(t, remoteClient.Get(ctx, req.NamespacedName, remote))
	remote.Finalizers = nil
	require.NoError(t, remoteClient.Update(ctx, remote))
	result, err := r.Reconcile(ctx, req)
	require.NoError(t, err)
	require.Equal(t, ctrl.Result{RequeueAfter: time.Second}, result)
	require.True(t, apierrors.IsNotFound(localClient.Get(ctx, req.NamespacedName, &v1alpha1.PodChaos{})))
	result, err = r.Reconcile(ctx, req)
	require.NoError(t, err)
	require.Equal(t, ctrl.Result{}, result)
}
