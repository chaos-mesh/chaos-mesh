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
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/shirou/gopsutil/process"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("background-process-manager")

type NsType string

const (
	MountNS NsType = "mnt"
	// uts namespace is not supported yet
	// UtsNS   NsType = "uts"
	IpcNS NsType = "ipc"
	NetNS NsType = "net"
	PidNS NsType = "pid"
	// user namespace is not supported yet
	// UserNS  NsType = "user"
)

var nsArgMap = map[NsType]string{
	MountNS: "m",
	// uts namespace is not supported by nsexec yet
	// UtsNS:   "u",
	IpcNS: "i",
	NetNS: "n",
	PidNS: "p",
	// user namespace is not supported by nsexec yet
	// UserNS:  "U",
}

const (
	pausePath  = "/usr/local/bin/pause"
	nsexecPath = "/usr/local/bin/nsexec"

	DefaultProcPrefix = "/proc"
)

// ProcessPair is an identifier for process
type ProcessPair struct {
	Pid        int
	CreateTime int64
}

// Stdio contains stdin, stdout and stderr
type Stdio struct {
	sync.Locker
	Stdin, Stdout, Stderr io.ReadWriteCloser
}

// BackgroundProcessManager manages all background processes
type BackgroundProcessManager struct {
	deathSig    *sync.Map
	identifiers *sync.Map
	stdio       *sync.Map
}

// NewBackgroundProcessManager creates a background process manager
func NewBackgroundProcessManager() BackgroundProcessManager {
	return BackgroundProcessManager{
		deathSig:    &sync.Map{},
		identifiers: &sync.Map{},
		stdio:       &sync.Map{},
	}
}

