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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/common"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/experiment"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/schedule"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/template"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/workflow"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/uiserver"
)

type apiClientPool struct {
	clientpool.Clients
	kube client.Client
	err  error
}

func (p apiClientPool) Client(string) (client.Client, error) { return p.kube, p.err }

type apiExperimentStore struct {
	core.ExperimentStore
	err error
}

func (s apiExperimentStore) FindByUID(_ context.Context, uid string) (*core.Experiment, error) {
	if s.err != nil {
		return nil, s.err
	}
	if uid != "uid" {
		return nil, gorm.ErrRecordNotFound
	}
	return &core.Experiment{ExperimentMeta: core.ExperimentMeta{UID: uid, Namespace: "target", Name: "example", Kind: "PodChaos"}}, nil
}

type apiScheduleStore struct {
	core.ScheduleStore
	err error
}

func (s apiScheduleStore) FindByUID(_ context.Context, uid string) (*core.Schedule, error) {
	if s.err != nil {
		return nil, s.err
	}
	if uid != "uid" {
		return nil, gorm.ErrRecordNotFound
	}
	return &core.Schedule{ScheduleMeta: core.ScheduleMeta{UID: uid, Namespace: "target", Name: "example"}}, nil
}

type apiWorkflowStore struct {
	core.WorkflowStore
	err error
}

func (s apiWorkflowStore) FindByUID(_ context.Context, uid string) (*core.WorkflowEntity, error) {
	if s.err != nil {
		return nil, s.err
	}
	if uid != "uid" {
		return nil, gorm.ErrRecordNotFound
	}
	return &core.WorkflowEntity{WorkflowMeta: core.WorkflowMeta{UID: uid, Namespace: "target", Name: "example"}, Workflow: apiResources()[2].body}, nil
}

func apiRouter(t *testing.T, objects []client.Object, faults interceptor.Funcs, storeErr error) (*gin.Engine, client.Client) {
	t.Helper()
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).WithInterceptorFuncs(faults).Build()
	original := clientpool.K8sClients
	clientpool.K8sClients = apiClientPool{kube: kube}
	t.Cleanup(func() { clientpool.K8sClients = original })
	conf := &config.ChaosDashboardConfig{TargetNamespace: "target"}
	router := newEngine(conf)
	api := router.Group("/api")
	experiment.Register(api, experiment.NewService(apiExperimentStore{err: storeErr}, nil, conf, scheme, logr.Discard()))
	schedule.Register(api, schedule.NewService(apiScheduleStore{err: storeErr}, nil, conf, scheme, logr.Discard()))
	workflow.Register(api, workflow.NewService(conf, apiWorkflowStore{err: storeErr}, logr.Discard()))
	template.Register(api, template.NewService(conf, logr.Discard()))
	common.Register(api, common.NewService(conf, kube, logr.Discard()))
	return router, kube
}

func apiRequest(t *testing.T, router http.Handler, method, path, body string, status int) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, "/api"+path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	require.Equal(t, status, response.Code, response.Body.String())
	return response
}

type apiResource struct {
	endpoint, body string
	object         client.Object
}

func apiResources() []apiResource {
	return []apiResource{
		{"experiments", `{"apiVersion":"chaos-mesh.org/v1alpha1","kind":"PodChaos","metadata":{"name":"example","namespace":"target","annotations":{"keep":"value"}},"spec":{"action":"pod-kill","mode":"one","selector":{"namespaces":["target"]}}}`, &v1alpha1.PodChaos{}},
		{"schedules", `{"apiVersion":"chaos-mesh.org/v1alpha1","kind":"Schedule","metadata":{"name":"example","namespace":"target","annotations":{"keep":"value"}},"spec":{"schedule":"*/5 * * * *","type":"PodChaos","podChaos":{"action":"pod-kill","mode":"one","selector":{"namespaces":["target"]}}}}`, &v1alpha1.Schedule{}},
		{"workflows", `{"apiVersion":"chaos-mesh.org/v1alpha1","kind":"Workflow","metadata":{"name":"example","namespace":"target","annotations":{"keep":"value"}},"spec":{"entry":"wait","templates":[{"name":"wait","templateType":"Suspend","deadline":"1s"}]}}`, &v1alpha1.Workflow{}},
	}
}

