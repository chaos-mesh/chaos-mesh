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
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/shirou/gopsutil/process"
	"golang.org/x/sys/unix"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

const (
	todaBin = "/usr/local/bin/toda"
)

func (s *daemonServer) ApplyIoChaos(ctx context.Context, in *pb.ApplyIoChaosRequest) (*pb.ApplyIoChaosResponse, error) {
	log.Info("applying io chaos", "Request", in)

	if in.Instance != 0 {
		err := s.killIoChaos(ctx, in.Instance, in.StartTime)
		if err != nil {
			return nil, err
		}
	}

	actions := []v1alpha1.IoChaosAction{}
	json.Unmarshal([]byte(in.Actions), &actions)
	log.Info("the length of actions", "length", len(actions))
	if len(actions) == 0 {
		return &pb.ApplyIoChaosResponse{
			Instance:  0,
			StartTime: 0,
		}, nil
	}

	pid, err := s.crClient.GetPidFromContainerID(ctx, in.ContainerId)
	if err != nil {
		log.Error(err, "error while getting PID")
		return nil, err
	}

	args := fmt.Sprintf("--path %s --pid %d --verbose info", in.Volume, pid)
	log.Info("executing", "cmd", todaBin+" "+args)
	cmd := exec.CommandContext(context.Background(), todaBin, strings.Split(args, " ")...)
	cmd.Stdin = strings.NewReader(in.Actions)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
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

	return &pb.ApplyIoChaosResponse{
		Instance:  int64(cmd.Process.Pid),
		StartTime: ct,
	}, nil
}

func (s *daemonServer) killIoChaos(ctx context.Context, pid int64, startTime int64) error {
	log.Info("killing toda", "pid", pid)

	ins, err := process.NewProcess(int32(pid))
	if err != nil {
		log.Info("cannot find process", "pid", pid)
		return nil
	}
	if ct, err := ins.CreateTime(); err == nil && ct == startTime {
		_, err := ins.Status()
		for err != nil {
			log.Error(err, "get status error", "pid", pid)
			// FIXME: only ignore NOT FOUND error
			return nil
		}

		err = ins.SendSignal(unix.SIGTERM)
		if err != nil {
			log.Error(err, "kill error", "pid", pid)
			// FIXME: only ignore NOT FOUND error
			return nil
		}

		log.Info("kill process and wait 1 second", "pid", pid)
		time.Sleep(1 * time.Second)
	} else {
		log.Info("find different process", "createTime", ct, "expectedCreateTime", startTime)
	}

	return nil
}
