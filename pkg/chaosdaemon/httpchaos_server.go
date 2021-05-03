// Copyright 2021 Chaos Mesh Authors.
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

	"github.com/shirou/gopsutil/process"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

const (
	tproxyBin    = "/usr/local/bin/tproxy"
	rustLog      = "RUST_LOG"
	rustLogLevel = "trace"
	pathEnv      = "PATH"
)

func (s *DaemonServer) ApplyHttpChaos(ctx context.Context, in *pb.ApplyHttpChaosRequest) (*pb.ApplyHttpChaosResponse, error) {
	log.Info("applying http chaos", "Request", in)

	if in.Instance == 0 {
		if err := s.createHttpChaos(ctx, in); err != nil {
			return nil, err
		}
	}

	stdin := s.backgroundProcessManager.Stdio(int(in.Instance), in.StartTime).Stdin
	if stdin == nil {
		return nil, fmt.Errorf("fail to get stdin of process")
	}

	rules := []*v1alpha1.PodHttpChaosRule{}
	err := json.Unmarshal([]byte(in.Rules), &rules)
	if err != nil {
		log.Error(err, "error while unmarshal json bytes")
		return nil, err
	}

	log.Info("the length of actions", "length", len(rules))

	httpChaosSpec := v1alpha1.PodHttpChaosSpec{
		ProxyPorts: make([]int32, 0, len(in.ProxyPorts)),
		Rules:      rules,
	}
	for _, port := range in.ProxyPorts {
		httpChaosSpec.ProxyPorts = append(httpChaosSpec.ProxyPorts, int32(port))
	}

	config, err := json.Marshal(&httpChaosSpec)
	if err != nil {
		return nil, err
	}

	_, err = stdin.Write([]byte(fmt.Sprintf("%s\n", string(config))))
	if err != nil {
		return nil, err
	}

	return &pb.ApplyHttpChaosResponse{
		Instance:  int64(in.Instance),
		StartTime: in.StartTime,
	}, nil
}

func (s *DaemonServer) createHttpChaos(ctx context.Context, in *pb.ApplyHttpChaosRequest) error {
	pid, err := s.crClient.GetPidFromContainerID(ctx, in.ContainerId)
	if err != nil {
		log.Error(err, "error while getting PID")
		return err
	}
	processBuilder := bpm.DefaultProcessBuilder(tproxyBin).
		EnableLocalMnt().
		SetIdentifier(in.ContainerId).
		SetEnv(rustLog, rustLogLevel).
		SetEnv(pathEnv, os.Getenv(pathEnv)).
		SetStdin(bpm.NewConcurrentBuffer())

	if in.EnterNS {
		processBuilder = processBuilder.SetNS(pid, bpm.PidNS).SetNS(pid, bpm.NetNS)
	}

	cmd := processBuilder.Build()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = s.backgroundProcessManager.StartProcess(cmd)
	if err != nil {
		return err
	}

	procState, err := process.NewProcess(int32(cmd.Process.Pid))
	if err != nil {
		log.Error(err, "new process failed")
		return err
	}
	ct, err := procState.CreateTime()
	if err != nil {
		log.Error(err, "get create time failed")
		if kerr := cmd.Process.Kill(); kerr != nil {
			log.Error(kerr, "kill tproxy failed", "request", in)
		}
		return err
	}

	in.Instance = int64(cmd.Process.Pid)
	in.StartTime = ct
	return nil
}
