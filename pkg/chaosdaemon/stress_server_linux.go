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
	"github.com/chaos-mesh/chaos-mesh/pkg/log"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/containerd/cgroups"
	"github.com/golang/protobuf/ptypes/empty"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	daemonCgroups "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/cgroups"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

func (s *DaemonServer) ExecStressors(ctx context.Context,
	req *pb.ExecStressRequest) (*pb.ExecStressResponse, error) {
	log.L().WithName(loggerNameDaemonServer).Info("Executing stressors", "request", req)

	// cpuStressors
	cpuInstance, cpuStartTime, err := s.ExecCPUStressors(ctx, req)
	if err != nil {
		return nil, err
	}

	// memoryStressor
	memoryInstance, memoryStartTime, err := s.ExecMemoryStressors(ctx, req)
	if err != nil {
		return nil, err
	}

	return &pb.ExecStressResponse{
		CpuInstance:     cpuInstance,
		CpuStartTime:    cpuStartTime,
		MemoryInstance:  memoryInstance,
		MemoryStartTime: memoryStartTime,
	}, nil
}

func (s *DaemonServer) CancelStressors(ctx context.Context,
	req *pb.CancelStressRequest) (*empty.Empty, error) {
	CpuPid, err := strconv.Atoi(req.CpuInstance)
	if err != nil {
		return nil, err
	}

	MemoryPid, err := strconv.Atoi(req.MemoryInstance)
	if err != nil {
		return nil, err
	}

	log.L().WithName(loggerNameDaemonServer).Info("Canceling stressors", "request", req)

	err = s.backgroundProcessManager.KillBackgroundProcess(ctx, CpuPid, req.CpuStartTime)
	if err != nil {
		return nil, err
	}

	err = s.backgroundProcessManager.KillBackgroundProcess(ctx, MemoryPid, req.MemoryStartTime)
	if err != nil {
		return nil, err
	}

	log.L().WithName(loggerNameDaemonServer).Info("killing stressor successfully")
	return &empty.Empty{}, nil
}

func (s *DaemonServer) ExecCPUStressors(ctx context.Context,
	req *pb.ExecStressRequest) (string, int64, error) {
	pid, err := s.crClient.GetPidFromContainerID(ctx, req.Target)
	if err != nil {
		return "", 0, err
	}
	control, err := cgroups.Load(daemonCgroups.V1, daemonCgroups.PidPath(int(pid)))
	if err != nil {
		return "", 0, err
	}

	processBuilder := bpm.DefaultProcessBuilder("stress-ng", strings.Fields(req.CpuStressors)...).
		EnablePause()
	if req.EnterNS {
		processBuilder = processBuilder.SetNS(pid, bpm.PidNS)
	}
	cmd := processBuilder.Build()

	procState, err := s.backgroundProcessManager.StartProcess(cmd)
	if err != nil {
		return "", 0, err
	}
	log.L().WithName(loggerNameDaemonServer).Info("Start stress-ng successfully")
	ct, err := procState.CreateTime()
	if err != nil {
		return "", 0, err
	}

	if err = control.Add(cgroups.Process{Pid: cmd.Process.Pid}); err != nil {
		if kerr := cmd.Process.Kill(); kerr != nil {
			log.L().WithName(loggerNameDaemonServer).Error(kerr, "kill stress-ng failed", "request", req)
		}
		return "", 0, err
	}

	for {
		// TODO: find a better way to resume pause process
		if err := cmd.Process.Signal(syscall.SIGCONT); err != nil {
			return "", 0, err
		}

		log.L().WithName(loggerNameDaemonServer).Info("send signal to resume process")
		time.Sleep(time.Millisecond)

		comm, err := ReadCommName(cmd.Process.Pid)
		if err != nil {
			return "", 0, err
		}
		if comm != "pause\n" {
			log.L().WithName(loggerNameDaemonServer).Info("pause has been resumed", "comm", comm)
			break
		}
		log.L().WithName(loggerNameDaemonServer).Info("the process hasn't resumed, step into the following loop", "comm", comm)
	}

	return strconv.Itoa(cmd.Process.Pid), ct, nil
}

func (s *DaemonServer) ExecMemoryStressors(ctx context.Context,
	req *pb.ExecStressRequest) (string, int64, error) {
	pid, err := s.crClient.GetPidFromContainerID(ctx, req.Target)
	if err != nil {
		return "", 0, err
	}
	control, err := cgroups.Load(daemonCgroups.V1, daemonCgroups.PidPath(int(pid)))
	if err != nil {
		return "", 0, err
	}
	processBuilder := bpm.DefaultProcessBuilder("memStress", strings.Fields(req.MemoryStressors)...).
		EnablePause()

	if req.EnterNS {
		processBuilder = processBuilder.SetNS(pid, bpm.PidNS)
	}
	cmd := processBuilder.Build()

	procState, err := s.backgroundProcessManager.StartProcess(cmd)
	if err != nil {
		return "", 0, err
	}
	log.L().WithName(loggerNameDaemonServer).Info("Start memStress successfully")
	ct, err := procState.CreateTime()
	if err != nil {
		return "", 0, err
	}

	if err = control.Add(cgroups.Process{Pid: cmd.Process.Pid}); err != nil {
		if kerr := cmd.Process.Kill(); kerr != nil {
			log.L().WithName(loggerNameDaemonServer).Error(kerr, "kill memStress failed", "request", req)
		}
		return "", 0, err
	}

	for {
		// TODO: find a better way to resume pause process
		if err := cmd.Process.Signal(syscall.SIGCONT); err != nil {
			return "", 0, err
		}

		log.L().WithName(loggerNameDaemonServer).Info("send signal to resume process")
		time.Sleep(time.Millisecond)

		comm, err := ReadCommName(cmd.Process.Pid)

		if err != nil {
			return "", 0, err
		}
		if comm != "pause\n" {
			log.L().WithName(loggerNameDaemonServer).Info("pause has been resumed", "comm", comm)
			break
		}
		log.L().WithName(loggerNameDaemonServer).Info("the process hasn't resumed, step into the following loop", "comm", comm)
	}

	return strconv.Itoa(cmd.Process.Pid), ct, nil
}
