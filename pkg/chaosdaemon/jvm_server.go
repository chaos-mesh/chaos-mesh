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
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/util"
	"io/ioutil"
	"os"
	"strings"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/golang/protobuf/ptypes/empty"
)

const (
	bmInstallCommand = "/usr/local/byteman/bin/bminstall.sh -b -Dorg.jboss.byteman.transform.all -Dorg.jboss.byteman.verbose -Dorg.jboss.byteman.compile.to.bytecode -p %d %d"
	bmSubmitCommand  = "/usr/local/byteman/bin/bmsubmit.sh -p %d -%s %s"
)

func (s *DaemonServer) InstallJVMRules(ctx context.Context,
	req *pb.InstallJVMRulesRequest) (*empty.Empty, error) {
	log.Info("InstallJVMRules", "request", req)
	// Use an backup version for linux versions less than 4.1
	less, err := util.CompareLinuxVersion(4, 1)
	if err != nil {
		log.Error(err, "Compare Linux Version")
		return nil, err
	}
	if less {
		return s.InstallJVMRulesBackUp(ctx, req)
	}
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

	//bytemanHome := os.Getenv("BYTEMAN_HOME")
	//if len(bytemanHome) == 0 {
	//	return nil, errors.New("environment variable BYTEMAN_HOME not set")
	//}
	//
	//// copy agent.jar to container's namespace
	////if req.EnterNS {
	//processBuilder := bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("mkdir -p %s/lib/", bytemanHome)).SetContext(ctx).SetNS(pid, bpm.MountNS)
	//output, err := processBuilder.Build().CombinedOutput()
	//if err != nil {
	//	return nil, err
	//}
	//if len(output) > 0 {
	//	log.Info("mkdir", "output", string(output))
	//}

	// todo: Need to write the BYTEMAN_HOME environment variable in bminstall.sh and bmsubmit.sh
	// or do this in code ?
	agentFile, err := os.Open("/usr/local/bin/byteman.tar.gz")
	if err != nil {
		return nil, err
	}
	//processBuilder = bpm.DefaultProcessBuilder("sh", "-c", "cat > /usr/local/byteman/lib/byteman.jar").SetContext(ctx)
	processBuilder := bpm.DefaultProcessBuilder("sh", "-c", "cat > /usr/local/byteman.tar.gz").SetContext(ctx)
	processBuilder = processBuilder.SetNS(pid, bpm.MountNS).SetStdin(agentFile)
	output, err := processBuilder.Build().CombinedOutput()
	if err != nil {
		return nil, err
	}
	if len(output) > 0 {
		log.Info("copy byteman.tar.gz", "output", string(output))
	}
	// uncompress byteman.tar.gz
	processBuilder = bpm.DefaultProcessBuilder("sh", "-c", "tar -zxf /usr/local/byteman.tar.gz -C /usr/local").SetContext(ctx).SetNS(pid, bpm.MountNS)
	output, err = processBuilder.Build().CombinedOutput()
	if err != nil {
		return nil, err
	}
	if len(output) > 0 {
		log.Info("tar byteman.tar.gz", "output", string(output))
	}

	bmInstallCmd := fmt.Sprintf(bmInstallCommand, req.Port, pid)

	output, err = s.crClient.ExecCommandByContainerID(ctx, req.ContainerId, []string{"sh", "-c", bmInstallCmd})

	if err != nil {
		// this error will occured when install agent more than once, and will ignore this error and continue to submit rule
		errMsg1 := "Agent JAR loaded but agent failed to initialize"

		// these two errors will occurred when java version less or equal to 1.8, and don't know why
		// but it can install agent success even with this error, so just ignore it now.
		// TODO: Investigate the cause of these two error
		errMsg2 := "Provider sun.tools.attach.LinuxAttachProvider not found"
		errMsg3 := "install java.io.IOException: Non-numeric value found"
		if !strings.Contains(string(output), errMsg1) && !strings.Contains(string(output), errMsg2) &&
			!strings.Contains(string(output), errMsg3) {
			log.Error(err, string(output))
			return nil, err
		}
		log.Info("exec comamnd", "cmd", bmInstallCmd, "output", string(output), "error", err.Error())
	}
	log.Info("exec comamnd", "cmd", bmInstallCmd, "output", string(output))

	filename, err := writeDataIntoFile(req.Rule, "rule.btm")
	if err != nil {
		return nil, err
	}
	ruleFile, err := os.Open(filename)

	processBuilder = bpm.DefaultProcessBuilder("sh", "-c", "cat > "+filename).SetContext(ctx)
	processBuilder = processBuilder.SetNS(pid, bpm.MountNS).SetStdin(ruleFile)
	output, err = processBuilder.Build().CombinedOutput()
	if err != nil {
		return nil, err
	}
	if len(output) > 0 {
		log.Info("copy ruleFile", "output", string(output))
	}

	bmSubmitCmd := fmt.Sprintf(bmSubmitCommand, req.Port, "l", filename)
	output, err = s.crClient.ExecCommandByContainerID(ctx, req.ContainerId, []string{"sh", "-c", bmSubmitCmd})
	if err != nil {
		log.Error(err, string(output))
		return nil, err
	}
	if len(output) > 0 {
		log.Info("submit rules", "output", string(output))
	}

	return &empty.Empty{}, nil

}

func (s *DaemonServer) UninstallJVMRules(ctx context.Context,
	req *pb.UninstallJVMRulesRequest) (*empty.Empty, error) {
	log.Info("InstallJVMRules", "request", req)
	// Use an backup version for linux versions less than 4.1
	less, err := util.CompareLinuxVersion(4, 1)
	if err != nil {
		log.Error(err, "Compare Linux Version")
		return nil, err
	}
	if less {
		return s.UninstallJVMRulesBackUp(ctx, req)
	}

	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "GetPidFromContainerID")
		return nil, err
	}

	filename, err := writeDataIntoFile(req.Rule, "rule.btm")
	if err != nil {
		return nil, err
	}
	ruleFile, err := os.Open(filename)

	processBuilder := bpm.DefaultProcessBuilder("sh", "-c", "cat > "+filename).SetContext(ctx)
	processBuilder = processBuilder.SetNS(pid, bpm.MountNS).SetStdin(ruleFile)
	output, err := processBuilder.Build().CombinedOutput()
	if err != nil {
		return nil, err
	}
	if len(output) > 0 {
		log.Info("copy ruleFile", "output", string(output))
	}

	bmSubmitCmd := fmt.Sprintf(bmSubmitCommand, req.Port, "u", filename)
	output, err = s.crClient.ExecCommandByContainerID(ctx, req.ContainerId, []string{"sh", "-c", bmSubmitCmd})
	if err != nil {
		log.Error(err, string(output))
		return nil, err
	}
	if len(output) > 0 {
		log.Info("submit rules", "output", string(output))
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
