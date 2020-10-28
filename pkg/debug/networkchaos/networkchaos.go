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

func debugEachChaos(chaosName string, ns string) error {
	p, err := common.GetPod("networkchaos", chaosName, ns)
	if err != nil {
		return err
	}

	// get nsenter path from log
	var nsenterPathList []string
	for _, tailNum := range []int64{100, 1000, 10000, -1} {
		log, err := common.GetLog(p.ChaosDaemonName, p.ChaosDaemonNamespace, tailNum)
		if err != nil {
			return fmt.Errorf("get log failed with: %s", err.Error())
		}
		nsenterPathList = regexp.MustCompile("(?:-n/proc/)(.*)(?:/ns/net)").FindStringSubmatch(log)
		if len(nsenterPathList) != 0 {
			break
		}
		if tailNum == -1 {
			return fmt.Errorf("could not found networkchaos related logs")
		}
	}
	nsenterPath := nsenterPathList[0]

	chaos, err := common.GetChaos("networkchaos", chaosName, ns)
	if err != nil {
		return fmt.Errorf("failed to get chaos %s: %s", chaosName, err.Error())
	}

	actionHier := []string{"spec", "action"}
	action, err := common.ExtractFromJson(chaos, actionHier)
	if err != nil {
		return fmt.Errorf("get action failed with: %s", err.Error())
	}
	var netemExpect string
	switch action.(string) {
	case "delay":
		latency, _ := common.ExtractFromJson(chaos, []string{"spec", "delay", "latency"})
		jitter, _ := common.ExtractFromJson(chaos, []string{"spec", "delay", "jitter"})
		correlation, _ := common.ExtractFromJson(chaos, []string{"spec", "delay", "correlation"})
		netemExpect = fmt.Sprintf("%v %v %v %v%%", action, latency, jitter, correlation)
	default:
		return fmt.Errorf("chaos not supported")
	}

	// print out debug info
	cmd := fmt.Sprintf("/usr/bin/nsenter %s -- ipset list", nsenterPath)
	out, err := common.ExecCommand(p.ChaosDaemonName, p.ChaosDaemonNamespace, cmd)
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	fmt.Println(string(common.ColorCyan), "1. [ipset list]", string(common.ColorReset))
	common.PrintWithTab(string(out))

	cmd = fmt.Sprintf("/usr/bin/nsenter %s -- tc qdisc list", nsenterPath)
	out, err = common.ExecCommand(p.ChaosDaemonName, p.ChaosDaemonNamespace, cmd)
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	fmt.Println(string(common.ColorCyan), "2. [tc qdisc list]", string(common.ColorReset))
	common.PrintWithTab(string(out))

	netemCurrent := regexp.MustCompile("(?:limit 1000)(.*)").FindStringSubmatch(string(out))
	if len(netemCurrent) == 0 {
		return fmt.Errorf("No NetworkChaos is applied")
	}
	for i, netem := range strings.Fields(netemCurrent[1]) {
		itemCurrent := netem
		itemExpect := strings.Fields(netemExpect)[i]
		if itemCurrent != itemExpect {
			r := regexp.MustCompile("([0-9]*[.])?[0-9]+")
			numCurrent, err := strconv.ParseFloat(r.FindString(itemCurrent), 64)
			if err != nil {
				return fmt.Errorf("parse itemCurrent failed: %s", err.Error())
			}
			numExpect, err := strconv.ParseFloat(r.FindString(itemExpect), 64)
			if err != nil {
				return fmt.Errorf("parse itemExpect failed: %s", err.Error())
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

	cmd = fmt.Sprintf("/usr/bin/nsenter %s -- iptables --list", nsenterPath)
	out, err = common.ExecCommand(p.ChaosDaemonName, p.ChaosDaemonNamespace, cmd)
	if err != nil {
		return fmt.Errorf("cmd.Run() failed with: %s", err.Error())
	}
	fmt.Println(string(common.ColorCyan), "3. [iptables list]", string(common.ColorReset))
	common.PrintWithTab(string(out))

	return nil
}
