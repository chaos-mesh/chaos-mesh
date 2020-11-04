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
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	cm "github.com/chaos-mesh/chaos-mesh/pkg/debug/common"
)

func Debug(ctx context.Context, chaos runtime.Object, c *cm.ClientSet) error {
	ioChaos, ok := chaos.(*v1alpha1.IoChaos)
	if !ok {
		return fmt.Errorf("chaos is not iochaos")
	}
	chaosStatus := ioChaos.Status.ChaosStatus
	chaosSelector := ioChaos.Spec.GetSelector()

	pods, daemons, err := cm.GetPods(ctx, chaosStatus, chaosSelector, c.CtrlClient)
	if err != nil {
		return err
	}

	for i := range pods {
		podName := pods[i].GetObjectMeta().GetName()
		cm.Print("[Pod]: "+podName, 0, cm.ColorBlue)
		err := debugEachPod(ctx, pods[i], daemons[i], ioChaos, c)
		if err != nil {
			return fmt.Errorf("for %s: %s", podName, err.Error())
		}
	}
	return nil
}

func debugEachPod(ctx context.Context, pod v1.Pod, daemon v1.Pod, chaos *v1alpha1.IoChaos, c *cm.ClientSet) error {
	daemonName := daemon.GetObjectMeta().GetName()
	daemonNamespace := daemon.GetObjectMeta().GetNamespace()

	// print out debug info
	cmd := fmt.Sprintf("ls /proc/1/fd -al")
	out, err := cm.Exec(daemonName, daemonNamespace, cmd, c.K8sClient)
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	cm.Print("1. [file discriptors]", 1, cm.ColorCyan)
	cm.Print(string(out), 1, "")

	cmd = fmt.Sprintf("mount")
	out, err = cm.Exec(daemonName, daemonNamespace, cmd, c.K8sClient)
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	cm.Print("2. [mount information]", 1, cm.ColorCyan)
	cm.Print(string(out), 1, "")

	return nil
}
