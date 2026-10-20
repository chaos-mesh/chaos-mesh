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
	"net/url"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/event"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
	workflowstore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/workflow"
)

const workflowEventsPath = "/events/workflow/workflow-id"

type workflowEventsFixture struct {
	archiveFixture
	workflow core.WorkflowEntity
	events   []core.Event
}

func newWorkflowEventsFixture(t *testing.T) workflowEventsFixture {
	t.Helper()
	conf := &config.ChaosDashboardConfig{TargetNamespace: "target"}
	fixture := newArchiveFixture(t, conf)
	workflows, err := workflowstore.NewStore(fixture.db)
	require.NoError(t, err)
	event.Register(fixture.router.Group("/api"), event.NewService(fixture.events.EventStore, workflows, conf, logr.Discard()))
	workflow := core.WorkflowEntity{WorkflowMeta: core.WorkflowMeta{
		UID: "workflow-id", Namespace: "target", Name: "example", CreatedAt: time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC),
	}, Workflow: apiResources()[2].body}
	require.NoError(t, workflows.Save(context.Background(), &workflow))
	kube, err := clientpool.K8sClients.Client("")
	require.NoError(t, err)
	for _, node := range []struct{ name, namespace, workflow string }{
		{"child-a", "target", "example"}, {"child-b", "target", "example"},
		{"unrelated", "target", "another-workflow"}, {"other-namespace", "other", "example"},
	} {
		require.NoError(t, kube.Create(context.Background(), &v1alpha1.WorkflowNode{ObjectMeta: metav1.ObjectMeta{
			Name: node.name, Namespace: node.namespace, UID: types.UID(node.name),
			Labels: map[string]string{v1alpha1.LabelWorkflow: node.workflow},
		}}))
	}
	for i, row := range []struct{ uid, namespace, timestamp string }{
		{"workflow-id", "target", "2026-01-02T12:00:00Z"},
		{"child-a", "target", "2026-01-02T13:00:00Z"},
		{"child-b", "target", "2026-01-02T20:30:00+08:00"},
		{"child-a", "target", "2026-01-02T11:00:00Z"},
		{"workflow-id", "target", "2026-01-02T14:00:00Z"},
		{"unrelated", "target", "2026-01-02T12:40:00Z"},
		{"other-namespace", "target", "2026-01-02T12:50:00Z"},
		{"child-a", "other", "2026-01-02T12:55:00Z"},
	} {
		created, err := time.Parse(time.RFC3339, row.timestamp)
		require.NoError(t, err)
		require.NoError(t, fixture.events.Create(context.Background(), &core.Event{
			ID: uint(i + 1), ObjectID: row.uid, Namespace: row.namespace, CreatedAt: created,
			Name: row.uid, Kind: "WorkflowNode", Type: "Normal", Reason: "Observed", Message: row.timestamp,
		}))
	}
	var rows []core.Event
	require.NoError(t, fixture.db.Order("id").Find(&rows).Error)
	return workflowEventsFixture{fixture, workflow, rows}
}

func (f workflowEventsFixture) assertResponse(t *testing.T, response *httptest.ResponseRecorder, ids ...uint) {
	t.Helper()
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	want := make([]core.Event, 0, len(ids))
	for _, id := range ids {
		want = append(want, f.events[id-1])
	}
	encoded, err := json.Marshal(want)
	require.NoError(t, err)
	require.JSONEq(t, string(encoded), response.Body.String())
}

func (f workflowEventsFixture) assertUnchanged(t *testing.T) {
	t.Helper()
	var workflows []core.WorkflowEntity
	require.NoError(t, f.db.Find(&workflows).Error)
	require.Equal(t, []core.WorkflowEntity{f.workflow}, workflows)
	var events []core.Event
	require.NoError(t, f.db.Order("id").Find(&events).Error)
	require.Equal(t, f.events, events)
}

func workflowEventsFilteredPath() string {
	return workflowEventsPath + "?" + url.Values{
		"start": {"2026-01-02T20:00:00+08:00"},
		"end":   {"2026-01-02T21:00:00+08:00"},
		"limit": {"2"},
	}.Encode()
}

func TestWorkflowEventsSelection(t *testing.T) {
	fixture := newWorkflowEventsFixture(t)
	for _, scenario := range []struct {
		name, path string
		ids        []uint
	}{
		{"all descendants sorted by time", workflowEventsPath, []uint{5, 2, 3, 1, 4}},
		{"offset bounds before limit", workflowEventsFilteredPath(), []uint{2, 3}},
		{"inclusive time bounds", workflowEventsPath + "?start=2026-01-02T12:00:00Z&end=2026-01-02T13:00:00Z", []uint{2, 3, 1}},
		{"no matching events", workflowEventsPath + "?start=2026-01-03T00:00:00Z", nil},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			response := apiRequest(t, fixture.router, http.MethodGet, scenario.path, "", http.StatusOK)
			fixture.assertResponse(t, response, scenario.ids...)
			fixture.assertUnchanged(t)
		})
	}
	kube, err := clientpool.K8sClients.Client("")
	require.NoError(t, err)
	for _, name := range []string{"child-a", "child-b"} {
		require.NoError(t, kube.Delete(context.Background(), &v1alpha1.WorkflowNode{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "target"}}))
	}
	response := apiRequest(t, fixture.router, http.MethodGet, workflowEventsPath, "", http.StatusOK)
	fixture.assertResponse(t, response, 5, 1)
	fixture.assertUnchanged(t)
}

