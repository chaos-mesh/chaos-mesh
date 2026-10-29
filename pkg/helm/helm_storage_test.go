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

package helm

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/go-logr/logr"
	"helm.sh/helm/v4/pkg/kube"
	"helm.sh/helm/v4/pkg/storage/driver"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestGetHelm3Release(t *testing.T) {
	const legacyRelease = `{"name":"chaos-mesh","namespace":"remote","version":2,"info":{"status":"deployed"},"chart":{"metadata":{"name":"chaos-mesh","version":"2.8.0","appVersion":"2.8.0","apiVersion":"v2"}},"config":{"dashboard":{"create":false}}}`
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	if _, err := writer.Write([]byte(legacyRelease)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sh.helm.release.v1.chaos-mesh.v2",
			Namespace: "remote",
			Labels:    map[string]string{"owner": "helm", "name": "chaos-mesh", "status": "deployed", "version": "2"},
		},
		Type: "helm.sh/release.v1",
		Data: map[string][]byte{"release": []byte(base64.StdEncoding.EncodeToString(compressed.Bytes()))},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/version":
			_, _ = w.Write([]byte(`{"major":"1","minor":"35","gitVersion":"v1.35.3"}`))
		case "/api/v1/namespaces/remote/secrets", "/api/v1/namespaces/other/secrets":
			result := corev1.SecretList{TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "SecretList"}}
			if r.URL.Path == "/api/v1/namespaces/remote/secrets" && r.URL.Query().Get("labelSelector") == "name=chaos-mesh,owner=helm" {
				result.Items = []corev1.Secret{secret}
			}
			if err := json.NewEncoder(w).Encode(result); err != nil {
				t.Error(err)
			}
		default:
			t.Errorf("unexpected Kubernetes request %s %s", r.Method, r.URL)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	config := clientcmdapi.Config{
		Clusters:       map[string]*clientcmdapi.Cluster{"remote": {Server: server.URL}},
		Contexts:       map[string]*clientcmdapi.Context{"remote": {Cluster: "remote"}},
		CurrentContext: "remote",
	}
	getter := NewRESTClientGetter(clientcmd.NewDefaultClientConfig(config, nil))
	client, err := NewHelmClient(getter, logr.Discard())
	if err != nil {
		t.Fatal(err)
	}
	release, err := client.GetRelease("remote", "chaos-mesh")
	if err != nil {
		t.Fatal(err)
	}
	if release.Name != "chaos-mesh" || release.Namespace != "remote" || release.Version != 2 || release.Chart.AppVersion() != "2.8.0" {
		t.Fatalf("unexpected legacy release: %#v", release)
	}
	if release.ApplyMethod != "" {
		t.Fatalf("legacy release acquired an apply method: %q", release.ApplyMethod)
	}
	_, err = client.GetRelease("other", "chaos-mesh")
	if !errors.Is(err, driver.ErrReleaseNotFound) {
		t.Fatalf("expected a missing release in the other namespace, got %v", err)
	}
}

func TestHelmFieldOwnership(t *testing.T) {
	var applied atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/api/v1/namespaces/remote/configmaps/release" {
			t.Errorf("unexpected Kubernetes request %s %s", r.Method, r.URL)
			http.NotFound(w, r)
			return
		}
		applied.Store(true)
		if got := r.URL.Query().Get("fieldManager"); got != "helm" {
			t.Errorf("field manager = %q, want helm to retain CLI field ownership", got)
		}
		if r.URL.Query().Get("force") == "true" {
			t.Error("must not force ownership of fields managed by other clients")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"release","namespace":"remote"}}`))
	}))
	defer server.Close()
	config := clientcmdapi.Config{
		Clusters:       map[string]*clientcmdapi.Cluster{"remote": {Server: server.URL}},
		Contexts:       map[string]*clientcmdapi.Context{"remote": {Cluster: "remote"}},
		CurrentContext: "remote",
	}
	getter := NewRESTClientGetter(clientcmd.NewDefaultClientConfig(config, nil))
	client, err := NewHelmClient(getter, logr.Discard())
	if err != nil {
		t.Fatal(err)
	}
	configuration, err := client.spawnConfigurationWithNamespace("remote")
	if err != nil {
		t.Fatal(err)
	}
	restConfig, err := getter.ToRESTConfig()
	if err != nil {
		t.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		t.Fatal(err)
	}
	resources := kube.ResourceList{&resource.Info{
		Client: clientset.CoreV1().RESTClient(),
		Mapping: &meta.RESTMapping{
			Resource:         schema.GroupVersionResource{Version: "v1", Resource: "configmaps"},
			GroupVersionKind: schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"},
			Scope:            meta.RESTScopeNamespace,
		},
		Name:      "release",
		Namespace: "remote",
		Object: &corev1.ConfigMap{
			TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"},
			ObjectMeta: metav1.ObjectMeta{Name: "release", Namespace: "remote"},
		},
	}}
	if _, err := configuration.KubeClient.Create(resources, kube.ClientCreateOptionServerSideApply(true, false)); err != nil {
		t.Fatal(err)
	}
	if !applied.Load() {
		t.Fatal("no apply request reached Kubernetes")
	}
}