func TestResourceCreateAndDelete(t *testing.T) {
	for _, resource := range apiResources() {
		t.Run(resource.endpoint, func(t *testing.T) {
			router, kube := apiRouter(t, nil, interceptor.Funcs{}, nil)
			response := apiRequest(t, router, http.MethodPost, "/"+resource.endpoint, resource.body, http.StatusOK)
			var result struct {
				Name     string `json:"name"`
				Metadata struct {
					Name string `json:"name"`
				} `json:"metadata"`
			}
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
			if resource.endpoint == "workflows" {
				require.Equal(t, "example", result.Name)
			} else {
				require.Equal(t, "example", result.Metadata.Name)
			}
			key := types.NamespacedName{Namespace: "target", Name: "example"}
			require.NoError(t, kube.Get(context.Background(), key, resource.object))
			require.Equal(t, "value", resource.object.GetAnnotations()["keep"])
			switch obj := resource.object.(type) {
			case *v1alpha1.PodChaos:
				require.Equal(t, v1alpha1.PodKillAction, obj.Spec.Action)
			case *v1alpha1.Schedule:
				require.Equal(t, "*/5 * * * *", obj.Spec.Schedule)
			case *v1alpha1.Workflow:
				require.Equal(t, "wait", obj.Spec.Entry)
			}
			apiRequest(t, router, http.MethodDelete, "/"+resource.endpoint+"/uid", "", http.StatusOK)
			require.True(t, apierrors.IsNotFound(kube.Get(context.Background(), key, resource.object)))
		})
	}
}

func TestResourceMalformedJSON(t *testing.T) {
	for _, resource := range apiResources() {
		t.Run(resource.endpoint, func(t *testing.T) {
			router, kube := apiRouter(t, nil, interceptor.Funcs{}, nil)
			apiRequest(t, router, http.MethodPost, "/"+resource.endpoint, `{"metadata":`, http.StatusBadRequest)
			require.True(t, apierrors.IsNotFound(kube.Get(context.Background(), types.NamespacedName{Namespace: "target", Name: "example"}, resource.object)))
		})
	}
}

func TestExperimentRequiresStringKind(t *testing.T) {
	for _, body := range []string{`{}`, `null`, `{"kind":null}`, `{"kind":4}`, `{"kind":{}}`, `{"kind":""}`, `{"kind":"Unknown"}`} {
		t.Run(body, func(t *testing.T) {
			router, kube := apiRouter(t, nil, interceptor.Funcs{}, nil)
			apiRequest(t, router, http.MethodPost, "/experiments", body, http.StatusBadRequest)
			var objects v1alpha1.PodChaosList
			require.NoError(t, kube.List(context.Background(), &objects))
			require.Empty(t, objects.Items)
		})
	}
}

func TestResourceCreateFailureDoesNotPersist(t *testing.T) {
	for _, resource := range apiResources() {
		t.Run(resource.endpoint, func(t *testing.T) {
			router, kube := apiRouter(t, nil, interceptor.Funcs{Create: func(context.Context, client.WithWatch, client.Object, ...client.CreateOption) error {
				return errors.New("kubernetes unavailable")
			}}, nil)
			apiRequest(t, router, http.MethodPost, "/"+resource.endpoint, resource.body, http.StatusInternalServerError)
			require.True(t, apierrors.IsNotFound(kube.Get(context.Background(), types.NamespacedName{Namespace: "target", Name: "example"}, resource.object)))
		})
	}
}

