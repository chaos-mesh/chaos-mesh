// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package bpm

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"syscall"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/process"

	"github.com/chaos-mesh/chaos-mesh/pkg/log"
)

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
// Keep compatible with v2.x
// TODO: remove in v3.x
//
// Currently, the bpm locate managed processes by both PID and create time, because the OS may reuse PID, we must check the create time to avoid locating the wrong process.
//
// However, the two-step locating is messy and the create time may be imprecise (we have fixed a [relevant bug](https://github.com/shirou/gopsutil/pull/1204)).
// In future version, we should completely remove the two-step locating and identify managed processes by UID only.
type ProcessPair struct {
	Pid        int
	CreateTime int64
}

type Process struct {
	Uid string

	// TODO: remove in v3.x
	// store create time, to keep compatible with v2.x
	Pair ProcessPair

	Cmd   *ManagedCommand
	Pipes Pipes

	ctx     context.Context
	stopped context.CancelFunc
}

// pipes that will be connected to the command's stdin/stdout
type Pipes struct {
	Stdin  io.WriteCloser
	Stdout io.ReadCloser
}

// BackgroundProcessManager manages all background processes
type BackgroundProcessManager struct {
	// deathChannel is a channel to receive Uid of dead processes
	deathChannel chan string

	// wait group to await all processes exit
	wg *sync.WaitGroup

	// identifiers is a map to prevent duplicated processes
	identifiers *sync.Map

	// Uid -> Process
	processes *sync.Map

	// TODO: remove in v3.x
	// PidPair -> Uid, to keep compatible with v2.x
	pidPairToUid *sync.Map

	rootLogger logr.Logger

	metricsCollector *metricsCollector
}

func startProcess(cmd *ManagedCommand) (*Process, error) {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, errors.Wrap(err, "create stdin pipe")
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "create stdout pipe")
	}

	err = cmd.Start()
	if err != nil {
		return nil, errors.Wrapf(err, "start command `%s`", cmd.String())
	}

	newProcess := &Process{
		Uid:   uuid.NewString(),
		Cmd:   cmd,
		Pipes: Pipes{Stdin: stdin, Stdout: stdout},
	}

	newProcess.ctx, newProcess.stopped = context.WithCancel(context.Background())

	// keep compatible with v2.x
	// TODO: remove in v3.x
	pid := cmd.Process.Pid
	proc, err := process.NewProcess(int32(cmd.Process.Pid))
	if err != nil {
		return nil, errors.Wrapf(err, "get process state for pid %d", pid)
	}

	ct, err := proc.CreateTime()
	if err != nil {
		return nil, errors.Wrapf(err, "get process create time for pid %d", pid)
	}

	newProcess.Pair = ProcessPair{
		Pid:        int(proc.Pid),
		CreateTime: ct,
	}
	return newProcess, nil
}

func (p *Process) Stopped() <-chan struct{} {
	return p.ctx.Done()
}

// StartBackgroundProcessManager creates a background process manager
func StartBackgroundProcessManager(registry prometheus.Registerer, rootLogger logr.Logger) *BackgroundProcessManager {
	backgroundProcessManager := &BackgroundProcessManager{
		deathChannel:     make(chan string, 1),
		wg:               &sync.WaitGroup{},
		identifiers:      &sync.Map{},
		processes:        &sync.Map{},
		pidPairToUid:     &sync.Map{},
		rootLogger:       rootLogger.WithName("background-process-manager"),
		metricsCollector: nil,
	}

	go func() {
		// return if deathChannel is closed
		for uid := range backgroundProcessManager.deathChannel {
			process, loaded := backgroundProcessManager.processes.LoadAndDelete(uid)
			if loaded {
				proc := process.(*Process)
				backgroundProcessManager.pidPairToUid.Delete(proc.Pair)
				if proc.Cmd.Identifier != nil {
					backgroundProcessManager.identifiers.Delete(*proc.Cmd.Identifier)
				}
				proc.stopped()
			}
			backgroundProcessManager.wg.Done()
		}
	}()

	if registry != nil {
		backgroundProcessManager.metricsCollector = newMetricsCollector(backgroundProcessManager, registry)
	}

	return backgroundProcessManager
}

