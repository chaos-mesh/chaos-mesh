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

	"github.com/containerd/cgroups"
	"github.com/pkg/errors"
)

func V1() ([]cgroups.Subsystem, error) {
	subsystems, err := defaults("/host-sys/fs/cgroup")
	if err != nil {
		return nil, err
	}
	var enabled []cgroups.Subsystem
	for _, s := range pathers(subsystems) {
		// check and remove the default groups that do not exist
		if _, err := os.Lstat(s.Path("/")); err == nil {
			enabled = append(enabled, s)
		}
	}
	return enabled, nil
}

func PidPath(pid int) cgroups.Path {
	p := fmt.Sprintf("/proc/%d/cgroup", pid)
	paths, err := cgroups.ParseCgroupFile(p)
	if err != nil {
		return func(_ cgroups.Name) (string, error) {
			return "", errors.Wrapf(err, "parse cgroup file %s", p)
		}
	}

	return func(name cgroups.Name) (string, error) {
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

// parseCgroupFromReader is copied from github.com/containerd/cgroups/v2/utils.go
func parseCgroupFromReader(r io.Reader) (string, error) {
	var (
		s = bufio.NewScanner(r)
	)
	for s.Scan() {
		var (
			text  = s.Text()
			parts = strings.SplitN(text, ":", 3)
		)
		if len(parts) < 3 {
			return "", fmt.Errorf("invalid cgroup entry: %q", text)
		}
		// text is like "0::/user.slice/user-1001.slice/session-1.scope"
		if parts[0] == "0" && parts[1] == "" {
			return parts[2], nil
		}
	}
	if err := s.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("cgroup path not found")
}
