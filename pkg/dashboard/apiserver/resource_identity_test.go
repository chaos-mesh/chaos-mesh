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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common/finalizers"
	apiservertypes "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/types"
)

func TestExperimentUIDRoutesRejectReplacementBeforeLookup(t *testing.T) {
	resource := apiResources()[0]
	require.NoError(t, json.Unmarshal([]byte(resource.body), resource.object))
	router, kube := apiRouter(t, []client.Object{resource.object}, interceptor.Funcs{}, nil)
	response := apiRequest(t, router, http.MethodGet, "/experiments/uid", "", http.StatusOK)
	var detail apiservertypes.ExperimentDetail
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &detail))
	require.Equal(t, "uid", detail.UID)
	require.Equal(t, "PodChaos", detail.Kind)
	require.Equal(t, "example", detail.KubeObject.Meta.Name)
	require.Equal(t, map[string]string{"keep": "value"}, detail.KubeObject.Meta.Annotations)

	replacement := resource.object.DeepCopyObject().(client.Object)
	require.NoError(t, kube.Delete(t.Context(), resource.object))
	replacement.SetUID("replacement")
	replacement.SetResourceVersion("")
	require.NoError(t, kube.Create(t.Context(), replacement))
	for _, request := range []struct{ method, path string }{
		{http.MethodGet, "/experiments/uid"},
		{http.MethodPut, "/experiments/pause/uid"},
		{http.MethodPut, "/experiments/start/uid"},
		{http.MethodDelete, "/experiments/uid"},
		{http.MethodDelete, "/experiments/uid?force=true"},
		{http.MethodDelete, "/experiments?uids=uid&force=true"},
	} {
		apiRequest(t, router, request.method, request.path, "", http.StatusNotFound)
	}
	var current v1alpha1.PodChaos
	require.NoError(t, kube.Get(t.Context(), client.ObjectKeyFromObject(resource.object), &current))
	require.Equal(t, types.UID("replacement"), current.UID)
	require.Equal(t, map[string]string{"keep": "value"}, current.Annotations)
	require.Nil(t, current.DeletionTimestamp)
}

func TestUIDScopedWritesRejectReplacementAfterLookup(t *testing.T) {
	for _, resource := range apiResources() {
		for _, method := range []string{http.MethodPut, http.MethodDelete} {
			t.Run(resource.endpoint+"/"+method, func(t *testing.T) {
				original := resource.object.DeepCopyObject().(client.Object)
				require.NoError(t, json.Unmarshal([]byte(resource.body), original))
				original.SetAnnotations(nil)
				key := client.ObjectKeyFromObject(original)
				path := "/" + resource.endpoint + "/uid"
				body := ""
				expectedStatus := http.StatusInternalServerError
				if method == http.MethodPut {
					if resource.endpoint == "workflows" {
						body = strings.Replace(resource.body, `"entry":"wait"`, `"entry":"other"`, 1)
					} else {
						path = "/" + resource.endpoint + "/pause/uid"
						expectedStatus = http.StatusNotFound
					}
				}

				var replacement client.Object
				replace := func(ctx context.Context, kube client.Client, observed client.Object) {
					t.Helper()
					require.Equal(t, types.UID("uid"), observed.GetUID())
					require.Nil(t, replacement, "mutation was retried against a different UID")
					require.NoError(t, kube.Delete(ctx, original))
					replacement = original.DeepCopyObject().(client.Object)
					replacement.SetUID("replacement")
					replacement.SetResourceVersion("")
					replacement.SetAnnotations(map[string]string{"owner": "replacement"})
					require.NoError(t, kube.Create(ctx, replacement))
					require.NotEqual(t, observed.GetResourceVersion(), replacement.GetResourceVersion())
				}
				router, kube := apiRouter(t, []client.Object{original}, interceptor.Funcs{
					Update: func(ctx context.Context, kube client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
						replace(ctx, kube, obj)
						return kube.Update(ctx, obj, opts...)
					},
					Patch: func(ctx context.Context, kube client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
						replace(ctx, kube, obj)
						return kube.Patch(ctx, obj, patch, opts...)
					},
					Delete: func(ctx context.Context, kube client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
						replace(ctx, kube, obj)
						// The fake client implements resource-version checks but skips delete UID preconditions.
						options := (&client.DeleteOptions{}).ApplyOptions(opts)
						require.NotNil(t, options.Preconditions)
						require.NotNil(t, options.Preconditions.UID)
						require.Equal(t, types.UID("uid"), *options.Preconditions.UID)
						return apierrors.NewConflict(schema.GroupResource{Group: v1alpha1.SchemeGroupVersion.Group, Resource: resource.endpoint}, key.Name, errors.New("UID precondition failed"))
					},
				}, nil)

				apiRequest(t, router, method, path, body, expectedStatus)
				require.NotNil(t, replacement, "request must reach the Kubernetes mutation")
				actual := replacement.DeepCopyObject().(client.Object)
				require.NoError(t, kube.Get(t.Context(), key, actual))
				require.Equal(t, replacement, actual)
			})
		}
	}
}

