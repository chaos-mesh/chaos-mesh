// Copyright 2020 Chaos Mesh Authors.
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

package bpm

import (
	"context"
	"os/exec"
	"strings"
	"syscall"

	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

// Build builds the process
func (b *ProcessBuilder) Build() *ManagedProcess {
	// The call routine is pause -> suicide -> nsenter --(fork)-> suicide -> process
	// so that when chaos-daemon killed the suicide process, the sub suicide process will
	// receive a signal and exit.
	// For example:
	// If you call `nsenter -p/proc/.../ns/pid bash -c "while true; do sleep 1; date; done"`
	// then even you kill the nsenter process, the subprocess of it will continue running
	// until it gets killed. The suicide program is used to make sure that the subprocess will
	// be terminated when its parent died.
	// But the `./bin/suicide nsenter -p/proc/.../ns/pid ./bin/suicide bash -c "while true; do sleep 1; date; done"`
	// can fix this problem. The first suicide is used to ensure when chaos-daemon is dead, the process is killed

	// I'm not sure this method is 100% reliable, but half a loaf is better than none.

	args := b.args
	cmd := b.cmd

	if b.suicide {
		args = append([]string{cmd}, args...)
		cmd = suicidePath
	}

	if len(b.nsOptions) > 0 {
		args = append([]string{"--", cmd}, args...)
		for _, option := range b.nsOptions {
			args = append([]string{"-" + nsArgMap[option.Typ], option.Path}, args...)
		}

		if b.localMnt {
			args = append([]string{"-l"}, args...)
		}
		cmd = nsexecPath
	}

	if b.pause {
		args = append([]string{cmd}, args...)
		cmd = pausePath
	}

	if c := mock.On("MockProcessBuild"); c != nil {
		f := c.(func(context.Context, string, ...string) *exec.Cmd)
		return &ManagedProcess{
			Cmd:        f(b.ctx, cmd, args...),
			Identifier: b.identifier,
		}
	}

	log.Info("build command", "command", cmd+" "+strings.Join(args, " "))

	command := exec.CommandContext(b.ctx, cmd, args...)
	command.SysProcAttr = &syscall.SysProcAttr{}

	if b.suicide {
		command.SysProcAttr.Pdeathsig = syscall.SIGTERM
	}

	return &ManagedProcess{
		Cmd:        command,
		Identifier: b.identifier,
	}
}
