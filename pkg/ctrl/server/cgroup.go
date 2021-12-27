package server

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"

	"github.com/chaos-mesh/chaos-mesh/pkg/ctrl/server/model"
)

// GetCgroups returns result of cat /proc/cgroups
func (r *Resolver) GetCgroups(ctx context.Context, obj *model.PodStressChaos) (*model.Cgroups, error) {
	cmd := "cat /proc/cgroups"
	raw, err := r.ExecBypass(ctx, obj.Pod, cmd)
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
	return r.ExecBypass(ctx, obj, cmd)
}

// GetCPUQuota returns result of cat cat /sys/fs/cgroup/:cpuMountType/cpu.cfs_quota_us
func (r *Resolver) GetCPUQuota(ctx context.Context, obj *v1.Pod, cpuMountType string) (int, error) {
	cmd := fmt.Sprintf("cat /sys/fs/cgroup/%s/cpu.cfs_quota_us", cpuMountType)
	out, err := r.ExecBypass(ctx, obj, cmd)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSuffix(string(out), "\n"))
}

// GetCPUPeriod returns result of cat cat /sys/fs/cgroup/:cpuMountType/cpu.cfs_period_us
func (r *Resolver) GetCPUPeriod(ctx context.Context, obj *v1.Pod, cpuMountType string) (int, error) {
	cmd := fmt.Sprintf("cat /sys/fs/cgroup/%s/cpu.cfs_period_us", cpuMountType)
	out, err := r.ExecBypass(ctx, obj, cmd)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSuffix(string(out), "\n"))
}

// GetMemoryLimit returns result of cat cat /sys/fs/cgroup/memory/memory.limit_in_bytes
func (r *Resolver) GetMemoryLimit(ctx context.Context, obj *v1.Pod) (string, error) {
	cmd := "cat /sys/fs/cgroup/memory/memory.limit_in_bytes"
	return r.ExecBypass(ctx, obj, cmd)
}
