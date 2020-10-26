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

package networkchaos

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/chaos-mesh/chaos-mesh/pkg/debug/common"
)

func Debug(chaos string, ns string) error {
	chaosList, err := common.Debug("networkchaos", chaos, ns)
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

func debugEachChaos(chaos string, ns string) error {
	p, err := common.GetPod("networkchaos", chaos, ns)
	if err != nil {
		return err
	}

	// get nsenter path from log
	var out []byte
	for _, tailNum := range []string{"--tail=100", "--tail=1000", "--tail=10000", ""} {
		cmd := fmt.Sprintf("kubectl logs %s -n %s %s | grep 'nsenter -n/proc/'", p.ChaosDaemonPodName, p.ChaosDaemonPodNamespace, tailNum)
		out, err = exec.Command("bash", "-c", cmd).CombinedOutput()
		if err != nil {
			return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
		}
		if len(out) != 0 {
			break
		}
		if tailNum == "" {
			return fmt.Errorf("could not found networkchaos related logs")
		}
	}
	line := strings.Split(string(out), "\n")[0]
	nsenterPath := regexp.MustCompile("(?:-n/proc/)(.*)(?:/ns/net)").FindStringSubmatch(line)[0]

	cmd := fmt.Sprintf("kubectl describe networkchaos %s -n %s", chaos, ns)
	out, err = exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	specHier := []string{"Spec", "Action"}
	action, err := common.ExtractFromYaml(string(out), specHier)
	if err != nil {
		fmt.Printf("get podName from '%s' failed with: %s", cmd, err.Error())
	}
	var netemExpect string
	switch action.(string) {
	case "delay":
		latency, _ := common.ExtractFromYaml(string(out), []string{"Spec", "Delay", "Latency"})
		jitter, _ := common.ExtractFromYaml(string(out), []string{"Spec", "Delay", "Jitter"})
		correlation, _ := common.ExtractFromYaml(string(out), []string{"Spec", "Delay", "Correlation"})
		netemExpect = fmt.Sprintf("%v %v %v %v%%", action, latency, jitter, correlation)
	default:
		return fmt.Errorf("chaos not supported")
	}

	// print out debug info

	cmd = fmt.Sprintf("kubectl exec %s -n %s -- /usr/bin/nsenter %s -- ipset list", p.ChaosDaemonPodName, p.ChaosDaemonPodNamespace, nsenterPath)
	out, err = exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	fmt.Println(string(common.ColorCyan), "1. [ipset list]", string(common.ColorReset))
	common.PrintWithTab(string(out))

	cmd = fmt.Sprintf("kubectl exec %s -n %s -- /usr/bin/nsenter %s -- tc qdisc list", p.ChaosDaemonPodName, p.ChaosDaemonPodNamespace, nsenterPath)
	out, err = exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	fmt.Println(string(common.ColorCyan), "2. [tc qdisc list]", string(common.ColorReset))
	common.PrintWithTab(string(out))

	netemCurrent := regexp.MustCompile("(?:limit 1000)(.*)").FindStringSubmatch(string(out))
	if len(netemCurrent) == 0 {
		return fmt.Errorf("No NetworkChaos is applied")
	}
	for i := range strings.Fields(netemCurrent[1]) {
		itemCurrent := strings.Fields(netemCurrent[1])[i]
		itemExpect := strings.Fields(netemExpect)[i]
		if itemCurrent != itemExpect {
			r := regexp.MustCompile("([0-9]*[.])?[0-9]+")
			numCurrent, err := strconv.ParseFloat(r.FindString(itemCurrent), 64)
			if err != nil {
				return fmt.Errorf("parse float failed: %s", err.Error())
			}
			numExpect, err := strconv.ParseFloat(r.FindString(itemExpect), 64)
			if err != nil {
				return fmt.Errorf("parse float failed: %s", err.Error())
			}
			if numCurrent == numExpect {
				continue
			}
			alpCurrent := regexp.MustCompile("[[:alpha:]]+").FindString(itemCurrent)
			alpExpect := regexp.MustCompile("[[:alpha:]]+").FindString(itemExpect)
			if alpCurrent == alpExpect {
				continue
			}
			errInfo := fmt.Sprintf("%sNetworkChaos didn't execute as expected, expect: %s, got: %s%s", string(common.ColorRed), netemExpect, netemCurrent, string(common.ColorReset))
			common.PrintWithTab(errInfo)
			return nil
		}
	}
	sucInfo := fmt.Sprintf("%sNetworkChaos execute as expected%s\n", string(common.ColorGreen), string(common.ColorReset))
	common.PrintWithTab(sucInfo)

	cmd = fmt.Sprintf("kubectl exec %s -n %s -- /usr/bin/nsenter %s -- iptables --list", p.ChaosDaemonPodName, p.ChaosDaemonPodNamespace, nsenterPath)
	out, err = exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("cmd.Run() failed with: %s", err.Error())
	}
	fmt.Println(string(common.ColorCyan), "3. [iptables list]", string(common.ColorReset))
	common.PrintWithTab(string(out))

	return nil
}
