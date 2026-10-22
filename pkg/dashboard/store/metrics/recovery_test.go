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

package metrics

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

type experimentSnapshot struct {
	core.ExperimentStore
	list func() ([]*core.ExperimentMeta, error)
}

func (s experimentSnapshot) ListMeta(context.Context, string, string, string, bool) ([]*core.ExperimentMeta, error) {
	return s.list()
}

type scheduleSnapshot struct {
	core.ScheduleStore
	list func() ([]*core.ScheduleMeta, error)
}

func (s scheduleSnapshot) ListMeta(context.Context, string, string, bool) ([]*core.ScheduleMeta, error) {
	return s.list()
}

type workflowSnapshot struct {
	core.WorkflowStore
	list func() ([]*core.WorkflowMeta, error)
}

func (s workflowSnapshot) ListMeta(context.Context, string, string, bool) ([]*core.WorkflowMeta, error) {
	return s.list()
}

func metricValues(families []*dto.MetricFamily) map[string]float64 {
	values := map[string]float64{}
	for _, family := range families {
		for _, metric := range family.Metric {
			key := family.GetName()
			for _, label := range metric.Label {
				key += "/" + label.GetName() + "=" + label.GetValue()
			}
			values[key] = metric.GetGauge().GetValue()
		}
	}
	return values
}

func TestConcurrentGatherKeepsIndependentSnapshots(t *testing.T) {
	collector := NewCollector(logr.Discard(), experimentSnapshot{list: func() ([]*core.ExperimentMeta, error) {
		runtime.Gosched()
		return []*core.ExperimentMeta{{Namespace: "a", Kind: "PodChaos"}, {Namespace: "a", Kind: "PodChaos"}, {Namespace: "b", Kind: "NetworkChaos"}}, nil
	}}, scheduleSnapshot{list: func() ([]*core.ScheduleMeta, error) {
		runtime.Gosched()
		return []*core.ScheduleMeta{{Namespace: "a"}, {Namespace: "b"}, {Namespace: "b"}}, nil
	}}, workflowSnapshot{list: func() ([]*core.WorkflowMeta, error) {
		runtime.Gosched()
		return []*core.WorkflowMeta{{Namespace: "a"}, {Namespace: "a"}, {Namespace: "b"}}, nil
	}})
	registry := prometheus.NewPedanticRegistry()
	registry.MustRegister(collector)
	want := map[string]float64{
		"chaos_dashboard_archived_experiments/namespace=a/type=PodChaos":     2,
		"chaos_dashboard_archived_experiments/namespace=b/type=NetworkChaos": 1,
		"chaos_dashboard_archived_schedules/namespace=a":                     1,
		"chaos_dashboard_archived_schedules/namespace=b":                     2,
		"chaos_dashboard_archived_workflows/namespace=a":                     2,
		"chaos_dashboard_archived_workflows/namespace=b":                     1,
	}
	const workers = 8
	results := make(chan error, workers)
	for range workers {
		go func() {
			for range 64 {
				families, err := registry.Gather()
				if err != nil {
					results <- err
					return
				}
				if got := metricValues(families); !reflect.DeepEqual(got, want) {
					results <- fmt.Errorf("incomplete concurrent scrape: got %v, want %v", got, want)
					return
				}
			}
			results <- nil
		}()
	}
	for range workers {
		select {
		case got := <-results:
			if got != nil {
				t.Error(got)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("gather did not complete")
		}
	}
}

func TestGatherStoreFailureAndRecovery(t *testing.T) {
	for _, failed := range []string{"experiments", "schedules", "workflows"} {
		t.Run(failed, func(t *testing.T) {
			failure := errors.New("database unavailable")
			failing := false
			collector := NewCollector(logr.Discard(), experimentSnapshot{list: func() ([]*core.ExperimentMeta, error) {
				if failing && failed == "experiments" {
					return nil, failure
				}
				return []*core.ExperimentMeta{{Namespace: "team", Kind: "PodChaos"}}, nil
			}}, scheduleSnapshot{list: func() ([]*core.ScheduleMeta, error) {
				if failing && failed == "schedules" {
					return nil, failure
				}
				return []*core.ScheduleMeta{{Namespace: "team"}}, nil
			}}, workflowSnapshot{list: func() ([]*core.WorkflowMeta, error) {
				if failing && failed == "workflows" {
					return nil, failure
				}
				return []*core.WorkflowMeta{{Namespace: "team"}}, nil
			}})
			registry := prometheus.NewPedanticRegistry()
			registry.MustRegister(collector)
			for _, broken := range []bool{false, true, false} {
				failing = broken
				got, err := registry.Gather()
				if err != nil {
					t.Fatal(err)
				}
				want := map[string]float64{}
				for _, kind := range []string{"experiments", "schedules", "workflows"} {
					if broken && kind == failed {
						continue
					}
					key := "chaos_dashboard_archived_" + kind + "/namespace=team"
					if kind == "experiments" {
						key += "/type=PodChaos"
					}
					want[key] = 1
				}
				if !reflect.DeepEqual(metricValues(got), want) {
					t.Fatalf("failing %v: got %v, want %v", broken, metricValues(got), want)
				}
			}
		})
	}
}
