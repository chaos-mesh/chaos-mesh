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

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	authorizationv1 "k8s.io/api/authorization/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/archive"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
	eventstore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/event"
	experimentstore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/experiment"
	schedulestore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/schedule"
	workflowstore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/workflow"
)

type recoveryEvents struct {
	core.EventStore
	failure error
}

func (s *recoveryEvents) DeleteByUID(ctx context.Context, uid string) error {
	if s.failure != nil {
		return s.failure
	}
	return s.EventStore.DeleteByUID(ctx, uid)
}

func (s *recoveryEvents) DeleteByUIDList(ctx context.Context, uids []string) error {
	if s.failure != nil {
		return s.failure
	}
	return s.EventStore.DeleteByUIDList(ctx, uids)
}

type archiveFixture struct {
	router *gin.Engine
	db     *gorm.DB
	events *recoveryEvents
}

func newArchiveFixture(t *testing.T, conf *config.ChaosDashboardConfig) archiveFixture {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	experiments, err := experimentstore.NewStore(db)
	require.NoError(t, err)
	schedules, err := schedulestore.NewStore(db)
	require.NoError(t, err)
	workflows, err := workflowstore.NewStore(db)
	require.NoError(t, err)
	events, err := eventstore.NewStore(db)
	require.NoError(t, err)
	failures := &recoveryEvents{EventStore: events}
	router, kube := apiRouter(t, nil, interceptor.Funcs{}, nil)
	auth := kubefake.NewClientset()
	auth.PrependReactor("create", "selfsubjectaccessreviews", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, &authorizationv1.SelfSubjectAccessReview{Status: authorizationv1.SubjectAccessReviewStatus{Allowed: true}}, nil
	})
	clientpool.K8sClients = apiArchivePool{kube: kube, auth: auth.AuthorizationV1()}
	archive.Register(router.Group("/api"), archive.NewService(experiments, schedules, failures, workflows, conf))
	return archiveFixture{router, db, failures}
}

type archiveKind struct {
	path   string
	record func(uid, namespace, snapshot string) any
}

func archiveKinds() []archiveKind {
	return []archiveKind{
		{"/archives", func(uid, namespace, snapshot string) any {
			return &core.Experiment{UID: uid, Namespace: namespace, Name: "example", Kind: "PodChaos", Archived: true, Experiment: snapshot}
		}},
		{"/archives/schedules", func(uid, namespace, snapshot string) any {
			return &core.Schedule{UID: uid, Namespace: namespace, Name: "example", Kind: "Schedule", Archived: true, Schedule: snapshot}
		}},
		{"/archives/workflows", func(uid, namespace, snapshot string) any {
			return &core.WorkflowEntity{UID: uid, Namespace: namespace, Name: "example", Archived: true, Workflow: snapshot}
		}},
	}
}

func TestArchiveCleanupRetryAfterEventFailure(t *testing.T) {
	for _, kind := range archiveKinds() {
		for _, batch := range []bool{false, true} {
			name := kind.path + map[bool]string{false: "/single", true: "/batch"}[batch]
			t.Run(name, func(t *testing.T) {
				fixture := newArchiveFixture(t, &config.ChaosDashboardConfig{TargetNamespace: "target"})
				uids := []string{"uid"}
				if batch {
					uids = append(uids, "second")
				}
				for _, uid := range uids {
					require.NoError(t, fixture.db.Create(kind.record(uid, "target", `{}`)).Error)
					require.NoError(t, fixture.events.Create(context.Background(), &core.Event{ObjectID: uid}))
				}
				require.NoError(t, fixture.events.Create(context.Background(), &core.Event{ObjectID: "untouched"}))
				path := kind.path + "/uid"
				if batch {
					path = kind.path + "?uids=uid,second,uid,missing"
				}
				fixture.events.failure = errors.New("event database unavailable")
				response := apiRequest(t, fixture.router, http.MethodDelete, path, "", 500)
				assertAPIError(t, response, 500)
				var count int64
				require.NoError(t, fixture.db.Model(kind.record("", "", "")).Where("uid IN ?", uids).Count(&count).Error)
				require.EqualValues(t, len(uids), count, "event cleanup failure must retain retry anchors")
				remaining, err := fixture.events.ListByUIDList(context.Background(), uids)
				require.NoError(t, err)
				require.Len(t, remaining, len(uids))
				fixture.events.failure = nil
				apiRequest(t, fixture.router, http.MethodDelete, path, "", 200)
				require.NoError(t, fixture.db.Model(kind.record("", "", "")).Where("uid IN ?", uids).Count(&count).Error)
				require.Zero(t, count)
				remaining, err = fixture.events.ListByUIDList(context.Background(), uids)
				require.NoError(t, err)
				require.Empty(t, remaining, "retry must clean events and archive rows")
				untouched, err := fixture.events.ListByUID(context.Background(), "untouched")
				require.NoError(t, err)
				require.Len(t, untouched, 1)
			})
		}
	}
}

