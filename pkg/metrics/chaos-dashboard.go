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
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

const chaosDashboardSubsystem = "chaos_dashboard"

// ChaosDashboardMetricsCollector implements prometheus.Collector interface
type ChaosDashboardMetricsCollector struct {
	httpRequestDuration *prometheus.HistogramVec
}

// NewChaosDashboardMetricsCollector initializes metrics and collector
func NewChaosDashboardMetricsCollector(engine *gin.Engine, registry *prometheus.Registry) *ChaosDashboardMetricsCollector {
	collector := &ChaosDashboardMetricsCollector{
		httpRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Subsystem: chaosDashboardSubsystem,
			Name:      "http_request_duration_seconds",
			Help:      "Time histogram for each HTTP query",
		}, []string{"path", "method", "status"}),
	}

	engine.Use(collector.ginMetricsCollector())
	registry.MustRegister(collector)

	return collector
}

// Describe implements the prometheus.Collector interface.
func (collector *ChaosDashboardMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	collector.httpRequestDuration.Describe(ch)
}

// Collect implements the prometheus.Collector interface.
func (collector *ChaosDashboardMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	collector.httpRequestDuration.Collect(ch)
}

func (collector *ChaosDashboardMetricsCollector) ginMetricsCollector() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		begin := time.Now()

		ctx.Next()

		collector.httpRequestDuration.WithLabelValues(ctx.FullPath(), ctx.Request.Method, strconv.Itoa(ctx.Writer.Status())).
			Observe(time.Since(begin).Seconds())
	}
}
