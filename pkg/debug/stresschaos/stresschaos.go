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
	"os/exec"
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
		fmt.Println(string(common.ColorRed), "[CHAOSNAME]:", string(common.ColorReset), chaosName)
		if err := debugEachChaos(chaosName, ns); err != nil {
			return fmt.Errorf("debug chaos failed with: %s", err.Error())
		}
	}
	return nil
}

func debugEachChaos(chaos string, ns string) error {
	p, err := common.GetPod("stresschaos", chaos, ns)
	if err != nil {
		return err
	}

	// cpu or memory chaos
	cmd := fmt.Sprintf("kubectl describe stresschaos %s -n %s", chaos, ns)
	out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	isCPU := regexp.MustCompile("(f:cpu)").MatchString(string(out))

	// get process path
	cmd = fmt.Sprintf("kubectl exec %s -n %s -- cat /proc/cgroups", p.PodName, p.PodNamespace)
	out, err = exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	var cpuMountType string
	if regexp.MustCompile("(cpu,cpuacct)").MatchString(string(out)) {
		cpuMountType = "cpu,cpuacct"
	} else {
		cpuMountType = "cpu"
	}

	cmd = fmt.Sprintf("kubectl exec %s -n %s -- cat /proc/1/cgroup", p.PodName, p.PodNamespace)
	out, err = exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	var expr string
	if isCPU {
		expr = "(?::" + cpuMountType + ":)(.*)"
	} else {
		expr = "(?::memory:)(.*)"
	}
	processPath := regexp.MustCompile(expr).FindStringSubmatch(string(out))[1]

	// print out debug info
	if isCPU {
		cmd = fmt.Sprintf("kubectl exec %s -n %s -- cat /sys/fs/cgroup/%s/%s/cpu.cfs_quota_us", p.ChaosDaemonPodName, p.ChaosDaemonPodNamespace, cpuMountType, processPath)
		out, err = exec.Command("bash", "-c", cmd).CombinedOutput()
		if err != nil {
			return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
		}
		fmt.Println(string(common.ColorGreen), "[cpu.cfs_quota_us]:", string(common.ColorReset), string(out))
		quota, err := strconv.Atoi(strings.TrimSuffix(string(out), "\n"))
		if err != nil {
			return fmt.Errorf("could not get cpu.cfs_quota_us with: %s", err.Error())
		}

		cmd = fmt.Sprintf("kubectl exec %s -n %s -- cat /sys/fs/cgroup/%s/%s/cpu.cfs_period_us", p.ChaosDaemonPodName, p.ChaosDaemonPodNamespace, cpuMountType, processPath)
		out, err = exec.Command("bash", "-c", cmd).CombinedOutput()
		if err != nil {
			return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
		}
		fmt.Println(string(common.ColorGreen), "[cpu.cfs_period_us]:", string(common.ColorReset), string(out))
		period, err := strconv.Atoi(strings.TrimSuffix(string(out), "\n"))
		if err != nil {
			return fmt.Errorf("could not get cpu.cfs_period_us with: %s", err.Error())
		}

		if quota == -1 {
			fmt.Println(string(common.ColorGreen), "no cpu limit is set for now", string(common.ColorReset))
		} else {
			fmt.Println(string(common.ColorGreen), "cpu limit is equals to", float64(quota)/float64(period), string(common.ColorReset))
		}
	} else {
		cmd = fmt.Sprintf("kubectl exec %s -n %s -- cat /sys/fs/cgroup/memory/%s/memory.limit_in_bytes", p.ChaosDaemonPodName, p.ChaosDaemonPodNamespace, processPath)
		out, err = exec.Command("bash", "-c", cmd).CombinedOutput()
		if err != nil {
			return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
		}
		limit, err := strconv.ParseUint(strings.TrimSuffix(string(out), "\n"), 10, 64)
		if err != nil {
			return fmt.Errorf("could not get memory.limit_in_bytes with: %s", err.Error())
		}
		fmt.Println(string(common.ColorGreen), "[memory.limit_in_bytes]: ", string(common.ColorReset), bytefmt.ByteSize(limit))
	}
	return nil
}
