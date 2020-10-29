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

	cm "github.com/chaos-mesh/chaos-mesh/pkg/debug/common"
)

func Debug(ctx context.Context, chaosName string, ns string, c *cm.ClientSet) error {
	p, err := cm.GetPod(ctx, "stresschaos", chaosName, ns, c.CtrlClient)
	if err != nil {
		return err
	}

	// cpu or memory chaos
	chaos, err := cm.GetChaos(ctx, "stresschaos", chaosName, ns, c.CtrlClient)
	if err != nil {
		return fmt.Errorf("failed to get chaos %s: %s", chaosName, err.Error())
	}

	isCPU := true
	cpuHier := []string{"spec", "stressors", "cpu"}
	_, err = cm.ExtractFromJson(chaos, cpuHier)
	if err != nil {
		isCPU = false
	}

	// get process path
	cmd := fmt.Sprintf("cat /proc/cgroups")
	out, err := cm.Exec(p.PodName, p.PodNamespace, cmd, c.K8sClient)
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
	out, err = cm.Exec(p.PodName, p.PodNamespace, cmd, c.K8sClient)
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	stressngLine := regexp.MustCompile("(.*)(stress-ng)").FindStringSubmatch(string(out))
	if len(stressngLine) == 0 {
		return fmt.Errorf("Could not find stress-ng, StressChaos failed")
	}
	stressngPid := strings.Fields(stressngLine[0])[0]

	cmd = fmt.Sprintf("cat /proc/1/cgroup")
	out, err = cm.Exec(p.PodName, p.PodNamespace, cmd, c.K8sClient)
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	fmt.Println(string(cm.ColorCyan), "1. [cat /proc/1/cgroup]:", string(cm.ColorReset))
	cm.PrintWithTab(string(out))

	var expr string
	if isCPU {
		expr = "(?::" + cpuMountType + ":)(.*)"
	} else {
		expr = "(?::memory:)(.*)"
	}
	processPath := regexp.MustCompile(expr).FindStringSubmatch(string(out))[1]

	cmd = fmt.Sprintf("cat /proc/%s/cgroup", stressngPid)
	outStress, err := cm.Exec(p.PodName, p.PodNamespace, cmd, c.K8sClient)
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	fmt.Println(string(cm.ColorCyan), "2. [cat /proc/(stress-ng pid)/cgroup]:", string(cm.ColorReset))
	cm.PrintWithTab(string(outStress))

	if string(out) != string(outStress) {
		errInfo := fmt.Sprintf("%sStressChaos failed to execute as expected%s\n", string(cm.ColorRed), string(cm.ColorReset))
		cm.PrintWithTab(errInfo)
		return nil
	}
	sucInfo := fmt.Sprintf("%scgroup is the same%s\n", string(cm.ColorGreen), string(cm.ColorReset))
	cm.PrintWithTab(sucInfo)

	// print out debug info
	if isCPU {
		cmd = fmt.Sprintf("cat /sys/fs/cgroup/%s/%s/cpu.cfs_quota_us", cpuMountType, processPath)
		out, err = cm.Exec(p.ChaosDaemonName, p.ChaosDaemonNamespace, cmd, c.K8sClient)
		if err != nil {
			return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
		}
		fmt.Println(string(cm.ColorCyan), "3. [cpu.cfs_quota_us]:", string(cm.ColorReset))
		cm.PrintWithTab(string(out))
		quota, err := strconv.Atoi(strings.TrimSuffix(string(out), "\n"))
		if err != nil {
			return fmt.Errorf("could not get cpu.cfs_quota_us with: %s", err.Error())
		}

		cmd = fmt.Sprintf("cat /sys/fs/cgroup/%s/%s/cpu.cfs_period_us", cpuMountType, processPath)
		out, err = cm.Exec(p.ChaosDaemonName, p.ChaosDaemonNamespace, cmd, c.K8sClient)
		if err != nil {
			return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
		}
		fmt.Println(string(cm.ColorCyan), "4. [cpu.cfs_period_us]:", string(cm.ColorReset))
		cm.PrintWithTab(string(out))
		period, err := strconv.Atoi(strings.TrimSuffix(string(out), "\n"))
		if err != nil {
			return fmt.Errorf("could not get cpu.cfs_period_us with: %s", err.Error())
		}

		if quota == -1 {
			fmt.Println(string(cm.ColorRed), "no cpu limit is set for now", string(cm.ColorReset))
		} else {
			fmt.Println(string(cm.ColorGreen), "cpu limit is equals to", float64(quota)/float64(period), string(cm.ColorReset))
		}
	} else {
		cmd = fmt.Sprintf("cat /sys/fs/cgroup/memory/%s/memory.limit_in_bytes", processPath)
		out, err = cm.Exec(p.ChaosDaemonName, p.ChaosDaemonNamespace, cmd, c.K8sClient)
		if err != nil {
			return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
		}
		limit, err := strconv.ParseUint(strings.TrimSuffix(string(out), "\n"), 10, 64)
		if err != nil {
			return fmt.Errorf("could not get memory.limit_in_bytes with: %s", err.Error())
		}
		fmt.Println(string(cm.ColorCyan), "3. [memory.limit_in_bytes]: ", string(cm.ColorReset))
		cm.PrintWithTab(bytefmt.ByteSize(limit) + "B")
	}
	return nil
}
