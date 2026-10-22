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
//

package ttlcontroller

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/go-logr/logr"
	"gorm.io/gorm"

	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
	eventstore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/event"
	experimentstore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/experiment"
	schedulestore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/schedule"
	workflowstore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/workflow"
)

func TestWorkerAppliesIndependentRetentionPolicies(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "ttl.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Error(err)
		}
	})
	events, err := eventstore.NewStore(db)
	if err != nil {
		t.Fatal(err)
	}
	experiments, err := experimentstore.NewStore(db)
	if err != nil {
		t.Fatal(err)
	}
	schedules, err := schedulestore.NewStore(db)
	if err != nil {
		t.Fatal(err)
	}
	workflows, err := workflowstore.NewStore(db)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	config := &config.TTLConfig{EventTTL: time.Hour, ExperimentTTL: 2 * time.Hour, ScheduleTTL: 3 * time.Hour, WorkflowTTL: 4 * time.Hour}
	for _, age := range []time.Duration{30 * time.Minute, 90 * time.Minute} {
		if err := events.Create(ctx, &core.Event{ObjectID: age.String(), CreatedAt: now.Add(-age)}); err != nil {
			t.Fatal(err)
		}
	}
	for _, state := range []struct {
		uid      string
		archived bool
		age      time.Duration
		noFinish bool
	}{
		{"expired", true, 5 * time.Hour, false}, {"retained", true, 30 * time.Minute, false}, {"active", false, 5 * time.Hour, false}, {"no-finish", true, 0, true},
	} {
		finish := now.Add(-state.age)
		var finishTime *time.Time
		if !state.noFinish {
			finishTime = &finish
		}
		if err := experiments.Save(ctx, &core.Experiment{UID: state.uid, Archived: state.archived, FinishTime: finishTime}); err != nil {
			t.Fatal(err)
		}
		if err := schedules.Save(ctx, &core.Schedule{UID: state.uid, Archived: state.archived, FinishTime: finishTime}); err != nil {
			t.Fatal(err)
		}
		if err := workflows.Save(ctx, &core.WorkflowEntity{UID: state.uid, Archived: state.archived, FinishTime: finishTime}); err != nil {
			t.Fatal(err)
		}
	}
	boundary := now.Add(-150 * time.Minute)
	if err := experiments.Save(ctx, &core.Experiment{UID: "policy", Archived: true, FinishTime: &boundary}); err != nil {
		t.Fatal(err)
	}
	if err := schedules.Save(ctx, &core.Schedule{UID: "policy", Archived: true, FinishTime: &boundary}); err != nil {
		t.Fatal(err)
	}
	workflowBoundary := now.Add(-210 * time.Minute)
	if err := schedules.Save(ctx, &core.Schedule{UID: "workflow-policy", Archived: true, FinishTime: &workflowBoundary}); err != nil {
		t.Fatal(err)
	}
	if err := workflows.Save(ctx, &core.WorkflowEntity{UID: "workflow-policy", Archived: true, FinishTime: &workflowBoundary}); err != nil {
		t.Fatal(err)
	}
	controller := NewController(events, experiments, schedules, workflows, config, logr.Discard())
	for range 2 {
		controller.runWorker(context.Background())
		gotEvents, err := events.List(ctx)
		if err != nil || len(gotEvents) != 1 || gotEvents[0].ObjectID != "30m0s" {
			t.Fatalf("events = %#v, error = %v", gotEvents, err)
		}
		for _, table := range []struct {
			model any
			want  []string
		}{
			{&core.Experiment{}, []string{"active", "no-finish", "retained"}},
			{&core.Schedule{}, []string{"active", "no-finish", "policy", "retained"}},
			{&core.WorkflowEntity{}, []string{"active", "no-finish", "retained", "workflow-policy"}},
		} {
			var uids []string
			if err := db.Model(table.model).Order("uid").Pluck("uid", &uids).Error; err != nil {
				t.Fatal(err)
			}
			if len(uids) != len(table.want) {
				t.Fatalf("%T UIDs = %v, want %v", table.model, uids, table.want)
			}
			for i := range uids {
				if uids[i] != table.want[i] {
					t.Fatalf("%T UIDs = %v, want %v", table.model, uids, table.want)
				}
			}
		}
	}
}
