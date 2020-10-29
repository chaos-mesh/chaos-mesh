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

	cm "github.com/chaos-mesh/chaos-mesh/pkg/debug/common"
)

func Debug(ctx context.Context, chaosName string, ns string, c *cm.ClientSet) error {
	p, err := cm.GetPod(ctx, "iochaos", chaosName, ns, c.CtrlClient)
	if err != nil {
		return err
	}

	// print out debug info
	cmd := fmt.Sprintf("ls /proc/1/fd -al")
	out, err := cm.Exec(p.ChaosDaemonName, p.ChaosDaemonNamespace, cmd, c.K8sClient)
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	fmt.Println(string(cm.ColorCyan), "1. [file discriptors]", string(cm.ColorReset))
	cm.PrintWithTab(string(out))

	cmd = fmt.Sprintf("mount")
	out, err = cm.Exec(p.ChaosDaemonName, p.ChaosDaemonNamespace, cmd, c.K8sClient)
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	fmt.Println(string(cm.ColorCyan), "2. [mount information]", string(cm.ColorReset))
	cm.PrintWithTab(string(out))

	return nil
}
