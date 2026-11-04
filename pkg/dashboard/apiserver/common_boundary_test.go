// Copyright Chaos Mesh Authors.
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

package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/common"
)

func commonBoundaryRouter(t *testing.T, conf *config.ChaosDashboardConfig, faults interceptor.Funcs) *gin.Engine {
	t.Helper()
	objects := []client.Object{
		&corev1.Namespace{Name: "target", Annotations: map[string]string{"chaos-mesh.org/inject": "enabled"}},
		&corev1.Namespace{Name: "other"},
	}
	for _, entry := range []struct{ name, namespace, zone string }{
		{"first", "target", "west"}, {"second", "target", "west"}, {"third", "target", "east"}, {"outside", "other", "outside"},
	} {
		objects = append(objects, &v1alpha1.PhysicalMachine{
			Name: entry.name, Namespace: entry.namespace, Labels: map[string]string{"zone": entry.zone}, Annotations: map[string]string{"owner": entry.zone},
			Spec: v1alpha1.PhysicalMachineSpec{Address: "http://" + entry.name + ":31767"},
		})
		objects = append(objects, &corev1.Pod{Name: entry.name, Namespace: entry.namespace})
	}
	_, kube := apiRouter(t, objects, faults, nil)
	router := newEngine(conf)
	common.Register(router.Group("/api"), common.NewService(conf, kube, logr.Discard()))
	return router
}

func TestPhysicalMachineSelectionAndMetadata(t *testing.T) {
	for _, scenario := range []struct {
		name, body      string
		cluster, filter bool
		want            []string
	}{
		{"labels", `{"namespaces":["target"],"labelSelectors":{"zone":"west"}}`, false, false, []string{"first", "second"}},
		{"explicit names", `{"physicalMachines":{"target":["first","missing"],"other":["outside"]}}`, false, false, []string{"first"}},
		{"explicit names with namespace filter", `{"physicalMachines":{"target":["first"],"other":["outside"]}}`, true, true, []string{"first"}},
		{"annotation", `{"namespaces":["target"],"annotationSelectors":{"owner":"east"}}`, false, false, []string{"third"}},
		{"namespace filter", `{"namespaces":["target","other"]}`, true, true, []string{"first", "second", "third"}},
		{"namespace filter without explicit namespace", `{}`, true, true, []string{"first", "second", "third"}},
		{"filter disabled", `{"namespaces":["target","other"]}`, true, false, []string{"first", "second", "third", "outside"}},
		{"empty matches", `{"namespaces":["target"],"labelSelectors":{"zone":"absent"}}`, false, false, []string{}},
		{"default namespace", `{}`, false, false, []string{"first", "second", "third"}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			router := commonBoundaryRouter(t, &config.ChaosDashboardConfig{TargetNamespace: "target", ClusterScoped: scenario.cluster, EnableFilterNamespace: scenario.filter}, interceptor.Funcs{})
			response := apiRequest(t, router, http.MethodPost, "/common/physicalmachines", scenario.body, 200)
			var machines []struct{ Name, Namespace, Address string }
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &machines))
			names := make([]string, 0, len(machines))
			for _, machine := range machines {
				names = append(names, machine.Name)
				require.Equal(t, "http://"+machine.Name+":31767", machine.Address)
				if machine.Name == "outside" {
					require.Equal(t, "other", machine.Namespace)
				} else {
					require.Equal(t, "target", machine.Namespace)
				}
			}
			require.ElementsMatch(t, scenario.want, names)
		})
	}
	for _, endpoint := range []struct{ path, key string }{{"physicalmachine-labels", "zone"}, {"physicalmachine-annotations", "owner"}} {
		t.Run(endpoint.path, func(t *testing.T) {
			router := commonBoundaryRouter(t, &config.ChaosDashboardConfig{TargetNamespace: "target", ClusterScoped: true, EnableFilterNamespace: true}, interceptor.Funcs{})
			response := apiRequest(t, router, http.MethodGet, "/common/"+endpoint.path+"?physicalMachineNamespaceList=target,other", "", 200)
			var values map[string][]string
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &values))
			require.ElementsMatch(t, []string{"west", "east"}, values[endpoint.key])
		})
	}
}