func TestArchiveDeleteFailureRetainsRetryAnchor(t *testing.T) {
	for _, kind := range archiveKinds() {
		for _, batch := range []bool{false, true} {
			t.Run(kind.path+map[bool]string{false: "/single", true: "/batch"}[batch], func(t *testing.T) {
				fixture := newArchiveFixture(t, &config.ChaosDashboardConfig{TargetNamespace: "target"})
				require.NoError(t, fixture.db.Create(kind.record("uid", "target", `{}`)).Error)
				require.NoError(t, fixture.events.Create(context.Background(), &core.Event{ObjectID: "uid"}))
				require.NoError(t, fixture.events.Create(context.Background(), &core.Event{ObjectID: "untouched"}))
				require.NoError(t, fixture.db.Callback().Delete().Before("gorm:delete").Register("fail_archive_delete", func(db *gorm.DB) {
					if db.Statement.Schema.Name != "Event" {
						db.AddError(errors.New("archive database unavailable"))
					}
				}))
				path := kind.path + "/uid"
				if batch {
					path = kind.path + "?uids=uid"
				}
				response := apiRequest(t, fixture.router, http.MethodDelete, path, "", 500)
				assertAPIError(t, response, 500)
				var count int64
				require.NoError(t, fixture.db.Model(kind.record("", "", "")).Where("uid = ?", "uid").Count(&count).Error)
				require.EqualValues(t, 1, count)
				remaining, err := fixture.events.ListByUID(context.Background(), "uid")
				require.NoError(t, err)
				require.Empty(t, remaining, "event cleanup committed before archive deletion failed")
				require.NoError(t, fixture.db.Callback().Delete().Remove("fail_archive_delete"))
				apiRequest(t, fixture.router, http.MethodDelete, path, "", 200)
				require.NoError(t, fixture.db.Model(kind.record("", "", "")).Where("uid = ?", "uid").Count(&count).Error)
				require.Zero(t, count)
				remaining, err = fixture.events.ListByUID(context.Background(), "uid")
				require.NoError(t, err)
				require.Empty(t, remaining)
				untouched, err := fixture.events.ListByUID(context.Background(), "untouched")
				require.NoError(t, err)
				require.Len(t, untouched, 1)
			})
		}
	}
}

func TestArchiveInvalidSnapshotReturnsErrorJSON(t *testing.T) {
	for _, kind := range archiveKinds() {
		for _, snapshot := range []string{`null`, `[]`, `{} garbage`, `{"spec":`} {
			t.Run(kind.path+"/"+snapshot, func(t *testing.T) {
				fixture := newArchiveFixture(t, &config.ChaosDashboardConfig{TargetNamespace: "target"})
				require.NoError(t, fixture.db.Create(kind.record("uid", "target", snapshot)).Error)
				response := apiRequest(t, fixture.router, http.MethodGet, kind.path+"/uid", "", 500)
				assertAPIError(t, response, 500)
			})
		}
	}
}

func TestArchiveListNamespaceConfiguration(t *testing.T) {
	for _, kind := range archiveKinds() {
		for _, scenario := range []struct {
			name    string
			cluster bool
			query   string
			want    []string
		}{
			{"namespaced default", false, "", []string{"target-id"}},
			{"cluster default", true, "", []string{"target-id", "other-id"}},
			{"explicit namespace", true, "?namespace=other", []string{"other-id"}},
			{"no matches", true, "?namespace=missing", []string{}},
		} {
			t.Run(kind.path+"/"+scenario.name, func(t *testing.T) {
				fixture := newArchiveFixture(t, &config.ChaosDashboardConfig{TargetNamespace: "target", ClusterScoped: scenario.cluster})
				require.NoError(t, fixture.db.Create(kind.record("target-id", "target", `{}`)).Error)
				require.NoError(t, fixture.db.Create(kind.record("other-id", "other", `{}`)).Error)
				response := apiRequest(t, fixture.router, http.MethodGet, kind.path+scenario.query, "", 200)
				var rows []struct {
					UID string `json:"uid"`
				}
				require.NoError(t, json.Unmarshal(response.Body.Bytes(), &rows))
				ids := make([]string, 0, len(rows))
				for _, row := range rows {
					ids = append(ids, row.UID)
				}
				require.ElementsMatch(t, scenario.want, ids)
			})
		}
	}
}

