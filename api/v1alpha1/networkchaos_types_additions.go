// Copyright 2024 Chaos Mesh Authors.
// Licensed under the Apache License, Version 2.0

package v1alpha1

const (
	CrossRegionLatencyAction NetworkChaosAction = "cross-region-latency"
	RedisClusterFailureAction NetworkChaosAction = "redis-cluster-failure"
)

type CrossRegionLatencySpec struct {
	Profiles []RegionLatencyProfile `json:"profiles"`
}

type RegionLatencyProfile struct {
	RegionName string `json:"regionName"`
	CIDRs []string `json:"cidrs"`
	Latency string `json:"latency"`
	Jitter string `json:"jitter,omitempty"`
	Correlation string `json:"correlation,omitempty"`
	BandwidthRate string `json:"bandwidthRate,omitempty"`
	Loss string `json:"loss,omitempty"`
}

type RedisClusterFailureSpec struct {
	Mode RedisFailureMode `json:"mode"`
	RedisPort uint32 `json:"redisPort,omitempty"`
	ClusterBusPort uint32 `json:"clusterBusPort,omitempty"`
	Latency *DelaySpec `json:"latency,omitempty"`
	Bandwidth *BandwidthSpec `json:"bandwidth,omitempty"`
	ExternalRedisAddresses []string `json:"externalRedisAddresses,omitempty"`
}

type RedisFailureMode string

const (
	RedisPartitionMode RedisFailureMode = "partition"
	RedisLatencyMode RedisFailureMode = "latency"
	RedisBandwidthMode RedisFailureMode = "bandwidth"
)
