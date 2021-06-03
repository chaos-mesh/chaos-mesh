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
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

const (
	tproxyBin = "/usr/local/bin/tproxy"
	pathEnv   = "PATH"
)

type stdioTransport struct {
	stdio *bpm.Stdio
}

type tproxyConfig struct {
	ProxyPorts []uint32                        `json:"proxy_ports"`
	Rules      []v1alpha1.PodHttpChaosBaseRule `json:"rules"`
}

func (t stdioTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	t.stdio.Lock()
	defer t.stdio.Unlock()

	if t.stdio.Stdin == nil {
		return nil, fmt.Errorf("fail to get stdin of process")
	}
	if t.stdio.Stdout == nil {
		return nil, fmt.Errorf("fail to get stdout of process")
	}

	err = req.Write(t.stdio.Stdin)
	if err != nil {
		return
	}

	resp, err = http.ReadResponse(bufio.NewReader(t.stdio.Stdout), req)
	return
}

func (s *DaemonServer) ApplyHttpChaos(ctx context.Context, in *pb.ApplyHttpChaosRequest) (*pb.ApplyHttpChaosResponse, error) {
	log := log.WithValues("Request", in)
	log.Info("applying http chaos")

	if in.Instance == 0 {
		if err := s.createHttpChaos(ctx, in); err != nil {
			return nil, err
		}
	}

	stdio := s.backgroundProcessManager.Stdio(int(in.Instance), in.StartTime)
	if stdio == nil {
		return nil, fmt.Errorf("fail to get stdio of process")
	}

	transport := stdioTransport{stdio: stdio}

	rules := []v1alpha1.PodHttpChaosBaseRule{}
	err := json.Unmarshal([]byte(in.Rules), &rules)
	if err != nil {
		log.Error(err, "error while unmarshal json bytes")
		return nil, err
	}

	log.Info("the length of actions", "length", len(rules))

	httpChaosSpec := tproxyConfig{
		ProxyPorts: append([]uint32{}, in.ProxyPorts...),
		Rules:      rules,
	}

	config, err := json.Marshal(&httpChaosSpec)
	if err != nil {
		return nil, err
	}

	log.Info("ready to apply", "config", string(config))

	req, err := http.NewRequest(http.MethodPut, "/", bytes.NewReader(config))
	if err != nil {
		return nil, err
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	log.Info("http chaos applied")

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &pb.ApplyHttpChaosResponse{
		Instance:   int64(in.Instance),
		StartTime:  in.StartTime,
		StatusCode: int32(resp.StatusCode),
		Error:      string(body),
	}, nil
}

func (s *DaemonServer) createHttpChaos(ctx context.Context, in *pb.ApplyHttpChaosRequest) error {
	pid, err := s.crClient.GetPidFromContainerID(ctx, in.ContainerId)
	if err != nil {
		log.Error(err, "error while getting PID")
		return err
	}
	processBuilder := bpm.DefaultProcessBuilder(tproxyBin, "-i", "-vv").
		EnableLocalMnt().
		SetIdentifier(in.ContainerId).
		SetEnv(pathEnv, os.Getenv(pathEnv)).
		SetStdin(bpm.NewBlockingBuffer()).
		SetStdout(bpm.NewBlockingBuffer())

	if in.EnterNS {
		processBuilder = processBuilder.SetNS(pid, bpm.PidNS).SetNS(pid, bpm.NetNS)
	}

	cmd := processBuilder.Build()
	cmd.Stderr = os.Stderr

	procState, err := s.backgroundProcessManager.StartProcess(cmd)
	if err != nil {
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
