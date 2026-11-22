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

package chaosdaemon

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/cgroups"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/util"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

func (s *DaemonServer) ExecStressors(ctx context.Context,
	req *pb.ExecStressRequest) (*pb.ExecStressResponse, error) {
	log := s.getLoggerFromContext(ctx)
	log.Info("Executing stressors", "request", req)

	// cpuStressors
	cpuProc, err := s.ExecCPUStressors(ctx, req)
	if err != nil {
		// this error would be resolved by grpc server framework, it's the top level to print it.
		s.rootLogger.Error(err, "exec cpu stressors", "containerID", req.Target, "cpuStressors", req.CpuStressors)
		return nil, err
	}

	// memoryStressor
	memoryProc, err := s.ExecMemoryStressors(ctx, req)
	if err != nil {
		if cpuProc != nil {
			err = errors.Join(err, stopUnreportedStressor(ctx, cpuProc))
		}
		return nil, err
	}

	resp := new(pb.ExecStressResponse)
	if cpuProc != nil {
		resp.CpuInstance = strconv.Itoa(cpuProc.Pair.Pid)
		resp.CpuStartTime = cpuProc.Pair.CreateTime
		resp.CpuInstanceUid = cpuProc.UID
	}
	if memoryProc != nil {
		resp.MemoryInstance = strconv.Itoa(memoryProc.Pair.Pid)
		resp.MemoryStartTime = memoryProc.Pair.CreateTime
		resp.MemoryInstanceUid = memoryProc.UID
	}

	return resp, nil
}

func (s *DaemonServer) CancelStressors(ctx context.Context,
	req *pb.CancelStressRequest) (*empty.Empty, error) {
	log := s.getLoggerFromContext(ctx)
	CPUPID, err := strconv.Atoi(req.CpuInstance)
	if req.CpuInstance != "" && err != nil {
		return nil, err
	}

	MemoryPid, err := strconv.Atoi(req.MemoryInstance)
	if req.MemoryInstance != "" && err != nil {
		return nil, err
	}

	if req.CpuInstanceUid == "" && CPUPID != 0 {
		if uid, ok := s.backgroundProcessManager.GetUID(bpm.ProcessPair{Pid: CPUPID, CreateTime: req.CpuStartTime}); ok {
			req.CpuInstanceUid = uid
		}
	}

	if req.MemoryInstanceUid == "" && MemoryPid != 0 {
		if uid, ok := s.backgroundProcessManager.GetUID(bpm.ProcessPair{Pid: MemoryPid, CreateTime: req.MemoryStartTime}); ok {
			req.MemoryInstanceUid = uid
		}
	}

	log.Info("Canceling stressors", "request", req)

	if req.CpuInstanceUid != "" {
		err = s.backgroundProcessManager.KillBackgroundProcess(ctx, req.CpuInstanceUid)
		if err != nil {
			return nil, err
		}
	}

	if req.MemoryInstanceUid != "" {
		err = s.backgroundProcessManager.KillBackgroundProcess(ctx, req.MemoryInstanceUid)
		if err != nil {
			return nil, err
		}
	}

	log.Info("killing stressor successfully")
	return &empty.Empty{}, nil
}

func (s *DaemonServer) ExecCPUStressors(ctx context.Context, req *pb.ExecStressRequest) (*bpm.Process, error) {
	return s.execStressor(ctx, req, "stress-ng", req.CpuStressors, 0)
}

func (s *DaemonServer) ExecMemoryStressors(ctx context.Context, req *pb.ExecStressRequest) (*bpm.Process, error) {
	return s.execStressor(ctx, req, "memStress", req.MemoryStressors, req.OomScoreAdj)
}

func (s *DaemonServer) execStressor(ctx context.Context, req *pb.ExecStressRequest, executable, args string, oomScoreAdj int32) (_ *bpm.Process, err error) {
	if args == "" {
		return nil, nil
	}
	pid, err := s.crClient.GetPidFromContainerID(ctx, req.Target)
	if err != nil {
		return nil, err
	}

	getAttacher := cgroups.GetAttacherForPID
	if resolve := mock.On("MockStressCGroupAttacher"); resolve != nil {
		getAttacher = resolve.(func(int) (cgroups.AttachCGroup, error))
	}
	attacher, err := getAttacher(int(pid))
	if err != nil {
		return nil, err
	}

	builder := bpm.DefaultProcessBuilder(executable, strings.Fields(args)...).EnablePause().SetOOMScoreAdj(int(oomScoreAdj))
	if req.EnterNS {
		builder = builder.SetNS(pid, bpm.PidNS)
	}
	cmd := builder.Build(ctx)
	proc, err := s.backgroundProcessManager.StartProcess(ctx, cmd)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, stopUnreportedStressor(ctx, proc))
		}
	}()

	log := s.getLoggerFromContext(ctx)
	log.Info("Started stressor", "command", cmd, "pid", proc.Pair.Pid, "uid", proc.UID)
	if err := attacher.AttachProcess(proc.Pair.Pid); err != nil {
		return nil, err
	}

	// The pause helper must resume after entering the target cgroup.
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if err := cmd.Process.Signal(syscall.SIGCONT); err != nil {
			return nil, err
		}
		time.Sleep(time.Millisecond)
		comm, err := util.ReadCommName(proc.Pair.Pid)
		if err != nil {
			return nil, err
		}
		if comm != "pause\n" {
			break
		}
	}

	return proc, nil
}

// stopUnreportedStressor terminates a process that the controller cannot
// recover because its identifier has not been sent yet.
func stopUnreportedStressor(ctx context.Context, proc *bpm.Process) error {
	// Give a running stressor time to stop its workers.
	if err := proc.Cmd.Process.Signal(syscall.SIGTERM); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return fmt.Errorf("stop unreported stressor %s: %w", proc.UID, err)
	}

	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	select {
	case <-proc.Stopped():
		return nil
	case <-cleanupCtx.Done():
		if err := proc.Cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			return fmt.Errorf("kill unreported stressor %s: %w", proc.UID, err)
		}
	}

	select {
	case <-proc.Stopped():
		return nil
	case <-time.After(time.Second):
		return fmt.Errorf("wait for unreported stressor %s: %w", proc.UID, cleanupCtx.Err())
	}
}
