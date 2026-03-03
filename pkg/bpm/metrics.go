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

package bpm

import (
	"github.com/prometheus/client_golang/prometheus"
)

type metricsCollector struct {
	backgroundProcessManager  *BackgroundProcessManager
	bpmControlledProcesses    prometheus.Gauge
	bpmControlledProcessTotal prometheus.Counter
}

// newMetricsCollector initializes metrics for each chaos daemon
func newMetricsCollector(backgroundProcessManager *BackgroundProcessManager, register prometheus.Registerer) *metricsCollector {
	collector := &metricsCollector{
		backgroundProcessManager: backgroundProcessManager,
		bpmControlledProcesses: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "chaos_daemon_bpm_controlled_processes",
			Help: "Current number of bpm controlled processes",
		}),
		bpmControlledProcessTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "chaos_daemon_bpm_controlled_process_total",
			Help: "Total count of bpm controlled processes",
		}),
	}
	register.MustRegister(collector)
	return collector
}

func (collector *metricsCollector) Describe(ch chan<- *prometheus.Desc) {
	collector.bpmControlledProcesses.Describe(ch)
	collector.bpmControlledProcessTotal.Describe(ch)
}

func (collector *metricsCollector) Collect(ch chan<- prometheus.Metric) {
	collector.collectBpmControlledProcesses()
	collector.bpmControlledProcesses.Collect(ch)
	collector.bpmControlledProcessTotal.Collect(ch)
}

func (collector *metricsCollector) collectBpmControlledProcesses() {
	ids := collector.backgroundProcessManager.GetIdentifiers()
	collector.bpmControlledProcesses.Set(float64(len(ids)))
}
