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

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/kubernetes/pkg/kubelet/types"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients"
	"github.com/chaos-mesh/chaos-mesh/pkg/metrics/utils"
)

// DefaultChaosDaemonMetricsCollector is the default metrics collector for chaos daemon
var DefaultChaosDaemonMetricsCollector = NewChaosDaemonMetricsCollector()

type ChaosDaemonMetricsCollector struct {
	crClient crclients.ContainerRuntimeInfoClient

	iptablesChains      *prometheus.GaugeVec
	iptablesRules       *prometheus.GaugeVec
	iptablesPackets     *prometheus.GaugeVec
	iptablesPacketBytes *prometheus.GaugeVec
	ipsetMembers        *prometheus.GaugeVec
	tcRules             *prometheus.GaugeVec
}

// NewChaosDaemonMetricsCollector initializes metrics for each chaos daemon
func NewChaosDaemonMetricsCollector() *ChaosDaemonMetricsCollector {
	return &ChaosDaemonMetricsCollector{
		iptablesChains: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "chaos_daemon_iptables_chains",
			Help: "Total number of iptables chains",
		}, []string{"namespace", "pod_name", "container_name"}),
		iptablesRules: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "chaos_daemon_iptables_rules",
			Help: "Total number of iptables rules",
		}, []string{"namespace", "pod_name", "container_name"}),
		iptablesPackets: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "chaos_daemon_iptables_packets",
			Help: "Total number of iptables packets",
		}, []string{"namespace", "pod_name", "container_name"}),
		iptablesPacketBytes: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "chaos_daemon_iptables_packet_bytes",
			Help: "Total bytes of iptables packets",
		}, []string{"namespace", "pod_name", "container_name"}),
		ipsetMembers: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "chaos_daemon_ipset_members",
			Help: "Total number of ipset members",
		}, []string{"namespace", "pod_name", "container_name"}),
		tcRules: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "chaos_daemon_tcs",
			Help: "Total number of tc rules",
		}, []string{"namespace", "pod_name", "container_name"}),
	}
}

func (collector *ChaosDaemonMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	collector.iptablesChains.Describe(ch)
	collector.iptablesRules.Describe(ch)
	collector.iptablesPackets.Describe(ch)
	collector.iptablesPacketBytes.Describe(ch)
	collector.ipsetMembers.Describe(ch)
	collector.tcRules.Describe(ch)
}

func (collector *ChaosDaemonMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	collector.collectNetworkMetrics()
	collector.iptablesChains.Collect(ch)
	collector.iptablesRules.Collect(ch)
	collector.iptablesPackets.Collect(ch)
	collector.iptablesPacketBytes.Collect(ch)
	collector.ipsetMembers.Collect(ch)
	collector.tcRules.Collect(ch)
}

func (collector *ChaosDaemonMetricsCollector) InjectCrClient(client crclients.ContainerRuntimeInfoClient) *ChaosDaemonMetricsCollector {
	collector.crClient = client
	return collector
}

func (collector *ChaosDaemonMetricsCollector) collectNetworkMetrics() {
	collector.iptablesChains.Reset()

	containerIDs, err := collector.crClient.ListContainerIDs(context.Background())
	if err != nil {
		log.Error(err, "fail to list all container process IDs")
		return
	}

	for _, containerID := range containerIDs {
		pid, err := collector.crClient.GetPidFromContainerID(context.Background(), containerID)
		if err != nil {
			log.Error(err, "fail to get pid from container ID")
			continue
		}

		labels, err := collector.crClient.GetLabelsFromContainerID(context.Background(), containerID)
		if err != nil {
			log.Error(err, "fail to get container labels", "containerID", containerID)
			continue
		}

		namespace, podName, containerName := labels[types.KubernetesPodNamespaceLabel],
			labels[types.KubernetesPodNameLabel], labels[types.KubernetesContainerNameLabel]

		labelValues := []string{namespace, podName, containerName}
		log := log.WithValues(
			"namespace", namespace,
			"podName", podName,
			"containerName", containerName,
			"containerID", containerID,
		)
		log.Info("collecting metrics", "pid", pid, "labels", labels)

		chains, rules, packets, packetBytes, err := utils.CollectIptablesMetrics(pid)
		if err != nil {
			log.Error(err, "fail to collect iptables metrics")
		}
		collector.iptablesChains.WithLabelValues(labelValues...).Set(float64(chains))
		collector.iptablesRules.WithLabelValues(labelValues...).Set(float64(rules))
		collector.iptablesPackets.WithLabelValues(labelValues...).Set(float64(packets))
		collector.iptablesPacketBytes.WithLabelValues(labelValues...).Set(float64(packetBytes))

		members, err := utils.CollectIPSetMembersMetric(pid)
		if err != nil {
			log.Error(err, "fail to collect ipset member metric")
		}
		collector.ipsetMembers.WithLabelValues(labelValues...).Set(float64(members))

		tcRules, err := utils.CollectTcRulesMetric(pid)
		if err != nil {
			log.Error(err, "fail to collect tc rules metric")
		}
		collector.tcRules.WithLabelValues(labelValues...).Set(float64(tcRules))
	}
}
