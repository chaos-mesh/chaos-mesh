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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/experiment"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/schedule"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/workflow"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
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

// Match the store's resource identity in the fake client, which does not assign UIDs.
func setFixtureUID(object client.Object) {
	if object.GetUID() == "" && object.GetNamespace() == "target" && object.GetName() == "example" {
		object.SetUID("uid")
	}
}

func apiRouter(t *testing.T, objects []client.Object, faults interceptor.Funcs, storeErr error) (*gin.Engine, client.Client) {
	t.Helper()
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	for _, object := range objects {
		setFixtureUID(object)
	}
	create := faults.Create
	faults.Create = func(ctx context.Context, kube client.WithWatch, object client.Object, opts ...client.CreateOption) error {
		setFixtureUID(object)
		if create != nil {
			return create(ctx, kube, object, opts...)
		}
		return kube.Create(ctx, object, opts...)
	}
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
