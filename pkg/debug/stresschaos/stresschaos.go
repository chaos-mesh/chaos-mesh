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
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"code.cloudfoundry.org/bytefmt"

	"github.com/chaos-mesh/chaos-mesh/pkg/debug/common"
)

func Debug(chaos string, ns string) error {
	chaosList, err := common.Debug("stresschaos", chaos, ns)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	for _, chaosName := range chaosList {
		fmt.Println(string(common.ColorCyan), "[CHAOSNAME]:", chaosName, string(common.ColorReset))
		if err := debugEachChaos(chaosName, ns); err != nil {
			return fmt.Errorf("debug chaos failed with: %s", err.Error())
		}
	}
	return nil
}

func debugEachChaos(chaosName string, ns string) error {
	p, err := common.GetPod("stresschaos", chaosName, ns)
	if err != nil {
		return err
	}

	// cpu or memory chaos
	chaos, err := common.GetChaos("stresschaos", chaosName, ns)
	if err != nil {
		return fmt.Errorf("failed to get chaos %s: %s", chaosName, err.Error())
	}

	isCPU := true
	cpuHier := []string{"spec", "stressors", "cpu"}
	_, err = common.ExtractFromJson(chaos, cpuHier)
	if err != nil {
		isCPU = false
	}

	// get process path
	cmd := fmt.Sprintf("cat /proc/cgroups")
	out, err := common.ExecCommand(p.PodName, p.PodNamespace, cmd)
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
	out, err = common.ExecCommand(p.PodName, p.PodNamespace, cmd)
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	stressngLine := regexp.MustCompile("(.*)(stress-ng)").FindStringSubmatch(string(out))
	if len(stressngLine) == 0 {
		return fmt.Errorf("Could not find stress-ng, StressChaos failed")
	}
	stressngPid := strings.Split(stressngLine[0], " ")[0]

	cmd = fmt.Sprintf("cat /proc/1/cgroup")
	out, err = common.ExecCommand(p.PodName, p.PodNamespace, cmd)
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	fmt.Println(string(common.ColorCyan), "1. [cat /proc/1/cgroup]:", string(common.ColorReset))
	common.PrintWithTab(string(out))

	var expr string
	if isCPU {
		expr = "(?::" + cpuMountType + ":)(.*)"
	} else {
		expr = "(?::memory:)(.*)"
	}
	processPath := regexp.MustCompile(expr).FindStringSubmatch(string(out))[1]

	cmd = fmt.Sprintf("cat /proc/%s/cgroup", stressngPid)
	outStress, err := common.ExecCommand(p.PodName, p.PodNamespace, cmd)
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	fmt.Println(string(common.ColorCyan), "2. [cat /proc/(stress-ng pid)/cgroup]:", string(common.ColorReset))
	common.PrintWithTab(string(outStress))

	if string(out) != string(outStress) {
		errInfo := fmt.Sprintf("%sStressChaos failed to execute as expected%s\n", string(common.ColorRed), string(common.ColorReset))
		common.PrintWithTab(errInfo)
		return nil
	}
	sucInfo := fmt.Sprintf("%scgroup is the same%s\n", string(common.ColorGreen), string(common.ColorReset))
	common.PrintWithTab(sucInfo)

	// print out debug info
	if isCPU {
		cmd = fmt.Sprintf("cat /sys/fs/cgroup/%s/%s/cpu.cfs_quota_us", cpuMountType, processPath)
		out, err = common.ExecCommand(p.ChaosDaemonName, p.ChaosDaemonNamespace, cmd)
		if err != nil {
			return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
		}
		fmt.Println(string(common.ColorCyan), "3. [cpu.cfs_quota_us]:", string(common.ColorReset))
		common.PrintWithTab(string(out))
		quota, err := strconv.Atoi(strings.TrimSuffix(string(out), "\n"))
		if err != nil {
			return fmt.Errorf("could not get cpu.cfs_quota_us with: %s", err.Error())
		}

		cmd = fmt.Sprintf("cat /sys/fs/cgroup/%s/%s/cpu.cfs_period_us", cpuMountType, processPath)
		out, err = common.ExecCommand(p.ChaosDaemonName, p.ChaosDaemonNamespace, cmd)
		if err != nil {
			return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
		}
		fmt.Println(string(common.ColorCyan), "4. [cpu.cfs_period_us]:", string(common.ColorReset))
		common.PrintWithTab(string(out))
		period, err := strconv.Atoi(strings.TrimSuffix(string(out), "\n"))
		if err != nil {
			return fmt.Errorf("could not get cpu.cfs_period_us with: %s", err.Error())
		}

		if quota == -1 {
			fmt.Println(string(common.ColorRed), "no cpu limit is set for now", string(common.ColorReset))
		} else {
			fmt.Println(string(common.ColorGreen), "cpu limit is equals to", float64(quota)/float64(period), string(common.ColorReset))
		}
	} else {
		cmd = fmt.Sprintf("cat /sys/fs/cgroup/memory/%s/memory.limit_in_bytes", processPath)
		out, err = common.ExecCommand(p.ChaosDaemonName, p.ChaosDaemonNamespace, cmd)
		if err != nil {
			return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
		}
		limit, err := strconv.ParseUint(strings.TrimSuffix(string(out), "\n"), 10, 64)
		if err != nil {
			return fmt.Errorf("could not get memory.limit_in_bytes with: %s", err.Error())
		}
		fmt.Println(string(common.ColorCyan), "3. [memory.limit_in_bytes]: ", string(common.ColorReset))
		common.PrintWithTab(bytefmt.ByteSize(limit) + "B")
	}
	return nil
}