func TestPauseAndStartPreserveAnnotations(t *testing.T) {
	for _, resource := range apiResources()[:2] {
		t.Run(resource.endpoint, func(t *testing.T) {
			require.NoError(t, json.Unmarshal([]byte(resource.body), resource.object))
			router, kube := apiRouter(t, []client.Object{resource.object}, interceptor.Funcs{}, nil)
			for _, action := range []struct{ name, value string }{{"pause", "true"}, {"start", "false"}} {
				apiRequest(t, router, http.MethodPut, "/"+resource.endpoint+"/"+action.name+"/uid", "", http.StatusOK)
				require.NoError(t, kube.Get(context.Background(), types.NamespacedName{Namespace: "target", Name: "example"}, resource.object))
				require.Equal(t, map[string]string{"keep": "value", v1alpha1.PauseAnnotationKey: action.value}, resource.object.GetAnnotations())
			}
		})
	}
}

func TestResourceLookupErrors(t *testing.T) {
	for _, resource := range apiResources() {
		for _, scenario := range []struct {
			name, uid string
			storeErr  error
			want      int
		}{
			{"missing-store", "absent", nil, http.StatusNotFound},
			{"store-error", "uid", errors.New("database unavailable"), http.StatusInternalServerError},
			{"missing-kubernetes", "uid", nil, http.StatusNotFound},
		} {
			t.Run(resource.endpoint+"/"+scenario.name, func(t *testing.T) {
				router, _ := apiRouter(t, nil, interceptor.Funcs{}, scenario.storeErr)
				requests := []struct{ method, path string }{{http.MethodGet, "/" + resource.endpoint + "/" + scenario.uid}, {http.MethodDelete, "/" + resource.endpoint + "/" + scenario.uid}}
				if resource.endpoint != "workflows" {
					requests = append(requests, struct{ method, path string }{http.MethodPut, "/" + resource.endpoint + "/pause/" + scenario.uid})
				}
				for _, request := range requests {
					t.Run(request.method, func(t *testing.T) {
						want := scenario.want
						if resource.endpoint == "workflows" && scenario.name == "missing-kubernetes" && request.method == http.MethodGet {
							want = http.StatusOK
						}
						apiRequest(t, router, request.method, request.path, "", want)
					})
				}
				if resource.endpoint == "workflows" {
					t.Run(http.MethodPut, func(t *testing.T) {
						apiRequest(t, router, http.MethodPut, "/workflows/"+scenario.uid, resource.body, scenario.want)
					})
				}
			})
		}
	}
}

func TestWorkflowUpdate(t *testing.T) {
	resource := apiResources()[2]
	require.NoError(t, json.Unmarshal([]byte(resource.body), resource.object))
	router, kube := apiRouter(t, []client.Object{resource.object}, interceptor.Funcs{}, nil)
	key := types.NamespacedName{Namespace: "target", Name: "example"}
	for _, body := range []string{strings.Replace(resource.body, `"namespace":"target"`, `"namespace":"other"`, 1), strings.Replace(resource.body, `"name":"example"`, `"name":"other"`, 1)} {
		apiRequest(t, router, http.MethodPut, "/workflows/uid", body, http.StatusBadRequest)
		var current v1alpha1.Workflow
		require.NoError(t, kube.Get(context.Background(), key, &current))
		require.Equal(t, "1s", *current.Spec.Templates[0].Deadline)
	}
	apiRequest(t, router, http.MethodPut, "/workflows/uid", strings.Replace(resource.body, `"deadline":"1s"`, `"deadline":"2s"`, 1), http.StatusOK)
	var updated v1alpha1.Workflow
	require.NoError(t, kube.Get(context.Background(), key, &updated))
	require.Equal(t, "2s", *updated.Spec.Templates[0].Deadline)
	apiRequest(t, router, http.MethodPut, "/workflows/uid", `{"spec":`, http.StatusBadRequest)
}

const statusCheckBody = `{"name":"probe","namespace":"target","description":"health probe","spec":{"type":"HTTP","http":{"url":"http://service/health","criteria":{"statusCode":"200"}}}}`