func TestPauseAndStartWithConcurrentUpdate(t *testing.T) {
	for _, resource := range apiResources()[:2] {
		for _, action := range []struct{ name, before, after string }{
			{"pause", "false", "true"},
			{"start", "true", "false"},
		} {
			t.Run(resource.endpoint+"/"+action.name, func(t *testing.T) {
				original := resource.object.DeepCopyObject().(client.Object)
				require.NoError(t, json.Unmarshal([]byte(resource.body), original))
				original.GetAnnotations()[v1alpha1.PauseAnnotationKey] = action.before
				key := client.ObjectKeyFromObject(original)

				firstPatch := true
				router, kube := apiRouter(t, []client.Object{original}, interceptor.Funcs{
					Patch: func(ctx context.Context, kube client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
						if firstPatch {
							firstPatch = false
							current := original.DeepCopyObject().(client.Object)
							require.NoError(t, kube.Get(ctx, key, current))
							current.GetAnnotations()["parallel"] = "preserved"
							require.NoError(t, kube.Update(ctx, current))
						}
						return kube.Patch(ctx, obj, patch, opts...)
					},
				}, nil)

				apiRequest(t, router, http.MethodPut, "/"+resource.endpoint+"/"+action.name+"/uid", "", http.StatusOK)
				require.NoError(t, kube.Get(t.Context(), key, original))
				require.Equal(t, map[string]string{
					"keep":                      "value",
					"parallel":                  "preserved",
					v1alpha1.PauseAnnotationKey: action.after,
				}, original.GetAnnotations())
			})
		}
	}
}

func TestForceDeleteExperimentRetriesOnlyAgainstOriginalUID(t *testing.T) {
	for _, mode := range []string{"concurrent update", "replacement"} {
		t.Run(mode, func(t *testing.T) {
			resource := apiResources()[0]
			require.NoError(t, json.Unmarshal([]byte(resource.body), resource.object))
			original := resource.object.(*v1alpha1.PodChaos)
			original.SetFinalizers([]string{finalizers.RecordFinalizer})
			key := client.ObjectKeyFromObject(original)

			updates, deletes := 0, 0
			router, kube := apiRouter(t, []client.Object{original}, interceptor.Funcs{
				Update: func(ctx context.Context, kube client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
					updates++
					require.Equal(t, types.UID("uid"), obj.GetUID())
					if updates == 1 {
						var concurrent v1alpha1.PodChaos
						require.NoError(t, kube.Get(ctx, key, &concurrent))
						if mode == "replacement" {
							concurrent.Finalizers = nil
							require.NoError(t, kube.Update(ctx, &concurrent))
							require.NoError(t, kube.Delete(ctx, &concurrent))
							concurrent.UID = "replacement"
							concurrent.ResourceVersion = ""
							require.NoError(t, kube.Create(ctx, &concurrent))
							return apierrors.NewConflict(schema.GroupResource{Group: v1alpha1.SchemeGroupVersion.Group, Resource: "podchaos"}, key.Name, errors.New("experiment was replaced"))
						}
						concurrent.Annotations["parallel"] = "preserved"
						require.NoError(t, kube.Update(ctx, &concurrent))
					}
					return kube.Update(ctx, obj, opts...)
				},
				Delete: func(ctx context.Context, kube client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
					deletes++
					return kube.Delete(ctx, obj, opts...)
				},
			}, nil)
			const path = "/experiments/uid?force=true"
			var actual v1alpha1.PodChaos
			if mode == "replacement" {
				apiRequest(t, router, http.MethodDelete, path, "", http.StatusNotFound)
				require.Equal(t, 1, updates)
				require.Zero(t, deletes)
				require.NoError(t, kube.Get(t.Context(), key, &actual))
				require.Equal(t, types.UID("replacement"), actual.UID)
				require.Equal(t, map[string]string{"keep": "value"}, actual.Annotations)
				return
			}

			apiRequest(t, router, http.MethodDelete, path, "", http.StatusOK)
			require.Equal(t, 2, updates)
			require.Equal(t, 1, deletes)
			require.NoError(t, kube.Get(t.Context(), key, &actual))
			require.NotNil(t, actual.DeletionTimestamp)
			require.Equal(t, map[string]string{
				"keep":                              "value",
				"parallel":                          "preserved",
				finalizers.AnnotationCleanFinalizer: finalizers.AnnotationCleanFinalizerForced,
			}, actual.Annotations)
		})
	}
}
