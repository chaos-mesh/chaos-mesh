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
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type registryStub struct {
	ensure func(context.Context, string, *clientcmdapi.Config) error
	stop   func(context.Context, string) error
}

func (r registryStub) Ensure(ctx context.Context, name string, cfg *clientcmdapi.Config) error {
	return r.ensure(ctx, name, cfg)
}
func (r registryStub) Stop(ctx context.Context, name string) error { return r.stop(ctx, name) }

func fixture(t *testing.T) (client.Client, *v1alpha1.RemoteCluster) {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	obj := &v1alpha1.RemoteCluster{Name: "remote", Generation: 2, Finalizers: []string{"other-owner"}}
	obj.Spec.KubeConfig.SecretRef = v1alpha1.RemoteClusterSecretRef{Namespace: "ns", Name: "config", Key: "kubeconfig"}
	obj.Status.CurrentVersion = "old-version"
	obj.Status.Conditions = []v1alpha1.RemoteClusterCondition{{Type: v1alpha1.RemoteClusterConditionReady, Status: corev1.ConditionTrue}}
	secret := &corev1.Secret{Namespace: "ns", Name: "config", Data: map[string][]byte{"kubeconfig": []byte(`apiVersion: v1
kind: Config
current-context: remote
contexts:
- name: remote
  context: {cluster: remote, user: remote}
clusters:
- name: remote
  cluster: {server: https://remote.example}
users:
- name: remote
  user: {token: old}
`)}}
	return fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(obj).WithObjects(obj, secret).Build(), obj
}

func TestReconcileManagerAndStatus(t *testing.T) {
	for _, test := range []struct {
		name     string
		observed int64
		fail     bool
		ready    corev1.ConditionStatus
	}{
		{"restart", 2, false, corev1.ConditionTrue},
		{"restart without Ready", 2, false, ""},
		{"new generation without Ready", 1, false, ""},
		{"new generation with Ready False", 1, false, corev1.ConditionFalse},
		{"replacement fails", 1, true, corev1.ConditionTrue},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			c, obj := fixture(t)
			obj.Status.ObservedGeneration = test.observed
			obj.Status.Conditions = nil
			if test.ready != "" {
				obj.Status.Conditions = []v1alpha1.RemoteClusterCondition{{Type: v1alpha1.RemoteClusterConditionReady, Status: test.ready}}
			}
			if err := c.Status().Update(ctx, obj); err != nil {
				t.Fatal(err)
			}
			helmCalls, ensureCalls := 0, 0
			assertFinalizer := func() {
				var latest v1alpha1.RemoteCluster
				if err := c.Get(ctx, client.ObjectKeyFromObject(obj), &latest); err != nil {
					t.Fatal(err)
				}
				if !controllerutil.ContainsFinalizer(&latest, remoteClusterControllerFinalizer) || !controllerutil.ContainsFinalizer(&latest, "other-owner") {
					t.Fatal("finalizers missing before external effect")
				}
			}
			r := Reconciler{Client: c, Log: logr.Discard()}
			r.ensureHelm = func(context.Context, *v1alpha1.RemoteCluster, clientcmd.ClientConfig) (string, error) {
				assertFinalizer()
				helmCalls++
				return "new-version", nil
			}
			r.registry = registryStub{ensure: func(context.Context, string, *clientcmdapi.Config) error {
				assertFinalizer()
				ensureCalls++
				if test.fail {
					return errors.New("start failed")
				}
				return nil
			}}
			result, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(obj)})
			if err != nil || result.Requeue != test.fail {
				t.Fatalf("result %v, err %v", result, err)
			}
			var latest v1alpha1.RemoteCluster
			if err := c.Get(ctx, client.ObjectKeyFromObject(obj), &latest); err != nil {
				t.Fatal(err)
			}
			wantObserved := int64(2)
			if test.fail {
				wantObserved = test.observed
			}
			if latest.Status.ObservedGeneration != wantObserved {
				t.Fatalf("observed %d", latest.Status.ObservedGeneration)
			}
			wantHelm := 1
			if test.observed == 2 {
				wantHelm = 0
			}
			if helmCalls != wantHelm || ensureCalls != 1 {
				t.Fatalf("helm=%d ensure=%d", helmCalls, ensureCalls)
			}
			if test.fail {
				return
			}
			if test.observed == 2 && latest.Status.CurrentVersion != "old-version" {
				t.Fatal("restart lost current version")
			}
			if len(latest.Status.Conditions) != 2 {
				t.Fatalf("missing Installed or Ready condition: %v", latest.Status.Conditions)
			}
			for _, cond := range latest.Status.Conditions {
				if cond.Status != corev1.ConditionTrue {
					t.Fatalf("condition lost: %+v", cond)
				}
			}
			before := latest.ResourceVersion
			if _, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(obj)}); err != nil {
				t.Fatal(err)
			}
			if err := c.Get(ctx, client.ObjectKeyFromObject(obj), &latest); err != nil {
				t.Fatal(err)
			}
			if latest.ResourceVersion != before {
				t.Fatal("unchanged reconcile wrote status or metadata")
			}
			if helmCalls != wantHelm || ensureCalls != 2 {
				t.Fatalf("repeat helm=%d ensure=%d", helmCalls, ensureCalls)
			}
		})
	}
}