func (m *BackgroundProcessManager) recycle(uid string) {
	m.deathChannel <- uid
}

// StartProcess manages a process in manager
func (m *BackgroundProcessManager) StartProcess(ctx context.Context, cmd *ManagedCommand) (*Process, error) {
	log := m.getLoggerFromContext(ctx)
	if cmd.Identifier != nil {
		_, loaded := m.identifiers.LoadOrStore(*cmd.Identifier, true)
		if loaded {
			return nil, errors.Errorf("process with identifier %s is running", *cmd.Identifier)
		}
	}

	process, err := startProcess(cmd)
	if err != nil {
		return nil, err
	}

	m.processes.Store(process.Uid, process)
	m.pidPairToUid.Store(process.Pair, process.Uid)
	// end

	if m.metricsCollector != nil {
		m.metricsCollector.bpmControlledProcessTotal.Inc()
	}

	m.wg.Add(1)
	log = log.WithValues("uid", process.Uid, "pid", process.Pair.Pid)

	go func() {
		err := cmd.Wait()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				status := exitErr.Sys().(syscall.WaitStatus)
				if status.Signaled() && status.Signal() == syscall.SIGTERM {
					log.Info("process stopped with SIGTERM signal")
				}
			} else {
				log.Error(err, "process exited accidentally")
			}
		}
		log.Info("process stopped")
		m.recycle(process.Uid)
	}()

	return process, nil
}

func (m *BackgroundProcessManager) Shutdown(ctx context.Context) {
	log := m.getLoggerFromContext(ctx)

	m.processes.Range(func(_, value interface{}) bool {
		process := value.(*Process)
		log := log.WithValues("uid", process.Uid, "pid", process.Pair.Pid)
		if err := process.Cmd.Process.Signal(syscall.SIGTERM); err != nil {
			log.Error(err, "send SIGTERM to process")
			return true
		}
		return true
	})
	m.wg.Wait()
	close(m.deathChannel)
}

func (m *BackgroundProcessManager) GetUID(pair ProcessPair) (string, bool) {
	if uid, loaded := m.pidPairToUid.Load(pair); loaded {
		return uid.(string), true
	}
	return "", false
}

func (m *BackgroundProcessManager) getProc(uid string) (*Process, bool) {
	if proc, loaded := m.processes.Load(uid); loaded {
		return proc.(*Process), true
	}
	return nil, false
}

func (m *BackgroundProcessManager) GetPipes(uid string) (Pipes, bool) {
	proc, ok := m.getProc(uid)
	if !ok {
		return Pipes{}, false
	}
	return proc.Pipes, true
}

// KillBackgroundProcess sends SIGTERM to process
func (m *BackgroundProcessManager) KillBackgroundProcess(ctx context.Context, uid string) error {
	log := m.getLoggerFromContext(ctx)

	log = log.WithValues("uid", uid)

	proc, loaded := m.getProc(uid)
	if !loaded {
		return errors.Errorf("failed to find process with uid %s", uid)
	}

	if err := proc.Cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return errors.Wrap(err, "send SIGTERM to process")
	}

	select {
	case <-proc.Stopped():
		log.Info("Successfully killed process")
	case <-ctx.Done():
		if err := ctx.Err(); err != nil {
			return errors.Wrap(err, "context closed")
		}
	}
	return nil
}

// GetIdentifiers finds all identifiers in BPM
func (m *BackgroundProcessManager) GetIdentifiers() []string {
	var identifiers []string
	m.identifiers.Range(func(key, value interface{}) bool {
		identifiers = append(identifiers, key.(string))
		return true
	})

	return identifiers
}

func (m *BackgroundProcessManager) getLoggerFromContext(ctx context.Context) logr.Logger {
	return log.EnrichLoggerWithContext(ctx, m.rootLogger)
}

