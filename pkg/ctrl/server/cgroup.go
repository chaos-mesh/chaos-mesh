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

// IsCgroupV2 detects if the system is using cgroup v2
func (r *Resolver) IsCgroupV2(ctx context.Context, obj *v1.Pod) (bool, error) {
	// Check if the unified cgroup hierarchy exists by testing for cgroup.controllers file
	cmd := "test -f /sys/fs/cgroup/cgroup.controllers && echo true || echo false"
	out, err := r.ExecBypass(ctx, obj, cmd, bpm.PidNS, bpm.MountNS)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) == "true", nil
}

func (r *Resolver) GetCgroup(ctx context.Context, obj *v1.Pod, pid string) (string, error) {
	cmd := fmt.Sprintf("cat /proc/%s/cgroup", pid)
	return r.ExecBypass(ctx, obj, cmd, bpm.PidNS, bpm.MountNS)
}

// GetCgroups also needs to be updated to handle cgroup v2
func (r *Resolver) GetCgroups(ctx context.Context, obj *model.PodStressChaos) (*model.Cgroups, error) {
	// Check cgroup version first
	isV2, err := r.IsCgroupV2(ctx, obj.Pod)
	if err != nil {
		return nil, err
	}

	var cmd string
	if isV2 {
		// In cgroup v2, controllers are listed in cgroup.controllers
		cmd = "cat /sys/fs/cgroup/cgroup.controllers"
	} else {
		// Original cgroup v1 command
		cmd = "cat /proc/cgroups"
	}

	raw, err := r.ExecBypass(ctx, obj.Pod, cmd, bpm.PidNS, bpm.MountNS)
	if err != nil {
		return nil, err
	}

	cgroups := &model.Cgroups{
		Raw: raw,
	}

	if obj.StressChaos.Spec.StressngStressors != "" || obj.StressChaos.Spec.Stressors == nil {
		return cgroups, nil
	}

	isCPU := true
	if obj.StressChaos.Spec.Stressors.CPUStressor == nil {
		isCPU = false
	}

	if isCPU {
		cgroups.CPU = &model.CgroupsCPU{}

		var cpuMountType string
		if !isV2 {
			if regexp.MustCompile("(cpu,cpuacct)").MatchString(string(raw)) {
				cpuMountType = "cpu,cpuacct"
			} else {
				// cgroup does not support cpuacct sub-system
				cpuMountType = "cpu"
			}
		}

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

// GetCPUQuota returns CPU quota based on cgroup version
func (r *Resolver) GetCPUQuota(ctx context.Context, obj *v1.Pod, cpuMountType string) (int, error) {
	isV2, err := r.IsCgroupV2(ctx, obj)
	if err != nil {
		return 0, err
	}

	if isV2 {
		// In cgroup v2, quota and period are in the same file (cpu.max)
		cmd := "cat /sys/fs/cgroup/cpu.max"
		out, err := r.ExecBypass(ctx, obj, cmd, bpm.PidNS, bpm.MountNS)
		if err != nil {
			return 0, err
		}
		// Format is "quota period"
		parts := strings.Fields(out)
		if len(parts) < 1 {
			return 0, fmt.Errorf("unexpected format in cpu.max: %s", out)
		}

		// Handle "max" value which means no limit
		if parts[0] == "max" {
			return -1, nil
		}

		return strconv.Atoi(parts[0])
	}

	// Original cgroup v1 code
	cmd := fmt.Sprintf("cat /sys/fs/cgroup/%s/cpu.cfs_quota_us", cpuMountType)
	out, err := r.ExecBypass(ctx, obj, cmd, bpm.PidNS, bpm.MountNS)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSuffix(string(out), "\n"))
}

// GetCPUPeriod returns CPU period based on cgroup version
func (r *Resolver) GetCPUPeriod(ctx context.Context, obj *v1.Pod, cpuMountType string) (int, error) {
	isV2, err := r.IsCgroupV2(ctx, obj)
	if err != nil {
		return 0, err
	}

	if isV2 {
		// In cgroup v2, quota and period are in the same file (cpu.max)
		cmd := "cat /sys/fs/cgroup/cpu.max"
		out, err := r.ExecBypass(ctx, obj, cmd, bpm.PidNS, bpm.MountNS)
		if err != nil {
			return 0, err
		}
		// Format is "quota period"
		parts := strings.Fields(out)
		if len(parts) < 2 {
			return 0, fmt.Errorf("unexpected format in cpu.max: %s", out)
		}
		return strconv.Atoi(parts[1])
	}

	// Original cgroup v1 code
	cmd := fmt.Sprintf("cat /sys/fs/cgroup/%s/cpu.cfs_period_us", cpuMountType)
	out, err := r.ExecBypass(ctx, obj, cmd, bpm.PidNS, bpm.MountNS)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSuffix(string(out), "\n"))
}

// GetMemoryLimit returns memory limit based on cgroup version
func (r *Resolver) GetMemoryLimit(ctx context.Context, obj *v1.Pod) (int64, error) {
	isV2, err := r.IsCgroupV2(ctx, obj)
	if err != nil {
		return 0, err
	}

	var cmd string
	if isV2 {
		cmd = "cat /sys/fs/cgroup/memory.max"
	} else {
		cmd = "cat /sys/fs/cgroup/memory/memory.limit_in_bytes"
	}

	rawLimit, err := r.ExecBypass(ctx, obj, cmd, bpm.PidNS, bpm.MountNS)
	if err != nil {
		return 0, errors.Wrap(err, "could not get memory limit")
	}

	// Handle "max" value in cgroup v2
	if strings.TrimSpace(rawLimit) == "max" {
		// Return -1 to indicate unlimited memory
		return -1, nil
	}

	limit, err := strconv.ParseUint(strings.TrimSpace(rawLimit), 10, 64)
	if err != nil {
		return 0, errors.Wrap(err, "could not parse memory limit")
	}
	return int64(limit), nil
}
