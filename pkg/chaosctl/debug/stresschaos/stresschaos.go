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
	chaosSelector := stressChaos.Spec.GetSelector()

	pods, daemons, err := cm.GetPods(ctx, chaosStatus, chaosSelector, c.CtrlCli)
	if err != nil {
		return err
	}

	if err := cm.CheckFailedMessage(ctx, chaosStatus.FailedMessage, daemons, c); err != nil {
		return err
	}

	for i := range pods {
		podName := pods[i].Name
		podResult := cm.PodResult{Name: podName}
		err := debugEachPod(ctx, pods[i], daemons[i], stressChaos, c, &podResult)
		result.Pods = append(result.Pods, podResult)
		if err != nil {
			return fmt.Errorf("for %s: %s", podName, err.Error())
		}
	}
	return nil
}

func debugEachPod(ctx context.Context, pod v1.Pod, daemon v1.Pod, chaos *v1alpha1.StressChaos, c *cm.ClientSet, result *cm.PodResult) error {
	// cpu or memory chaos
	isCPU := true
	if cpuSpec := chaos.Spec.Stressors.CPUStressor; cpuSpec == nil {
		isCPU = false
	}

	// get process path
	cmd := fmt.Sprintf("cat /proc/cgroups")
	out, err := cm.ExecBypass(ctx, pod, daemon, cmd, c.KubeCli)
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	var cpuMountType string
	if regexp.MustCompile("(cpu,cpuacct)").MatchString(string(out)) {
		cpuMountType = "cpu,cpuacct"
	} else {
		cpuMountType = "cpu"
	}

	cmd = fmt.Sprintf("ps")
	out, err = cm.ExecBypass(ctx, pod, daemon, cmd, c.KubeCli)
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	result.Items = append(result.Items, cm.ItemResult{Name: "ps", Value: string(out)})
	stressngLine := regexp.MustCompile("(.*)(stress-ng)").FindStringSubmatch(string(out))
	if len(stressngLine) == 0 {
		return fmt.Errorf("Could not find stress-ng, StressChaos failed")
	}
	stressngPid := strings.Fields(stressngLine[0])[0]

	cmd = fmt.Sprintf("cat /proc/1/cgroup")
	out, err = cm.ExecBypass(ctx, pod, daemon, cmd, c.KubeCli)
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	result.Items = append(result.Items, cm.ItemResult{Name: "/proc/1/cgroup", Value: string(out)})

	var expr string
	if isCPU {
		expr = "(?::" + cpuMountType + ":)(.*)"
	} else {
		expr = "(?::memory:)(.*)"
	}
	processPath := regexp.MustCompile(expr).FindStringSubmatch(string(out))[1]

	cmd = fmt.Sprintf("cat /proc/%s/cgroup", stressngPid)
	outStress, err := cm.ExecBypass(ctx, pod, daemon, cmd, c.KubeCli)
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	itemResult := cm.ItemResult{Name: "/proc/(stress-ng pid)/cgroup", Value: string(outStress)}

	if string(out) != string(outStress) {
		itemResult.Status = cm.ItemFailure
		itemResult.ErrInfo = "Cgroup of stress-ng and init process not the same"
	} else {
		itemResult.Status = cm.ItemSuccess
	}
	result.Items = append(result.Items, itemResult)

	// print out debug info
	if isCPU {
		cmd = fmt.Sprintf("cat /sys/fs/cgroup/%s/%s/cpu.cfs_quota_us", cpuMountType, processPath)
		out, err = cm.Exec(ctx, daemon, cmd, c.KubeCli)
		if err != nil {
			return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
		}
		result.Items = append(result.Items, cm.ItemResult{Name: "cpu.cfs_quota_us", Value: string(out)})
		quota, err := strconv.Atoi(strings.TrimSuffix(string(out), "\n"))
		if err != nil {
			return fmt.Errorf("could not get cpu.cfs_quota_us with: %s", err.Error())
		}

		cmd = fmt.Sprintf("cat /sys/fs/cgroup/%s/%s/cpu.cfs_period_us", cpuMountType, processPath)
		out, err = cm.Exec(ctx, daemon, cmd, c.KubeCli)
		if err != nil {
			return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
		}
		period, err := strconv.Atoi(strings.TrimSuffix(string(out), "\n"))
		if err != nil {
			return fmt.Errorf("could not get cpu.cfs_period_us with: %s", err.Error())
		}
		itemResult = cm.ItemResult{Name: "cpu.cfs_period_us", Value: string(out)}

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
			return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
		}
		limit, err := strconv.ParseUint(strings.TrimSuffix(string(out), "\n"), 10, 64)
		if err != nil {
			return fmt.Errorf("could not get memory.limit_in_bytes with: %s", err.Error())
		}
		result.Items = append(result.Items, cm.ItemResult{Name: "memory.limit_in_bytes", Value: bytefmt.ByteSize(limit) + "B"})
	}
	return nil
}