func TestStatusCheckTemplateLifecycle(t *testing.T) {
	router, kube := apiRouter(t, nil, interceptor.Funcs{}, nil)
	apiRequest(t, router, http.MethodPost, "/templates/statuschecks", statusCheckBody, http.StatusOK)
	key := types.NamespacedName{Namespace: "target", Name: "template-status-check-probe"}
	var cm corev1.ConfigMap
	require.NoError(t, kube.Get(context.Background(), key, &cm))
	require.Equal(t, "chaos-mesh", cm.Labels["app.kubernetes.io/managed-by"])
	require.Equal(t, "StatusCheck", cm.Labels["template.chaos-mesh.org/type"])
	require.Equal(t, "probe", cm.Annotations["template.chaos-mesh.org/name"])
	require.Contains(t, cm.Data["spec"], "http://service/health")
	cm.Data["unrelated"] = "preserved"
	cm.Annotations["unrelated"] = "preserved"
	require.NoError(t, kube.Update(context.Background(), &cm))
	response := apiRequest(t, router, http.MethodGet, "/templates/statuschecks/statuscheck?name=probe", "", http.StatusOK)
	var detail struct {
		Name, Namespace, Description string
		Spec                         struct{ Mode string }
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &detail))
	require.Equal(t, "probe", detail.Name)
	require.Equal(t, "target", detail.Namespace)
	require.Equal(t, "health probe", detail.Description)
	require.Equal(t, "Synchronous", detail.Spec.Mode)
	updated := strings.Replace(statusCheckBody, "health probe", "updated probe", 1)
	updated = strings.Replace(updated, "http://service/health", "http://service/ready", 1)
	apiRequest(t, router, http.MethodPut, "/templates/statuschecks/statuscheck", updated, http.StatusOK)
	require.NoError(t, kube.Get(context.Background(), key, &cm))
	require.Equal(t, "updated probe", cm.Annotations["template.chaos-mesh.org/description"])
	require.Equal(t, "preserved", cm.Annotations["unrelated"])
	require.Equal(t, "preserved", cm.Data["unrelated"])
	require.Contains(t, cm.Data["spec"], "http://service/ready")
	apiRequest(t, router, http.MethodDelete, "/templates/statuschecks/statuscheck?name=probe", "", http.StatusOK)
	require.True(t, apierrors.IsNotFound(kube.Get(context.Background(), key, &cm)))
}

func TestStatusCheckTemplateErrors(t *testing.T) {
	router, kube := apiRouter(t, nil, interceptor.Funcs{}, nil)
	for _, method := range []string{http.MethodGet, http.MethodDelete} {
		apiRequest(t, router, method, "/templates/statuschecks/statuscheck", "", http.StatusBadRequest)
		apiRequest(t, router, method, "/templates/statuschecks/statuscheck?name=missing", "", http.StatusNotFound)
	}
	for _, request := range []struct{ method, path string }{{http.MethodPost, "/templates/statuschecks"}, {http.MethodPut, "/templates/statuschecks/statuscheck"}} {
		apiRequest(t, router, request.method, request.path, `{"spec":`, http.StatusBadRequest)
	}
	apiRequest(t, router, http.MethodPut, "/templates/statuschecks/statuscheck", statusCheckBody, http.StatusNotFound)
	var cms corev1.ConfigMapList
	require.NoError(t, kube.List(context.Background(), &cms))
	require.Empty(t, cms.Items)
}

