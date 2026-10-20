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
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/event"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

type requestContextKey struct{}

func stoppedRequestContext(t *testing.T, deadline bool) context.Context {
	t.Helper()
	ctx := context.WithValue(context.Background(), requestContextKey{}, "request")
	if deadline {
		stopped, cancel := context.WithDeadline(ctx, time.Unix(1, 0))
		t.Cleanup(cancel)
		require.ErrorIs(t, stopped.Err(), context.DeadlineExceeded)
		return stopped
	}
	stopped, cancel := context.WithCancel(ctx)
	cancel()
	return stopped
}

func liveRequestContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.WithValue(context.Background(), requestContextKey{}, "request"), time.Hour)
	t.Cleanup(cancel)
	return ctx
}

func requestWithContext(ctx context.Context, router http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/api"+path, strings.NewReader(body)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	return response
}

type contextKubeProbe struct {
	mu        sync.Mutex
	contexts  []context.Context
	operation string
	enabled   bool
}

func (p *contextKubeProbe) check(ctx context.Context, operation string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.enabled || operation != p.operation {
		return nil
	}
	p.contexts = append(p.contexts, ctx)
	return ctx.Err()
}

func (p *contextKubeProbe) interceptors() interceptor.Funcs {
	return interceptor.Funcs{
		Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			if err := p.check(ctx, "get"); err != nil {
				return err
			}
			return c.Get(ctx, key, obj, opts...)
		},
		List: func(ctx context.Context, c client.WithWatch, obj client.ObjectList, opts ...client.ListOption) error {
			if err := p.check(ctx, "list"); err != nil {
				return err
			}
			return c.List(ctx, obj, opts...)
		},
		Create: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
			if err := p.check(ctx, "create"); err != nil {
				return err
			}
			return c.Create(ctx, obj, opts...)
		},
		Update: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
			if err := p.check(ctx, "update"); err != nil {
				return err
			}
			return c.Update(ctx, obj, opts...)
		},
		Patch: func(ctx context.Context, c client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
			if err := p.check(ctx, "patch"); err != nil {
				return err
			}
			return c.Patch(ctx, obj, patch, opts...)
		},
		Delete: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
			if err := p.check(ctx, "delete"); err != nil {
				return err
			}
			return c.Delete(ctx, obj, opts...)
		},
	}
}

func (p *contextKubeProbe) assertForwarded(ctx context.Context, t *testing.T) {
	t.Helper()
	p.mu.Lock()
	defer p.mu.Unlock()
	if ctx.Err() == nil {
		require.NotEmpty(t, p.contexts)
	}
	for _, received := range p.contexts {
		require.Equal(t, "request", received.Value(requestContextKey{}))
		if ctx.Err() != nil {
			require.ErrorIs(t, received.Err(), ctx.Err())
		}
		deadline, ok := received.Deadline()
		wantDeadline, wantOK := ctx.Deadline()
		require.Equal(t, wantOK, ok)
		require.Equal(t, wantDeadline, deadline)
	}
}

