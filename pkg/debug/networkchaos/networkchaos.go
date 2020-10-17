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
		fmt.Println(string(common.ColorRed), "[CHAOSNAME]\n", chaosName, "\n", string(common.ColorReset))
		if err := debugEachChaos(chaosName, ns); err != nil {
			return fmt.Errorf("debug chaos failed with: %s", err.Error())
		}
	}
	return nil
}

func debugEachChaos(chaos string, ns string) error {
	// get podName
	out, err := exec.Command("kubectl", "describe", "networkchaos", chaos, "-n", ns).CombinedOutput()
	if err != nil {
		return fmt.Errorf("run command 'kubectl describe networkchaos' failed with: %s", err.Error())
	}
	podHier := []string{"Status", "Experiment", "Pod Records", "Name"}
	podName, err := common.ExtractFromYaml(string(out), podHier)
	if err != nil {
		return fmt.Errorf("get podName failed with: %s", err.Error())
	}
	podHier = []string{"Status", "Experiment", "Pod Records", "Namespace"}
	podNamespace, err := common.ExtractFromYaml(string(out), podHier)
	if err != nil {
		return fmt.Errorf("get podNamespace failed with: %s", err.Error())
	}

	// get nodeName
	out, err = exec.Command("kubectl", "get", "pods", "-o", "wide", podName, "-n", podNamespace).CombinedOutput()
	if err != nil {
		return fmt.Errorf("cmd.Run() failed with: %s", err.Error())
	}
	nodeName, err := common.ExtractFromGet(string(out), "NODE")
	if err != nil {
		return fmt.Errorf("get nodeName failed with: %s", err.Error())
	}

	// get chaos daemon
	fullCmd := "kubectl get pods -A -o wide"
	out, err = exec.Command("bash", "-c", fullCmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("cmd.Run() failed with: %s", err.Error())
	}
	title := strings.Split(string(out), "\n")[0]
	fullCmd = "kubectl get pods -A -o wide | grep chaos-daemon | grep " + nodeName
	out, err = exec.Command("bash", "-c", fullCmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("cmd.Run() failed with: %s", err.Error())
	}
	chaosDaemonPodName, err := common.ExtractFromGet(title+"\n"+string(out), "NAME")
	chaosDaemonPodNamespace, err := common.ExtractFromGet(title+"\n"+string(out), "NAMESPACE")

	// get nsenter path from log
	for _, tailNum := range []string{"--tail=20", "--tail=100", "--tail=500", ""} {
		fullCmd := "kubectl logs " + chaosDaemonPodName + " -n " + chaosDaemonPodNamespace + " " + tailNum + " | grep 'nsenter -n/proc/'"
		out, err = exec.Command("bash", "-c", fullCmd).CombinedOutput()
		if err != nil {
			return fmt.Errorf("cmd.Run() failed with: %s", err.Error())
		}
		if len(out) != 0 {
			break
		}
		if tailNum == "" {
			return fmt.Errorf("could not found networkchaos related logs")
		}
	}

	line := strings.Split(string(out), "\n")[0]
	r, _ := regexp.Compile("(?:-n/proc/)(.*)(?:/ns/net)")
	nsenterPath := r.FindStringSubmatch(line)[0]

	// print out result

	fullCmd = "kubectl exec " + chaosDaemonPodName + " -n" + chaosDaemonPodNamespace + " -- /usr/bin/nsenter " + nsenterPath + " -- ipset list"
	out, err = exec.Command("bash", "-c", fullCmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("cmd.Run() failed with: %s", err.Error())
	}
	fmt.Println(string(common.ColorGreen), "[ipset list]", string(common.ColorReset))
	fmt.Println(string(out))

	fullCmd = "kubectl exec " + chaosDaemonPodName + " -n" + chaosDaemonPodNamespace + " -- /usr/bin/nsenter " + nsenterPath + " -- tc qdisc list"
	out, err = exec.Command("bash", "-c", fullCmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("cmd.Run() failed with: %s", err.Error())
	}
	fmt.Println(string(common.ColorGreen), "[tc qdisc list]", string(common.ColorReset))
	fmt.Println(string(out))

	fullCmd = "kubectl exec " + chaosDaemonPodName + " -n" + chaosDaemonPodNamespace + " -- /usr/bin/nsenter " + nsenterPath + " -- iptables --list"
	out, err = exec.Command("bash", "-c", fullCmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("cmd.Run() failed with: %s", err.Error())
	}
	fmt.Println(string(common.ColorGreen), "[iptables list]", string(common.ColorReset))
	fmt.Println(string(out))

	return nil
}
