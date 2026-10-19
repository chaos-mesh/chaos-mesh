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

	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/utils"
)

func TestResourceJSONBoundaryNoMutation(t *testing.T) {
	for _, resource := range apiResources() {
		for _, body := range []struct{ name, value string }{
			{"null", "null"}, {"array", "[]"},
			{"trailing garbage", resource.body + " garbage"},
			{"second object", resource.body + " {}"},
		} {
			t.Run(resource.endpoint+"/"+body.name, func(t *testing.T) {
				router, kube := apiRouter(t, nil, interceptor.Funcs{}, nil)
				response := apiRequest(t, router, http.MethodPost, "/"+resource.endpoint, body.value, 400)
				assertAPIError(t, response, 400)
				err := kube.Get(context.Background(), types.NamespacedName{Namespace: "target", Name: "example"}, resource.object)
				require.True(t, apierrors.IsNotFound(err), "invalid JSON must not create a resource")
			})
		}
	}
	for _, body := range []string{"null", "[]", apiResources()[2].body + " garbage", apiResources()[2].body + " {}"} {
		t.Run("workflow update/"+body, func(t *testing.T) {
			resource := apiResources()[2]
			require.NoError(t, json.Unmarshal([]byte(resource.body), resource.object))
			router, kube := apiRouter(t, []client.Object{resource.object}, interceptor.Funcs{}, nil)
			before := resource.object.DeepCopyObject().(client.Object)
			require.NoError(t, kube.Get(context.Background(), client.ObjectKeyFromObject(before), before))
			response := apiRequest(t, router, http.MethodPut, "/workflows/uid", body, 400)
			assertAPIError(t, response, 400)
			after := resource.object.DeepCopyObject().(client.Object)
			require.NoError(t, kube.Get(context.Background(), client.ObjectKeyFromObject(before), after))
			require.Equal(t, before, after)
		})
	}
}

func TestResourceDeleteFailurePreservesObject(t *testing.T) {
	failures := []struct {
		name      string
		err       error
		code      int
		errorType string
	}{
		{"not found", apierrors.NewNotFound(schema.GroupResource{Group: "chaos-mesh.org", Resource: "experiments"}, "example"), 404, "error.api.resource_not_found"},
		{"namespace forbidden", apierrors.NewForbidden(schema.GroupResource{Group: "chaos-mesh.org", Resource: "experiments"}, "example", errors.New("denied in the namespace target")), 401, "error.api.no_namespace_privilege"},
		{"conflict", apierrors.NewConflict(schema.GroupResource{Group: "chaos-mesh.org", Resource: "experiments"}, "example", errors.New("resource version changed")), 500, "error.api.internal_server_error"},
		{"transient", errors.New("kubernetes unavailable"), 500, "error.api.internal_server_error"},
	}
	for _, resource := range apiResources() {
		for _, failure := range failures {
			t.Run(resource.endpoint+"/"+failure.name, func(t *testing.T) {
				require.NoError(t, json.Unmarshal([]byte(resource.body), resource.object))
				fail := true
				router, kube := apiRouter(t, []client.Object{resource.object}, interceptor.Funcs{Delete: func(ctx context.Context, kube client.WithWatch, object client.Object, opts ...client.DeleteOption) error {
					if fail {
						return failure.err
					}
					return kube.Delete(ctx, object, opts...)
				}}, nil)
				before := resource.object.DeepCopyObject().(client.Object)
				key := client.ObjectKeyFromObject(before)
				require.NoError(t, kube.Get(context.Background(), key, before))
				response := apiRequest(t, router, http.MethodDelete, "/"+resource.endpoint+"/uid", "", failure.code)
				var body utils.APIError
				require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
				require.Equal(t, failure.code, body.Code)
				require.Equal(t, failure.errorType, body.Type)
				require.NotEmpty(t, body.Message)
				after := resource.object.DeepCopyObject().(client.Object)
				require.NoError(t, kube.Get(context.Background(), key, after))
				require.Equal(t, before, after)
				fail = false
				apiRequest(t, router, http.MethodDelete, "/"+resource.endpoint+"/uid", "", 200)
				require.True(t, apierrors.IsNotFound(kube.Get(context.Background(), key, after)))
			})
		}
	}
}

func TestHTTPTaskJSONBoundary(t *testing.T) {
	for _, path := range []string{"/workflows/render-task/http", "/workflows/parse-task/http", "/workflows/validate-task/http"} {
		for _, body := range []string{"null", "[]", `{} garbage`, `{} {}`} {
			t.Run(path+"/"+body, func(t *testing.T) {
				router, _ := apiRouter(t, nil, interceptor.Funcs{}, nil)
				response := apiRequest(t, router, http.MethodPost, path, body, 400)
				assertAPIError(t, response, 400)
			})
		}
	}
}
