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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

const (
	todaBin                      = "/usr/local/bin/toda"
	todaUnixSocketFilePath       = "/toda.sock"
	todaClientUnixScoketFilePath = "/proc/%d/root/toda.sock"
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
		if err := s.killIOChaos(ctx, in.InstanceUid); err != nil {
			// ignore this error
			log.Error(err, "kill background process", "uid", in.InstanceUid)
		}
	}

	if err := s.createIOChaos(ctx, in); err != nil {
		return nil, errors.Wrap(err, "create IO chaos")
	}

	log.Info("Waiting for toda to start")
	resp, err := s.applyIOChaos(ctx, in)
	if err != nil {
		if kerr := s.killIOChaos(ctx, in.InstanceUid); kerr != nil {
			log.Error(kerr, "kill toda", "request", in)
		}
		return nil, errors.Wrap(err, "apply config")
	}
	return resp, err
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

func (s *DaemonServer) applyIOChaos(ctx context.Context, in *pb.ApplyIOChaosRequest) (*pb.ApplyIOChaosResponse, error) {
	log := s.getLoggerFromContext(ctx)

	pid, err := s.crClient.GetPidFromContainerID(ctx, in.ContainerId)
	if err != nil {
		return nil, errors.Wrap(err, "getting PID")
	}

	transport := &unixSocketTransport{
		addr: fmt.Sprintf(todaClientUnixScoketFilePath, pid),
	}

	req, err := http.NewRequest(http.MethodPut, "http://psedo-host/update", bytes.NewReader([]byte(in.Actions)))
	if err != nil {
		return nil, errors.Wrap(err, "create http://psedo-host/update request")
	}

	_, _ = transport.RoundTrip(req)

	req, err = http.NewRequest(http.MethodPut, "http://psedo-host/get_status", bytes.NewReader([]byte("ping")))
	if err != nil {
		return nil, errors.Wrap(err, "create http://psedo-host/get_status request")
	}
	resp, err := transport.RoundTrip(req)
	if err != nil {
		return nil, errors.Wrap(err, "send http request")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil || string(body) != "ok" {
		return nil, errors.Wrap(err, "toda startup takes too long or an error occurs")
	}

	log.Info("http chaos applied")

	return &pb.ApplyIOChaosResponse{
		Instance:    in.Instance,
		StartTime:   in.StartTime,
		InstanceUid: in.InstanceUid,
	}, nil
}

func (s *DaemonServer) createIOChaos(ctx context.Context, in *pb.ApplyIOChaosRequest) error {
	log := s.getLoggerFromContext(ctx)

	var actions []v1alpha1.IOChaosAction
	err := json.Unmarshal([]byte(in.Actions), &actions)
	if err != nil {
		return errors.Wrap(err, "unmarshal json bytes")
	}

	log.Info("the length of actions", "length", len(actions))
	if len(actions) == 0 {
		return nil
	}

	pid, err := s.crClient.GetPidFromContainerID(ctx, in.ContainerId)
	if err != nil {
		return errors.Wrap(err, "getting PID")
	}

	// TODO: make this log level configurable
	args := fmt.Sprintf("--path %s --verbose info --unix-socket-path %s", in.Volume, fmt.Sprintf(todaUnixSocketFilePath))
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
		return errors.Wrapf(err, "start process `%s`", cmd)
	}

	in.Instance = int64(proc.Pair.Pid)
	in.StartTime = proc.Pair.CreateTime
	in.InstanceUid = proc.Uid
	return nil

}
