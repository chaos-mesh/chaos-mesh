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

package remotecluster

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func changeRemoteCluster(t *testing.T, c client.Client, key client.ObjectKey, change string) *v1alpha1.RemoteCluster {
	t.Helper()
	ctx := context.Background()
	var latest v1alpha1.RemoteCluster
	if err := c.Get(ctx, key, &latest); err != nil {
		t.Fatal(err)
	}
	switch change {
	case "replacement":
		replacement := latest.DeepCopy()
		latest.Finalizers = nil
		if err := c.Update(ctx, &latest); err != nil {
			t.Fatal(err)
		}
		if err := c.Delete(ctx, &latest); client.IgnoreNotFound(err) != nil {
			t.Fatal(err)
		}
		replacement.UID = "replacement"
		replacement.ResourceVersion = ""
		replacement.DeletionTimestamp = nil
		if err := c.Create(ctx, replacement); err != nil {
			t.Fatal(err)
		}
	case "generation":
		latest.Generation++
		if err := c.Update(ctx, &latest); err != nil {
			t.Fatal(err)
		}
	case "deletion":
		if err := c.Delete(ctx, &latest); err != nil {
			t.Fatal(err)
		}
	case "observed generation":
		latest.Status.ObservedGeneration = latest.Generation + 1
		if err := c.Status().Update(ctx, &latest); err != nil {
			t.Fatal(err)
		}
	}
	if err := c.Get(ctx, key, &latest); err != nil {
		t.Fatal(err)
	}
	return &latest
}

func TestStaleSetupDoesNotPublishStatus(t *testing.T) {
	for _, change := range []string{"replacement", "generation", "deletion", "observed generation"} {
		for _, failed := range []bool{false, true} {
			t.Run(change+map[bool]string{false: "/success", true: "/failure"}[failed], func(t *testing.T) {
				ctx := context.Background()
				c, obj := fixture(t)
				key := client.ObjectKeyFromObject(obj)
				var changed *v1alpha1.RemoteCluster
				r := Reconciler{Client: c, Log: logr.Discard(),
					ensureHelm: func(context.Context, *v1alpha1.RemoteCluster, clientcmd.ClientConfig) (string, error) {
						return "new-version", nil
					},
					registry: registryStub{ensure: func(context.Context, string, *clientcmdapi.Config) error {
						changed = changeRemoteCluster(t, c, key, change)
						if failed {
							return errors.New("manager failed")
						}
						return nil
					}},
				}
				result, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: key})
				if err != nil || !result.Requeue {
					t.Fatalf("result=%v err=%v", result, err)
				}
				if err := c.Get(ctx, key, obj); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(obj.Status, changed.Status) || obj.ResourceVersion != changed.ResourceVersion {
					t.Fatalf("stale setup changed newer object: %+v", obj.Status)
				}
			})
		}
	}
}

func TestStatusConflictPreservesConcurrentCondition(t *testing.T) {
	ctx := context.Background()
	c, obj := fixture(t)
	writes := 0
	foreign := v1alpha1.RemoteClusterCondition{Type: "ExternalCheck", Status: corev1.ConditionTrue, Reason: "ConcurrentWriter"}
	wrapped := interceptor.NewClient(c.(client.WithWatch), interceptor.Funcs{
		SubResourceUpdate: func(ctx context.Context, c client.Client, subresource string, object client.Object, opts ...client.SubResourceUpdateOption) error {
			writes++
			if writes == 1 {
				var latest v1alpha1.RemoteCluster
				if err := c.Get(ctx, client.ObjectKeyFromObject(object), &latest); err != nil {
					t.Fatal(err)
				}
				latest.Status.Conditions = append(latest.Status.Conditions, foreign)
				if err := c.Status().Update(ctx, &latest); err != nil {
					t.Fatal(err)
				}
				return apierrors.NewConflict(schema.GroupResource{Resource: "remoteclusters"}, object.GetName(), errors.New("concurrent status update"))
			}
			return c.SubResource(subresource).Update(ctx, object, opts...)
		},
	})
	r := Reconciler{Client: wrapped, Log: logr.Discard(),
		ensureHelm: func(context.Context, *v1alpha1.RemoteCluster, clientcmd.ClientConfig) (string, error) {
			return "new-version", nil
		},
		registry: registryStub{ensure: func(context.Context, string, *clientcmdapi.Config) error { return nil }},
	}
	for range 2 {
		result, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(obj)})
		if err != nil || result.Requeue {
			t.Fatalf("result=%v err=%v", result, err)
		}
	}
	if err := c.Get(ctx, client.ObjectKeyFromObject(obj), obj); err != nil {
		t.Fatal(err)
	}
	assertReadyCondition(t, obj, corev1.ConditionTrue)
	found := false
	for _, condition := range obj.Status.Conditions {
		found = found || condition == foreign
	}
	if !found || writes != 2 {
		t.Fatalf("condition preserved=%v, status writes=%d", found, writes)
	}
}

