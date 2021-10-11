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
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

const (
	bmInstallCommand = "bminstall.sh -b -Dorg.jboss.byteman.transform.all -Dorg.jboss.byteman.verbose -p %d %d"
	bmSubmitCommand  = "bmsubmit.sh -p %d -%s %s"
)

func (s *DaemonServer) InstallJVMRules(ctx context.Context,
	req *pb.InstallJVMRulesRequest) (*empty.Empty, error) {
	log.Info("InstallJVMRules", "request", req)
	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "GetPidFromContainerID")
		return nil, err
	}

	if req.Enable {
		return s.installJVMRules(ctx, req, pid)
	}

	return s.uninstallJVMRules(ctx, req, pid)
}

func (s *DaemonServer) installJVMRules(ctx context.Context,
	req *pb.InstallJVMRulesRequest, pid uint32) (*empty.Empty, error) {
	bytemanHome := os.Getenv("BYTEMAN_HOME")
	if len(bytemanHome) == 0 {
		return nil, fmt.Errorf("environment variable BYTEMAN_HOME not set")
	}

	// copy agent.jar to container's namespace
	if req.EnterNS {
		processBuilder := bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("mkdir -p %s/lib/", bytemanHome)).SetContext(ctx).SetNS(pid, bpm.MountNS)
		output, err := processBuilder.Build().CombinedOutput()
		if err != nil {
			return nil, err
		}
		if len(output) > 0 {
			log.Info("mkdir", "output", string(output))
		}

		agentFile, err := os.Open(fmt.Sprintf("%s/lib/byteman.jar", bytemanHome))
		if err != nil {
			return nil, err
		}
		processBuilder = bpm.DefaultProcessBuilder("sh", "-c", "cat > /usr/local/byteman/lib/byteman.jar").SetContext(ctx)
		processBuilder = processBuilder.SetNS(pid, bpm.MountNS).SetStdin(agentFile)
		output, err = processBuilder.Build().CombinedOutput()
		if err != nil {
			return nil, err
		}
		if len(output) > 0 {
			log.Info("copy agent.jar", "output", string(output))
		}
	}

	bmInstallCmd := fmt.Sprintf(bmInstallCommand, req.Port, pid)
	processBuilder := bpm.DefaultProcessBuilder("sh", "-c", bmInstallCmd).SetContext(ctx)
	if req.EnterNS {
		processBuilder = processBuilder.EnableLocalMnt()
	}

	cmd := processBuilder.Build()
	output, err := cmd.CombinedOutput()
	if err != nil {
		// this error will occured when install agent more than once, and will ignore this error and continue to submit rule
		errMsg1 := "Agent JAR loaded but agent failed to initialize"

		// these two errors will occured when java version less or euqal to 1.8, and don't know why
		// but it can install agent success even with this error, so just ignore it now.
		// TODO: Investigate the cause of these two error
		errMsg2 := "Provider sun.tools.attach.LinuxAttachProvider not found"
		errMsg3 := "install java.io.IOException: Non-numeric value found"
		if !strings.Contains(string(output), errMsg1) && !strings.Contains(string(output), errMsg2) &&
			!strings.Contains(string(output), errMsg3) {
			log.Error(err, string(output))
			return nil, err
		}
		log.Info("exec comamnd", "cmd", cmd.String(), "output", string(output), "error", err.Error())
	}

	// submit rules
	filename, err := writeDataIntoFile(req.Rule, "rule.btm")
	if err != nil {
		return nil, err
	}

	bmSubmitCmd := fmt.Sprintf(bmSubmitCommand, req.Port, "l", filename)
	processBuilder = bpm.DefaultProcessBuilder("sh", "-c", bmSubmitCmd).SetContext(ctx)
	if req.EnterNS {
		processBuilder = processBuilder.SetNS(pid, bpm.NetNS)
	}
	output, err = processBuilder.Build().CombinedOutput()
	if err != nil {
		log.Error(err, string(output))
		return nil, err
	}
	if len(output) > 0 {
		log.Info("submit rules", "output", string(output))
	}

	return &empty.Empty{}, nil
}

func (s *DaemonServer) uninstallJVMRules(ctx context.Context,
	req *pb.InstallJVMRulesRequest, pid uint32) (*empty.Empty, error) {
	filename, err := writeDataIntoFile(req.Rule, "rule.btm")
	if err != nil {
		return nil, err
	}
	log.Info("create btm file", "file", filename)

	bmSubmitCmd := fmt.Sprintf(bmSubmitCommand, req.Port, "u", filename)
	processBuilder := bpm.DefaultProcessBuilder("sh", "-c", bmSubmitCmd).SetContext(ctx)
	if req.EnterNS {
		processBuilder = processBuilder.SetNS(pid, bpm.NetNS)
	}
	output, err := processBuilder.Build().CombinedOutput()
	if err != nil {
		log.Error(err, string(output))
		if strings.Contains(string(output), "No rule scripts to remove") {
			return &empty.Empty{}, nil
		}
		return nil, err
	}

	if len(output) > 0 {
		log.Info(string(output))
	}

	return &empty.Empty{}, nil
}

func writeDataIntoFile(data string, filename string) (string, error) {
	tmpfile, err := ioutil.TempFile("", filename)
	if err != nil {
		return "", err
	}

	if _, err := tmpfile.WriteString(data); err != nil {
		return "", err
	}

	if err := tmpfile.Close(); err != nil {
		return "", err
	}

	return tmpfile.Name(), err
}