func TestListRespectsNamespaceAndName(t *testing.T) {
	for _, resource := range apiResources()[:2] {
		t.Run(resource.endpoint, func(t *testing.T) {
			require.NoError(t, json.Unmarshal([]byte(resource.body), resource.object))
			otherNamespace := resource.object.DeepCopyObject().(client.Object)
			otherNamespace.SetNamespace("other")
			otherName := resource.object.DeepCopyObject().(client.Object)
			otherName.SetName("other")
			router, _ := apiRouter(t, []client.Object{resource.object, otherNamespace, otherName}, interceptor.Funcs{}, nil)
			response := apiRequest(t, router, http.MethodGet, "/"+resource.endpoint+"?name=example", "", http.StatusOK)
			var list []struct{ Name, Namespace string }
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &list))
			require.Equal(t, []struct{ Name, Namespace string }{{"example", "target"}}, list)
			response = apiRequest(t, router, http.MethodGet, "/"+resource.endpoint+"?namespace=other&name=example", "", http.StatusOK)
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &list))
			require.Equal(t, []struct{ Name, Namespace string }{{"example", "other"}}, list)
			response = apiRequest(t, router, http.MethodGet, "/"+resource.endpoint+"?name=absent", "", http.StatusOK)
			require.JSONEq(t, `[]`, response.Body.String())
		})
	}
}

func TestCommonPodSelectionAndMetadata(t *testing.T) {
	var selected corev1.Pod
	require.NoError(t, json.Unmarshal([]byte(`{"metadata":{"name":"selected","namespace":"target","labels":{"app":"web"},"annotations":{"owner":"team"}},"status":{"podIP":"10.0.0.1","phase":"Running"}}`), &selected))
	duplicate := selected.DeepCopy()
	duplicate.Name = "duplicate"
	excluded := selected.DeepCopy()
	excluded.Name = "excluded"
	excluded.Namespace = "other"
	excluded.Labels["app"] = "outside"
	router, _ := apiRouter(t, []client.Object{&selected, duplicate, excluded}, interceptor.Funcs{}, nil)
	response := apiRequest(t, router, http.MethodPost, "/common/pods", `{"namespaces":["target"],"labelSelectors":{"app":"web"}}`, http.StatusOK)
	var pods []struct{ Name, Namespace, IP, State string }
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &pods))
	require.ElementsMatch(t, []struct{ Name, Namespace, IP, State string }{{"selected", "target", "10.0.0.1", "Running"}, {"duplicate", "target", "10.0.0.1", "Running"}}, pods)
	response = apiRequest(t, router, http.MethodGet, "/common/labels?podNamespaceList=target", "", http.StatusOK)
	require.JSONEq(t, `{"app":["web"]}`, response.Body.String())
	response = apiRequest(t, router, http.MethodGet, "/common/annotations?podNamespaceList=target", "", http.StatusOK)
	require.JSONEq(t, `{"owner":["team"]}`, response.Body.String())
	apiRequest(t, router, http.MethodPost, "/common/pods", `{"namespaces":`, http.StatusBadRequest)
	response = apiRequest(t, router, http.MethodPost, "/common/pods", `{"namespaces":["target"],"labelSelectors":{"app":"absent"}}`, http.StatusOK)
	require.JSONEq(t, `[]`, response.Body.String())
}

func TestResourceMutationFailuresPreserveObject(t *testing.T) {
	for _, resource := range apiResources() {
		t.Run(resource.endpoint, func(t *testing.T) {
			require.NoError(t, json.Unmarshal([]byte(resource.body), resource.object))
			router, kube := apiRouter(t, []client.Object{resource.object}, interceptor.Funcs{
				Delete: func(context.Context, client.WithWatch, client.Object, ...client.DeleteOption) error {
					return errors.New("delete unavailable")
				},
				Patch: func(context.Context, client.WithWatch, client.Object, client.Patch, ...client.PatchOption) error {
					return errors.New("patch unavailable")
				},
				Update: func(context.Context, client.WithWatch, client.Object, ...client.UpdateOption) error {
					return errors.New("update unavailable")
				},
			}, nil)
			apiRequest(t, router, http.MethodDelete, "/"+resource.endpoint+"/uid", "", http.StatusInternalServerError)
			if resource.endpoint == "workflows" {
				apiRequest(t, router, http.MethodPut, "/workflows/uid", resource.body, http.StatusInternalServerError)
			} else {
				apiRequest(t, router, http.MethodPut, "/"+resource.endpoint+"/pause/uid", "", http.StatusInternalServerError)
			}
			require.NoError(t, kube.Get(context.Background(), types.NamespacedName{Namespace: "target", Name: "example"}, resource.object))
			require.Equal(t, map[string]string{"keep": "value"}, resource.object.GetAnnotations())
		})
	}
}

