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
	"time"

	"github.com/prometheus/client_golang/prometheus"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

// ChaosControllerManagerMetricsCollector implements prometheus.Collector interface
type ChaosControllerManagerMetricsCollector struct {
	store               cache.Cache
	SidecarTemplates    prometheus.Gauge
	ConfigTemplates     *prometheus.GaugeVec
	InjectionConfigs    *prometheus.GaugeVec
	TemplateNotExist    *prometheus.CounterVec
	TemplateLoadError   prometheus.Counter
	ConfigNameDuplicate *prometheus.CounterVec
	InjectRequired      *prometheus.CounterVec
	Injections          *prometheus.CounterVec
	reconcileDuration   *prometheus.HistogramVec
	EmittedEvents       *prometheus.CounterVec
}

// NewChaosControllerManagerMetricsCollector initializes metrics and collector
func NewChaosControllerManagerMetricsCollector(manager ctrl.Manager, registerer *prometheus.Registry) *ChaosControllerManagerMetricsCollector {
	var store cache.Cache
	if manager != nil {
		store = manager.GetCache()
	}

	c := &ChaosControllerManagerMetricsCollector{
		store: store,
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
		reconcileDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "chaos_controller_manager_reconcile_duration_seconds",
			Help:    "Duration histogram for each reconcile request",
			Buckets: prometheus.DefBuckets,
		}, []string{"type"}),
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

// NewTestChaosControllerManagerMetricsCollector provides metrics collector for testing
func NewTestChaosControllerManagerMetricsCollector() *ChaosControllerManagerMetricsCollector {
	return NewChaosControllerManagerMetricsCollector(nil, nil)
}

// Describe implements the prometheus.Collector interface.
func (collector *ChaosControllerManagerMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	collector.SidecarTemplates.Describe(ch)
	collector.ConfigTemplates.Describe(ch)
	collector.InjectionConfigs.Describe(ch)
	collector.TemplateNotExist.Describe(ch)
	collector.ConfigNameDuplicate.Describe(ch)
	collector.TemplateLoadError.Describe(ch)
	collector.InjectRequired.Describe(ch)
	collector.Injections.Describe(ch)
	collector.reconcileDuration.Describe(ch)
	collector.EmittedEvents.Describe(ch)
}

// Collect implements the prometheus.Collector interface.
func (collector *ChaosControllerManagerMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	collector.SidecarTemplates.Collect(ch)
	collector.ConfigTemplates.Collect(ch)
	collector.InjectionConfigs.Collect(ch)
	collector.TemplateNotExist.Collect(ch)
	collector.ConfigNameDuplicate.Collect(ch)
	collector.TemplateLoadError.Collect(ch)
	collector.InjectRequired.Collect(ch)
	collector.Injections.Collect(ch)
	collector.reconcileDuration.Collect(ch)
	collector.EmittedEvents.Collect(ch)
}

func (collector *ChaosControllerManagerMetricsCollector) CollectReconcileDuration(typeLabel string, before time.Time) {
	after := time.Now()
	duration := after.Sub(before).Seconds()
	collector.reconcileDuration.WithLabelValues(typeLabel).Observe(duration)
}
