// Copyright 2024 Chaos Mesh Authors.
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

package utils

import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	v1 "k8s.io/cri-api/pkg/apis/runtime/v1"
	"strings"
)

// BuildRuntimeServiceClient creates a new RuntimeServiceClient from the given endpoint
func BuildRuntimeServiceClient(endpoint string) (v1.RuntimeServiceClient, error) {
	addr := endpoint
	if !strings.HasPrefix(addr, "unix://") {
		addr = fmt.Sprintf("unix://%s", addr)
	}
	conn, err := grpc.Dial(addr,
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := v1.NewRuntimeServiceClient(conn)
	return client, err
}

// BuildContainerStatsFromCRIResponse creates a new ContainerStats from the response of runtimeClient.ContainerStats
func BuildContainerStatsFromCRIResponse(resp *v1.ContainerStatsResponse) *ContainerStats {
	result := &ContainerStats{}

	stats := resp.Stats
	if stats == nil {
		return result
	}

	cpu := stats.Cpu
	if cpu != nil {
		result.Cpu.UsageCoreNanoSeconds = cpu.UsageCoreNanoSeconds.GetValue()
	}

	memory := stats.Memory
	if memory != nil {
		result.Memory.WorkingSetBytes = memory.WorkingSetBytes.GetValue()
		result.Memory.AvailableBytes = memory.AvailableBytes.GetValue()
		result.Memory.UsageBytes = memory.UsageBytes.GetValue()
		result.Memory.RssBytes = memory.RssBytes.GetValue()
		result.Memory.PageFaults = memory.PageFaults.GetValue()
		result.Memory.MajorPageFaults = memory.MajorPageFaults.GetValue()
	}

	swap := stats.Swap
	if swap != nil {
		result.Swap.AvailableBytes = swap.SwapAvailableBytes.GetValue()
		result.Swap.UsageBytes = swap.SwapUsageBytes.GetValue()
	}

	return result
}
