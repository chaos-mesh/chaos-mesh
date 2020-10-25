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

package iochaos

import (
	"fmt"
	"os/exec"

	"github.com/chaos-mesh/chaos-mesh/pkg/debug/common"
)

func Debug(chaos string, ns string) error {
	chaosList, err := common.Debug("iochaos", chaos, ns)
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
	p, err := common.GetPod("iochaos", chaos, ns)
	if err != nil {
		return err
	}

	// print out debug info
	cmd := fmt.Sprintf("kubectl exec %s -n %s -- ls /proc/1/fd -al", p.ChaosDaemonPodName, p.ChaosDaemonPodNamespace)
	out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	fmt.Println(string(common.ColorGreen), "[file discriptors]", string(common.ColorReset))
	fmt.Println(string(out))

	cmd = fmt.Sprintf("kubectl exec %s -n %s -- mount", p.ChaosDaemonPodName, p.ChaosDaemonPodNamespace)
	out, err = exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	fmt.Println(string(common.ColorGreen), "[mount information]", string(common.ColorReset))
	fmt.Println(string(out))

	return nil
}