func TestWorkflowEventsDatabaseRecovery(t *testing.T) {
	for _, stage := range []struct{ name, table string }{
		{"workflow lookup", "workflow_entities"}, {"event lookup", "events"},
	} {
		t.Run(stage.name, func(t *testing.T) {
			fixture := newWorkflowEventsFixture(t)
			path := workflowEventsFilteredPath()
			fixture.assertResponse(t, apiRequest(t, fixture.router, http.MethodGet, path, "", http.StatusOK), 2, 3)
			require.NoError(t, fixture.db.Exec("ALTER TABLE "+stage.table+" RENAME TO unavailable").Error)
			response := apiRequest(t, fixture.router, http.MethodGet, path, "", http.StatusInternalServerError)
			assertAPIError(t, response, http.StatusInternalServerError)
			require.NoError(t, fixture.db.Exec("ALTER TABLE unavailable RENAME TO "+stage.table).Error)
			fixture.assertUnchanged(t)
			fixture.assertResponse(t, apiRequest(t, fixture.router, http.MethodGet, path, "", http.StatusOK), 2, 3)
			fixture.assertUnchanged(t)
		})
	}
}

func TestWorkflowEventsWorkflowStateRecovery(t *testing.T) {
	for _, scenario := range []struct {
		name string
		code int
	}{
		{"missing", http.StatusNotFound}, {"archived", http.StatusBadRequest},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			fixture := newWorkflowEventsFixture(t)
			if scenario.name == "missing" {
				require.NoError(t, fixture.db.Delete(&fixture.workflow).Error)
			} else {
				require.NoError(t, fixture.db.Model(&fixture.workflow).Update("archived", true).Error)
			}
			response := apiRequest(t, fixture.router, http.MethodGet, workflowEventsPath, "", scenario.code)
			assertAPIError(t, response, scenario.code)
			if scenario.name == "missing" {
				require.NoError(t, fixture.db.Create(&fixture.workflow).Error)
			} else {
				require.NoError(t, fixture.db.Model(&fixture.workflow).Update("archived", false).Error)
			}
			fixture.assertResponse(t, apiRequest(t, fixture.router, http.MethodGet, workflowEventsPath, "", http.StatusOK), 5, 2, 3, 1, 4)
			fixture.assertUnchanged(t)
		})
	}
}

func TestWorkflowEventsRequestContextRecovery(t *testing.T) {
	for _, scenario := range []string{"canceled before workflow", "deadline before workflow", "canceled before events"} {
		t.Run(scenario, func(t *testing.T) {
			fixture := newWorkflowEventsFixture(t)
			fixture.assertResponse(t, apiRequest(t, fixture.router, http.MethodGet, workflowEventsPath, "", http.StatusOK), 5, 2, 3, 1, 4)
			ctx := stoppedRequestContext(t, scenario == "deadline before workflow")
			var cancel context.CancelFunc
			cancelBeforeEvents := scenario == "canceled before events"
			if cancelBeforeEvents {
				ctx, cancel = context.WithCancel(liveRequestContext(t))
				defer cancel()
			}
			received := map[string]context.Context{}
			require.NoError(t, fixture.db.Callback().Query().Before("gorm:query").Register("workflow_event_context", func(db *gorm.DB) {
				received[db.Statement.Table] = db.Statement.Context
				if cancelBeforeEvents && db.Statement.Table == "events" {
					cancel()
				}
			}))
			response := requestWithContext(ctx, fixture.router, http.MethodGet, workflowEventsPath, "")
			require.Equal(t, http.StatusInternalServerError, response.Code, response.Body.String())
			assertAPIError(t, response, http.StatusInternalServerError)
			require.Error(t, ctx.Err())
			require.Contains(t, received, "workflow_entities")
			if cancelBeforeEvents {
				require.Contains(t, received, "events")
			}
			for _, got := range received {
				require.Equal(t, "request", got.Value(requestContextKey{}))
				require.ErrorIs(t, got.Err(), ctx.Err())
				deadline, ok := got.Deadline()
				wantDeadline, wantOK := ctx.Deadline()
				require.Equal(t, wantOK, ok)
				require.Equal(t, wantDeadline, deadline)
			}
			cancelBeforeEvents = false
			fixture.assertUnchanged(t)
			received = map[string]context.Context{}
			liveCtx := liveRequestContext(t)
			fixture.assertResponse(t, requestWithContext(liveCtx, fixture.router, http.MethodGet, workflowEventsPath, ""), 5, 2, 3, 1, 4)
			require.Contains(t, received, "workflow_entities")
			require.Contains(t, received, "events")
			for _, got := range received {
				require.Equal(t, "request", got.Value(requestContextKey{}))
				require.NoError(t, got.Err())
				deadline, ok := got.Deadline()
				wantDeadline, wantOK := liveCtx.Deadline()
				require.Equal(t, wantOK, ok)
				require.Equal(t, wantDeadline, deadline)
			}
			require.NoError(t, fixture.db.Callback().Query().Remove("workflow_event_context"))
			fixture.assertUnchanged(t)
		})
	}
}
