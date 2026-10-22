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

package metrics

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
	experimentstore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/experiment"
	schedulestore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/schedule"
	workflowstore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/workflow"
)

func TestArchivedMetricsRefreshAfterDeletion(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "metrics.db")), &gorm.Config{})
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
	for _, row := range []core.ExperimentMeta{
		{UID: "a", Namespace: "default", Kind: "PodChaos", Archived: true},
		{UID: "b", Namespace: "default", Kind: "PodChaos", Archived: true},
		{UID: "c", Namespace: "other", Kind: "NetworkChaos", Archived: true},
		{UID: "active", Namespace: "active", Kind: "PodChaos", Archived: false},
	} {
		if err := experiments.Save(ctx, &core.Experiment{ExperimentMeta: row}); err != nil {
			t.Fatal(err)
		}
	}
	for _, archived := range []bool{false, true, true} {
		if err := schedules.Save(ctx, &core.Schedule{Namespace: "default", Archived: archived}); err != nil {
			t.Fatal(err)
		}
		if err := workflows.Save(ctx, &core.WorkflowEntity{Namespace: "default", Archived: archived}); err != nil {
			t.Fatal(err)
		}
	}
	registry := prometheus.NewPedanticRegistry()
	registry.MustRegister(NewCollector(logr.Discard(), experiments, schedules, workflows))
	metrics, err := registry.Gather()
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]float64{}
	for _, family := range metrics {
		for _, metric := range family.Metric {
			key := family.GetName()
			for _, label := range metric.Label {
				key += "/" + label.GetName() + "=" + label.GetValue()
			}
			got[key] = metric.GetGauge().GetValue()
		}
	}
	want := map[string]float64{
		"chaos_dashboard_archived_experiments/namespace=default/type=PodChaos":   2,
		"chaos_dashboard_archived_experiments/namespace=other/type=NetworkChaos": 1,
		"chaos_dashboard_archived_schedules/namespace=default":                   2,
		"chaos_dashboard_archived_workflows/namespace=default":                   2,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("metrics = %v, want %v", got, want)
	}
	for _, model := range []any{&core.Experiment{}, &core.Schedule{}, &core.WorkflowEntity{}} {
		if err := db.Where("archived = ?", true).Delete(model).Error; err != nil {
			t.Fatal(err)
		}
	}
	metrics, err = registry.Gather()
	if err != nil {
		t.Fatal(err)
	}
	if len(metrics) != 0 {
		t.Fatalf("deleted archives remain in metrics: %v", metrics)
	}
}
