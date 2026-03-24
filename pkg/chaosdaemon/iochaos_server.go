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
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/creachadair/jrpc2"
	"github.com/creachadair/jrpc2/channel"
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

const (
	todaBin = "/usr/local/bin/toda"
)

func (s *DaemonServer) ApplyIOChaos(ctx context.Context, in *pb.ApplyIOChaosRequest) (*pb.ApplyIOChaosResponse, error) {
	log := s.getLoggerFromContext(ctx)
	log.Info("applying io chaos", "Request", in)

	if in.InstanceUid == "" {
		if uid, ok := s.backgroundProcessManager.GetUID(bpm.ProcessPair{Pid: int(in.Instance), CreateTime: in.StartTime}); ok {
			in.InstanceUid = uid
		}
	}

	if in.InstanceUid != "" {
		err := s.killIOChaos(ctx, in.InstanceUid)
		if err != nil {
			return nil, err
		}
	}

	actions := []v1alpha1.IOChaosAction{}
	err := json.Unmarshal([]byte(in.Actions), &actions)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal json bytes")
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
		return nil, errors.Wrap(err, "getting PID")
	}

	// TODO: make this log level configurable
	args := fmt.Sprintf("--path %s --verbose info", in.Volume)
	log.Info("executing", "cmd", todaBin+" "+args)

	processBuilder := bpm.DefaultProcessBuilder(todaBin, strings.Split(args, " ")...).
		EnableLocalMnt().
		SetIdentifier(fmt.Sprintf("toda-%s", in.ContainerId))

	if in.EnterNS {
		processBuilder = processBuilder.SetNS(pid, bpm.MountNS).SetNS(pid, bpm.PidNS)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := processBuilder.Build(ctx)
	cmd.Stderr = os.Stderr
	proc, err := s.backgroundProcessManager.StartProcess(ctx, cmd)
	if err != nil {
		return nil, errors.Wrapf(err, "start process `%s`", cmd)
	}

	ch := channel.Line(proc.Pipes.Stdout, proc.Pipes.Stdin)
	client := jrpc2.NewClient(ch, nil)

	var result string
	log.Info("Waiting for toda to start")
	var rpcError error
	maxWaitTime := time.Millisecond * 2000
	timeOut, cancel := context.WithTimeout(ctx, maxWaitTime)
	defer cancel()
	if err := client.CallResult(timeOut, "update", []any{actions}, &result); err != nil {
		log.Error(err, "update RPC call failed")

		if kerr := s.killIOChaos(ctx, proc.Uid); kerr != nil {
			log.Error(kerr, "kill toda", "request", in)
		}

		return nil, errors.Wrap(err, "toda update RPC failed")
	}
	rpcError = client.CallResult(timeOut, "get_status", []any{"ping"}, &result)
	if rpcError != nil || result != "ok" {
		log.Info("Starting toda takes too long or encounter an error")

		if kerr := s.killIOChaos(ctx, proc.Uid); kerr != nil {
			log.Error(kerr, "kill toda", "request", in)
		}

		return nil, errors.Errorf("toda startup takes too long or an error occurs: %s", result)
	}

	return &pb.ApplyIOChaosResponse{
		Instance:    int64(proc.Pair.Pid),
		StartTime:   proc.Pair.CreateTime,
		InstanceUid: proc.Uid,
	}, nil
}

func (s *DaemonServer) killIOChaos(ctx context.Context, uid string) error {
	log := s.getLoggerFromContext(ctx)

	err := s.backgroundProcessManager.KillBackgroundProcess(ctx, uid)
	if err != nil {
		return errors.Wrapf(err, "kill toda %s", uid)
	}
	log.Info("kill toda successfully", "uid", uid)
	return nil
}
