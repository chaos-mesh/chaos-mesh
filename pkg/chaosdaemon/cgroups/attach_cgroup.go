// Copyright 2022 Chaos Mesh Authors.
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

// Package cgroups locates process cgroups and attaches processes to cgroup v1 or v2 groups.
package cgroups

import (
	"fmt"
	"os/exec"

	"github.com/containerd/cgroups/v3"
	"github.com/containerd/cgroups/v3/cgroup1"
	"github.com/pkg/errors"
)

type CGroupInfo struct {
	CGMode       cgroups.CGMode
	V1Path       cgroup1.Path
	V2CGroupPath string
}

type AttachCGroup interface {
	TargetCGroup() CGroupInfo
	AttachProcess(pid int) error
}

var _ AttachCGroup = (*AttachCGroupV1)(nil)

type AttachCGroupV1 struct {
	mode cgroups.CGMode
	path cgroup1.Path
}

func (a *AttachCGroupV1) TargetCGroup() CGroupInfo {
	return CGroupInfo{
		CGMode:       a.mode,
		V1Path:       a.path,
		V2CGroupPath: "",
	}
}

func (a *AttachCGroupV1) AttachProcess(pid int) error {
	return a.attachProcess(pid, V1)
}

func (a *AttachCGroupV1) attachProcess(pid int, hierarchy cgroup1.Hierarchy) error {
	cgroupv1, err := cgroup1.Load(a.path, cgroup1.WithHierarchy(hierarchy))
	if err != nil {
		cpuCGroupPath, _ := a.path("cpu")
		memoryCGroupPath, _ := a.path("memory")
		return errors.Wrapf(err, "load cgroup v1 manager, pid %d, cpu path %s, memory path %s", pid, cpuCGroupPath, memoryCGroupPath)
	}
	err = cgroupv1.Add(cgroup1.Process{Pid: pid})
	if err != nil {
		cpuCGroupPath, _ := a.path("cpu")
		memoryCGroupPath, _ := a.path("memory")
		return errors.Wrapf(err, "add process to cgroup, pid %d, cpu path %s, memory path %s", pid, cpuCGroupPath, memoryCGroupPath)
	}
	return nil
}

var _ AttachCGroup = (*AttachCGroupV2)(nil)

type AttachCGroupV2 struct {
	mode cgroups.CGMode
	path string
}

func (a *AttachCGroupV2) TargetCGroup() CGroupInfo {
	return CGroupInfo{
		CGMode:       a.mode,
		V1Path:       nil,
		V2CGroupPath: a.path,
	}
}

func (a *AttachCGroupV2) AttachProcess(pid int) error {
	// escape the CGroup Namespace, we could not modify cgroups across different cgroups namespace,
	// resolve https://github.com/chaos-mesh/chaos-mesh/pull/2928#issuecomment-1049465242
	targetFile := fmt.Sprintf("/host-sys/fs/cgroup%s/cgroup.procs", a.path)
	command := exec.Command("nsenter", "-C", "-t", "1", "--", "sh", "-c", fmt.Sprintf("echo %d >> %s", pid, targetFile))
	output, err := command.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "attach process to cgroup, pid %d, target cgourp file %s, output %s", pid, targetFile, string(output))
	}
	return nil
}

// GetAttacherForPID return a AttachCGroup, which could attach a process to the same cgroup of the target pid
func GetAttacherForPID(targetPID int) (AttachCGroup, error) {
	return getAttacherForPID(targetPID, cgroups.Mode())
}

func getAttacherForPID(targetPID int, mode cgroups.CGMode) (AttachCGroup, error) {
	if mode == cgroups.Unified {
		groupPath, err := V2PidGroupPath(targetPID)
		if err != nil {
			return nil, err
		}
		return &AttachCGroupV2{
			mode: cgroups.Unified,
			path: groupPath,
		}, nil
	}

	// By default it's cgroup v1
	return &AttachCGroupV1{
		mode: mode,
		path: PidPath(targetPID),
	}, nil
}