func TestDeletionStopsBeforeReadingSecret(t *testing.T) {
	ctx := context.Background()
	c, obj := fixture(t)
	obj.Finalizers = append(obj.Finalizers, remoteClusterControllerFinalizer)
	if err := c.Update(ctx, obj); err != nil {
		t.Fatal(err)
	}
	if err := c.Delete(ctx, obj); err != nil {
		t.Fatal(err)
	}
	if err := c.Delete(ctx, &corev1.Secret{Namespace: "ns", Name: "config"}); err != nil {
		t.Fatal(err)
	}
	stops := 0
	r := Reconciler{Client: c, Log: logr.Discard(), registry: registryStub{stop: func(context.Context, string) error { stops++; return nil }}}
	result, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(obj)})
	if err != nil || !result.Requeue || stops != 1 {
		t.Fatalf("result=%v err=%v stops=%d", result, err, stops)
	}
	var latest v1alpha1.RemoteCluster
	if err := c.Get(ctx, client.ObjectKeyFromObject(obj), &latest); err != nil {
		t.Fatal(err)
	}
	if !controllerutil.ContainsFinalizer(&latest, remoteClusterControllerFinalizer) {
		t.Fatal("cleanup finalizer removed before uninstall")
	}
	assertReadyCondition(t, &latest, corev1.ConditionFalse)
}

func TestSecretRequests(t *testing.T) {
	ctx := context.Background()
	c, obj := fixture(t)
	other := obj.DeepCopy()
	other.Name = "other"
	other.ResourceVersion = ""
	other.Spec.KubeConfig.SecretRef.Namespace = "other"
	if err := c.Create(ctx, other); err != nil {
		t.Fatal(err)
	}
	r := Reconciler{Client: c, Log: logr.Discard()}
	for _, key := range []types.NamespacedName{{Namespace: "ns", Name: "config"}, {Namespace: "ns", Name: "missing"}, {Namespace: "missing", Name: "config"}} {
		got := r.requestsForSecret(ctx, &corev1.Secret{Namespace: key.Namespace, Name: key.Name})
		if key.Namespace == "ns" && key.Name == "config" {
			if len(got) != 1 || got[0].Name != "remote" || got[0].Namespace != "" {
				t.Fatalf("requests=%v", got)
			}
		} else if len(got) != 0 {
			t.Fatalf("unexpected requests=%v", got)
		}
	}
}

