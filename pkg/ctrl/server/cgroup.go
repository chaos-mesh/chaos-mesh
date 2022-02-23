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

package server

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	"github.com/chaos-mesh/chaos-mesh/pkg/ctrl/server/model"
)

// GetCgroups returns result of cat /proc/cgroups
func (r *Resolver) GetCgroups(ctx context.Context, obj *model.PodStressChaos) (*model.Cgroups, error) {
	cmd := "cat /proc/cgroups"
	raw, err := r.ExecBypass(ctx, obj.Pod, cmd, bpm.PidNS, bpm.MountNS)
	if err != nil {
		return nil, err
	}

	cgroups := &model.Cgroups{
		Raw: raw,
	}

	// no more info for StressngStressors
	if obj.StressChaos.Spec.StressngStressors != "" || obj.StressChaos.Spec.Stressors == nil {
		return cgroups, nil
	}

	isCPU := true
	if obj.StressChaos.Spec.Stressors.CPUStressor == nil {
		isCPU = false
	}

	if isCPU {
		var cpuMountType string
		if regexp.MustCompile("(cpu,cpuacct)").MatchString(string(raw)) {
			cpuMountType = "cpu,cpuacct"
		} else {
			cpuMountType = "cpu"
		}
		cgroups.CPU = &model.CgroupsCPU{}
		cgroups.CPU.Quota, err = r.GetCPUQuota(ctx, obj.Pod, cpuMountType)
		if err != nil {
			return nil, err
		}
		cgroups.CPU.Period, err = r.GetCPUPeriod(ctx, obj.Pod, cpuMountType)
		if err != nil {
			return nil, err
		}
	} else {
		cgroups.Memory = &model.CgroupsMemory{}
		cgroups.Memory.Limit, err = r.GetMemoryLimit(ctx, obj.Pod)
		if err != nil {
			return nil, err
		}
	}

	return cgroups, nil
}

// GetCgroup returns result of cat /proc/:pid/cgroup
func (r *Resolver) GetCgroup(ctx context.Context, obj *v1.Pod, pid string) (string, error) {
	cmd := fmt.Sprintf("cat /proc/%s/cgroup", pid)
	return r.ExecBypass(ctx, obj, cmd, bpm.PidNS, bpm.MountNS)
}

// GetCPUQuota returns result of cat cat /sys/fs/cgroup/:cpuMountType/cpu.cfs_quota_us
func (r *Resolver) GetCPUQuota(ctx context.Context, obj *v1.Pod, cpuMountType string) (int, error) {
	cmd := fmt.Sprintf("cat /sys/fs/cgroup/%s/cpu.cfs_quota_us", cpuMountType)
	out, err := r.ExecBypass(ctx, obj, cmd, bpm.PidNS, bpm.MountNS)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSuffix(string(out), "\n"))
}

// GetCPUPeriod returns result of cat cat /sys/fs/cgroup/:cpuMountType/cpu.cfs_period_us
func (r *Resolver) GetCPUPeriod(ctx context.Context, obj *v1.Pod, cpuMountType string) (int, error) {
	cmd := fmt.Sprintf("cat /sys/fs/cgroup/%s/cpu.cfs_period_us", cpuMountType)
	out, err := r.ExecBypass(ctx, obj, cmd, bpm.PidNS, bpm.MountNS)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSuffix(string(out), "\n"))
}

// GetMemoryLimit returns result of cat cat /sys/fs/cgroup/memory/memory.limit_in_bytes
func (r *Resolver) GetMemoryLimit(ctx context.Context, obj *v1.Pod) (int64, error) {
	cmd := "cat /sys/fs/cgroup/memory/memory.limit_in_bytes"
	rawLimit, err := r.ExecBypass(ctx, obj, cmd, bpm.PidNS, bpm.MountNS)
	if err != nil {
		return 0, errors.Wrap(err, "could not get memory.limit_in_bytes")
	}
	limit, err := strconv.ParseUint(strings.TrimSuffix(rawLimit, "\n"), 10, 64)
	if err != nil {
		return 0, errors.Wrap(err, "could not parse memory.limit_in_bytes")
	}
	return int64(limit), nil
}
