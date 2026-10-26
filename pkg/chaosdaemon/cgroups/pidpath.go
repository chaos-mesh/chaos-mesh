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

package cgroups

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/containerd/cgroups/v3/cgroup1"
	"github.com/pkg/errors"
)

func V1() ([]cgroup1.Subsystem, error) {
	return v1Hierarchy("/host-sys/fs/cgroup")
}

func v1Hierarchy(root string) ([]cgroup1.Subsystem, error) {
	subsystems, err := defaults(root)
	if err != nil {
		return nil, err
	}
	var enabled []cgroup1.Subsystem
	for _, s := range pathers(subsystems) {
		// check and remove the default groups that do not exist
		if _, err := os.Lstat(s.Path("/")); err == nil {
			enabled = append(enabled, s)
		}
	}
	return enabled, nil
}

func PidPath(pid int) cgroup1.Path {
	return pidPathFromFile(fmt.Sprintf("/proc/%d/cgroup", pid))
}

func pidPathFromFile(p string) cgroup1.Path {
	paths, err := cgroup1.ParseCgroupFile(p)
	if err != nil {
		return func(_ cgroup1.Name) (string, error) {
			return "", errors.Wrapf(err, "parse cgroup file %s", p)
		}
	}

	return func(name cgroup1.Name) (string, error) {
		root, ok := paths[string(name)]
		if !ok {
			if root, ok = paths["name="+string(name)]; !ok {
				return "", errors.New("controller is not supported")
			}
		}

		return root, nil
	}
}

func V2PidGroupPath(pid int) (string, error) {
	// escape the CGroup Namespace, resolve https://github.com/chaos-mesh/chaos-mesh/pull/2928#issuecomment-1049465242
	// nsenter -C -t 1 cat /proc/$pid/cgroup
	command := exec.Command("nsenter", "-C", "-t", "1", "cat", fmt.Sprintf("/proc/%d/cgroup", pid))
	var buffer bytes.Buffer
	command.Stdout = &buffer

	err := command.Run()
	if err != nil {
		return "", errors.Wrapf(err, "get cgroup path of pid %d", pid)
	}
	return parseCgroupFromReader(&buffer)
}

// parseCgroupFromReader is adapted from containerd/cgroups/v3/cgroup2 v3.1.3.
// https://github.com/containerd/cgroups/blob/v3.1.3/cgroup2/utils.go
func parseCgroupFromReader(r io.Reader) (string, error) {
	s := bufio.NewScanner(r)
	for s.Scan() {
		if path, ok := strings.CutPrefix(s.Text(), "0::"); ok {
			return path, nil
		}
	}
	if err := s.Err(); err != nil {
		return "", err
	}
	return "", errors.New("cgroup path not found")
}
