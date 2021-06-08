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
	"io"
	"os"
	"strings"
	"time"

	jrpc "github.com/ethereum/go-ethereum/rpc"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

const (
	todaBin = "/usr/local/bin/toda"
)

func (s *DaemonServer) ApplyIOChaos(ctx context.Context, in *pb.ApplyIOChaosRequest) (*pb.ApplyIOChaosResponse, error) {
	log.Info("applying io chaos", "Request", in)

	if in.Instance != 0 {
		err := s.killIOChaos(ctx, in.Instance, in.StartTime)
		if err != nil {
			return nil, err
		}
	}

	actions := []v1alpha1.IOChaosAction{}
	err := json.Unmarshal([]byte(in.Actions), &actions)
	if err != nil {
		log.Error(err, "error while unmarshal json bytes")
		return nil, err
	}

	log.Info("the length of actions", "length", len(actions))
	if len(actions) == 0 {
		return &pb.ApplyIOChaosResponse{
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
	caller, receiver := bpm.NewBlockingBuffer(), bpm.NewBlockingBuffer()
	defer caller.Close()
	defer receiver.Close()
	client, err := jrpc.DialIO(ctx, receiver, caller)
	if err != nil {
		return nil, err
	}

	cmd := processBuilder.Build()
	cmd.Stdin = caller
	cmd.Stdout = io.MultiWriter(receiver, os.Stdout)
	cmd.Stderr = os.Stderr
	procState, err := s.backgroundProcessManager.StartProcess(cmd)
	if err != nil {
		return nil, err
	}
	var ret string
	ct, err := procState.CreateTime()
	if err != nil {
		log.Error(err, "get create time failed")
		if kerr := cmd.Process.Kill(); kerr != nil {
			log.Error(kerr, "kill toda failed", "request", in)
		}
		return nil, err
	}

	log.Info("Waiting for toda to start")
	var rpcError error
	maxWaitTime := time.Millisecond * 2000
	timeOut, cancel := context.WithTimeout(ctx, maxWaitTime)
	defer cancel()
	_ = client.CallContext(timeOut, &ret, "update", actions)
	rpcError = client.CallContext(timeOut, &ret, "get_status", "ping")
	if rpcError != nil || ret != "ok" {
		log.Info("Starting toda takes too long or encounter an error")
		caller.Close()
		receiver.Close()
		if kerr := s.killIOChaos(ctx, int64(cmd.Process.Pid), ct); kerr != nil {
			log.Error(kerr, "kill toda failed", "request", in)
		}
		return nil, fmt.Errorf("toda startup takes too long or an error occurs: %s", ret)
	}

	return &pb.ApplyIOChaosResponse{
		Instance:  int64(cmd.Process.Pid),
		StartTime: ct,
	}, nil
}

func (s *DaemonServer) killIOChaos(ctx context.Context, pid int64, startTime int64) error {
	log.Info("killing toda", "pid", pid)

	err := s.backgroundProcessManager.KillBackgroundProcess(ctx, int(pid), startTime)
	if err != nil {
		return err
	}
	log.Info("kill toda successfully")
	return nil
}