func TestRequestContextKubernetesMutationRecovery(t *testing.T) {
	for _, deadline := range []bool{false, true} {
		for _, resource := range append(apiResources()[:2], apiResource{endpoint: "templates/statuschecks", body: statusCheckBody, object: &corev1.ConfigMap{}}) {
			for _, operation := range []string{"create", "delete", "update"} {
				t.Run(resource.endpoint+"/"+operation+map[bool]string{false: "/canceled", true: "/deadline"}[deadline], func(t *testing.T) {
					probe := &contextKubeProbe{operation: operation}
					router, kube := apiRouter(t, nil, probe.interceptors(), nil)
					isTemplate := resource.endpoint == "templates/statuschecks"
					method, path, body := http.MethodPost, "/"+resource.endpoint, resource.body
					key := types.NamespacedName{Namespace: "target", Name: "example"}
					if isTemplate {
						key.Name = v1alpha1.GenerateTemplateName("probe")
					}
					var before client.Object
					if operation != "create" {
						apiRequest(t, router, http.MethodPost, path, body, 200)
						require.NoError(t, kube.Get(context.Background(), key, resource.object))
						before = resource.object.DeepCopyObject().(client.Object)
						if operation == "delete" {
							method, body = http.MethodDelete, ""
							path += "/uid"
							if isTemplate {
								path = "/templates/statuschecks/statuscheck?name=probe"
							}
						} else {
							method = http.MethodPut
							if isTemplate {
								path = "/templates/statuschecks/statuscheck"
								body = strings.Replace(body, "health probe", "updated probe", 1)
							} else {
								path = "/" + resource.endpoint + "/pause/uid"
								probe.operation = "patch"
							}
						}
					}
					probe.enabled = true
					ctx := stoppedRequestContext(t, deadline)
					response := requestWithContext(ctx, router, method, path, body)
					probe.enabled = false
					after := resource.object.DeepCopyObject().(client.Object)
					err := kube.Get(context.Background(), key, after)
					if operation == "create" {
						require.True(t, apierrors.IsNotFound(err), "canceled create changed state: %v", err)
					} else {
						require.NoError(t, err)
						require.Equal(t, before, after, "canceled mutation changed stored object")
					}
					require.Equal(t, 500, response.Code, response.Body.String())
					assertAPIError(t, response, 500)
					probe.assertForwarded(ctx, t)
					probe.contexts = nil
					probe.enabled = true
					liveCtx := liveRequestContext(t)
					retryResponse := requestWithContext(liveCtx, router, method, path, body)
					require.Equal(t, 200, retryResponse.Code, retryResponse.Body.String())
					probe.assertForwarded(liveCtx, t)
					probe.enabled = false
					err = kube.Get(context.Background(), key, after)
					if operation == "delete" {
						require.True(t, apierrors.IsNotFound(err))
						return
					}
					require.NoError(t, err)
					if operation == "update" {
						if isTemplate {
							require.Equal(t, "updated probe", after.GetAnnotations()[v1alpha1.TemplateDescriptionAnnotationKey])
						} else {
							require.Equal(t, "true", after.GetAnnotations()[v1alpha1.PauseAnnotationKey])
						}
					}
				})
			}
		}
	}
}

func TestRequestContextArchiveCleanupRecovery(t *testing.T) {
	for _, deadline := range []bool{false, true} {
		for _, kind := range archiveKinds() {
			for _, batch := range []bool{false, true} {
				t.Run(kind.path+map[bool]string{false: "/single", true: "/batch"}[batch]+map[bool]string{false: "/canceled", true: "/deadline"}[deadline], func(t *testing.T) {
					fixture := newArchiveFixture(t, &config.ChaosDashboardConfig{TargetNamespace: "target"})
					require.NoError(t, fixture.db.Create(kind.record("uid", "target", `{}`)).Error)
					require.NoError(t, fixture.events.Create(context.Background(), &core.Event{ObjectID: "uid"}))
					path := kind.path + "/uid"
					if batch {
						path = kind.path + "?uids=uid"
					}
					response := requestWithContext(stoppedRequestContext(t, deadline), fixture.router, http.MethodDelete, path, "")
					var count int64
					require.NoError(t, fixture.db.Model(kind.record("", "", "")).Count(&count).Error)
					require.EqualValues(t, 1, count, "canceled delete removed archive")
					remaining, err := fixture.events.ListByUID(context.Background(), "uid")
					require.NoError(t, err)
					require.Len(t, remaining, 1, "canceled delete removed event")
					require.Equal(t, 500, response.Code, response.Body.String())
					assertAPIError(t, response, 500)
					apiRequest(t, fixture.router, http.MethodDelete, path, "", 200)
					require.NoError(t, fixture.db.Model(kind.record("", "", "")).Count(&count).Error)
					require.Zero(t, count)
					remaining, err = fixture.events.ListByUID(context.Background(), "uid")
					require.NoError(t, err)
					require.Empty(t, remaining)
				})
			}
		}
	}
}

