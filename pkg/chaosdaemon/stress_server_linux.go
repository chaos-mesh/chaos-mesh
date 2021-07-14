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
)

func (s *DaemonServer) ExecStressors(ctx context.Context,
	req *pb.ExecStressRequest) (*pb.ExecStressResponse, error) {
	log.Info("Executing stressors", "request", req)
	pid, err := s.crClient.GetPidFromContainerID(ctx, req.Target)
	if err != nil {
		return nil, err
	}
	control, err := cgroups.Load(daemonCgroups.V1, daemonCgroups.PidPath(int(pid)))
	if err != nil {
		return nil, err
	}

	processBuilder := bpm.DefaultProcessBuilder("stress-ng", strings.Fields(req.Stressors)...).
		EnablePause()
	if req.EnterNS {
		processBuilder = processBuilder.SetNS(pid, bpm.PidNS)
	}
	cmd := processBuilder.Build()

	procState, err := s.backgroundProcessManager.StartProcess(cmd)
	if err != nil {
		return nil, err
	}
	log.Info("Start process successfully")
	ct, err := procState.CreateTime()
	if err != nil {
		return nil, err
	}

	if err = control.Add(cgroups.Process{Pid: cmd.Process.Pid}); err != nil {
		if kerr := cmd.Process.Kill(); kerr != nil {
			log.Error(kerr, "kill stressors failed", "request", req)
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

		comm, err := ReadCommName(cmd.Process.Pid)
		if err != nil {
			return nil, err
		}
		if comm != "pause\n" {
			log.Info("pause has been resumed", "comm", comm)
			break
		}
		log.Info("the process hasn't resumed, step into the following loop", "comm", comm)
	}

	return &pb.ExecStressResponse{
		Instance:  strconv.Itoa(cmd.Process.Pid),
		StartTime: ct,
	}, nil
}

func (s *DaemonServer) CancelStressors(ctx context.Context,
	req *pb.CancelStressRequest) (*empty.Empty, error) {
	pid, err := strconv.Atoi(req.Instance)
	if err != nil {
		return nil, err
	}
	log.Info("Canceling stressors", "request", req)

	err = s.backgroundProcessManager.KillBackgroundProcess(ctx, pid, req.StartTime)
	if err != nil {
		return nil, err
	}
	log.Info("killing stressor successfully")
	return &empty.Empty{}, nil
}
