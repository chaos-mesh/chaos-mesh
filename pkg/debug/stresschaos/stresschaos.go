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
	cm "github.com/chaos-mesh/chaos-mesh/pkg/debug/common"
)

func Debug(ctx context.Context, chaos runtime.Object, c *cm.ClientSet) error {
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

	for i := range pods {
		podName := pods[i].GetObjectMeta().GetName()
		cm.Print("[Pod]: "+podName, 0, cm.ColorBlue)
		err := debugEachPod(ctx, pods[i], daemons[i], stressChaos, c)
		if err != nil {
			return fmt.Errorf("for %s: %s", podName, err.Error())
		}
	}
	return nil
}

func debugEachPod(ctx context.Context, pod v1.Pod, daemon v1.Pod, chaos *v1alpha1.StressChaos, c *cm.ClientSet) error {
	podName := pod.GetObjectMeta().GetName()
	podNamespace := pod.GetObjectMeta().GetNamespace()
	daemonName := daemon.GetObjectMeta().GetName()
	daemonNamespace := daemon.GetObjectMeta().GetNamespace()

	// cpu or memory chaos
	isCPU := true
	if cpuSpec := chaos.Spec.Stressors.CPUStressor; cpuSpec == nil {
		isCPU = false
	}

	// get process path
	cmd := fmt.Sprintf("cat /proc/cgroups")
	out, err := cm.Exec(podName, podNamespace, cmd, c.KubeCli)
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
	out, err = cm.Exec(podName, podNamespace, cmd, c.KubeCli)
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	stressngLine := regexp.MustCompile("(.*)(stress-ng)").FindStringSubmatch(string(out))
	if len(stressngLine) == 0 {
		return fmt.Errorf("Could not find stress-ng, StressChaos failed")
	}
	stressngPid := strings.Fields(stressngLine[0])[0]

	cmd = fmt.Sprintf("cat /proc/1/cgroup")
	out, err = cm.Exec(podName, podNamespace, cmd, c.KubeCli)
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	cm.Print("1. [cat /proc/1/cgroup]", 1, cm.ColorCyan)
	cm.Print(string(out), 1, "")

	var expr string
	if isCPU {
		expr = "(?::" + cpuMountType + ":)(.*)"
	} else {
		expr = "(?::memory:)(.*)"
	}
	processPath := regexp.MustCompile(expr).FindStringSubmatch(string(out))[1]

	cmd = fmt.Sprintf("cat /proc/%s/cgroup", stressngPid)
	outStress, err := cm.Exec(podName, podNamespace, cmd, c.KubeCli)
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	cm.Print("2. [cat /proc/(stress-ng pid)/cgroup]", 1, cm.ColorCyan)
	cm.Print(string(outStress), 1, "")

	if string(out) != string(outStress) {
		cm.Print("StressChaos failed to execute as expected", 1, cm.ColorRed)
		return nil
	}
	cm.Print("cgroup is the same", 1, cm.ColorGreen)

	// print out debug info
	if isCPU {
		cmd = fmt.Sprintf("cat /sys/fs/cgroup/%s/%s/cpu.cfs_quota_us", cpuMountType, processPath)
		out, err = cm.Exec(daemonName, daemonNamespace, cmd, c.KubeCli)
		if err != nil {
			return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
		}
		cm.Print("3. [cpu.cfs_quota_us]", 1, cm.ColorCyan)
		cm.Print(string(out), 1, "")
		quota, err := strconv.Atoi(strings.TrimSuffix(string(out), "\n"))
		if err != nil {
			return fmt.Errorf("could not get cpu.cfs_quota_us with: %s", err.Error())
		}

		cmd = fmt.Sprintf("cat /sys/fs/cgroup/%s/%s/cpu.cfs_period_us", cpuMountType, processPath)
		out, err = cm.Exec(daemonName, daemonNamespace, cmd, c.KubeCli)
		if err != nil {
			return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
		}
		cm.Print("4. [cpu.cfs_period_us]", 1, cm.ColorCyan)
		cm.Print(string(out), 1, "")
		period, err := strconv.Atoi(strings.TrimSuffix(string(out), "\n"))
		if err != nil {
			return fmt.Errorf("could not get cpu.cfs_period_us with: %s", err.Error())
		}

		if quota == -1 {
			cm.Print("no cpu limit is set for now", 1, cm.ColorRed)
		} else {
			cpuLimitStr := fmt.Sprintf("cpu limit is equals to %.2f", float64(quota)/float64(period))
			cm.Print(cpuLimitStr, 1, cm.ColorGreen)
		}
	} else {
		cmd = fmt.Sprintf("cat /sys/fs/cgroup/memory/%s/memory.limit_in_bytes", processPath)
		out, err = cm.Exec(daemonName, daemonNamespace, cmd, c.KubeCli)
		if err != nil {
			return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
		}
		limit, err := strconv.ParseUint(strings.TrimSuffix(string(out), "\n"), 10, 64)
		if err != nil {
			return fmt.Errorf("could not get memory.limit_in_bytes with: %s", err.Error())
		}
		cm.Print("3. [memory.limit_in_bytes]", 1, cm.ColorCyan)
		cm.Print(bytefmt.ByteSize(limit)+"B", 1, "")
	}
	return nil
}