func TestDeleteStatusFailuresPreserveCleanup(t *testing.T) {
	for _, failure := range []string{"initial read", "initial write", "final"} {
		t.Run(failure, func(t *testing.T) {
			ctx := context.Background()
			c, obj := fixture(t)
			obj.Finalizers = append(obj.Finalizers, remoteClusterControllerFinalizer)
			if err := c.Update(ctx, obj); err != nil {
				t.Fatal(err)
			}
			obj.Status.Conditions = append(obj.Status.Conditions, v1alpha1.RemoteClusterCondition{Type: v1alpha1.RemoteClusterConditionInstalled, Status: corev1.ConditionTrue})
			if err := c.Status().Update(ctx, obj); err != nil {
				t.Fatal(err)
			}
			if err := c.Delete(ctx, obj); err != nil {
				t.Fatal(err)
			}
			gets, writes, stops, uninstalls := 0, 0, 0, 0
			wrapped := interceptor.NewClient(c.(client.WithWatch), interceptor.Funcs{
				Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, object client.Object, opts ...client.GetOption) error {
					if _, ok := object.(*v1alpha1.RemoteCluster); ok {
						gets++
						if failure == "initial read" && gets == 2 {
							return errors.New("status read unavailable")
						}
					}
					return c.Get(ctx, key, object, opts...)
				},
				SubResourceUpdate: func(ctx context.Context, c client.Client, subresource string, object client.Object, opts ...client.SubResourceUpdateOption) error {
					writes++
					if failure == "initial write" && writes == 1 || failure == "final" && writes == 2 {
						return errors.New("status unavailable")
					}
					return c.SubResource(subresource).Update(ctx, object, opts...)
				},
			})
			r := Reconciler{Client: wrapped, Log: logr.Discard(),
				registry:      registryStub{stop: func(context.Context, string) error { stops++; return nil }},
				uninstallHelm: func(context.Context, *v1alpha1.RemoteCluster, clientcmd.ClientConfig) error { uninstalls++; return nil },
			}
			result, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(obj)})
			if err != nil || result.Requeue != (failure == "final") || stops != 1 || uninstalls != 1 {
				t.Fatalf("result=%v err=%v stops=%d uninstalls=%d", result, err, stops, uninstalls)
			}
			if err := c.Get(ctx, client.ObjectKeyFromObject(obj), obj); err != nil {
				t.Fatal(err)
			}
			if controllerutil.ContainsFinalizer(obj, remoteClusterControllerFinalizer) != (failure == "final") {
				t.Fatalf("finalizers=%v", obj.Finalizers)
			}
			if ready := assertReadyCondition(t, obj, corev1.ConditionFalse); ready.Reason != "Deleting" {
				t.Fatalf("deletion reason=%q", ready.Reason)
			}
		})
	}
}

func TestStaleDeletionDoesNotRemoveFinalizer(t *testing.T) {
	for _, change := range []string{"replacement", "generation"} {
		t.Run(change, func(t *testing.T) {
			ctx := context.Background()
			c, obj := fixture(t)
			obj.Finalizers = append(obj.Finalizers, remoteClusterControllerFinalizer)
			if err := c.Update(ctx, obj); err != nil {
				t.Fatal(err)
			}
			if err := c.Delete(ctx, obj); err != nil {
				t.Fatal(err)
			}
			key := client.ObjectKeyFromObject(obj)
			var changed *v1alpha1.RemoteCluster
			r := Reconciler{Client: c, Log: logr.Discard(),
				registry: registryStub{stop: func(context.Context, string) error { return nil }},
				uninstallHelm: func(context.Context, *v1alpha1.RemoteCluster, clientcmd.ClientConfig) error {
					changed = changeRemoteCluster(t, c, key, change)
					return nil
				},
			}
			result, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: key})
			if err != nil || !result.Requeue {
				t.Fatalf("result=%v err=%v", result, err)
			}
			if err := c.Get(ctx, key, obj); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(obj.Status, changed.Status) || obj.ResourceVersion != changed.ResourceVersion || !controllerutil.ContainsFinalizer(obj, remoteClusterControllerFinalizer) {
				t.Fatal("stale deletion changed status or removed finalizer")
			}
		})
	}
}

