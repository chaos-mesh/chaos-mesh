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
	"github.com/prometheus/client_golang/prometheus"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
)

// DefaultChaosDaemonMetricsCollector is the default metrics collector for chaos daemon
var DefaultChaosDaemonMetricsCollector = NewChaosDaemonMetricsCollector()

type ChaosDaemonMetricsCollector struct {
	backgroundProcessManager  bpm.BackgroundProcessManager
	bpmControlledProcessTotal prometheus.Counter
	bpmControlledProcesses    prometheus.Gauge
}

// NewChaosDaemonMetricsCollector initializes metrics for each chaos daemon
func NewChaosDaemonMetricsCollector() *ChaosDaemonMetricsCollector {
	return &ChaosDaemonMetricsCollector{
		backgroundProcessManager: bpm.NewBackgroundProcessManager(nil),
		bpmControlledProcessTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "chaos_daemon_bpm_controlled_process_total",
			Help: "Total count of bpm controlled processes",
		}),
		bpmControlledProcesses: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "chaos_daemon_bpm_controlled_processes",
			Help: "Current number of bpm controlled processes",
		}),
	}
}

func (collector *ChaosDaemonMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	collector.bpmControlledProcessTotal.Describe(ch)
	collector.bpmControlledProcesses.Describe(ch)
}

func (collector *ChaosDaemonMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	collector.collectBpmControlledProcesses()
	collector.bpmControlledProcessTotal.Collect(ch)
	collector.bpmControlledProcesses.Collect(ch)
}

func (collector *ChaosDaemonMetricsCollector) IncreaseControlledProcesses() {
	collector.bpmControlledProcessTotal.Inc()
}

func (collector *ChaosDaemonMetricsCollector) collectBpmControlledProcesses() {
	ids, err := collector.backgroundProcessManager.GetControlledProcessIDs()
	if err != nil {
		return
	}

	collector.bpmControlledProcesses.Set(float64(len(ids)))
}
