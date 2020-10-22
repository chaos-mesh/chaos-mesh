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
	"strings"

	"github.com/chaos-mesh/chaos-mesh/pkg/debug/common"
)

func Debug(chaos string, ns string) error {
	chaosList, err := common.Debug("networkchaos", chaos, ns)
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

	// print out debug info
	cmd := fmt.Sprintf("kubectl exec %s -n %s -- /usr/bin/nsenter %s -- ipset list", p.ChaosDaemonPodName, p.ChaosDaemonPodNamespace, nsenterPath)
	out, err = exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	fmt.Println(string(common.ColorGreen), "[ipset list]", string(common.ColorReset))
	fmt.Println(string(out))

	cmd = fmt.Sprintf("kubectl exec %s -n %s -- /usr/bin/nsenter %s -- tc qdisc list", p.ChaosDaemonPodName, p.ChaosDaemonPodNamespace, nsenterPath)
	out, err = exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	fmt.Println(string(common.ColorGreen), "[tc qdisc list]", string(common.ColorReset))
	fmt.Println(string(out))

	cmd = fmt.Sprintf("kubectl exec %s -n %s -- /usr/bin/nsenter %s -- iptables --list", p.ChaosDaemonPodName, p.ChaosDaemonPodNamespace, nsenterPath)
	out, err = exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("cmd.Run() failed with: %s", err.Error())
	}
	fmt.Println(string(common.ColorGreen), "[iptables list]", string(common.ColorReset))
	fmt.Println(string(out))

	return nil
}
