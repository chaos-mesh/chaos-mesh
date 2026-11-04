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
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	authorizationv1 "k8s.io/api/authorization/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
	typedauth "k8s.io/client-go/kubernetes/typed/authorization/v1"
	k8stesting "k8s.io/client-go/testing"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/archive"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

func assertAPIError(t *testing.T, response *httptest.ResponseRecorder, code int) {
	t.Helper()
	var body utils.APIError
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.Equal(t, code, body.Code)
	require.NotEmpty(t, body.Message)
	wantType := map[int]string{400: "error.api.bad_request", 404: "error.api.resource_not_found", 500: "error.api.internal_server_error"}[code]
	require.Equal(t, wantType, body.Type)
}

func TestClientPoolErrorsAreHTTPFailures(t *testing.T) {
	cases := []struct{ method, path, body string }{
		{http.MethodGet, "/workflows", ""}, {http.MethodGet, "/workflows/uid", ""}, {http.MethodDelete, "/workflows/uid", ""},
		{http.MethodPost, "/common/pods", `{}`}, {http.MethodGet, "/common/labels?podNamespaceList=target", ""}, {http.MethodGet, "/common/annotations?podNamespaceList=target", ""},
	}
	for _, scenario := range cases {
		t.Run(scenario.method+scenario.path, func(t *testing.T) {
			router, _ := apiRouter(t, nil, interceptor.Funcs{}, nil)
			clientpool.K8sClients = apiClientPool{err: errors.New("invalid token")}
			response := apiRequest(t, router, scenario.method, scenario.path, scenario.body, http.StatusBadRequest)
			assertAPIError(t, response, http.StatusBadRequest)
		})
	}
}

func TestCommonErrorResponseBodies(t *testing.T) {
	for _, scenario := range []struct {
		method, path, body string
		code               int
	}{
		{http.MethodGet, "/common/rbac-config?role=invalid", "", 400},
		{http.MethodGet, "/common/labels", "", 500},
		{http.MethodGet, "/common/annotations", "", 500},
		{http.MethodPost, "/common/pods", `{"namespaces":`, 400},
	} {
		t.Run(scenario.path, func(t *testing.T) {
			router, _ := apiRouter(t, nil, interceptor.Funcs{}, nil)
			response := apiRequest(t, router, scenario.method, scenario.path, scenario.body, scenario.code)
			assertAPIError(t, response, scenario.code)
		})
	}
	for _, path := range []string{"/common/namespaces", "/common/pods", "/common/labels?podNamespaceList=target", "/common/annotations?podNamespaceList=target"} {
		t.Run(path+"/list-error", func(t *testing.T) {
			router, _ := apiRouter(t, nil, interceptor.Funcs{List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
				return errors.New("list unavailable")
			}}, nil)
			method := http.MethodGet
			body := ""
			if path == "/common/pods" {
				method = http.MethodPost
				body = `{"namespaces":["target"]}`
			}
			response := apiRequest(t, router, method, path, body, 500)
			assertAPIError(t, response, 500)
		})
	}
}

type apiArchiveExperiments struct {
	apiExperimentStore
	deleteErr error
	deleted   []string
}

func (s *apiArchiveExperiments) Delete(_ context.Context, exp *core.Experiment) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}
	s.deleted = append(s.deleted, exp.UID)
	return nil
}
func (s *apiArchiveExperiments) DeleteByUIDs(_ context.Context, uids []string) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}
	s.deleted = append(s.deleted, uids...)
	return nil
}

type apiArchiveEvents struct {
	core.EventStore
	deleteErr error
	deleted   []string
}

func (s *apiArchiveEvents) DeleteByUID(_ context.Context, uid string) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}
	s.deleted = append(s.deleted, uid)
	return nil
}
func (s *apiArchiveEvents) DeleteByUIDList(_ context.Context, uids []string) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}
	s.deleted = append(s.deleted, uids...)
	return nil
}

type apiArchivePool struct {
	apiClientPool
	auth typedauth.AuthorizationV1Interface
}

