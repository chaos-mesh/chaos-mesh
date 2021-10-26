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

	v1 "k8s.io/api/core/v1"

	"github.com/prometheus/client_golang/prometheus"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/status"
)

var log = ctrl.Log.WithName("metrics-collector")

// ChaosControllerManagerMetricsCollector implements prometheus.Collector interface
type ChaosControllerManagerMetricsCollector struct {
	store                cache.Cache
	chaosExperimentCount *prometheus.GaugeVec
	SidecarTemplates     prometheus.Gauge
	ConfigTemplates      *prometheus.GaugeVec
	InjectionConfigs     *prometheus.GaugeVec
	TemplateNotExist     *prometheus.CounterVec
	TemplateLoadError    prometheus.Counter
	ConfigNameDuplicate  *prometheus.CounterVec
	InjectRequired       *prometheus.CounterVec
	Injections           *prometheus.CounterVec
	chaosScheduleCount   *prometheus.GaugeVec
	chaosWorkflowCount   *prometheus.GaugeVec
	emittedEventCount    *prometheus.GaugeVec
}

// NewChaosControllerManagerMetricsCollector initializes metrics and collector
func NewChaosControllerManagerMetricsCollector(store cache.Cache, registerer prometheus.Registerer) *ChaosControllerManagerMetricsCollector {
	c := &ChaosControllerManagerMetricsCollector{
		store: store,
		chaosExperimentCount: prometheus.NewGaugeVec(prometheus.GaugeOpts{
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
		chaosScheduleCount: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "chaos_controller_manager_chaos_schedules",
			Help: "Total number of chaos schedules",
		}, []string{"namespace"}),
		chaosWorkflowCount: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "chaos_controller_manager_chaos_workflows",
			Help: "Total number of chaos workflows",
		}, []string{"namespace"}),
		emittedEventCount: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "chaos_controller_manager_emitted_events",
			Help: "Total number of the emitted events by chaos-controller-manager",
		}, []string{"type", "reason"}),
	}
	registerer.MustRegister(c)
	return c
}

// Describe implements the prometheus.Collector interface.
func (collector *ChaosControllerManagerMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	collector.chaosExperimentCount.Describe(ch)
	collector.SidecarTemplates.Describe(ch)
	collector.ConfigTemplates.Describe(ch)
	collector.InjectionConfigs.Describe(ch)
	collector.TemplateNotExist.Describe(ch)
	collector.ConfigNameDuplicate.Describe(ch)
	collector.TemplateLoadError.Describe(ch)
	collector.InjectRequired.Describe(ch)
	collector.Injections.Describe(ch)
	collector.chaosScheduleCount.Describe(ch)
	collector.chaosWorkflowCount.Describe(ch)
	collector.emittedEventCount.Describe(ch)
}

// Collect implements the prometheus.Collector interface.
func (collector *ChaosControllerManagerMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	collector.collectChaosExperimentCount()
	collector.collectChaosScheduleCount()
	collector.collectChaosWorkflowCount()
	collector.collectEmittedEventCount()
	collector.SidecarTemplates.Collect(ch)
	collector.ConfigTemplates.Collect(ch)
	collector.InjectionConfigs.Collect(ch)
	collector.TemplateNotExist.Collect(ch)
	collector.ConfigNameDuplicate.Collect(ch)
	collector.TemplateLoadError.Collect(ch)
	collector.InjectRequired.Collect(ch)
	collector.Injections.Collect(ch)
	collector.chaosExperimentCount.Collect(ch)
	collector.chaosScheduleCount.Collect(ch)
	collector.chaosWorkflowCount.Collect(ch)
	collector.emittedEventCount.Collect(ch)
}

func (collector *ChaosControllerManagerMetricsCollector) collectChaosExperimentCount() {
	// TODO(yeya24) if there is an error in List
	// the experiment status will be lost
	collector.chaosExperimentCount.Reset()

	for kind, obj := range v1alpha1.AllKinds() {
		expCache := map[string]map[string]int{}
		chaosList := obj.SpawnList()
		if err := collector.store.List(context.TODO(), chaosList); err != nil {
			log.Error(err, "failed to list chaos", "kind", kind)
			return
		}

		items := reflect.ValueOf(chaosList).Elem().FieldByName("Items")
		for i := 0; i < items.Len(); i++ {
			item := items.Index(i).Addr().Interface().(v1alpha1.InnerObject)
			if _, ok := expCache[item.GetNamespace()]; !ok {
				// There is only 4 supported phases
				expCache[item.GetNamespace()] = make(map[string]int, 4)
			}
			expCache[item.GetNamespace()][string(status.GetChaosStatus(item))]++
		}

		for ns, v := range expCache {
			for phase, count := range v {
				collector.chaosExperimentCount.WithLabelValues(ns, kind, phase).Set(float64(count))
			}
		}
	}
}

func (collector *ChaosControllerManagerMetricsCollector) collectChaosScheduleCount() {
	collector.chaosScheduleCount.Reset()

	schedules := v1alpha1.AllKindsIncludeScheduleAndWorkflow()[v1alpha1.KindSchedule].SpawnList()
	if err := collector.store.List(context.TODO(), schedules); err != nil {
		log.Error(err, "failed to list schedules")
		return
	}

	countByNamespace := make(map[string]int)
	items := reflect.ValueOf(schedules).Elem().FieldByName("Items")
	for i := 0; i < items.Len(); i++ {
		item := items.Index(i).Addr().Interface().(*v1alpha1.Schedule)
		countByNamespace[item.GetNamespace()]++
	}

	for namespace, count := range countByNamespace {
		collector.chaosScheduleCount.WithLabelValues(namespace).Set(float64(count))
	}
}

func (collector *ChaosControllerManagerMetricsCollector) collectChaosWorkflowCount() {
	collector.chaosWorkflowCount.Reset()

	workflows := v1alpha1.AllKindsIncludeScheduleAndWorkflow()[v1alpha1.KindWorkflow].SpawnList()
	if err := collector.store.List(context.TODO(), workflows); err != nil {
		log.Error(err, "failed to list workflows")
		return
	}

	countByNamespace := make(map[string]int)
	items := reflect.ValueOf(workflows).Elem().FieldByName("Items")
	for i := 0; i < items.Len(); i++ {
		item := items.Index(i).Addr().Interface().(*v1alpha1.Workflow)
		countByNamespace[item.GetNamespace()]++
	}

	for namespace, count := range countByNamespace {
		collector.chaosWorkflowCount.WithLabelValues(namespace).Set(float64(count))
	}
}

func (collector *ChaosControllerManagerMetricsCollector) collectEmittedEventCount() {
	collector.emittedEventCount.Reset()

	events := &v1.EventList{}
	if err := collector.store.List(context.TODO(), events); err != nil {
		log.Error(err, "failed to list events")
		return
	}

	countByTypeAndReason := make(map[string]map[string]int)
	for _, item := range events.Items {
		if _, ok := countByTypeAndReason[item.Type]; !ok {
			countByTypeAndReason[item.Type] = make(map[string]int)
		}
		countByTypeAndReason[item.Type][item.Reason]++
	}

	for typeLabel, countByReason := range countByTypeAndReason {
		for reason, count := range countByReason {
			collector.emittedEventCount.WithLabelValues(typeLabel, reason).Set(float64(count))
		}
	}
}
