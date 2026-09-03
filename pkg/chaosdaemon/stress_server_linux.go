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
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/cgroups"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/util"
)

const stressRollbackTimeout = 10 * time.Second

type stressorExecutor interface {
	ExecCPUStressors(context.Context, *pb.ExecStressRequest) (*bpm.Process, error)
	ExecMemoryStressors(context.Context, *pb.ExecStressRequest) (*bpm.Process, error)
	CancelStressors(context.Context, *pb.CancelStressRequest) (*empty.Empty, error)
}

func (s *DaemonServer) ExecStressors(ctx context.Context,
	req *pb.ExecStressRequest) (*pb.ExecStressResponse, error) {
	log := s.getLoggerFromContext(ctx)
	log.Info("Executing stressors", "request", req)

	resp, err := execStressors(ctx, req, s)
	if err != nil {
		log.Error(err, "exec stressors", "containerID", req.Target)
	}
	return resp, err
}

func execStressors(ctx context.Context, req *pb.ExecStressRequest, executor stressorExecutor) (*pb.ExecStressResponse, error) {
	cpuProc, err := executor.ExecCPUStressors(ctx, req)
	if err != nil {
		return nil, err
	}

	memoryProc, err := executor.ExecMemoryStressors(ctx, req)
	if err != nil {
		if cpuProc != nil {
			// No process handles reach the controller on failure. Stop the CPU
			// worker even if cancellation of the request caused memory startup to fail.
			cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), stressRollbackTimeout)
			defer cancel()
			_, cleanupErr := executor.CancelStressors(cleanupCtx, &pb.CancelStressRequest{CpuInstanceUid: cpuProc.Uid})
			if cleanupErr != nil {
				return nil, errors.Wrapf(err, "roll back CPU stressor %s: %v", cpuProc.Uid, cleanupErr)
			}
		}
		return nil, err
	}

	resp := new(pb.ExecStressResponse)
	if cpuProc != nil {
		resp.CpuInstance = strconv.Itoa(cpuProc.Pair.Pid)
		resp.CpuStartTime = cpuProc.Pair.CreateTime
		resp.CpuInstanceUid = cpuProc.Uid
	}
	if memoryProc != nil {
		resp.MemoryInstance = strconv.Itoa(memoryProc.Pair.Pid)
		resp.MemoryStartTime = memoryProc.Pair.CreateTime
		resp.MemoryInstanceUid = memoryProc.Uid
	}

	return resp, nil
}

func (s *DaemonServer) CancelStressors(ctx context.Context,
	req *pb.CancelStressRequest) (*empty.Empty, error) {
	log := s.getLoggerFromContext(ctx)
	CpuPid, err := strconv.Atoi(req.CpuInstance)
	if req.CpuInstance != "" && err != nil {
		return nil, err
	}

	MemoryPid, err := strconv.Atoi(req.MemoryInstance)
	if req.MemoryInstance != "" && err != nil {
		return nil, err
	}

	if req.CpuInstanceUid == "" && CpuPid != 0 {
		if uid, ok := s.backgroundProcessManager.GetUID(bpm.ProcessPair{Pid: CpuPid, CreateTime: req.CpuStartTime}); ok {
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

func (s *DaemonServer) ExecCPUStressors(ctx context.Context,
	req *pb.ExecStressRequest) (*bpm.Process, error) {
	log := s.getLoggerFromContext(ctx)
	if req.CpuStressors == "" {
		return nil, nil
	}
	pid, err := s.crClient.GetPidFromContainerID(ctx, req.Target)
	if err != nil {
		return nil, err
	}

	attachCGroup, err := cgroups.GetAttacherForPID(int(pid))
	if err != nil {
		return nil, err
	}

	processBuilder := bpm.DefaultProcessBuilder("stress-ng", strings.Fields(req.CpuStressors)...).
		EnablePause()
	if req.EnterNS {
		processBuilder = processBuilder.SetNS(pid, bpm.PidNS)
	}
	cmd := processBuilder.Build(ctx)

	proc, err := s.backgroundProcessManager.StartProcess(ctx, cmd)
	if err != nil {
		return nil, err
	}
	log.Info("Start stress-ng successfully", "command", cmd, "pid", proc.Pair.Pid, "uid", proc.Uid)

	if err = attachCGroup.AttachProcess(proc.Pair.Pid); err != nil {
		if kerr := cmd.Process.Kill(); kerr != nil {
			log.Error(kerr, "kill stress-ng failed", "request", req)
		}
		return nil, err
	}

	for {
		// TODO: find a better way to resume pause process
		if err := cmd.Process.Signal(syscall.SIGCONT); err != nil {
			return nil, err
		}

		log.Info("send signal to resume process")
		time.Sleep(time.Millisecond)

		// TODO: integrate the resume process into the bpm
		comm, err := util.ReadCommName(cmd.Process.Pid)
		if err != nil {
			return nil, err
		}
		if comm != "pause\n" {
			log.Info("pause has been resumed", "comm", comm)
			break
		}
		log.Info("the process hasn't resumed, step into the following loop", "comm", comm)
	}

	return proc, nil
}

func (s *DaemonServer) ExecMemoryStressors(ctx context.Context,
	req *pb.ExecStressRequest) (*bpm.Process, error) {
	log := s.getLoggerFromContext(ctx)
	if req.MemoryStressors == "" {
		return nil, nil
	}
	pid, err := s.crClient.GetPidFromContainerID(ctx, req.Target)
	if err != nil {
		return nil, err
	}

	attachCGroup, err := cgroups.GetAttacherForPID(int(pid))
	if err != nil {
		return nil, err
	}

	processBuilder := bpm.DefaultProcessBuilder("memStress", strings.Fields(req.MemoryStressors)...).
		EnablePause()

	if req.OomScoreAdj != 0 {
		processBuilder = processBuilder.SetOOMScoreAdj(int(req.OomScoreAdj))
	}
	if req.EnterNS {
		processBuilder = processBuilder.SetNS(pid, bpm.PidNS)
	}
	cmd := processBuilder.Build(ctx)

	proc, err := s.backgroundProcessManager.StartProcess(ctx, cmd)
	if err != nil {
		return nil, err
	}
	log.Info("Start memStress successfully", "command", cmd, "pid", proc.Pair.Pid, "uid", proc.Uid)

	if err = attachCGroup.AttachProcess(proc.Pair.Pid); err != nil {
		if kerr := cmd.Process.Kill(); kerr != nil {
			log.Error(kerr, "kill memStress failed", "request", req)
		}
		return nil, err
	}

	for {
		// TODO: find a better way to resume pause process
		if err := cmd.Process.Signal(syscall.SIGCONT); err != nil {
			return nil, err
		}

		log.Info("send signal to resume process")
		time.Sleep(time.Millisecond)
		comm, err := util.ReadCommName(proc.Pair.Pid)

		if err != nil {
			return nil, err
		}
		if comm != "pause\n" {
			log.Info("pause has been resumed", "comm", comm)
			break
		}
		log.Info("the process hasn't resumed, step into the following loop", "comm", comm)
	}

	return proc, nil
}
