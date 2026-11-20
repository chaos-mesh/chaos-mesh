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

package metrics

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func TestMetrics(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Metrics")
}

type metricsManager struct {
	ctrl.Manager
	store cache.Cache
}

func (m metricsManager) GetCache() cache.Cache { return m.store }

type metricsCache struct {
	cache.Cache
	reader client.Reader
	err    error
}

func (c *metricsCache) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	if c.err != nil {
		return c.err
	}
	return c.reader.List(ctx, list, opts...)
}

func gaugeSamples(registry *prometheus.Registry, names ...string) map[string]map[string]float64 {
	families, err := registry.Gather()
	Expect(err).NotTo(HaveOccurred())
	values := map[string]map[string]float64{}
	for _, family := range families {
		name := family.GetName()
		if !slices.Contains(names, name) {
			continue
		}
		values[name] = map[string]float64{}
		for _, sample := range family.GetMetric() {
			var labels []string
			for _, label := range sample.GetLabel() {
				labels = append(labels, label.GetName()+"="+label.GetValue())
			}
			slices.Sort(labels)
			values[name][strings.Join(labels, ",")] = sample.GetGauge().GetValue()
		}
	}
	return values
}

var _ = Describe("Controller manager metrics", func() {
	const experimentMetric = "chaos_controller_manager_chaos_experiments"
	const scheduleMetric = "chaos_controller_manager_chaos_schedules"
	const workflowMetric = "chaos_controller_manager_chaos_workflows"

	It("groups resources by namespace, experiment kind and phase and refreshes them after changes", func(ctx SpecContext) {
		scheme := runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		pod := &v1alpha1.PodChaos{ObjectMeta: metav1.ObjectMeta{Name: "example", Namespace: "team-a"},
			Spec: v1alpha1.PodChaosSpec{Action: v1alpha1.PodKillAction}}
		otherPod := pod.DeepCopy()
		otherPod.Name = "another"
		paused := &v1alpha1.NetworkChaos{ObjectMeta: metav1.ObjectMeta{Name: "pause", Namespace: "team-b"}}
		paused.Status.Conditions = []v1alpha1.ChaosCondition{{Type: v1alpha1.ConditionPaused, Status: corev1.ConditionTrue}}
		schedule := &v1alpha1.Schedule{ObjectMeta: metav1.ObjectMeta{Name: "daily", Namespace: "team-a"}}
		workflow := &v1alpha1.Workflow{ObjectMeta: metav1.ObjectMeta{Name: "rollout", Namespace: "team-b"}}
		kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod, otherPod, paused, schedule, workflow).
			WithStatusSubresource(pod).Build()
		store := &metricsCache{reader: kube}
		registry := prometheus.NewRegistry()
		NewChaosControllerManagerMetricsCollector(metricsManager{store: store}, registry, logr.Discard())
		names := []string{experimentMetric, scheduleMetric, workflowMetric}
		Expect(gaugeSamples(registry, names...)).To(Equal(map[string]map[string]float64{
			experimentMetric: {"kind=NetworkChaos,namespace=team-b,phase=paused": 1, "kind=PodChaos,namespace=team-a,phase=injecting": 2},
			scheduleMetric:   {"namespace=team-a": 1},
			workflowMetric:   {"namespace=team-b": 1},
		}))

		Expect(kube.Delete(ctx, schedule)).To(Succeed())
		Expect(kube.Delete(ctx, workflow)).To(Succeed())
		Expect(kube.Delete(ctx, paused)).To(Succeed())
		Expect(kube.Delete(ctx, otherPod)).To(Succeed())
		pod.Status.Conditions = []v1alpha1.ChaosCondition{
			{Type: v1alpha1.ConditionSelected, Status: corev1.ConditionTrue},
			{Type: v1alpha1.ConditionAllInjected, Status: corev1.ConditionTrue},
		}
		Expect(kube.Status().Update(ctx, pod)).To(Succeed())
		updated := map[string]map[string]float64{
			experimentMetric: {"kind=PodChaos,namespace=team-a,phase=running": 1},
		}
		Expect(gaugeSamples(registry, names...)).To(Equal(updated))

		store.err = errors.New("cache unavailable")
		Expect(gaugeSamples(registry, names...)).To(BeEmpty())
		store.err = nil
		Expect(gaugeSamples(registry, names...)).To(Equal(updated))
	})
})

var _ = Describe("Dashboard metrics", func() {
	It("groups observations by route template, method and response status", func() {
		const metric = "chaos_dashboard_http_request_duration_seconds"
		router := gin.New()
		registry := prometheus.NewRegistry()
		NewChaosDashboardMetricsCollector(router, registry)
		router.GET("/experiments/:uid", func(ctx *gin.Context) {
			if ctx.Param("uid") == "missing" {
				ctx.Status(http.StatusNotFound)
				return
			}
			ctx.Status(http.StatusOK)
		})
		for _, path := range []string{"/experiments/first", "/experiments/second", "/experiments/missing"} {
			router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))
		}
		families, err := registry.Gather()
		Expect(err).NotTo(HaveOccurred())
		Expect(families).To(HaveLen(1))
		Expect(families[0].GetName()).To(Equal(metric))
		observations := map[string]uint64{}
		for _, observation := range families[0].GetMetric() {
			labels := map[string]string{}
			for _, label := range observation.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}
			Expect(labels).To(HaveLen(3))
			Expect(labels).To(HaveKeyWithValue("path", "/experiments/:uid"))
			Expect(labels).To(HaveKeyWithValue("method", http.MethodGet))
			observations[labels["status"]] = observation.GetHistogram().GetSampleCount()
		}
		Expect(observations).To(Equal(map[string]uint64{"200": 2, "404": 1}))
	})
})
