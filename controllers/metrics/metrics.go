package metrics

import (
	"context"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

var log = ctrl.Log.WithName("metrics-collector")

// ChaosCollector implements prometheus.Collector interface
type ChaosCollector struct {
	store               cache.Cache
	experimentStatus    *prometheus.GaugeVec
	SidecarTemplates    prometheus.Gauge
	ConfigTemplates     *prometheus.GaugeVec
	InjectionConfigs    *prometheus.GaugeVec
	TemplateNotExist    *prometheus.CounterVec
	TemplateLoadError   *prometheus.CounterVec
	ConfigNameDuplicate *prometheus.CounterVec
	InjectRequired      *prometheus.CounterVec
	Injections          *prometheus.CounterVec
}

// NewChaosCollector initializes metrics and collector
func NewChaosCollector(store cache.Cache, registerer prometheus.Registerer) *ChaosCollector {
	c := &ChaosCollector{
		store: store,
		experimentStatus: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "chaos_mesh_experiments",
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
		TemplateLoadError: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "chaos_mesh_template_load_failed_total",
			Help: "Total number of failures when rendering config args to template",
		}, []string{"namespace", "template", "config"}),
		InjectRequired: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "chaos_mesh_inject_required_total",
			Help: "Total number of injections required",
		}, []string{"namespace", "config"}),
		Injections: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "chaos_mesh_injections_total",
			Help: "Total number of sidecar injections performed on the webhook",
		}, []string{"namespace", "config"}),
	}
	registerer.MustRegister(c)
	return c
}

// Describe implements the prometheus.Collector interface.
func (c *ChaosCollector) Describe(ch chan<- *prometheus.Desc) {
	c.experimentStatus.Describe(ch)
	c.SidecarTemplates.Describe(ch)
	c.ConfigTemplates.Describe(ch)
	c.InjectionConfigs.Describe(ch)
	c.TemplateNotExist.Describe(ch)
	c.ConfigNameDuplicate.Describe(ch)
	c.TemplateLoadError.Describe(ch)
	c.InjectRequired.Describe(ch)
	c.Injections.Describe(ch)
}

// Collect implements the prometheus.Collector interface.
func (c *ChaosCollector) Collect(ch chan<- prometheus.Metric) {
	c.collect()
	c.SidecarTemplates.Collect(ch)
	c.ConfigTemplates.Collect(ch)
	c.InjectionConfigs.Collect(ch)
	c.TemplateNotExist.Collect(ch)
	c.ConfigNameDuplicate.Collect(ch)
	c.TemplateLoadError.Collect(ch)
	c.InjectRequired.Collect(ch)
	c.Injections.Collect(ch)
	c.experimentStatus.Collect(ch)
}

func (c *ChaosCollector) collect() {
	kinds := map[string]v1alpha1.ChaosList{
		v1alpha1.KindPodChaos:     &v1alpha1.PodChaosList{},
		v1alpha1.KindNetworkChaos: &v1alpha1.NetworkChaosList{},
		v1alpha1.KindStressChaos:  &v1alpha1.StressChaosList{},
		v1alpha1.KindIOChaos:      &v1alpha1.IoChaosList{},
		v1alpha1.KindKernelChaos:  &v1alpha1.KernelChaosList{},
		v1alpha1.KindTimeChaos:    &v1alpha1.TimeChaosList{},
	}

	// TODO(yeya24) if there is an error in List
	// the experiment status will be lost
	c.experimentStatus.Reset()

	for kind, obj := range kinds {
		expCache := map[string]map[string]int{}
		if err := c.store.List(context.TODO(), obj); err != nil {
			log.Error(err, "failed to list chaos", "kind", kind)
			return
		}
		for _, chaos := range obj.ListChaos() {
			if _, ok := expCache[chaos.Namespace]; !ok {
				// There is only 4 supported phases
				expCache[chaos.Namespace] = make(map[string]int, 4)
			}
			expCache[chaos.Namespace][chaos.Status]++
		}

		for ns, v := range expCache {
			for phase, count := range v {
				c.experimentStatus.WithLabelValues(ns, kind, phase).Set(float64(count))
			}
		}
	}
}
