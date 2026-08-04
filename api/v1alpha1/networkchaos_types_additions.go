// Copyright 2024 Chaos Mesh Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

const (
	CrossRegionLatencyAction NetworkChaosAction = "cross-region-latency"
	RedisClusterFailureAction NetworkChaosAction = "redis-cluster-failure"
)

type CrossRegionLatencySpec struct {
	Profiles []RegionLatencyProfile `json:"profiles"`
}

type RegionLatencyProfile struct {
	RegionName    string   `json:"regionName"`
	CIDRs         []string `json:"cidrs"`
	Latency       string   `json:"latency" webhook:"Duration"`
	Jitter        string   `json:"jitter,omitempty" webhook:"Duration"`
	Correlation   string   `json:"correlation,omitempty"`
	BandwidthRate string   `json:"bandwidthRate,omitempty"`
	Loss          string   `json:"loss,omitempty"`
}

type RedisClusterFailureSpec struct {
	Mode                   RedisFailureMode `json:"mode"`
	RedisPort              uint32           `json:"redisPort,omitempty"`
	ClusterBusPort         uint32           `json:"clusterBusPort,omitempty"`
	Latency                *DelaySpec       `json:"latency,omitempty"`
	Bandwidth              *BandwidthSpec   `json:"bandwidth,omitempty"`
	ExternalRedisAddresses []string         `json:"externalRedisAddresses,omitempty"`
}

type RedisFailureMode string

const (
	RedisPartitionMode  RedisFailureMode = "partition"
	RedisLatencyMode    RedisFailureMode = "latency"
	RedisBandwidthMode  RedisFailureMode = "bandwidth"
)