func TestWorkflowListNewestFirst(t *testing.T) {
	old := &v1alpha1.Workflow{ObjectMeta: metav1.ObjectMeta{Name: "a-old", Namespace: "target", CreationTimestamp: metav1.NewTime(time.Unix(100, 0))}}
	recent := &v1alpha1.Workflow{ObjectMeta: metav1.ObjectMeta{Name: "z-recent", Namespace: "target", CreationTimestamp: metav1.NewTime(time.Unix(200, 0))}}
	old.Status.StartTime = &old.CreationTimestamp
	recent.Status.StartTime = &recent.CreationTimestamp
	outside := recent.DeepCopy()
	outside.Namespace = "other"
	router, _ := apiRouter(t, []client.Object{old, recent, outside}, interceptor.Funcs{}, nil)
	response := apiRequest(t, router, http.MethodGet, "/workflows", "", http.StatusOK)
	var list []core.WorkflowMeta
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &list))
	require.Len(t, list, 2)
	require.Equal(t, "z-recent", list[0].Name)
	require.Equal(t, "a-old", list[1].Name)
	for _, item := range list {
		require.Equal(t, "target", item.Namespace)
	}
}

func TestWorkflowHTTPTaskRoundTrip(t *testing.T) {
	router, _ := apiRouter(t, nil, interceptor.Funcs{}, nil)
	body := `{"name":"health","method":"POST","url":"http://service/health","header":{"X-Test":["value"]},"body":"hello","followLocation":true,"jsonContent":false}`
	rendered := apiRequest(t, router, http.MethodPost, "/workflows/render-task/http", body, http.StatusOK)
	response := apiRequest(t, router, http.MethodPost, "/workflows/validate-task/http", rendered.Body.String(), http.StatusOK)
	require.JSONEq(t, `true`, response.Body.String())
	response = apiRequest(t, router, http.MethodPost, "/workflows/parse-task/http", rendered.Body.String(), http.StatusOK)
	require.JSONEq(t, body, response.Body.String())
	response = apiRequest(t, router, http.MethodPost, "/workflows/validate-task/http", `{"templateType":"Suspend","name":"wait"}`, http.StatusOK)
	require.JSONEq(t, `false`, response.Body.String())
	for _, action := range []string{"render", "parse", "validate"} {
		apiRequest(t, router, http.MethodPost, "/workflows/"+action+"-task/http", `{"name":`, http.StatusBadRequest)
	}
}

func TestServerRoutes(t *testing.T) {
	router := newEngine(&config.ChaosDashboardConfig{})
	newAPIRouter(router)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, http.StatusOK, response.Code)
	if uiserver.AssetsFS() == nil {
		require.Contains(t, response.Body.String(), "Dashboard UI is not built.")
	}
	response = apiRequest(t, router, http.MethodGet, "/swagger/doc.json", "", http.StatusOK)
	var swagger struct {
		Swagger string
		Paths   map[string]json.RawMessage
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &swagger))
	require.Equal(t, "2.0", swagger.Swagger)
	require.Contains(t, swagger.Paths, "/experiments")
}