func (p apiArchivePool) AuthClient(string) (typedauth.AuthorizationV1Interface, error) {
	return p.auth, nil
}

func TestArchiveDeleteEffectsAndErrors(t *testing.T) {
	for _, batch := range []bool{false, true} {
		for _, failure := range []string{"", "archive", "events", "missing"} {
			t.Run(failure+map[bool]string{true: "/batch", false: "/single"}[batch], func(t *testing.T) {
				router, kube := apiRouter(t, nil, interceptor.Funcs{}, nil)
				auth := kubefake.NewClientset()
				auth.PrependReactor("create", "selfsubjectaccessreviews", func(k8stesting.Action) (bool, runtime.Object, error) {
					return true, &authorizationv1.SelfSubjectAccessReview{Status: authorizationv1.SubjectAccessReviewStatus{Allowed: true}}, nil
				})
				clientpool.K8sClients = apiArchivePool{kube: kube, auth: auth.AuthorizationV1()}
				experiments := &apiArchiveExperiments{}
				events := &apiArchiveEvents{}
				if failure == "archive" {
					experiments.deleteErr = errors.New("archive unavailable")
				}
				if failure == "events" {
					events.deleteErr = errors.New("events unavailable")
				}
				archive.Register(router.Group("/api"), archive.NewService(experiments, nil, events, nil, &config.ChaosDashboardConfig{TargetNamespace: "target"}))
				path := "/archives/uid"
				want := 200
				if batch {
					path = "/archives?uids=uid,second"
				}
				if failure == "archive" || failure == "events" {
					want = 500
				}
				if failure == "missing" {
					if batch {
						path = "/archives"
						want = 400
					} else {
						path = "/archives/absent"
						want = 404
					}
				}
				response := apiRequest(t, router, http.MethodDelete, path, "", want)
				if want != 200 {
					assertAPIError(t, response, want)
				} else {
					require.JSONEq(t, `{"status":"success"}`, response.Body.String())
				}
				if failure == "" {
					require.Equal(t, []string{"uid"}, experiments.deleted)
				} else {
					require.Empty(t, experiments.deleted)
				}
				if failure == "" || failure == "archive" {
					require.Equal(t, []string{"uid"}, events.deleted)
				} else {
					require.Empty(t, events.deleted)
				}
				if failure == "archive" || failure == "events" {
					experiments.deleteErr = nil
					events.deleteErr = nil
					events.deleted = nil
					apiRequest(t, router, http.MethodDelete, path, "", 200)
					require.Equal(t, []string{"uid"}, experiments.deleted)
					require.Equal(t, []string{"uid"}, events.deleted)
				}
			})
		}
	}
}

type apiArchivedRecord struct {
	core.ExperimentStore
	record *core.Experiment
}

func (s apiArchivedRecord) FindByUID(context.Context, string) (*core.Experiment, error) {
	return s.record, nil
}

func TestArchiveRejectsUnreadableSnapshot(t *testing.T) {
	for _, record := range []*core.Experiment{
		{UID: "uid", Kind: "PodChaos", Experiment: `{"spec":`},
		{UID: "uid", Kind: "Unknown", Experiment: `{}`},
	} {
		t.Run(record.Kind, func(t *testing.T) {
			router, kube := apiRouter(t, nil, interceptor.Funcs{}, nil)
			auth := kubefake.NewClientset()
			auth.PrependReactor("create", "selfsubjectaccessreviews", func(k8stesting.Action) (bool, runtime.Object, error) {
				return true, &authorizationv1.SelfSubjectAccessReview{Status: authorizationv1.SubjectAccessReviewStatus{Allowed: true}}, nil
			})
			clientpool.K8sClients = apiArchivePool{kube: kube, auth: auth.AuthorizationV1()}
			archive.Register(router.Group("/api"), archive.NewService(apiArchivedRecord{record: record}, nil, nil, nil, &config.ChaosDashboardConfig{TargetNamespace: "target"}))
			response := apiRequest(t, router, http.MethodGet, "/archives/uid", "", 500)
			assertAPIError(t, response, 500)
		})
	}
}