// DefaultProcessBuilder returns the default process builder
func DefaultProcessBuilder(cmd string, args ...string) *CommandBuilder {
	return &CommandBuilder{
		cmd:        cmd,
		args:       args,
		nsOptions:  []nsOption{},
		pause:      false,
		identifier: nil,
		ctx:        context.Background(),
	}
}

// CommandBuilder builds a exec.Cmd for daemon
type CommandBuilder struct {
	cmd  string
	args []string
	env  []string

	nsOptions []nsOption

	pause    bool
	localMnt bool

	identifier *string
	stdin      io.Reader
	stdout     io.Writer
	stderr     io.Writer

	oomScoreAdj int

	// the context is used to kill the process and will be passed into
	// `exec.CommandContext`
	ctx context.Context
}

// GetNsPath returns corresponding namespace path
func GetNsPath(pid uint32, typ NsType) string {
	return fmt.Sprintf("%s/%d/ns/%s", DefaultProcPrefix, pid, string(typ))
}

// SetEnv sets the environment variables of the process
func (b *CommandBuilder) SetEnv(key, value string) *CommandBuilder {
	b.env = append(b.env, fmt.Sprintf("%s=%s", key, value))
	return b
}

// SetNS sets the namespace of the process
func (b *CommandBuilder) SetNS(pid uint32, typ NsType) *CommandBuilder {
	return b.SetNSOpt([]nsOption{{
		Typ:  typ,
		Path: GetNsPath(pid, typ),
	}})
}

// SetNSOpt sets the namespace of the process
func (b *CommandBuilder) SetNSOpt(options []nsOption) *CommandBuilder {
	b.nsOptions = append(b.nsOptions, options...)

	return b
}

// SetIdentifier sets the identifier of the process
//
// The identifier is used to identify the process in BPM, to confirm only one identified process is running.
// If one identified process is already running, new processes with the same identifier will be blocked by lock.
func (b *CommandBuilder) SetIdentifier(id string) *CommandBuilder {
	b.identifier = &id

	return b
}

// EnablePause enables pause for process
func (b *CommandBuilder) EnablePause() *CommandBuilder {
	b.pause = true

	return b
}

func (b *CommandBuilder) EnableLocalMnt() *CommandBuilder {
	b.localMnt = true

	return b
}

// SetContext sets context for process
func (b *CommandBuilder) SetContext(ctx context.Context) *CommandBuilder {
	b.ctx = ctx

	return b
}

// SetStdin sets stdin for process
func (b *CommandBuilder) SetStdin(stdin io.Reader) *CommandBuilder {
	b.stdin = stdin

	return b
}

// SetStdout sets stdout for process
func (b *CommandBuilder) SetStdout(stdout io.Writer) *CommandBuilder {
	b.stdout = stdout

	return b
}

// SetStderr sets stderr for process
func (b *CommandBuilder) SetStderr(stderr io.Writer) *CommandBuilder {
	b.stderr = stderr

	return b
}

// SetOOMScoreAdj sets the oom_score_adj for a process
// oom_score_adj ranges from -1000 to 1000
func (b *CommandBuilder) SetOOMScoreAdj(scoreAdj int) *CommandBuilder {
	b.oomScoreAdj = scoreAdj
	return b
}

func (b *CommandBuilder) getLoggerFromContext(ctx context.Context) logr.Logger {
	// this logger is inherited from the global one
	// TODO: replace it with a specific logger by passing in one or creating a new one
	logger := log.L().WithName("background-process-manager.process-builder")
	return log.EnrichLoggerWithContext(ctx, logger)
}

type nsOption struct {
	Typ  NsType
	Path string
}

// ManagedCommand is a process which can be managed by backgroundProcessManager
type ManagedCommand struct {
	*exec.Cmd

	// If the identifier is not nil, process manager should make sure no other
	// process with this identifier is running when executing this command
	Identifier *string
}