func TestArchiveReadAndDatabaseErrors(t *testing.T) {
	for _, kind := range archiveKinds() {
		t.Run(kind.path, func(t *testing.T) {
			fixture := newArchiveFixture(t, &config.ChaosDashboardConfig{TargetNamespace: "target"})
			missing := apiRequest(t, fixture.router, http.MethodGet, kind.path+"/missing", "", 404)
			assertAPIError(t, missing, 404)
			require.NoError(t, fixture.db.Create(kind.record("uid", "target", `{"apiVersion":"chaos-mesh.org/v1alpha1","kind":"Snapshot","metadata":{"name":"saved","namespace":"target"},"spec":{}}`)).Error)
			detail := apiRequest(t, fixture.router, http.MethodGet, kind.path+"/uid", "", 200)
			var decoded struct {
				UID        string `json:"uid"`
				KubeObject struct {
					Kind string                           `json:"kind"`
					Meta struct{ Name, Namespace string } `json:"metadata"`
				} `json:"kube_object"`
			}
			require.NoError(t, json.Unmarshal(detail.Body.Bytes(), &decoded))
			require.Equal(t, "uid", decoded.UID)
			require.Equal(t, "Snapshot", decoded.KubeObject.Kind)
			require.Equal(t, "saved", decoded.KubeObject.Meta.Name)
			require.Equal(t, "target", decoded.KubeObject.Meta.Namespace)
			require.NoError(t, fixture.db.Callback().Query().Before("gorm:query").Register("fail_archive_read", func(db *gorm.DB) { db.AddError(errors.New("database unavailable")) }))
			for _, path := range []string{kind.path, kind.path + "/uid"} {
				response := apiRequest(t, fixture.router, http.MethodGet, path, "", 500)
				assertAPIError(t, response, 500)
			}
			if kind.path != "/archives/workflows" {
				response := apiRequest(t, fixture.router, http.MethodDelete, kind.path+"/uid", "", 500)
				assertAPIError(t, response, 500)
			}
			require.NoError(t, fixture.db.Callback().Query().Remove("fail_archive_read"))
			var count int64
			require.NoError(t, fixture.db.Model(kind.record("", "", "")).Where("uid = ?", "uid").Count(&count).Error)
			require.EqualValues(t, 1, count)
			apiRequest(t, fixture.router, http.MethodGet, kind.path+"/uid", "", 200)
		})
	}
}

func TestArchiveEmptyBatchPreservesRecords(t *testing.T) {
	for _, kind := range archiveKinds() {
		for _, query := range []string{"", "?uids="} {
			t.Run(kind.path+query, func(t *testing.T) {
				fixture := newArchiveFixture(t, &config.ChaosDashboardConfig{TargetNamespace: "target"})
				require.NoError(t, fixture.db.Create(kind.record("uid", "target", `{}`)).Error)
				require.NoError(t, fixture.events.Create(context.Background(), &core.Event{ObjectID: "uid"}))
				response := apiRequest(t, fixture.router, http.MethodDelete, kind.path+query, "", 400)
				assertAPIError(t, response, 400)
				var count int64
				require.NoError(t, fixture.db.Model(kind.record("", "", "")).Count(&count).Error)
				require.EqualValues(t, 1, count)
				events, err := fixture.events.ListByUID(context.Background(), "uid")
				require.NoError(t, err)
				require.Len(t, events, 1)
			})
		}
	}
}

