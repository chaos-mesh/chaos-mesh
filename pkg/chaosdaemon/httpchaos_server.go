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
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/chaos-mesh/chaos-mesh/pkg/log"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/tproxyconfig"
)

const (
	tproxyBin = "/usr/local/bin/tproxy"
	pathEnv   = "PATH"
)

type stdioTransport struct {
	stdio *bpm.Stdio
}

func (t stdioTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	t.stdio.Lock()
	defer t.stdio.Unlock()

	if t.stdio.Stdin == nil {
		return nil, errors.New("fail to get stdin of process")
	}
	if t.stdio.Stdout == nil {
		return nil, errors.New("fail to get stdout of process")
	}

	err = req.Write(t.stdio.Stdin)
	if err != nil {
		return
	}

	resp, err = http.ReadResponse(bufio.NewReader(t.stdio.Stdout), req)
	return
}

func (s *DaemonServer) ApplyHttpChaos(ctx context.Context, in *pb.ApplyHttpChaosRequest) (*pb.ApplyHttpChaosResponse, error) {
	logger := log.L().WithName(loggerNameDaemonServer).WithValues("Request", in)
	logger.Info("applying http chaos")

	if s.backgroundProcessManager.Stdio(int(in.Instance), in.StartTime) == nil {
		// chaos daemon may restart, create another tproxy instance
		if err := s.backgroundProcessManager.KillBackgroundProcess(ctx, int(in.Instance), in.StartTime); err != nil {
			return nil, errors.Wrapf(err, "kill background process(%d)", in.Instance)
		}
		if err := s.createHttpChaos(ctx, in); err != nil {
			return nil, err
		}
	}

	resp, err := s.applyHttpChaos(ctx, logger, in)
	if err != nil {
		killError := s.backgroundProcessManager.KillBackgroundProcess(ctx, int(in.Instance), in.StartTime)
		logger.Error(killError, "kill tproxy", "instance", in.Instance)
		return nil, errors.Wrap(err, "apply config")
	}
	return resp, err
}

func (s *DaemonServer) applyHttpChaos(ctx context.Context, logger logr.Logger, in *pb.ApplyHttpChaosRequest) (*pb.ApplyHttpChaosResponse, error) {
	stdio := s.backgroundProcessManager.Stdio(int(in.Instance), in.StartTime)
	if stdio == nil {
		return nil, errors.Errorf("fail to get stdio of instance(%d)", in.Instance)
	}

	transport := stdioTransport{stdio: stdio}

	var rules []tproxyconfig.PodHttpChaosBaseRule
	err := json.Unmarshal([]byte(in.Rules), &rules)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal rules")
	}

	logger.Info("the length of actions", "length", len(rules))

	httpChaosSpec := tproxyconfig.Config{
		ProxyPorts: in.ProxyPorts,
		Rules:      rules,
	}

	config, err := json.Marshal(&httpChaosSpec)
	if err != nil {
		return nil, err
	}

	logger.Info("ready to apply", "config", string(config))

	// TODO: use new new request with context
	req, err := http.NewRequest(http.MethodPut, "/", bytes.NewReader(config))
	if err != nil {
		return nil, errors.Wrap(err, "create http request")
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		return nil, errors.Wrap(err, "send http request")
	}

	logger.Info("http chaos applied")

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read response body")
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
		return errors.Wrapf(err, "get PID of container(%s)", in.ContainerId)
	}
	processBuilder := bpm.DefaultProcessBuilder(tproxyBin, "-i", "-vv").
		EnableLocalMnt().
		SetIdentifier(fmt.Sprintf("tproxy-%s", in.ContainerId)).
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
		return errors.Wrapf(err, "execute command(%s)", cmd)
	}
	ct, err := procState.CreateTime()
	if err != nil {
		if kerr := cmd.Process.Kill(); kerr != nil {
			log.L().WithName(loggerNameDaemonServer).Error(kerr, "kill tproxy", "request", in)
		}
		return errors.Wrap(err, "get create time")
	}

	in.Instance = int64(cmd.Process.Pid)
	in.StartTime = ct
	return nil
}