func TestDeletionPreservesOtherFinalizers(t *testing.T) {
	ctx := context.Background()
	c, obj := fixture(t)
	obj.Finalizers = append(obj.Finalizers, remoteClusterControllerFinalizer)
	if err := c.Update(ctx, obj); err != nil {
		t.Fatal(err)
	}
	if err := c.Delete(ctx, obj); err != nil {
		t.Fatal(err)
	}
	var events []string
	r := Reconciler{Client: c, Log: logr.Discard(), registry: registryStub{stop: func(context.Context, string) error {
		events = append(events, "stop")
		return nil
	}}, uninstallHelm: func(context.Context, *v1alpha1.RemoteCluster, clientcmd.ClientConfig) error {
		events = append(events, "uninstall")
		return nil
	}}
	result, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(obj)})
	if err != nil || result.Requeue {
		t.Fatalf("result=%v err=%v", result, err)
	}
	if len(events) != 2 || events[0] != "stop" || events[1] != "uninstall" {
		t.Fatalf("events=%v", events)
	}
	var latest v1alpha1.RemoteCluster
	if err := c.Get(ctx, client.ObjectKeyFromObject(obj), &latest); err != nil {
		t.Fatal(err)
	}
	if len(latest.Finalizers) != 1 || latest.Finalizers[0] != "other-owner" {
		t.Fatalf("finalizers=%v", latest.Finalizers)
	}
	for _, cond := range latest.Status.Conditions {
		if cond.Type == v1alpha1.RemoteClusterConditionInstalled && cond.Status != corev1.ConditionFalse {
			t.Fatal("Installed remained true")
		}
		if cond.Type == v1alpha1.RemoteClusterConditionReady && cond.Status != corev1.ConditionFalse {
			t.Fatal("Ready remained true during deletion")
		}
	}
}

func TestInvalidInitialCredentialsDoNotAddFinalizer(t *testing.T) {
	for _, invalid := range []string{"missing secret", "missing key", "invalid context"} {
		t.Run(invalid, func(t *testing.T) {
			ctx := context.Background()
			c, obj := fixture(t)
			var secret corev1.Secret
			if err := c.Get(ctx, types.NamespacedName{Namespace: "ns", Name: "config"}, &secret); err != nil {
				t.Fatal(err)
			}
			if invalid == "missing secret" {
				if err := c.Delete(ctx, &secret); err != nil {
					t.Fatal(err)
				}
			} else {
				if invalid == "missing key" {
					secret.Data = nil
				} else {
					secret.Data["kubeconfig"] = []byte("apiVersion: v1\nkind: Config\ncurrent-context: missing\n")
				}
				if err := c.Update(ctx, &secret); err != nil {
					t.Fatal(err)
				}
			}
			r := Reconciler{Client: c, Log: logr.Discard()}
			result, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(obj)})
			if err != nil || !result.Requeue {
				t.Fatalf("result=%v err=%v", result, err)
			}
			var latest v1alpha1.RemoteCluster
			if err := c.Get(ctx, client.ObjectKeyFromObject(obj), &latest); err != nil {
				t.Fatal(err)
			}
			if controllerutil.ContainsFinalizer(&latest, remoteClusterControllerFinalizer) {
				t.Fatal("invalid initial credentials gained cleanup finalizer")
			}
		})
	}
}

func assertReadyCondition(t *testing.T, obj *v1alpha1.RemoteCluster, status corev1.ConditionStatus) v1alpha1.RemoteClusterCondition {
	t.Helper()
	for _, condition := range obj.Status.Conditions {
		if condition.Type == v1alpha1.RemoteClusterConditionReady {
			if condition.Status != status {
				t.Fatalf("Ready=%s, want %s", condition.Status, status)
			}
			return condition
		}
	}
	t.Fatal("Ready condition missing")
	return v1alpha1.RemoteClusterCondition{}
}