func TestAvailableNamespaceConfiguration(t *testing.T) {
	for _, scenario := range []struct {
		name            string
		cluster, filter bool
		want            []string
	}{
		{"namespaced", false, false, []string{"target"}},
		{"cluster", true, false, []string{"other", "target"}},
		{"filtered cluster", true, true, []string{"target"}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			router := commonBoundaryRouter(t, &config.ChaosDashboardConfig{TargetNamespace: "target", ClusterScoped: scenario.cluster, EnableFilterNamespace: scenario.filter}, interceptor.Funcs{})
			response := apiRequest(t, router, http.MethodGet, "/common/chaos-available-namespaces", "", 200)
			var namespaces []string
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &namespaces))
			require.Equal(t, scenario.want, namespaces)
		})
	}
}

func TestPhysicalMachineErrorBodies(t *testing.T) {
	for _, endpoint := range []struct{ method, path, body string }{
		{http.MethodPost, "/common/physicalmachines", `{"namespaces":["target"]}`},
		{http.MethodGet, "/common/physicalmachine-labels?physicalMachineNamespaceList=target", ""},
		{http.MethodGet, "/common/physicalmachine-annotations?physicalMachineNamespaceList=target", ""},
	} {
		for _, failure := range []string{"client", "list"} {
			t.Run(endpoint.path+"/"+failure, func(t *testing.T) {
				router := commonBoundaryRouter(t, &config.ChaosDashboardConfig{TargetNamespace: "target"}, interceptor.Funcs{List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
					return errors.New("list unavailable")
				}})
				status := 500
				if failure == "client" {
					clientpool.K8sClients = apiClientPool{err: errors.New("invalid token")}
					status = 400
				}
				response := apiRequest(t, router, endpoint.method, endpoint.path, endpoint.body, status)
				assertAPIError(t, response, status)
			})
		}
	}
	for _, path := range []string{"/common/physicalmachine-labels", "/common/physicalmachine-annotations"} {
		t.Run(path+"/missing namespace", func(t *testing.T) {
			router := commonBoundaryRouter(t, &config.ChaosDashboardConfig{TargetNamespace: "target"}, interceptor.Funcs{})
			response := apiRequest(t, router, http.MethodGet, path, "", 500)
			assertAPIError(t, response, 500)
		})
	}
}

func TestSelectorRequestJSONBoundary(t *testing.T) {
	for _, path := range []string{"/common/pods", "/common/physicalmachines"} {
		for _, body := range []string{"null", "[]", `{} garbage`, `{} {}`} {
			t.Run(path+"/"+body, func(t *testing.T) {
				router := commonBoundaryRouter(t, &config.ChaosDashboardConfig{TargetNamespace: "target"}, interceptor.Funcs{})
				response := apiRequest(t, router, http.MethodPost, path, body, 400)
				assertAPIError(t, response, 400)
			})
		}
	}
}

func TestPodNamespaceFiltering(t *testing.T) {
	for _, scenario := range []struct {
		name, body, target string
		cluster, filter    bool
		want               []string
	}{
		{"implicit cluster filter", `{}`, "target", true, true, []string{"first", "second", "third"}},
		{"explicit cluster filter", `{"namespaces":["target","other"]}`, "target", true, true, []string{"first", "second", "third"}},
		{"explicit pods filter", `{"pods":{"target":["first"],"other":["outside"]}}`, "target", true, true, []string{"first"}},
		{"filter disabled", `{}`, "target", true, false, []string{"first", "second", "third", "outside"}},
		{"namespaced scope", `{}`, "target", false, false, []string{"first", "second", "third"}},
		{"disabled namespaced scope", `{}`, "other", false, true, []string{}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			router := commonBoundaryRouter(t, &config.ChaosDashboardConfig{TargetNamespace: scenario.target, ClusterScoped: scenario.cluster, EnableFilterNamespace: scenario.filter}, interceptor.Funcs{})
			response := apiRequest(t, router, http.MethodPost, "/common/pods", scenario.body, 200)
			var pods []struct{ Name string }
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &pods))
			names := make([]string, 0, len(pods))
			for _, pod := range pods {
				names = append(names, pod.Name)
			}
			require.ElementsMatch(t, scenario.want, names)
		})
	}
}
