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

package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/joomcode/errorx"
	authorizationv1 "k8s.io/api/authorization/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	authclient "k8s.io/client-go/kubernetes/typed/authorization/v1"
	clienttesting "k8s.io/client-go/testing"

	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
)

func TestAPIErrorResponses(t *testing.T) {
	for _, tc := range []struct {
		kind   *errorx.Type
		status int
	}{
		{ErrBadRequest, 400}, {ErrNotFound, 404}, {ErrNoClusterPrivilege, 401},
		{ErrNoNamespacePrivilege, 401}, {ErrUnknown, 500}, {ErrInternalServer, 500},
		{ErrNS.NewType("unmapped_test_error"), 500},
	} {
		t.Run(tc.kind.FullName(), func(t *testing.T) {
			r := gin.New()
			continued := false
			r.Use(func(c *gin.Context) { SetAPIError(c, tc.kind.New("request failed")) })
			r.GET("/", func(c *gin.Context) { continued = true; c.Status(204) })
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
			var body APIError
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if w.Code != tc.status || body.Code != tc.status || body.Type != tc.kind.FullName() || !strings.Contains(body.Message, "request failed") || body.FullText == "" || continued {
				t.Fatalf("status %d, body %+v, continued %v", w.Code, body, continued)
			}
		})
	}
}

func TestKubernetesErrorResponses(t *testing.T) {
	resource := schema.GroupResource{Group: "chaos-mesh.org", Resource: "podchaos"}
	for _, tc := range []struct {
		name   string
		err    error
		status int
		kind   *errorx.Type
	}{
		{"missing", apierrors.NewNotFound(resource, "missing"), 404, ErrNotFound},
		{"cluster forbidden", apierrors.NewForbidden(resource, "test", errors.New("at the cluster scope")), 401, ErrNoClusterPrivilege},
		{"namespace forbidden", apierrors.NewForbidden(resource, "test", errors.New("in the namespace default")), 401, ErrNoNamespacePrivilege},
		{"other", errors.New("unavailable"), 500, ErrInternalServer},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			SetAPImachineryError(c, tc.err)
			var body APIError
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if w.Code != tc.status || body.Type != tc.kind.FullName() || !c.IsAborted() {
				t.Fatalf("status %d, body %+v", w.Code, body)
			}
		})
	}
}

func TestJSONBinding(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
	}{
		{"valid", `{"name":"test"}`, 200}, {"malformed", `{"name":`, 400},
		{"missing required", `{}`, 400}, {"wrong type", `{"name":123}`, 400}, {"empty", "", 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.POST("/", func(c *gin.Context) {
				var first, second struct {
					Name string `json:"name" binding:"required"`
				}
				if ShouldBindBodyWithJSON(c, &first) != nil {
					return
				}
				if ShouldBindBodyWithJSON(c, &second) != nil {
					return
				}
				if first.Name != "test" || second.Name != "test" {
					t.Errorf("bound values %q %q", first.Name, second.Name)
				}
				c.JSON(200, second)
			})
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.body)))
			if w.Code != tc.status {
				t.Fatalf("status %d, body %s", w.Code, w.Body)
			}
		})
	}
}

func TestJSONBodyIsOneCompleteDocument(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
	}{
		{"single object", "  {\"name\":\"example\"}\n", 204},
		{"trailing garbage", `{"name":"example"} invalid`, 400},
		{"second document", `{"name":"example"} {}`, 400},
		{"null object", `null`, 400},
		{"array instead of object", `[]`, 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			continued := false
			router.POST("/", func(c *gin.Context) {
				var payload struct {
					Name string `json:"name"`
				}
				if err := ShouldBindBodyWithJSON(c, &payload); err != nil {
					return
				}
				continued = true
				c.Status(204)
			})
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.body)))
			if response.Code != tc.status || continued != (tc.status == 204) {
				t.Fatalf("status %d, continued %v, body %s", response.Code, continued, response.Body)
			}
			if tc.status != 204 {
				var body APIError
				if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if body.Code != tc.status || body.Type != ErrBadRequest.FullName() {
					t.Fatalf("error body %+v", body)
				}
			}
		})
	}
}

type authClients struct {
	clientpool.Clients
	auth  authclient.AuthorizationV1Interface
	err   error
	token string
}

func (c *authClients) AuthClient(token string) (authclient.AuthorizationV1Interface, error) {
	c.token = token
	return c.auth, c.err
}

func TestAuthMiddleware(t *testing.T) {
	for _, tc := range []struct {
		name, method, query, target, namespace, verb string
		cluster, allowed                             bool
		clientErr, reviewErr                         error
		status                                       int
	}{
		{name: "cluster read", method: "GET", cluster: true, allowed: true, verb: "list", status: 204},
		{name: "namespace read", method: "GET", query: "?namespace=team", cluster: true, namespace: "team", allowed: true, verb: "list", status: 204},
		{name: "target default", method: "GET", target: "target", namespace: "target", allowed: true, verb: "list", status: 204},
		{name: "explicit namespace", method: "POST", query: "?namespace=team", target: "target", namespace: "team", allowed: true, verb: "patch", status: 204},
		{name: "write", method: "DELETE", cluster: true, allowed: true, verb: "patch", status: 204},
		{name: "denied cluster", method: "GET", cluster: true, verb: "list", status: 401},
		{name: "denied namespace", method: "GET", query: "?namespace=team", namespace: "team", verb: "list", status: 401},
		{name: "client error", method: "GET", clientErr: errors.New("no client"), status: 400},
		{name: "review error", method: "GET", reviewErr: errors.New("offline"), verb: "list", status: 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			kube := fake.NewClientset()
			reviews := 0
			kube.PrependReactor("create", "selfsubjectaccessreviews", func(action clienttesting.Action) (bool, runtime.Object, error) {
				reviews++
				attrs := action.(clienttesting.CreateAction).GetObject().(*authorizationv1.SelfSubjectAccessReview).Spec.ResourceAttributes
				if attrs.Namespace != tc.namespace || attrs.Verb != tc.verb || attrs.Group != "chaos-mesh.org" || attrs.Resource != "*" {
					t.Errorf("unexpected authorization request %+v", attrs)
				}
				return true, &authorizationv1.SelfSubjectAccessReview{Status: authorizationv1.SubjectAccessReviewStatus{Allowed: tc.allowed}}, tc.reviewErr
			})
			pool := &authClients{auth: kube.AuthorizationV1(), err: tc.clientErr}
			previous := clientpool.K8sClients
			clientpool.K8sClients = pool
			t.Cleanup(func() { clientpool.K8sClients = previous })
			r := gin.New()
			conf := &config.ChaosDashboardConfig{ClusterScoped: tc.cluster, TargetNamespace: tc.target}
			r.Use(func(c *gin.Context) { AuthMiddleware(c, conf) })
			continued := false
			r.Handle(tc.method, "/", func(c *gin.Context) { continued = true; c.Status(204) })
			req := httptest.NewRequest(tc.method, "/"+tc.query, nil)
			req.Header.Set("Authorization", "Bearer test-token")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.status || continued != (tc.status == 204) || pool.token != "test-token" {
				t.Fatalf("status %d, continued %v, token %q, body %s", w.Code, continued, pool.token, w.Body)
			}
			if tc.clientErr == nil && reviews != 1 {
				t.Fatalf("authorization reviews %d", reviews)
			}
			if tc.clientErr != nil && reviews != 0 {
				t.Fatalf("authorization attempted despite client error")
			}
		})
	}
}
