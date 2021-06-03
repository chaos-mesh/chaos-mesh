// Copyright 2019 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package stresschaos

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"code.cloudfoundry.org/bytefmt"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	cm "github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/common"
)

// Debug get chaos debug information
func Debug(ctx context.Context, chaos runtime.Object, c *cm.ClientSet, result *cm.ChaosResult) error {
	stressChaos, ok := chaos.(*v1alpha1.StressChaos)
	if !ok {
		return fmt.Errorf("chaos is not stresschaos")
	}
	chaosStatus := stressChaos.Status.ChaosStatus
	chaosSelector := stressChaos.Spec.Selector

	pods, daemons, err := cm.GetPods(ctx, stressChaos.GetName(), chaosStatus, chaosSelector, c.CtrlCli)
	if err != nil {
		return err
	}

	for i := range pods {
		podName := pods[i].Name
		podResult := cm.PodResult{Name: podName}
		_ = debugEachPod(ctx, pods[i], daemons[i], stressChaos, c, &podResult)
		result.Pods = append(result.Pods, podResult)
		// TODO: V(4) log when err != nil, wait for #1433
	}
	return nil
}

func debugEachPod(ctx context.Context, pod v1.Pod, daemon v1.Pod, chaos *v1alpha1.StressChaos, c *cm.ClientSet, result *cm.PodResult) error {
	// get process path
	cmd := "cat /proc/cgroups"
	out, err := cm.ExecBypass(ctx, pod, daemon, cmd, c.KubeCli)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("run command %s failed", cmd))
	}
	result.Items = append(result.Items, cm.ItemResult{Name: "cat /proc/cgroups", Value: string(out)})

	cmd = "ps"
	out, err = cm.ExecBypass(ctx, pod, daemon, cmd, c.KubeCli)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("run command %s failed", cmd))
	}
	result.Items = append(result.Items, cm.ItemResult{Name: "ps", Value: string(out)})
	stressngLine := regexp.MustCompile("(.*)(stress-ng)").FindStringSubmatch(string(out))
	if len(stressngLine) == 0 {
		return fmt.Errorf("could not find stress-ng, StressChaos failed")
	}

	pids, commands, err := cm.GetPidFromPS(ctx, pod, daemon, c.KubeCli)
	if err != nil {
		return errors.Wrap(err, "get pid from ps failed")
	}

	for i := range pids {
		cmd = fmt.Sprintf("cat /proc/%s/cgroup", pids[i])
		out, err = cm.ExecBypass(ctx, pod, daemon, cmd, c.KubeCli)
		if err != nil {
			cm.L().WithName("stress-chaos").V(2).Info("failed to fetch cgroup ofr certain process",
				"pod", fmt.Sprintf("%s/%s", pod.Namespace, pod.Name),
				"pid", i,
			)
			result.Items = append(result.Items, cm.ItemResult{Name: fmt.Sprintf("/proc/%s/cgroup of %s", pids[i], commands[i]), Value: "No cgroup found"})
		} else {
			result.Items = append(result.Items, cm.ItemResult{Name: fmt.Sprintf("/proc/%s/cgroup of %s", pids[i], commands[i]), Value: string(out)})
		}
	}

	// no more info for StressngStressors
	if chaos.Spec.StressngStressors != "" {
		return nil
	}

	isCPU := true
	if cpuSpec := chaos.Spec.Stressors.CPUStressor; cpuSpec == nil {
		isCPU = false
	}

	var expr, cpuMountType string
	if isCPU {
		if regexp.MustCompile("(cpu,cpuacct)").MatchString(string(out)) {
			cpuMountType = "cpu,cpuacct"
		} else {
			cpuMountType = "cpu"
		}
		expr = "(?::" + cpuMountType + ":)(.*)"
	} else {
		expr = "(?::memory:)(.*)"
	}
	processPath := regexp.MustCompile(expr).FindStringSubmatch(string(out))[1]

	// print out debug info
	if isCPU {
		cmd = fmt.Sprintf("cat /sys/fs/cgroup/%s/%s/cpu.cfs_quota_us", cpuMountType, processPath)
		out, err = cm.Exec(ctx, daemon, cmd, c.KubeCli)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("run command %s failed", cmd))
		}
		result.Items = append(result.Items, cm.ItemResult{Name: "cpu.cfs_quota_us", Value: string(out)})
		quota, err := strconv.Atoi(strings.TrimSuffix(string(out), "\n"))
		if err != nil {
			return errors.Wrap(err, "could not get cpu.cfs_quota_us")
		}

		cmd = fmt.Sprintf("cat /sys/fs/cgroup/%s/%s/cpu.cfs_period_us", cpuMountType, processPath)
		out, err = cm.Exec(ctx, daemon, cmd, c.KubeCli)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("run command %s failed", cmd))
		}
		period, err := strconv.Atoi(strings.TrimSuffix(string(out), "\n"))
		if err != nil {
			return errors.Wrap(err, "could not get cpu.cfs_period_us")
		}
		itemResult := cm.ItemResult{Name: "cpu.cfs_period_us", Value: string(out)}

		if quota == -1 {
			itemResult.Status = cm.ItemFailure
			itemResult.ErrInfo = "no cpu limit is set for now"
		} else {
			itemResult.Status = cm.ItemSuccess
			itemResult.SucInfo = fmt.Sprintf("cpu limit is equals to %.2f", float64(quota)/float64(period))
		}
		result.Items = append(result.Items, itemResult)
	} else {
		cmd = fmt.Sprintf("cat /sys/fs/cgroup/memory/%s/memory.limit_in_bytes", processPath)
		out, err = cm.Exec(ctx, daemon, cmd, c.KubeCli)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("run command %s failed", cmd))
		}
		limit, err := strconv.ParseUint(strings.TrimSuffix(string(out), "\n"), 10, 64)
		if err != nil {
			return errors.Wrap(err, "could not get memory.limit_in_bytes")
		}
		result.Items = append(result.Items, cm.ItemResult{Name: "memory.limit_in_bytes", Value: bytefmt.ByteSize(limit) + "B"})
	}
	return nil
}
