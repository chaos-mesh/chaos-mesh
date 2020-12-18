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
	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	cm "github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/common"
)

// Debug get chaos debug information
func Debug(ctx context.Context, chaos runtime.Object, c *cm.ClientSet, result *cm.ChaosResult) error {
	ioChaos, ok := chaos.(*v1alpha1.IoChaos)
	if !ok {
		return fmt.Errorf("chaos is not iochaos")
	}
	chaosStatus := ioChaos.Status.ChaosStatus
	chaosSelector := ioChaos.Spec.GetSelector()

	pods, daemons, err := cm.GetPods(ctx, chaosStatus, chaosSelector, c.CtrlCli)
	if err != nil {
		return err
	}

	if err := cm.CheckFailedMessage(ctx, chaosStatus.FailedMessage, daemons, c); err != nil {
		return err
	}

	for i := range pods {
		podName := pods[i].Name
		podResult := cm.PodResult{Name: podName}
		err := debugEachPod(ctx, pods[i], daemons[i], ioChaos, c, &podResult)
		result.Pods = append(result.Pods, podResult)
		if err != nil {
			return fmt.Errorf("for %s: %s", podName, err.Error())
		}
	}
	return nil
}

func debugEachPod(ctx context.Context, pod v1.Pod, daemon v1.Pod, chaos *v1alpha1.IoChaos, c *cm.ClientSet, result *cm.PodResult) error {
	// print out debug info
	cmd := fmt.Sprintf("cat /proc/mounts")
	out, err := cm.ExecBypass(ctx, pod, daemon, cmd, c.KubeCli, bpm.PidNS)
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	result.Items = append(result.Items, cm.ItemResult{Name: "mount information", Value: string(out)})

	pids, commands, err := cm.GetPidFromPS(ctx, pod, daemon, c.KubeCli)
	if err != nil {
		return fmt.Errorf("get pid from ps failed with: %s", err.Error())
	}

	for i := range pids {
		cmd = fmt.Sprintf("ls -l /proc/%s/fd", pids[i])
		out, err = cm.ExecBypass(ctx, pod, daemon, cmd, c.KubeCli)
		if err == nil {
			result.Items = append(result.Items, cm.ItemResult{Name: fmt.Sprintf("file discriptors of PID: %s, COMMAND: %s", pids[i], commands[i]), Value: string(out)})
		}
	}

	return nil
}
