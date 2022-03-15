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
	"reflect"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/status"
)

// ChaosControllerManagerMetricsCollector implements prometheus.Collector interface
type ChaosControllerManagerMetricsCollector struct {
	logger              logr.Logger
	store               cache.Cache
	chaosExperiments    *prometheus.GaugeVec
	SidecarTemplates    prometheus.Gauge
	ConfigTemplates     *prometheus.GaugeVec
	InjectionConfigs    *prometheus.GaugeVec
	TemplateNotExist    *prometheus.CounterVec
	TemplateLoadError   prometheus.Counter
	ConfigNameDuplicate *prometheus.CounterVec
	InjectRequired      *prometheus.CounterVec
	Injections          *prometheus.CounterVec
	chaosSchedules      *prometheus.GaugeVec
	chaosWorkflows      *prometheus.GaugeVec
	EmittedEvents       *prometheus.CounterVec
}

// NewChaosControllerManagerMetricsCollector initializes metrics and collector
func NewChaosControllerManagerMetricsCollector(manager ctrl.Manager, registerer *prometheus.Registry, logger logr.Logger) *ChaosControllerManagerMetricsCollector {
	var store cache.Cache
	if manager != nil {
		store = manager.GetCache()
	}

	c := &ChaosControllerManagerMetricsCollector{
		logger: logger,
		store:  store,
		chaosExperiments: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "chaos_controller_manager_chaos_experiments",
			Help: "Total number of chaos experiments and their phases",
		}, []string{"namespace", "kind", "phase"}),
		SidecarTemplates: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "chaos_mesh_templates",
			Help: "Total number of injection templates",
		}),
		ConfigTemplates: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "chaos_mesh_config_templates",
			Help: "Total number of config templates",
		}, []string{"namespace", "template"}),
		InjectionConfigs: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "chaos_mesh_injection_configs",
			Help: "Total number of injection configs",
		}, []string{"namespace", "template"}),
		TemplateNotExist: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "chaos_mesh_template_not_exist_total",
			Help: "Total number of template not exist error",
		}, []string{"namespace", "template"}),
		ConfigNameDuplicate: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "chaos_mesh_config_name_duplicate_total",
			Help: "Total number of config name duplication error",
		}, []string{"namespace", "config"}),
		TemplateLoadError: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "chaos_mesh_template_load_failed_total",
			Help: "Total number of failures when rendering config args to template",
		}),
		InjectRequired: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "chaos_mesh_inject_required_total",
			Help: "Total number of injections required",
		}, []string{"namespace", "config"}),
		Injections: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "chaos_mesh_injections_total",
			Help: "Total number of sidecar injections performed on the webhook",
		}, []string{"namespace", "config"}),
		chaosSchedules: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "chaos_controller_manager_chaos_schedules",
			Help: "Total number of chaos schedules",
		}, []string{"namespace"}),
		chaosWorkflows: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "chaos_controller_manager_chaos_workflows",
			Help: "Total number of chaos workflows",
		}, []string{"namespace"}),
		EmittedEvents: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "chaos_controller_manager_emitted_event_total",
			Help: "Total number of the emitted event by chaos-controller-manager",
		}, []string{"type", "reason", "namespace"}),
	}

	if registerer != nil {
		registerer.MustRegister(c)
	}
	return c
}

// Describe implements the prometheus.Collector interface.
func (collector *ChaosControllerManagerMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	collector.chaosExperiments.Describe(ch)
	collector.SidecarTemplates.Describe(ch)
	collector.ConfigTemplates.Describe(ch)
	collector.InjectionConfigs.Describe(ch)
	collector.TemplateNotExist.Describe(ch)
	collector.ConfigNameDuplicate.Describe(ch)
	collector.TemplateLoadError.Describe(ch)
	collector.InjectRequired.Describe(ch)
	collector.Injections.Describe(ch)
	collector.EmittedEvents.Describe(ch)
	collector.chaosSchedules.Describe(ch)
	collector.chaosWorkflows.Describe(ch)
}

// Collect implements the prometheus.Collector interface.
func (collector *ChaosControllerManagerMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	collector.collectChaosExperiments()
	collector.collectChaosSchedules()
	collector.collectChaosWorkflows()
	collector.SidecarTemplates.Collect(ch)
	collector.ConfigTemplates.Collect(ch)
	collector.InjectionConfigs.Collect(ch)
	collector.TemplateNotExist.Collect(ch)
	collector.ConfigNameDuplicate.Collect(ch)
	collector.TemplateLoadError.Collect(ch)
	collector.InjectRequired.Collect(ch)
	collector.Injections.Collect(ch)
	collector.chaosExperiments.Collect(ch)
	collector.chaosSchedules.Collect(ch)
	collector.chaosWorkflows.Collect(ch)
	collector.EmittedEvents.Collect(ch)
}

func (collector *ChaosControllerManagerMetricsCollector) collectChaosExperiments() {
	// TODO(yeya24) if there is an error in List
	// the experiment status will be lost
	collector.chaosExperiments.Reset()

	for kind, obj := range v1alpha1.AllKinds() {
		expCache := map[string]map[string]int{}
		chaosList := obj.SpawnList()
		if err := collector.store.List(context.TODO(), chaosList); err != nil {
			collector.logger.Error(err, "failed to list chaos", "kind", kind)
			return
		}

		items := chaosList.GetItems()
		for _, item := range items {
			if _, ok := expCache[item.GetNamespace()]; !ok {
				// There is only 4 supported phases
				expCache[item.GetNamespace()] = make(map[string]int, 4)
			}
			innerObject := reflect.ValueOf(item).Interface().(v1alpha1.InnerObject)
			expCache[item.GetNamespace()][string(status.GetChaosStatus(innerObject))]++
		}

		for ns, v := range expCache {
			for phase, count := range v {
				collector.chaosExperiments.WithLabelValues(ns, kind, phase).Set(float64(count))
			}
		}
	}
}

func (collector *ChaosControllerManagerMetricsCollector) collectChaosSchedules() {
	collector.chaosSchedules.Reset()

	schedules := &v1alpha1.ScheduleList{}
	if err := collector.store.List(context.TODO(), schedules); err != nil {
		collector.logger.Error(err, "failed to list schedules")
		return
	}

	countByNamespace := make(map[string]int)
	items := schedules.GetItems()
	for _, item := range items {
		countByNamespace[item.GetNamespace()]++
	}

	for namespace, count := range countByNamespace {
		collector.chaosSchedules.WithLabelValues(namespace).Set(float64(count))
	}
}

func (collector *ChaosControllerManagerMetricsCollector) collectChaosWorkflows() {
	collector.chaosWorkflows.Reset()

	workflows := &v1alpha1.WorkflowList{}
	if err := collector.store.List(context.TODO(), workflows); err != nil {
		collector.logger.Error(err, "failed to list workflows")
		return
	}

	countByNamespace := make(map[string]int)
	items := workflows.GetItems()
	for _, item := range items {
		countByNamespace[item.GetNamespace()]++
	}

	for namespace, count := range countByNamespace {
		collector.chaosWorkflows.WithLabelValues(namespace).Set(float64(count))
	}
}