func TestSetupFailureClearsReadyAndRecovers(t *testing.T) {
	for _, failure := range []string{"missing secret", "missing key", "malformed kubeconfig", "invalid context", "helm", "manager"} {
		t.Run(failure, func(t *testing.T) {
			ctx := context.Background()
			c, obj := fixture(t)
			obj.Status.ObservedGeneration = 1
			foreign := v1alpha1.RemoteClusterCondition{Type: "ExternalCheck", Status: corev1.ConditionTrue, Reason: "OtherOwner"}
			installed := v1alpha1.RemoteClusterCondition{Type: v1alpha1.RemoteClusterConditionInstalled, Status: corev1.ConditionTrue}
			obj.Status.Conditions = append(obj.Status.Conditions, installed, foreign)
			if err := c.Status().Update(ctx, obj); err != nil {
				t.Fatal(err)
			}
			var secret corev1.Secret
			if err := c.Get(ctx, types.NamespacedName{Namespace: "ns", Name: "config"}, &secret); err != nil {
				t.Fatal(err)
			}
			validData := secret.DeepCopy().Data
			switch failure {
			case "missing secret":
				if err := c.Delete(ctx, &secret); err != nil {
					t.Fatal(err)
				}
			case "missing key", "malformed kubeconfig", "invalid context":
				switch failure {
				case "missing key":
					secret.Data = nil
				case "malformed kubeconfig":
					secret.Data["kubeconfig"] = []byte("clusters: [")
				case "invalid context":
					secret.Data["kubeconfig"] = []byte("apiVersion: v1\nkind: Config\ncurrent-context: missing\n")
				}
				if err := c.Update(ctx, &secret); err != nil {
					t.Fatal(err)
				}
			}
			failed := true
			helmCalls, ensureCalls := 0, 0
			r := Reconciler{Client: c, Log: logr.Discard(),
				ensureHelm: func(context.Context, *v1alpha1.RemoteCluster, clientcmd.ClientConfig) (string, error) {
					helmCalls++
					if failed && failure == "helm" {
						return "new-version", errors.New("helm failure with private-value")
					}
					return "new-version", nil
				},
				registry: registryStub{ensure: func(context.Context, string, *clientcmdapi.Config) error {
					ensureCalls++
					if failed && failure == "manager" {
						return errors.New("manager failure with private-value")
					}
					return nil
				}},
			}
			req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(obj)}
			var previous v1alpha1.RemoteClusterStatus
			var previousVersion string
			for attempt := range 2 {
				result, err := r.Reconcile(ctx, req)
				if err != nil || !result.Requeue {
					t.Fatalf("result=%v err=%v", result, err)
				}
				if err := c.Get(ctx, req.NamespacedName, obj); err != nil {
					t.Fatal(err)
				}
				ready := assertReadyCondition(t, obj, corev1.ConditionFalse)
				if ready.Reason == "" || strings.Contains(ready.Reason, "private-value") {
					t.Fatalf("failure reason must be stable and exclude raw error details: %q", ready.Reason)
				}
				if obj.Status.ObservedGeneration != 1 || obj.Status.CurrentVersion != "old-version" {
					t.Fatalf("failure advanced successful setup status: %+v", obj.Status)
				}
				for _, want := range []v1alpha1.RemoteClusterCondition{installed, foreign} {
					found := false
					for _, condition := range obj.Status.Conditions {
						found = found || condition == want
					}
					if !found {
						t.Fatalf("failure changed unrelated condition %+v: %+v", want, obj.Status.Conditions)
					}
				}
				if !controllerutil.ContainsFinalizer(obj, "other-owner") {
					t.Fatal("failure lost another owner's finalizer")
				}
				if attempt > 0 && (!reflect.DeepEqual(previous, obj.Status) || previousVersion != obj.ResourceVersion) {
					t.Fatal("repeated failure changed status or metadata")
				}
				previous, previousVersion = obj.Status, obj.ResourceVersion
			}
			if failure != "helm" && failure != "manager" && (helmCalls != 0 || ensureCalls != 0) {
				t.Fatal("invalid credentials reached remote setup")
			}
			if failure == "helm" && ensureCalls != 0 {
				t.Fatal("failed Helm setup started a manager")
			}
			failed = false
			secret.Data = validData
			if failure == "missing secret" {
				secret.ResourceVersion = ""
				if err := c.Create(ctx, &secret); err != nil {
					t.Fatal(err)
				}
			} else if err := c.Update(ctx, &secret); err != nil {
				t.Fatal(err)
			}
			result, err := r.Reconcile(ctx, req)
			if err != nil || result.Requeue {
				t.Fatalf("recovery result=%v err=%v", result, err)
			}
			if err := c.Get(ctx, req.NamespacedName, obj); err != nil {
				t.Fatal(err)
			}
			ready := assertReadyCondition(t, obj, corev1.ConditionTrue)
			if ready.Reason != "" || obj.Status.ObservedGeneration != 2 || obj.Status.CurrentVersion != "new-version" {
				t.Fatalf("recovery did not publish successful setup: %+v", obj.Status)
			}
		})
	}
}