func TestRequestContextKubernetesReadRecovery(t *testing.T) {
	for _, scenario := range []struct{ method, path, body, operation string }{
		{"GET", "/experiments", "", "list"}, {"GET", "/experiments/state", "", "list"},
		{"GET", "/schedules", "", "list"}, {"GET", "/experiments/uid", "", "get"},
		{"GET", "/common/namespaces", "", "list"},
		{"POST", "/common/pods", `{"namespaces":["target"]}`, "list"},
		{"POST", "/common/physicalmachines", `{"namespaces":["target"]}`, "list"},
		{"GET", "/templates/statuschecks", "", "list"},
	} {
		for _, deadline := range []bool{false, true} {
			t.Run(scenario.path+map[bool]string{false: "/canceled", true: "/deadline"}[deadline], func(t *testing.T) {
				probe := &contextKubeProbe{operation: scenario.operation}
				router, _ := apiRouter(t, nil, probe.interceptors(), nil)
				apiRequest(t, router, "POST", "/experiments", apiResources()[0].body, 200)
				probe.enabled = true
				ctx := stoppedRequestContext(t, deadline)
				response := requestWithContext(ctx, router, scenario.method, scenario.path, scenario.body)
				require.Equal(t, 500, response.Code, response.Body.String())
				assertAPIError(t, response, 500)
				probe.assertForwarded(ctx, t)
				probe.contexts = nil
				liveCtx := liveRequestContext(t)
				response = requestWithContext(liveCtx, router, scenario.method, scenario.path, scenario.body)
				require.Equal(t, 200, response.Code, response.Body.String())
				probe.assertForwarded(liveCtx, t)
			})
		}
	}
}

func TestRequestContextDatabaseReadRecovery(t *testing.T) {
	for _, path := range []string{"/archives", "/archives/uid", "/archives/schedules", "/archives/schedules/uid", "/archives/workflows", "/archives/workflows/uid", "/events", "/events/1"} {
		for _, deadline := range []bool{false, true} {
			t.Run(path+map[bool]string{false: "/canceled", true: "/deadline"}[deadline], func(t *testing.T) {
				conf := &config.ChaosDashboardConfig{TargetNamespace: "target"}
				fixture := newArchiveFixture(t, conf)
				for _, kind := range archiveKinds() {
					require.NoError(t, fixture.db.Create(kind.record("uid", "target", `{}`)).Error)
				}
				require.NoError(t, fixture.events.Create(context.Background(), &core.Event{ObjectID: "uid", Namespace: "target"}))
				event.Register(fixture.router.Group("/api"), event.NewService(fixture.events, nil, conf, logr.Discard()))
				ctx := stoppedRequestContext(t, deadline)
				var received []context.Context
				require.NoError(t, fixture.db.Callback().Query().Before("gorm:query").Register("capture_request", func(db *gorm.DB) { received = append(received, db.Statement.Context) }))
				response := requestWithContext(ctx, fixture.router, http.MethodGet, path, "")
				require.Equal(t, 500, response.Code, response.Body.String())
				assertAPIError(t, response, 500)
				for _, got := range received {
					require.Equal(t, "request", got.Value(requestContextKey{}))
					require.ErrorIs(t, got.Err(), ctx.Err())
				}
				received = nil
				liveCtx := liveRequestContext(t)
				response = requestWithContext(liveCtx, fixture.router, http.MethodGet, path, "")
				require.Equal(t, 200, response.Code, response.Body.String())
				require.NotEmpty(t, received)
				for _, got := range received {
					require.Equal(t, "request", got.Value(requestContextKey{}))
					require.NoError(t, got.Err())
					deadline, ok := got.Deadline()
					wantDeadline, wantOK := liveCtx.Deadline()
					require.Equal(t, wantOK, ok)
					require.Equal(t, wantDeadline, deadline)
				}
				require.NoError(t, fixture.db.Callback().Query().Remove("capture_request"))
				var value any
				require.NoError(t, json.Unmarshal(response.Body.Bytes(), &value))
				require.NotEmpty(t, value)
			})
		}
	}
}

