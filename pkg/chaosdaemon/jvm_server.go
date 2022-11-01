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
	"fmt"
	"os"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/util"
)

const (
	bmInstallCommand = "bminstall.sh -b -Dorg.jboss.byteman.transform.all -Dorg.jboss.byteman.verbose -Dorg.jboss.byteman.compileToBytecode -p %d %d"
	bmSubmitCommand  = "bmsubmit.sh -p %d -%s %s"
)

func (s *DaemonServer) InstallJVMRules(ctx context.Context,
	req *pb.InstallJVMRulesRequest) (*empty.Empty, error) {
	log := s.getLoggerFromContext(ctx)
	log.Info("InstallJVMRules", "request", req)
	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "GetPidFromContainerID")
		return nil, err
	}

	containerPids := []uint32{pid}
	childPids, err := util.GetChildProcesses(pid, log)
	if err != nil {
		log.Error(err, "GetChildProcesses")
	}
	containerPids = append(containerPids, childPids...)
	for _, containerPid := range containerPids {
		name, err := util.ReadCommName(int(containerPid))
		if err != nil {
			log.Error(err, "ReadCommName")
			continue
		}
		if name == "java\n" {
			pid = containerPid
			break
		}
	}

	bytemanHome := os.Getenv("BYTEMAN_HOME")
	if len(bytemanHome) == 0 {
		return nil, errors.New("environment variable BYTEMAN_HOME not set")
	}

	// Copy byteman.jar, byteman-helper.jar and chaos-agent.jar into container's namespace.
	if req.EnterNS {
		processBuilder := bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("mkdir -p %s/lib/", bytemanHome)).SetContext(ctx).SetNS(pid, bpm.MountNS)
		output, err := processBuilder.Build(ctx).CombinedOutput()
		if err != nil {
			return nil, err
		}
		if len(output) > 0 {
			log.Info("mkdir", "output", string(output))
		}

		jars := []string{"byteman.jar", "byteman-helper.jar", "chaos-agent.jar"}

		for _, jar := range jars {
			source := fmt.Sprintf("%s/lib/%s", bytemanHome, jar)
			dest := fmt.Sprintf("/usr/local/byteman/lib/%s", jar)

			output, err = copyFileAcrossNS(ctx, source, dest, pid)
			if err != nil {
				return nil, err
			}

			log.Info("copy", "jar name", jar, "from source", source, "to destination", dest, "output", string(output))
		}
	}

	bmInstallCmd := fmt.Sprintf(bmInstallCommand, req.Port, pid)
	processBuilder := bpm.DefaultProcessBuilder("sh", "-c", bmInstallCmd).SetContext(ctx)
	if req.EnterNS {
		processBuilder = processBuilder.EnableLocalMnt()
	}

	cmd := processBuilder.Build(ctx)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// this error will occured when install agent more than once, and will ignore this error and continue to submit rule
		errMsg1 := "Agent JAR loaded but agent failed to initialize"

		// these two errors will occured when java version less or euqal to 1.8, and don't know why
		// but it can install agent success even with this error, so just ignore it now.
		// TODO: Investigate the cause of these two error
		errMsg2 := "Provider sun.tools.attach.LinuxAttachProvider not found"
		errMsg3 := "install java.io.IOException: Non-numeric value found"

		// this error is caused by the different attach result codes in different java versions. In fact, the agent has attached success, just ignore it here.
		// refer to https://stackoverflow.com/questions/54340438/virtualmachine-attach-throws-com-sun-tools-attach-agentloadexception-0-when-usi/54454418#54454418
		errMsg4 := "com.sun.tools.attach.AgentLoadException"
		if !strings.Contains(string(output), errMsg1) && !strings.Contains(string(output), errMsg2) &&
			!strings.Contains(string(output), errMsg3) && !strings.Contains(string(output), errMsg4) {
			log.Error(err, string(output))
			return nil, errors.Wrap(err, string(output))
		}
		log.Info("exec comamnd", "cmd", cmd.String(), "output", string(output), "error", err.Error())
	}

	// submit helper jar
	bmSubmitCmd := fmt.Sprintf(bmSubmitCommand, req.Port, "b", fmt.Sprintf("%s/lib/byteman-helper.jar", os.Getenv("BYTEMAN_HOME")))
	processBuilder = bpm.DefaultProcessBuilder("sh", "-c", bmSubmitCmd).SetContext(ctx)
	if req.EnterNS {
		processBuilder = processBuilder.SetNS(pid, bpm.NetNS)
	}
	output, err = processBuilder.Build(ctx).CombinedOutput()
	if err != nil {
		log.Error(err, string(output))
		return nil, err
	}
	if len(output) > 0 {
		log.Info("submit helper jar", "output", string(output))
	}

	// submit rules
	filename, err := writeDataIntoFile(req.Rule, "rule.btm")
	if err != nil {
		return nil, err
	}

	bmSubmitCmd = fmt.Sprintf(bmSubmitCommand, req.Port, "l", filename)
	processBuilder = bpm.DefaultProcessBuilder("sh", "-c", bmSubmitCmd).SetContext(ctx)
	if req.EnterNS {
		processBuilder = processBuilder.SetNS(pid, bpm.NetNS)
	}
	output, err = processBuilder.Build(ctx).CombinedOutput()
	if err != nil {
		log.Error(err, string(output))
		return nil, errors.Wrap(err, string(output))
	}
	if len(output) > 0 {
		log.Info("submit rules", "output", string(output))
	}

	return &empty.Empty{}, nil
}

func (s *DaemonServer) UninstallJVMRules(ctx context.Context,
	req *pb.UninstallJVMRulesRequest) (*empty.Empty, error) {
	log := s.getLoggerFromContext(ctx)
	log.Info("InstallJVMRules", "request", req)
	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "GetPidFromContainerID")
		return nil, err
	}

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
	output, err := processBuilder.Build(ctx).CombinedOutput()
	if err != nil {
		log.Error(err, string(output))
		if strings.Contains(string(output), "No rule scripts to remove") {
			return &empty.Empty{}, nil
		}
		return nil, errors.Wrap(err, string(output))
	}

	if len(output) > 0 {
		log.Info(string(output))
	}

	return &empty.Empty{}, nil
}

func writeDataIntoFile(data string, filename string) (string, error) {
	tmpfile, err := os.CreateTemp("", filename)
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

func copyFileAcrossNS(ctx context.Context, source string, dest string, pid uint32) ([]byte, error) {
	sourceFile, err := os.Open(source)
	if err != nil {
		return nil, err
	}
	defer sourceFile.Close()

	processBuilder := bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("cat > %s", dest)).SetContext(ctx)
	processBuilder = processBuilder.SetNS(pid, bpm.MountNS).SetStdin(sourceFile)
	output, err := processBuilder.Build(ctx).CombinedOutput()
	if err != nil {
		return nil, err
	}

	return output, nil
}