// StartProcess manages a process in manager
func (m *BackgroundProcessManager) StartProcess(cmd *ManagedProcess) (*process.Process, error) {
	var identifierLock *sync.Mutex
	if cmd.Identifier != nil {
		lock, _ := m.identifiers.LoadOrStore(*cmd.Identifier, &sync.Mutex{})

		identifierLock = lock.(*sync.Mutex)

		identifierLock.Lock()
	}

	err := cmd.Start()
	if err != nil {
		log.Error(err, "fail to start process")
		return nil, err
	}

	pid := cmd.Process.Pid
	procState, err := process.NewProcess(int32(cmd.Process.Pid))
	if err != nil {
		return nil, err
	}
	ct, err := procState.CreateTime()
	if err != nil {
		return nil, err
	}

	pair := ProcessPair{
		Pid:        pid,
		CreateTime: ct,
	}

	channel, _ := m.deathSig.LoadOrStore(pair, make(chan bool, 1))
	deathChannel := channel.(chan bool)

	stdio := &Stdio{Locker: &sync.Mutex{}}
	if cmd.Stdin != nil {
		if stdin, ok := cmd.Stdin.(io.ReadWriteCloser); ok {
			stdio.Stdin = stdin
		}
	}

	if cmd.Stdout != nil {
		if stdout, ok := cmd.Stdout.(io.ReadWriteCloser); ok {
			stdio.Stdout = stdout
		}
	}

	if cmd.Stderr != nil {
		if stderr, ok := cmd.Stderr.(io.ReadWriteCloser); ok {
			stdio.Stderr = stderr
		}
	}

	m.stdio.Store(pair, stdio)

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
		if io, loaded := m.stdio.LoadAndDelete(pair); loaded {
			if stdio, ok := io.(*Stdio); ok {
				stdio.Lock()
				if stdio.Stdin != nil {
					if err = stdio.Stdin.Close(); err != nil {
						log.Error(err, "stdin fails to be closed")
					}
				}
				if stdio.Stdout != nil {
					if err = stdio.Stdout.Close(); err != nil {
						log.Error(err, "stdout fails to be closed")
					}
				}
				if stdio.Stderr != nil {
					if err = stdio.Stderr.Close(); err != nil {
						log.Error(err, "stderr fails to be closed")
					}
				}
				stdio.Unlock()
			}
		}

		if identifierLock != nil {
			identifierLock.Unlock()
			m.identifiers.Delete(*cmd.Identifier)
		}
	}()

	return procState, nil
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

	// There is a bug in calculating CreateTime in the new version of
	// gopsutils. This is a temporary solution before the upstream fixes it.
	if startTime-ct > 1000 || ct-startTime > 1000 {
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
		CreateTime: startTime,
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

func (m *BackgroundProcessManager) Stdio(pid int, startTime int64) *Stdio {
	log := log.WithValues("pid", pid)

	procState, err := process.NewProcess(int32(pid))
	if err != nil {
		log.Info("fail to get process information", "pid", pid)
		// return successfully as the process has exited
		return nil
	}
	ct, err := procState.CreateTime()
	if err != nil {
		log.Error(err, "fail to read create time")
		// return successfully as the process has exited
		return nil
	}

	// There is a bug in calculating CreateTime in the new version of
	// gopsutils. This is a temporary solution before the upstream fixes it.
	if startTime-ct > 1000 || ct-startTime > 1000 {
		log.Info("process has exited", "startTime", ct, "expectedStartTime", startTime)
		// return successfully as the process has exited
		return nil
	}

	pair := ProcessPair{
		Pid:        pid,
		CreateTime: startTime,
	}

	io, ok := m.stdio.Load(pair)
	if !ok {
		log.Info("fail to load with pair", "pair", pair)
		// stdio is not stored
		return nil
	}

	return io.(*Stdio)
}

// DefaultProcessBuilder returns the default process builder
func DefaultProcessBuilder(cmd string, args ...string) *ProcessBuilder {
	return &ProcessBuilder{
		cmd:        cmd,
		args:       args,
		nsOptions:  []nsOption{},
		pause:      false,
		identifier: nil,
		ctx:        context.Background(),
	}
}

// ProcessBuilder builds a exec.Cmd for daemon
type ProcessBuilder struct {
	cmd  string
	args []string
	env  []string

	nsOptions []nsOption

	pause    bool
	localMnt bool

	identifier *string
	stdin      io.ReadWriteCloser
	stdout     io.ReadWriteCloser
	stderr     io.ReadWriteCloser

	ctx context.Context
}

// GetNsPath returns corresponding namespace path
func GetNsPath(pid uint32, typ NsType) string {
	return fmt.Sprintf("%s/%d/ns/%s", DefaultProcPrefix, pid, string(typ))
}

// SetEnv sets the environment variables of the process
func (b *ProcessBuilder) SetEnv(key, value string) *ProcessBuilder {
	b.env = append(b.env, fmt.Sprintf("%s=%s", key, value))
	return b
}

// SetNS sets the namespace of the process
func (b *ProcessBuilder) SetNS(pid uint32, typ NsType) *ProcessBuilder {
	return b.SetNSOpt([]nsOption{{
		Typ:  typ,
		Path: GetNsPath(pid, typ),
	}})
}

// SetNSOpt sets the namespace of the process
func (b *ProcessBuilder) SetNSOpt(options []nsOption) *ProcessBuilder {
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

func (b *ProcessBuilder) EnableLocalMnt() *ProcessBuilder {
	b.localMnt = true

	return b
}

// SetContext sets context for process
func (b *ProcessBuilder) SetContext(ctx context.Context) *ProcessBuilder {
	b.ctx = ctx

	return b
}

// SetStdin sets stdin for process
func (b *ProcessBuilder) SetStdin(stdin io.ReadWriteCloser) *ProcessBuilder {
	b.stdin = stdin

	return b
}

// SetStdout sets stdout for process
func (b *ProcessBuilder) SetStdout(stdout io.ReadWriteCloser) *ProcessBuilder {
	b.stdout = stdout

	return b
}

// SetStderr sets stderr for process
func (b *ProcessBuilder) SetStderr(stderr io.ReadWriteCloser) *ProcessBuilder {
	b.stderr = stderr

	return b
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