func TestWorkflowDetailFallsBackToArchive(t *testing.T) {
	router, _ := apiRouter(t, nil, interceptor.Funcs{}, nil)
	response := apiRequest(t, router, http.MethodGet, "/workflows/uid", "", http.StatusOK)
	var detail core.WorkflowDetail
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &detail))
	require.Equal(t, "example", detail.Name)
	require.Equal(t, "target", detail.Namespace)
	require.Equal(t, "uid", detail.UID)
	require.Equal(t, "Workflow", detail.KubeObject.Kind)
	spec, err := json.Marshal(detail.KubeObject.Spec)
	require.NoError(t, err)
	require.JSONEq(t, `{"entry":"wait","templates":[{"name":"wait","templateType":"Suspend","deadline":"1s"}]}`, string(spec))
}

func TestBatchDeleteResources(t *testing.T) {
	for _, resource := range apiResources()[:2] {
		t.Run(resource.endpoint, func(t *testing.T) {
			require.NoError(t, json.Unmarshal([]byte(resource.body), resource.object))
			router, kube := apiRouter(t, []client.Object{resource.object}, interceptor.Funcs{}, nil)
			apiRequest(t, router, http.MethodDelete, "/"+resource.endpoint+"?uids=absent,uid", "", http.StatusNotFound)
			require.NoError(t, kube.Get(context.Background(), types.NamespacedName{Namespace: "target", Name: "example"}, resource.object))
			apiRequest(t, router, http.MethodDelete, "/"+resource.endpoint+"?uids=uid", "", http.StatusOK)
			require.True(t, apierrors.IsNotFound(kube.Get(context.Background(), types.NamespacedName{Namespace: "target", Name: "example"}, resource.object)))
		})
	}
}

func TestForceDeleteExperiment(t *testing.T) {
	for _, force := range []string{"false", "true"} {
		t.Run(force, func(t *testing.T) {
			resource := apiResources()[0]
			require.NoError(t, json.Unmarshal([]byte(resource.body), resource.object))
			resource.object.SetFinalizers([]string{"chaos-mesh/records"})
			router, kube := apiRouter(t, []client.Object{resource.object}, interceptor.Funcs{}, nil)
			apiRequest(t, router, http.MethodDelete, "/experiments/uid?force="+force, "", http.StatusOK)
			require.NoError(t, kube.Get(context.Background(), types.NamespacedName{Namespace: "target", Name: "example"}, resource.object))
			require.NotNil(t, resource.object.GetDeletionTimestamp())
			require.Equal(t, "value", resource.object.GetAnnotations()["keep"])
			if force == "true" {
				require.Equal(t, "forced", resource.object.GetAnnotations()["chaos-mesh.chaos-mesh.org/cleanFinalizer"])
			} else {
				require.NotContains(t, resource.object.GetAnnotations(), "chaos-mesh.chaos-mesh.org/cleanFinalizer")
			}
		})
	}
}

func TestTemplateListFiltersAndSorts(t *testing.T) {
	old := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "template-status-check-old", Namespace: "target", CreationTimestamp: metav1.NewTime(time.Unix(100, 0)), Labels: map[string]string{"app.kubernetes.io/managed-by": "chaos-mesh", "template.chaos-mesh.org/type": "StatusCheck"}, Annotations: map[string]string{"template.chaos-mesh.org/name": "old"}}}
	recent := old.DeepCopy()
	recent.Name = "template-status-check-recent"
	recent.Annotations["template.chaos-mesh.org/name"] = "recent"
	recent.CreationTimestamp = metav1.NewTime(time.Unix(200, 0))
	outside := recent.DeepCopy()
	outside.Namespace = "other"
	unmanaged := recent.DeepCopy()
	unmanaged.Name = "unmanaged"
	unmanaged.Labels = nil
	invalid := recent.DeepCopy()
	invalid.Name = "invalid"
	invalid.Annotations = nil
	router, _ := apiRouter(t, []client.Object{old, recent, outside, unmanaged, invalid}, interceptor.Funcs{}, nil)
	response := apiRequest(t, router, http.MethodGet, "/templates/statuschecks", "", http.StatusOK)
	var list []struct{ Name, Namespace string }
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &list))
	require.Equal(t, []struct{ Name, Namespace string }{{"recent", "target"}, {"old", "target"}}, list)
	response = apiRequest(t, router, http.MethodGet, "/templates/statuschecks?name=old", "", http.StatusOK)
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &list))
	require.Equal(t, []struct{ Name, Namespace string }{{"old", "target"}}, list)
	response = apiRequest(t, router, http.MethodGet, "/templates/statuschecks?name=absent", "", http.StatusOK)
	require.JSONEq(t, `[]`, response.Body.String())
}