func TestBlockedDeletionClearsReady(t *testing.T) {
	for _, failure := range []string{"stop", "uninstall"} {
		t.Run(failure, func(t *testing.T) {
			ctx := context.Background()
			c, obj := fixture(t)
			obj.Finalizers = append(obj.Finalizers, remoteClusterControllerFinalizer)
			if err := c.Update(ctx, obj); err != nil {
				t.Fatal(err)
			}
			if err := c.Delete(ctx, obj); err != nil {
				t.Fatal(err)
			}
			uninstalls := 0
			r := Reconciler{Client: c, Log: logr.Discard(),
				registry: registryStub{stop: func(context.Context, string) error {
					if failure == "stop" {
						return errors.New("stop blocked")
					}
					return nil
				}},
				uninstallHelm: func(context.Context, *v1alpha1.RemoteCluster, clientcmd.ClientConfig) error {
					uninstalls++
					return errors.New("uninstall blocked")
				},
			}
			var version string
			for attempt := range 2 {
				result, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(obj)})
				if err != nil || !result.Requeue {
					t.Fatalf("result=%v err=%v", result, err)
				}
				if err := c.Get(ctx, client.ObjectKeyFromObject(obj), obj); err != nil {
					t.Fatal(err)
				}
				assertReadyCondition(t, obj, corev1.ConditionFalse)
				if !controllerutil.ContainsFinalizer(obj, remoteClusterControllerFinalizer) || !controllerutil.ContainsFinalizer(obj, "other-owner") {
					t.Fatalf("blocked deletion lost finalizers: %v", obj.Finalizers)
				}
				if attempt > 0 && version != obj.ResourceVersion {
					t.Fatal("repeated blocked deletion wrote unchanged status")
				}
				version = obj.ResourceVersion
			}
			if failure == "stop" && uninstalls != 0 {
				t.Fatal("uninstalled while manager stop was blocked")
			}
		})
	}
}

func TestHelmLifecycleManaged(t *testing.T) {
	RegisterTestingT(t)

	tests := []struct {
		name        string
		annotations map[string]string
		expected    bool
	}{
		{
			name:        "nil annotations defaults to managed",
			annotations: nil,
			expected:    true,
		},
		{
			name:        "empty annotations defaults to managed",
			annotations: map[string]string{},
			expected:    true,
		},
		{
			name: "annotation absent defaults to managed",
			annotations: map[string]string{
				"some-other-annotation": "value",
			},
			expected: true,
		},
		{
			name: "annotation set to true means managed",
			annotations: map[string]string{
				v1alpha1.AnnotationManagedHelmLifecycle: "true",
			},
			expected: true,
		},
		{
			name: "annotation set to false means not managed",
			annotations: map[string]string{
				v1alpha1.AnnotationManagedHelmLifecycle: "false",
			},
			expected: false,
		},
		{
			name: "annotation set to empty string means managed",
			annotations: map[string]string{
				v1alpha1.AnnotationManagedHelmLifecycle: "",
			},
			expected: true,
		},
		{
			name: "annotation set to arbitrary value means managed",
			annotations: map[string]string{
				v1alpha1.AnnotationManagedHelmLifecycle: "yes",
			},
			expected: true,
		},
		{
			name: "annotation set to False (capitalized) means managed",
			annotations: map[string]string{
				v1alpha1.AnnotationManagedHelmLifecycle: "False",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterTestingT(t)
			obj := &v1alpha1.RemoteCluster{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tt.annotations,
				},
			}
			Expect(helmLifecycleManaged(obj)).To(Equal(tt.expected))
		})
	}
}