func TestStaleDeletionDoesNotStopReplacement(t *testing.T) {
	ctx := context.Background()
	c, obj := fixture(t)
	obj.Finalizers = append(obj.Finalizers, remoteClusterControllerFinalizer)
	if err := c.Update(ctx, obj); err != nil {
		t.Fatal(err)
	}
	if err := c.Delete(ctx, obj); err != nil {
		t.Fatal(err)
	}
	gets, stops, uninstalls := 0, 0, 0
	wrapped := interceptor.NewClient(c.(client.WithWatch), interceptor.Funcs{
		Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, object client.Object, opts ...client.GetOption) error {
			if _, ok := object.(*v1alpha1.RemoteCluster); ok {
				gets++
				if gets == 2 {
					changeRemoteCluster(t, c, key, "replacement")
				}
			}
			return c.Get(ctx, key, object, opts...)
		},
	})
	r := Reconciler{Client: wrapped, Log: logr.Discard(),
		registry:      registryStub{stop: func(context.Context, string) error { stops++; return nil }},
		uninstallHelm: func(context.Context, *v1alpha1.RemoteCluster, clientcmd.ClientConfig) error { uninstalls++; return nil },
	}
	result, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(obj)})
	if err != nil || !result.Requeue || stops != 0 || uninstalls != 0 {
		t.Fatalf("result=%v err=%v stops=%d uninstalls=%d", result, err, stops, uninstalls)
	}
	if err := c.Get(ctx, client.ObjectKeyFromObject(obj), obj); err != nil {
		t.Fatal(err)
	}
	assertReadyCondition(t, obj, corev1.ConditionTrue)
}

func TestFinalizerConflictDoesNotChangeNewerObject(t *testing.T) {
	for _, deleting := range []bool{false, true} {
		for _, change := range []string{"replacement", "generation"} {
			t.Run(map[bool]string{false: "add/", true: "remove/"}[deleting]+change, func(t *testing.T) {
				ctx := context.Background()
				c, obj := fixture(t)
				if deleting {
					obj.Finalizers = append(obj.Finalizers, remoteClusterControllerFinalizer)
					if err := c.Update(ctx, obj); err != nil {
						t.Fatal(err)
					}
					if err := c.Delete(ctx, obj); err != nil {
						t.Fatal(err)
					}
				}
				key := client.ObjectKeyFromObject(obj)
				updates, setups := 0, 0
				var changed *v1alpha1.RemoteCluster
				wrapped := interceptor.NewClient(c.(client.WithWatch), interceptor.Funcs{
					Update: func(_ context.Context, c client.WithWatch, object client.Object, _ ...client.UpdateOption) error {
						updates++
						changed = changeRemoteCluster(t, c, key, change)
						return apierrors.NewConflict(schema.GroupResource{Resource: "remoteclusters"}, object.GetName(), errors.New("concurrent metadata update"))
					},
				})
				r := Reconciler{Client: wrapped, Log: logr.Discard(),
					ensureHelm: func(context.Context, *v1alpha1.RemoteCluster, clientcmd.ClientConfig) (string, error) {
						setups++
						return "new-version", nil
					},
					registry: registryStub{
						ensure: func(context.Context, string, *clientcmdapi.Config) error { setups++; return nil },
						stop:   func(context.Context, string) error { return nil },
					},
					uninstallHelm: func(context.Context, *v1alpha1.RemoteCluster, clientcmd.ClientConfig) error { return nil },
				}
				result, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: key})
				if err != nil || !result.Requeue || updates != 1 || setups != 0 {
					t.Fatalf("result=%v err=%v updates=%d setups=%d", result, err, updates, setups)
				}
				if err := c.Get(ctx, key, obj); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(obj.Status, changed.Status) || !reflect.DeepEqual(obj.Finalizers, changed.Finalizers) || obj.ResourceVersion != changed.ResourceVersion {
					t.Fatal("stale retry changed newer object")
				}
			})
		}
	}
}

func TestFinalizerFailureClearsReady(t *testing.T) {
	ctx := context.Background()
	c, obj := fixture(t)
	before := obj.Status.DeepCopy()
	wrapped := interceptor.NewClient(c.(client.WithWatch), interceptor.Funcs{
		Update: func(context.Context, client.WithWatch, client.Object, ...client.UpdateOption) error {
			return errors.New("metadata unavailable with private-value")
		},
	})
	r := Reconciler{Client: wrapped, Log: logr.Discard()}
	result, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(obj)})
	if err != nil || !result.Requeue {
		t.Fatalf("result=%v err=%v", result, err)
	}
	if err := c.Get(ctx, client.ObjectKeyFromObject(obj), obj); err != nil {
		t.Fatal(err)
	}
	ready := assertReadyCondition(t, obj, corev1.ConditionFalse)
	if ready.Reason != "FinalizerUpdateFailed" || obj.Status.CurrentVersion != before.CurrentVersion || obj.Status.ObservedGeneration != before.ObservedGeneration || controllerutil.ContainsFinalizer(obj, remoteClusterControllerFinalizer) {
		t.Fatalf("metadata failure status=%+v finalizers=%v", obj.Status, obj.Finalizers)
	}
}
