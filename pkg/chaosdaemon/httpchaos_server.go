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
	"encoding/json"
	"os"
	"strings"

	"github.com/shirou/gopsutil/process"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

const (
	tproxyBin = "/usr/local/bin/tproxy"
)

func (s *DaemonServer) ApplyHttpChaos(ctx context.Context, in *pb.ApplyHttpChaosRequest) (*pb.ApplyHttpChaosResponse, error) {
	log.Info("applying io chaos", "Request", in)

	if in.Instance != 0 {
		err := s.killHttpChaos(ctx, in.Instance, in.StartTime)
		if err != nil {
			return nil, err
		}
	}

	rules := []*v1alpha1.PodHttpChaosRule{}
	err := json.Unmarshal([]byte(in.Rules), &rules)
	if err != nil {
		log.Error(err, "error while unmarshal json bytes")
		return nil, err
	}

	log.Info("the length of actions", "length", len(rules))
	if len(rules) == 0 {
		return &pb.ApplyHttpChaosResponse{
			Instance:  0,
			StartTime: 0,
		}, nil
	}

	pid, err := s.crClient.GetPidFromContainerID(ctx, in.ContainerId)
	if err != nil {
		log.Error(err, "error while getting PID")
		return nil, err
	}

	// TODO: make this log level configurable
	log.Info("executing", "cmd", tproxyBin)

	processBuilder := bpm.DefaultProcessBuilder(tproxyBin).
		EnableLocalMnt().
		SetIdentifier(in.ContainerId)

	if in.EnterNS {
		processBuilder = processBuilder.SetNS(pid, bpm.MountNS).SetNS(pid, bpm.PidNS)
	}

	cmd := processBuilder.Build()
	cmd.Stdin = strings.NewReader("")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = s.backgroundProcessManager.StartProcess(cmd)
	if err != nil {
		return nil, err
	}

	procState, err := process.NewProcess(int32(cmd.Process.Pid))
	if err != nil {
		return nil, err
	}
	ct, err := procState.CreateTime()
	if err != nil {
		if kerr := cmd.Process.Kill(); kerr != nil {
			log.Error(kerr, "kill toda failed", "request", in)
		}
		return nil, err
	}

	return &pb.ApplyHttpChaosResponse{
		Instance:  int64(cmd.Process.Pid),
		StartTime: ct,
	}, nil
}

func (s *DaemonServer) killHttpChaos(ctx context.Context, pid int64, startTime int64) error {
	log.Info("killing tproxy", "pid", pid)

	err := s.backgroundProcessManager.KillBackgroundProcess(ctx, int(pid), startTime)
	if err != nil {
		return err
	}
	log.Info("kill tproxy successfully")
	return nil
}