func TestRequestContextArchiveEventCleanupRecovery(t *testing.T) {
	for _, kind := range archiveKinds() {
		for _, batch := range []bool{false, true} {
			t.Run(kind.path+map[bool]string{false: "/single", true: "/batch"}[batch], func(t *testing.T) {
				fixture := newArchiveFixture(t, &config.ChaosDashboardConfig{TargetNamespace: "target"})
				require.NoError(t, fixture.db.Create(kind.record("uid", "target", `{}`)).Error)
				require.NoError(t, fixture.events.Create(context.Background(), &core.Event{ObjectID: "uid"}))
				ctx, cancel := context.WithCancel(context.WithValue(context.Background(), requestContextKey{}, "request"))
				defer cancel()
				require.NoError(t, fixture.db.Callback().Delete().After("gorm:commit_or_rollback_transaction").Register("cancel_after_event_delete", func(db *gorm.DB) {
					if db.Statement.Schema.Name == "Event" {
						cancel()
					}
				}))
				path := kind.path + "/uid"
				if batch {
					path = kind.path + "?uids=uid"
				}
				response := requestWithContext(ctx, fixture.router, http.MethodDelete, path, "")
				require.ErrorIs(t, ctx.Err(), context.Canceled)
				var count int64
				require.NoError(t, fixture.db.Model(kind.record("", "", "")).Count(&count).Error)
				require.EqualValues(t, 1, count, "cancellation must retain the archive retry anchor")
				remaining, err := fixture.events.ListByUID(context.Background(), "uid")
				require.NoError(t, err)
				require.Empty(t, remaining, "event cleanup committed before cancellation")
				require.Equal(t, 500, response.Code, response.Body.String())
				assertAPIError(t, response, 500)
				require.NoError(t, fixture.db.Callback().Delete().Remove("cancel_after_event_delete"))
				apiRequest(t, fixture.router, http.MethodDelete, path, "", 200)
				require.NoError(t, fixture.db.Model(kind.record("", "", "")).Count(&count).Error)
				require.Zero(t, count)
				remaining, err = fixture.events.ListByUID(context.Background(), "uid")
				require.NoError(t, err)
				require.Empty(t, remaining)
			})
		}
	}
}

func TestRequestContextForcedDeletionRecovery(t *testing.T) {
	for _, deadline := range []bool{false, true} {
		t.Run(map[bool]string{false: "canceled", true: "deadline"}[deadline], func(t *testing.T) {
			probe := &contextKubeProbe{operation: "update"}
			router, kube := apiRouter(t, nil, probe.interceptors(), nil)
			apiRequest(t, router, http.MethodPost, "/experiments", apiResources()[0].body, 200)
			key := types.NamespacedName{Namespace: "target", Name: "example"}
			var before v1alpha1.PodChaos
			require.NoError(t, kube.Get(context.Background(), key, &before))
			probe.enabled = true
			ctx := stoppedRequestContext(t, deadline)
			response := requestWithContext(ctx, router, http.MethodDelete, "/experiments/uid?force=true", "")
			probe.assertForwarded(ctx, t)
			require.Equal(t, 500, response.Code, response.Body.String())
			assertAPIError(t, response, 500)
			var after v1alpha1.PodChaos
			require.NoError(t, kube.Get(context.Background(), key, &after))
			require.Equal(t, before, after)
			probe.contexts = nil
			liveCtx := liveRequestContext(t)
			response = requestWithContext(liveCtx, router, http.MethodDelete, "/experiments/uid?force=true", "")
			require.Equal(t, 200, response.Code, response.Body.String())
			probe.assertForwarded(liveCtx, t)
			require.True(t, apierrors.IsNotFound(kube.Get(context.Background(), key, &after)))
		})
	}
}
