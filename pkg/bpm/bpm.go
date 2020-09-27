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
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/shirou/gopsutil/process"

	"github.com/chaos-mesh/chaos-mesh/pkg/mock"

	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("background-process-manager")

type NsType string

const (
	MountNS NsType = "mnt"
	UtsNS   NsType = "uts"
	IpcNS   NsType = "ipc"
	NetNS   NsType = "net"
	PidNS   NsType = "pid"
	UserNS  NsType = "user"
)

var nsArgMap = map[NsType]string{
	MountNS: "m",
	UtsNS:   "u",
	IpcNS:   "i",
	NetNS:   "n",
	PidNS:   "p",
	UserNS:  "U",
}

const (
	pausePath   = "/usr/local/bin/pause"
	suicidePath = "/usr/local/bin/suicide"
	ignorePath  = "/usr/local/bin/ignore"
)

// ProcessPair is an identifier for process
type ProcessPair struct {
	Pid        int
	CreateTime int64
}

// BackgroundProcessManager manages all background processes
type BackgroundProcessManager struct {
	deathSig    *sync.Map
	identifiers *sync.Map
}

// NewBackgroundProcessManager creates a background process manager
func NewBackgroundProcessManager() BackgroundProcessManager {
	return BackgroundProcessManager{
		deathSig:    &sync.Map{},
		identifiers: &sync.Map{},
	}
}

// StartProcess manages a process in manager
func (m *BackgroundProcessManager) StartProcess(cmd *ManagedProcess) error {
	var identifierLock *sync.Mutex
	if cmd.Identifier != nil {
		lock, _ := m.identifiers.LoadOrStore(*cmd.Identifier, &sync.Mutex{})

		identifierLock = lock.(*sync.Mutex)

		identifierLock.Lock()
	}

	err := cmd.Start()
	if err != nil {
		log.Error(err, "fail to start process")
		return err
	}

	pid := cmd.Process.Pid
	procState, err := process.NewProcess(int32(cmd.Process.Pid))
	if err != nil {
		return err
	}
	ct, err := procState.CreateTime()

	pair := ProcessPair{
		Pid:        pid,
		CreateTime: ct,
	}

	channel, _ := m.deathSig.LoadOrStore(pair, make(chan bool, 1))
	deathChannel := channel.(chan bool)

	log := log.WithValues("pid", pid)

	go func() {
		err := cmd.Wait()
		if err != nil {
			err, ok := err.(*exec.ExitError)
			if ok {
				status := err.Sys().(syscall.WaitStatus)
				if status.Signaled() && status.Signal() == syscall.SIGTERM {
					log.Info("process stopped with SIGTERM signal")
				}
			} else {
				log.Error(err, "process exited accidentally")
			}
		}

		log.Info("process stopped")

		deathChannel <- true
		m.deathSig.Delete(pair)

		if identifierLock != nil {
			identifierLock.Unlock()
			m.identifiers.Delete(*cmd.Identifier)
		}
	}()

	return nil
}

// KillBackgroundProcess sends SIGTERM to process
func (m *BackgroundProcessManager) KillBackgroundProcess(ctx context.Context, pid int, startTime int64) error {
	log := log.WithValues("pid", pid)

	p, err := os.FindProcess(int(pid))
	if err != nil {
		log.Error(err, "unreachable path. `os.FindProcess` will never return an error on unix")
		return err
	}

	procState, err := process.NewProcess(int32(pid))
	if err != nil {
		// return successfully as the process has exited
		return nil
	}
	ct, err := procState.CreateTime()
	if err != nil {
		log.Error(err, "fail to read create time")
		// return successfully as the process has exited
		return nil
	}
	if startTime != ct {
		log.Info("process has already been killed", "startTime", ct, "expectedStartTime", startTime)
		// return successfully as the process has exited
		return nil
	}

	ppid, err := procState.Ppid()
	if err != nil {
		log.Error(err, "fail to read parent id")
		// return successfully as the process has exited
		return nil
	}
	if ppid != int32(os.Getpid()) {
		log.Info("process has already been killed", "ppid", ppid)
		// return successfully as the process has exited
		return nil
	}

	err = p.Signal(syscall.SIGTERM)

	if err != nil && err.Error() != "os: process already finished" {
		log.Error(err, "error while killing process")
		return err
	}

	pair := ProcessPair{
		Pid:        pid,
		CreateTime: ct,
	}
	channel, ok := m.deathSig.Load(pair)
	if ok {
		deathChannel := channel.(chan bool)
		select {
		case <-deathChannel:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	log.Info("Successfully killed process")
	return nil
}

// DefaultProcessBuilder returns the default process builder
func DefaultProcessBuilder(cmd string, args ...string) *ProcessBuilder {
	return &ProcessBuilder{
		cmd:        cmd,
		args:       args,
		nsOptions:  []nsOption{},
		pause:      false,
		suicide:    false,
		identifier: nil,
		ctx:        context.Background(),
	}
}

// ProcessBuilder builds a exec.Cmd for daemon
type ProcessBuilder struct {
	cmd  string
	args []string

	nsOptions []nsOption

	pause   bool
	suicide bool

	identifier *string

	ctx context.Context
}

// SetNetNS sets the net namespace of the process
func (b *ProcessBuilder) SetNetNS(nsPath string) *ProcessBuilder {
	return b.SetNS([]nsOption{{
		Typ:  NetNS,
		Path: nsPath,
	}})
}

// SetPidNS sets the pid namespace of the process
func (b *ProcessBuilder) SetPidNS(nsPath string) *ProcessBuilder {
	return b.SetNS([]nsOption{{
		Typ:  PidNS,
		Path: nsPath,
	}})
}

// SetNS sets the namespace of the process
func (b *ProcessBuilder) SetNS(options []nsOption) *ProcessBuilder {
	b.nsOptions = append(b.nsOptions, options...)

	return b
}

// SetIdentifier sets the identifier of the process
func (b *ProcessBuilder) SetIdentifier(id string) *ProcessBuilder {
	b.identifier = &id

	return b
}

// EnablePause enables pause for process
func (b *ProcessBuilder) EnablePause() *ProcessBuilder {
	b.pause = true

	return b
}

// EnableSuicide enables suicide for process
func (b *ProcessBuilder) EnableSuicide() *ProcessBuilder {
	b.suicide = true

	return b
}

// SetContext sets context for process
func (b *ProcessBuilder) SetContext(ctx context.Context) *ProcessBuilder {
	b.ctx = ctx

	return b
}

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
			args = append([]string{"-" + nsArgMap[option.Typ] + option.Path}, args...)
		}
		cmd = "nsenter"
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

type nsOption struct {
	Typ  NsType
	Path string
}

// ManagedProcess is a process which can be managed by backgroundProcessManager
type ManagedProcess struct {
	*exec.Cmd

	// If the identifier is not nil, process manager should make sure no other
	// process with this identifier is running when executing this command
	Identifier *string
}
