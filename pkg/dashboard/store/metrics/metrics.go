// Copyright 2021 Chaos Mesh Authors.
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

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

const chaosDashboardSubsystem = "chaos_dashboard"

// Collector implements prometheus.Collector interface
type Collector struct {
	log             logr.Logger
	experimentStore core.ExperimentStore
	scheduleStore   core.ScheduleStore
	workflowStore   core.WorkflowStore

	archivedExperiments *prometheus.Desc
	archivedSchedules   *prometheus.Desc
	archivedWorkflows   *prometheus.Desc
}

// NewCollector initializes metrics and collector
func NewCollector(log logr.Logger, experimentStore core.ExperimentStore, scheduleStore core.ScheduleStore, workflowStore core.WorkflowStore) *Collector {
	return &Collector{
		log:             log,
		experimentStore: experimentStore,
		scheduleStore:   scheduleStore,
		workflowStore:   workflowStore,
		archivedExperiments: prometheus.NewDesc(
			prometheus.BuildFQName("", chaosDashboardSubsystem, "archived_experiments"),
			"Total number of archived chaos experiments", []string{"namespace", "type"}, nil),
		archivedSchedules: prometheus.NewDesc(
			prometheus.BuildFQName("", chaosDashboardSubsystem, "archived_schedules"),
			"Total number of archived chaos schedules", []string{"namespace"}, nil),
		archivedWorkflows: prometheus.NewDesc(
			prometheus.BuildFQName("", chaosDashboardSubsystem, "archived_workflows"),
			"Total number of archived chaos workflows", []string{"namespace"}, nil),
	}
}

// Describe implements the prometheus.Collector interface.
func (collector *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.archivedExperiments
	ch <- collector.archivedSchedules
	ch <- collector.archivedWorkflows
}

// Collect implements the prometheus.Collector interface.
func (collector *Collector) Collect(ch chan<- prometheus.Metric) {
	collector.collectArchivedExperiments(ch)
	collector.collectArchivedSchedules(ch)
	collector.collectArchivedWorkflows(ch)
}

func (collector *Collector) collectArchivedExperiments(ch chan<- prometheus.Metric) {
	metas, err := collector.experimentStore.ListMeta(context.TODO(), "", "", "", true)
	if err != nil {
		collector.log.Error(err, "fail to list all archived chaos experiments")
		return
	}

	countByNamespaceAndKind := map[string]map[string]int{}
	for _, meta := range metas {
		if _, ok := countByNamespaceAndKind[meta.Namespace]; !ok {
			countByNamespaceAndKind[meta.Namespace] = map[string]int{}
		}

		countByNamespaceAndKind[meta.Namespace][meta.Kind]++
	}

	for namespace, countByKind := range countByNamespaceAndKind {
		for kind, count := range countByKind {
			ch <- prometheus.MustNewConstMetric(collector.archivedExperiments, prometheus.GaugeValue, float64(count), namespace, kind)
		}
	}
}

func (collector *Collector) collectArchivedSchedules(ch chan<- prometheus.Metric) {
	metas, err := collector.scheduleStore.ListMeta(context.TODO(), "", "", true)
	if err != nil {
		collector.log.Error(err, "fail to list all archived schedules")
		return
	}

	countByNamespace := map[string]int{}
	for _, meta := range metas {
		countByNamespace[meta.Namespace]++
	}

	for namespace, count := range countByNamespace {
		ch <- prometheus.MustNewConstMetric(collector.archivedSchedules, prometheus.GaugeValue, float64(count), namespace)
	}
}

func (collector *Collector) collectArchivedWorkflows(ch chan<- prometheus.Metric) {
	metas, err := collector.workflowStore.ListMeta(context.TODO(), "", "", true)
	if err != nil {
		collector.log.Error(err, "fail to list all archived workflows")
		return
	}

	countByNamespace := map[string]int{}
	for _, meta := range metas {
		countByNamespace[meta.Namespace]++
	}

	for namespace, count := range countByNamespace {
		ch <- prometheus.MustNewConstMetric(collector.archivedWorkflows, prometheus.GaugeValue, float64(count), namespace)
	}
}

type Params struct {
	fx.In
	Log             logr.Logger
	Registry        *prometheus.Registry
	ExperimentStore core.ExperimentStore
	ScheduleStore   core.ScheduleStore
	WorkflowStore   core.WorkflowStore
}

func Register(params Params) {
	collector := NewCollector(params.Log, params.ExperimentStore, params.ScheduleStore, params.WorkflowStore)
	params.Registry.MustRegister(collector)
}
