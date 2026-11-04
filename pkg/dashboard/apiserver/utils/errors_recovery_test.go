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
	"testing"

	"github.com/gin-gonic/gin"
	authorizationv1 "k8s.io/api/authorization/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"

	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
)

func TestAuthMiddlewareReviewErrorRecovery(t *testing.T) {
	resource := schema.GroupResource{Group: "chaos-mesh.org", Resource: "podchaos"}
	for _, tc := range []struct {
		name   string
		err    error
		status int
		kind   string
	}{
		{"cluster forbidden", apierrors.NewForbidden(resource, "test", errors.New("at the cluster scope")), 401, ErrNoClusterPrivilege.FullName()},
		{"namespace forbidden", apierrors.NewForbidden(resource, "test", errors.New("in the namespace team")), 401, ErrNoNamespacePrivilege.FullName()},
		{"not found", apierrors.NewNotFound(resource, "test"), 404, ErrNotFound.FullName()},
		{"transport failure", errors.New("connection reset"), 500, ErrInternalServer.FullName()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			kube := fake.NewClientset()
			reviewErr := tc.err
			reviews := 0
			kube.PrependReactor("create", "selfsubjectaccessreviews", func(clienttesting.Action) (bool, runtime.Object, error) {
				reviews++
				if reviewErr != nil {
					return true, nil, reviewErr
				}
				return true, &authorizationv1.SelfSubjectAccessReview{Status: authorizationv1.SubjectAccessReviewStatus{Allowed: true}}, nil
			})
			previous := clientpool.K8sClients
			clientpool.K8sClients = &authClients{auth: kube.AuthorizationV1()}
			t.Cleanup(func() { clientpool.K8sClients = previous })
			router := gin.New()
			router.Use(func(c *gin.Context) { AuthMiddleware(c, &config.ChaosDashboardConfig{ClusterScoped: true}) })
			downstream := 0
			router.GET("/", func(c *gin.Context) { downstream++; c.Status(http.StatusNoContent) })
			request := func() *httptest.ResponseRecorder {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Authorization", "Bearer test-token")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				return w
			}
			failed := request()
			var body APIError
			if err := json.Unmarshal(failed.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if failed.Code != tc.status || body.Code != tc.status || body.Type != tc.kind || downstream != 0 || len(failed.Result().Cookies()) != 0 {
				t.Errorf("failed review status %d, body %+v, downstream %d", failed.Code, body, downstream)
			}
			reviewErr = nil
			recovered := request()
			if recovered.Code != http.StatusNoContent || downstream != 1 || reviews != 2 {
				t.Fatalf("recovery status %d, downstream %d, reviews %d", recovered.Code, downstream, reviews)
			}
		})
	}
}
