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

	"github.com/containerd/cgroups"
	"github.com/golang/protobuf/ptypes/empty"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	daemonCgroups "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/cgroups"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/util"
)

func (s *DaemonServer) ExecStressors(ctx context.Context,
	req *pb.ExecStressRequest) (*pb.ExecStressResponse, error) {
	log.Info("Executing stressors", "request", req)

	// cpuStressors
	cpuProc, err := s.ExecCPUStressors(ctx, req)
	if err != nil {
		return nil, err
	}

	// memoryStressor
	memoryProc, err := s.ExecMemoryStressors(ctx, req)
	if err != nil {
		return nil, err
	}

	return &pb.ExecStressResponse{
		CpuInstance:       strconv.Itoa(cpuProc.Pair.Pid),
		CpuStartTime:      cpuProc.Pair.CreateTime,
		CpuInstanceUid:    cpuProc.Uid,
		MemoryInstance:    strconv.Itoa(memoryProc.Pair.Pid),
		MemoryStartTime:   memoryProc.Pair.CreateTime,
		MemoryInstanceUid: memoryProc.Uid,
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

	log.Info("Canceling stressors", "request", req)

	if req.CpuInstanceUid == "" {
		if uid, ok := s.backgroundProcessManager.GetUID(bpm.ProcessPair{Pid: CpuPid, CreateTime: req.CpuStartTime}); ok {
			req.CpuInstanceUid = uid
		}
	}

	if req.MemoryInstanceUid == "" {
		if uid, ok := s.backgroundProcessManager.GetUID(bpm.ProcessPair{Pid: MemoryPid, CreateTime: req.MemoryStartTime}); ok {
			req.MemoryInstanceUid = uid
		}
	}

	err = s.backgroundProcessManager.KillBackgroundProcess(ctx, req.CpuInstanceUid)
	if err != nil {
		return nil, err
	}

	err = s.backgroundProcessManager.KillBackgroundProcess(ctx, req.MemoryInstanceUid)
	if err != nil {
		return nil, err
	}

	log.Info("killing stressor successfully")
	return &empty.Empty{}, nil
}

func (s *DaemonServer) ExecCPUStressors(ctx context.Context,
	req *pb.ExecStressRequest) (*bpm.Process, error) {
	pid, err := s.crClient.GetPidFromContainerID(ctx, req.Target)
	if err != nil {
		return nil, err
	}
	control, err := cgroups.Load(daemonCgroups.V1, daemonCgroups.PidPath(int(pid)))
	if err != nil {
		return nil, err
	}

	processBuilder := bpm.DefaultProcessBuilder("stress-ng", strings.Fields(req.CpuStressors)...).
		EnablePause()
	if req.EnterNS {
		processBuilder = processBuilder.SetNS(pid, bpm.PidNS)
	}
	cmd := processBuilder.Build()

	proc, err := s.backgroundProcessManager.StartProcess(cmd)
	if err != nil {
		return nil, err
	}
	log.Info("Start stress-ng successfully")

	if err = control.Add(cgroups.Process{Pid: proc.Pair.Pid}); err != nil {
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
	pid, err := s.crClient.GetPidFromContainerID(ctx, req.Target)
	if err != nil {
		return nil, err
	}
	control, err := cgroups.Load(daemonCgroups.V1, daemonCgroups.PidPath(int(pid)))
	if err != nil {
		return nil, err
	}
	processBuilder := bpm.DefaultProcessBuilder("memStress", strings.Fields(req.MemoryStressors)...).
		EnablePause()

	if req.EnterNS {
		processBuilder = processBuilder.SetNS(pid, bpm.PidNS)
	}
	cmd := processBuilder.Build()

	proc, err := s.backgroundProcessManager.StartProcess(cmd)
	if err != nil {
		return nil, err
	}
	log.Info("Start memStress successfully")

	if err = control.Add(cgroups.Process{Pid: proc.Pair.Pid}); err != nil {
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
