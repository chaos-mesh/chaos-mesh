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
	grpcprometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients"
	"github.com/chaos-mesh/chaos-mesh/pkg/log"
	"github.com/chaos-mesh/chaos-mesh/pkg/metrics/utils"
)

var (
	// DefaultChaosDaemonMetricsCollector is the default metrics collector for chaos daemon
	DefaultChaosDaemonMetricsCollector = NewChaosDaemonMetricsCollector(log.L().WithName("chaos-daemon").WithName("metrics"))

	// ChaosDaemonGrpcServerBuckets is the buckets for gRPC server handling histogram metrics
	ChaosDaemonGrpcServerBuckets = []float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 10}
)

const (
	// kubernetesPodNameLabel, kubernetesPodNamespaceLabel and kubernetesContainerNameLabel are the label keys
	//   indicating the kubernetes information of the container under `k8s.io/kubernetes` package
	// And it is best not to set `k8s.io/kubernetes` as dependency, see more: https://github.com/kubernetes/kubernetes/issues/90358#issuecomment-617859364.
	kubernetesPodNameLabel       = "io.kubernetes.pod.name"
	kubernetesPodNamespaceLabel  = "io.kubernetes.pod.namespace"
	kubernetesContainerNameLabel = "io.kubernetes.container.name"
)

func WithHistogramName(name string) grpcprometheus.HistogramOption {
	return func(opts *prometheus.HistogramOpts) {
		opts.Name = name
	}
}

type ChaosDaemonMetricsCollector struct {
	crClient            crclients.ContainerRuntimeInfoClient
	logger              logr.Logger
	iptablesPackets     *prometheus.GaugeVec
	iptablesPacketBytes *prometheus.GaugeVec
	ipsetMembers        *prometheus.GaugeVec
	tcRules             *prometheus.GaugeVec
}

// NewChaosDaemonMetricsCollector initializes metrics for each chaos daemon
func NewChaosDaemonMetricsCollector(logger logr.Logger) *ChaosDaemonMetricsCollector {
	return &ChaosDaemonMetricsCollector{
		logger: logger,
		iptablesPackets: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "chaos_daemon_iptables_packets",
			Help: "Total number of iptables packets",
		}, []string{"namespace", "pod", "container", "table", "chain", "policy", "rule"}),
		iptablesPacketBytes: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "chaos_daemon_iptables_packet_bytes",
			Help: "Total bytes of iptables packets",
		}, []string{"namespace", "pod", "container", "table", "chain", "policy", "rule"}),
		ipsetMembers: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "chaos_daemon_ipset_members",
			Help: "Total number of ipset members",
		}, []string{"namespace", "pod", "container"}),
		tcRules: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "chaos_daemon_tcs",
			Help: "Total number of tc rules",
		}, []string{"namespace", "pod", "container"}),
	}
}

func (collector *ChaosDaemonMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	collector.iptablesPackets.Describe(ch)
	collector.iptablesPacketBytes.Describe(ch)
	collector.ipsetMembers.Describe(ch)
	collector.tcRules.Describe(ch)
}

func (collector *ChaosDaemonMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	collector.collectNetworkMetrics()
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
	collector.iptablesPackets.Reset()
	collector.iptablesPacketBytes.Reset()
	collector.ipsetMembers.Reset()
	collector.tcRules.Reset()

	containerIDs, err := collector.crClient.ListContainerIDs(context.Background())
	if err != nil {
		collector.logger.Error(err, "fail to list all container process IDs")
		return
	}

	for _, containerID := range containerIDs {
		pid, err := collector.crClient.GetPidFromContainerID(context.Background(), containerID)
		if err != nil {
			collector.logger.Error(err, "fail to get pid from container ID")
			continue
		}

		labels, err := collector.crClient.GetLabelsFromContainerID(context.Background(), containerID)
		if err != nil {
			collector.logger.Error(err, "fail to get container labels", "containerID", containerID)
			continue
		}

		namespace, podName, containerName := labels[kubernetesPodNamespaceLabel],
			labels[kubernetesPodNameLabel], labels[kubernetesContainerNameLabel]

		labelValues := []string{namespace, podName, containerName}
		log := collector.logger.WithValues(
			"namespace", namespace,
			"podName", podName,
			"containerName", containerName,
			"containerID", containerID,
		)

		tables, err := utils.GetIptablesContentByNetNS(pid)
		if err != nil {
			log.Error(err, "fail to collect iptables metrics")
		}
		for tableName, table := range tables {
			for chainName, chain := range table {
				for _, rule := range chain.Rules {
					collector.iptablesPackets.
						WithLabelValues(namespace, podName, containerName, tableName, chainName, chain.Policy, rule.Rule).
						Set(float64(rule.Packets))

					collector.iptablesPacketBytes.
						WithLabelValues(namespace, podName, containerName, tableName, chainName, chain.Policy, rule.Rule).
						Set(float64(rule.Bytes))
				}
			}
		}

		members, err := utils.GetIPSetRulesNumberByNetNS(pid)
		if err != nil {
			log.Error(err, "fail to collect ipset member metric")
		}
		collector.ipsetMembers.WithLabelValues(labelValues...).Set(float64(members))

		tcRules, err := utils.GetTcRulesNumberByNetNS(pid)
		if err != nil {
			log.Error(err, "fail to collect tc rules metric")
		}
		collector.tcRules.WithLabelValues(labelValues...).Set(float64(tcRules))
	}
}
