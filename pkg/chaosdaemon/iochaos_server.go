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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/shirou/gopsutil/process"

	jrpc "github.com/ethereum/go-ethereum/rpc"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

const (
	todaBin = "/usr/local/bin/toda"
)

func (s *DaemonServer) ApplyIoChaos(ctx context.Context, in *pb.ApplyIoChaosRequest) (*pb.ApplyIoChaosResponse, error) {
	log.Info("applying io chaos", "Request", in)

	if in.Instance != 0 {
		err := s.killIoChaos(ctx, in.Instance, in.StartTime)
		if err != nil {
			return nil, err
		}
	}

	actions := []v1alpha1.IoChaosAction{}
	err := json.Unmarshal([]byte(in.Actions), &actions)
	if err != nil {
		log.Error(err, "error while unmarshal json bytes")
		return nil, err
	}

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

	// TODO: make this log level configurable
	args := fmt.Sprintf("--path %s --verbose info", in.Volume)
	log.Info("executing", "cmd", todaBin+" "+args)

	processBuilder := bpm.DefaultProcessBuilder(todaBin, strings.Split(args, " ")...).
		EnableLocalMnt().
		SetIdentifier(in.ContainerId)

	if in.EnterNS {
		processBuilder = processBuilder.SetNS(pid, bpm.MountNS).SetNS(pid, bpm.PidNS)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var caller, receiver bytes.Buffer
	client, err := jrpc.DialIO(ctx, &receiver, &caller)
	if err != nil {
		return nil, err
	}

	cmd := processBuilder.Build()
	cmd.Stdin = io.MultiReader(strings.NewReader(in.Actions), &caller)
	cmd.Stdout = io.MultiWriter(&receiver, os.Stdout)
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

	log.Info("Waiting for toda to start")
	time.Sleep(time.Second * 3)
	var ret string
	err = client.Call(&ret, "ping", struct{}{})
	if err != nil || ret != "pong" {
		log.Info("Starting toda takes too long")
		return nil, fmt.Errorf("Toda startup takes too long or an error occurs")
	}

	return &pb.ApplyIoChaosResponse{
		Instance:  int64(cmd.Process.Pid),
		StartTime: ct,
	}, nil
}

func (s *DaemonServer) killIoChaos(ctx context.Context, pid int64, startTime int64) error {
	log.Info("killing toda", "pid", pid)

	err := s.backgroundProcessManager.KillBackgroundProcess(ctx, int(pid), startTime)
	if err != nil {
		return err
	}
	log.Info("kill toda successfully")
	return nil
}