func TestReconcileWithoutManagedHelm(t *testing.T) {
	for _, observed := range []int64{0, 2} {
		t.Run(fmt.Sprint(observed), func(t *testing.T) {
			ctx := context.Background()
			c, obj := fixture(t)
			obj.Annotations = map[string]string{v1alpha1.AnnotationManagedHelmLifecycle: "false"}
			obj.Spec.Version = "self-managed-version"
			if err := c.Update(ctx, obj); err != nil {
				t.Fatal(err)
			}
			obj.Status.ObservedGeneration = observed
			if err := c.Status().Update(ctx, obj); err != nil {
				t.Fatal(err)
			}
			ensures := 0
			r := Reconciler{Client: c, Log: logr.Discard(),
				ensureHelm: func(context.Context, *v1alpha1.RemoteCluster, clientcmd.ClientConfig) (string, error) {
					t.Fatal("Helm install/upgrade called for self-managed cluster")
					return "", nil
				},
				registry: registryStub{ensure: func(context.Context, string, *clientcmdapi.Config) error {
					ensures++
					return nil
				}},
			}
			result, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(obj)})
			if err != nil || result.Requeue || ensures != 1 {
				t.Fatalf("result=%v err=%v ensures=%d", result, err, ensures)
			}
			var latest v1alpha1.RemoteCluster
			if err := c.Get(ctx, client.ObjectKeyFromObject(obj), &latest); err != nil {
				t.Fatal(err)
			}
			if latest.Status.CurrentVersion != obj.Spec.Version || latest.Status.ObservedGeneration != obj.Generation {
				t.Fatalf("unexpected status: %+v", latest.Status)
			}
			if !controllerutil.ContainsFinalizer(&latest, remoteClusterControllerFinalizer) || !controllerutil.ContainsFinalizer(&latest, "other-owner") {
				t.Fatalf("finalizers=%v", latest.Finalizers)
			}
			assertReadyCondition(t, &latest, corev1.ConditionTrue)
		})
	}
}

func TestDeleteWithoutManagedHelm(t *testing.T) {
	for _, missingSecret := range []bool{false, true} {
		t.Run(fmt.Sprint(missingSecret), func(t *testing.T) {
			ctx := context.Background()
			c, obj := fixture(t)
			obj.Annotations = map[string]string{v1alpha1.AnnotationManagedHelmLifecycle: "false"}
			obj.Finalizers = append(obj.Finalizers, remoteClusterControllerFinalizer)
			if err := c.Update(ctx, obj); err != nil {
				t.Fatal(err)
			}
			if err := c.Delete(ctx, obj); err != nil {
				t.Fatal(err)
			}
			if missingSecret {
				if err := c.Delete(ctx, &corev1.Secret{Namespace: "ns", Name: "config"}); err != nil {
					t.Fatal(err)
				}
			}
			stops := 0
			r := Reconciler{Client: c, Log: logr.Discard(),
				uninstallHelm: func(context.Context, *v1alpha1.RemoteCluster, clientcmd.ClientConfig) error {
					t.Fatal("Helm uninstall called for self-managed cluster")
					return nil
				},
				registry: registryStub{stop: func(context.Context, string) error { stops++; return nil }},
			}
			result, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(obj)})
			if err != nil || result.Requeue || stops != 1 {
				t.Fatalf("result=%v err=%v stops=%d", result, err, stops)
			}
			var latest v1alpha1.RemoteCluster
			if err := c.Get(ctx, client.ObjectKeyFromObject(obj), &latest); err != nil {
				t.Fatal(err)
			}
			if len(latest.Finalizers) != 1 || latest.Finalizers[0] != "other-owner" {
				t.Fatalf("finalizers=%v", latest.Finalizers)
			}
			assertReadyCondition(t, &latest, corev1.ConditionFalse)
		})
	}
}