func TestArchiveDeletePreservesUnownedHistories(t *testing.T) {
	for _, kind := range archiveKinds() {
		for _, mode := range []string{"single", "batch missing", "batch mixed"} {
			t.Run(kind.path+"/"+mode, func(t *testing.T) {
				fixture := newArchiveFixture(t, &config.ChaosDashboardConfig{TargetNamespace: "target"})
				foreignUIDs := []string{"workflow-node", "orphan"}
				eventKinds := map[string]string{"workflow-node": "WorkflowNode", "orphan": "PodChaos"}
				for _, foreign := range archiveKinds() {
					if foreign.path == kind.path {
						continue
					}
					uid := strings.ReplaceAll(foreign.path, "/", "-")
					foreignUIDs = append(foreignUIDs, uid)
					eventKinds[uid] = map[string]string{"/archives": "PodChaos", "/archives/schedules": "Schedule", "/archives/workflows": "Workflow"}[foreign.path]
					require.NoError(t, fixture.db.Create(foreign.record(uid, "target", `{}`)).Error)
				}
				for _, uid := range foreignUIDs {
					for _, reason := range []string{"Started", "Completed"} {
						require.NoError(t, fixture.events.Create(context.Background(), &core.Event{ObjectID: uid, Kind: eventKinds[uid], Reason: reason, Message: "preserve history"}))
					}
				}
				before, err := fixture.events.List(context.Background())
				require.NoError(t, err)
				if mode == "single" {
					for _, uid := range append(foreignUIDs, "absent") {
						status := 404
						if kind.path == "/archives/workflows" {
							status = 200
						}
						response := apiRequest(t, fixture.router, http.MethodDelete, kind.path+"/"+uid, "", status)
						if status == 404 {
							assertAPIError(t, response, status)
						}
					}
				} else {
					uids := append([]string{}, foreignUIDs...)
					uids = append(uids, "absent")
					if mode == "batch mixed" {
						for _, uid := range []string{"owned", "owned-second"} {
							require.NoError(t, fixture.db.Create(kind.record(uid, "target", `{}`)).Error)
							require.NoError(t, fixture.events.Create(context.Background(), &core.Event{ObjectID: uid}))
							uids = append(uids, uid)
						}
						uids = append(uids, "owned")
					}
					apiRequest(t, fixture.router, http.MethodDelete, kind.path+"?uids="+strings.Join(uids, ","), "", 200)
				}
				after, err := fixture.events.List(context.Background())
				require.NoError(t, err)
				require.ElementsMatch(t, before, after, "a route may only delete events for records in its own store")
				for _, recordKind := range archiveKinds() {
					var count int64
					require.NoError(t, fixture.db.Model(recordKind.record("", "", "")).Count(&count).Error)
					if recordKind.path == kind.path {
						require.Zero(t, count)
					} else {
						require.EqualValues(t, 1, count)
					}
				}
			})
		}
	}
}

func TestArchiveDeleteLookupFailurePreservesRecordsAndEvents(t *testing.T) {
	for _, kind := range archiveKinds() {
		for _, batch := range []bool{false, true} {
			t.Run(kind.path+map[bool]string{false: "/single", true: "/batch"}[batch], func(t *testing.T) {
				fixture := newArchiveFixture(t, &config.ChaosDashboardConfig{TargetNamespace: "target"})
				require.NoError(t, fixture.db.Create(kind.record("uid", "target", `{}`)).Error)
				require.NoError(t, fixture.events.Create(context.Background(), &core.Event{ObjectID: "uid"}))
				before, err := fixture.events.List(context.Background())
				require.NoError(t, err)
				calls := 0
				require.NoError(t, fixture.db.Callback().Query().Before("gorm:query").Register("fail_archive_lookup", func(db *gorm.DB) {
					calls++
					if !batch || calls == 2 {
						db.AddError(errors.New("archive lookup unavailable"))
					}
				}))
				path := kind.path + "/uid"
				if batch {
					path = kind.path + "?uids=uid,unknown"
				}
				response := apiRequest(t, fixture.router, http.MethodDelete, path, "", 500)
				assertAPIError(t, response, 500)
				require.NoError(t, fixture.db.Callback().Query().Remove("fail_archive_lookup"))
				var count int64
				require.NoError(t, fixture.db.Model(kind.record("", "", "")).Count(&count).Error)
				require.EqualValues(t, 1, count)
				after, err := fixture.events.List(context.Background())
				require.NoError(t, err)
				require.Equal(t, before, after)
				apiRequest(t, fixture.router, http.MethodDelete, path, "", 200)
				require.NoError(t, fixture.db.Model(kind.record("", "", "")).Count(&count).Error)
				require.Zero(t, count)
				after, err = fixture.events.List(context.Background())
				require.NoError(t, err)
				require.Empty(t, after)
			})
		}
	}
}