func TestTemplateRejectsUnmanagedConfigMap(t *testing.T) {
	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "template-status-check-probe", Namespace: "target"}, Data: map[string]string{"keep": "value"}}
	router, kube := apiRouter(t, []client.Object{cm}, interceptor.Funcs{}, nil)
	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		apiRequest(t, router, method, "/templates/statuschecks/statuscheck?name=probe", statusCheckBody, http.StatusInternalServerError)
	}
	require.NoError(t, kube.Get(context.Background(), types.NamespacedName{Namespace: "target", Name: cm.Name}, cm))
	require.Equal(t, map[string]string{"keep": "value"}, cm.Data)
	require.Empty(t, cm.Annotations)
}

func TestCommonRBACScopeAndPermissions(t *testing.T) {
	for _, namespace := range []string{"", "target", "cluster"} {
		for _, role := range []string{"viewer", "manager"} {
			t.Run(namespace+"/"+role, func(t *testing.T) {
				router, _ := apiRouter(t, nil, interceptor.Funcs{}, nil)
				response := apiRequest(t, router, http.MethodGet, "/common/rbac-config?namespace="+namespace+"&role="+role, "", http.StatusOK)
				var configs map[string]string
				require.NoError(t, json.Unmarshal(response.Body.Bytes(), &configs))
				require.Len(t, configs, 1)
				for accountName, manifest := range configs {
					decoder := k8syaml.NewYAMLOrJSONDecoder(strings.NewReader(manifest), 4096)
					var objects []map[string]interface{}
					for {
						var object map[string]interface{}
						err := decoder.Decode(&object)
						if err == io.EOF {
							break
						}
						require.NoError(t, err)
						objects = append(objects, object)
					}
					require.Len(t, objects, 3)
					account, permission, binding := objects[0], objects[1], objects[2]
					require.Equal(t, "ServiceAccount", account["kind"])
					require.Equal(t, accountName, account["metadata"].(map[string]interface{})["name"])
					expectedNamespace := namespace
					expectedKind := "Role"
					if namespace == "" {
						expectedNamespace = "default"
						expectedKind = "ClusterRole"
					}
					require.Equal(t, expectedNamespace, account["metadata"].(map[string]interface{})["namespace"])
					require.Equal(t, expectedKind, permission["kind"])
					require.Equal(t, expectedKind+"Binding", binding["kind"])
					ref := binding["roleRef"].(map[string]interface{})
					require.Equal(t, expectedKind, ref["kind"])
					require.Equal(t, permission["metadata"].(map[string]interface{})["name"], ref["name"])
					subjects := binding["subjects"].([]interface{})
					require.Len(t, subjects, 1)
					require.Equal(t, accountName, subjects[0].(map[string]interface{})["name"])
					require.Equal(t, expectedNamespace, subjects[0].(map[string]interface{})["namespace"])
					rules := permission["rules"].([]interface{})
					require.Len(t, rules, 2)
					verbs := rules[1].(map[string]interface{})["verbs"]
					expectedVerbs := []interface{}{"get", "list", "watch"}
					if role == "manager" {
						expectedVerbs = append(expectedVerbs, "create", "delete", "patch", "update")
					}
					require.ElementsMatch(t, expectedVerbs, verbs)
				}
			})
		}
	}
	router, _ := apiRouter(t, nil, interceptor.Funcs{}, nil)
	apiRequest(t, router, http.MethodGet, "/common/rbac-config?role=invalid", "", http.StatusBadRequest)
}
