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

package time

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type timeProcessStat struct {
	Identity processIdentity
	Parent   int
}

func parseTimeProcessStat(pid int, data []byte) (timeProcessStat, error) {
	// comm is parenthesized and can contain spaces and closing parentheses.
	end := strings.LastIndexByte(string(data), ')')
	if end < 0 {
		return timeProcessStat{}, errors.New("malformed process stat")
	}
	fields := strings.Fields(string(data[end+1:]))
	if len(fields) <= 19 {
		return timeProcessStat{}, errors.New("truncated process stat")
	}
	if fields[0] == "Z" || fields[0] == "X" {
		return timeProcessStat{}, os.ErrNotExist
	}
	parent, err := strconv.Atoi(fields[1])
	if err != nil {
		return timeProcessStat{}, errors.Wrap(err, "invalid parent PID")
	}
	start, err := strconv.ParseUint(fields[19], 10, 64)
	if err != nil {
		return timeProcessStat{}, errors.Wrap(err, "invalid process start time")
	}
	return timeProcessStat{processIdentity{PID: pid, StartTime: strconv.FormatUint(start, 10)}, parent}, nil
}

func readProcessIdentity(pid int) (processIdentity, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return processIdentity{}, err
	}
	stat, err := parseTimeProcessStat(pid, data)
	return stat.Identity, err
}

// Root cgroups are shared by unrelated processes and must not be used to infer
// container membership. The ancestry walk is still available in that case.
func isContainerCgroup(cgroup string) bool {
	nonRoot := false
	for _, line := range strings.Split(strings.TrimSpace(cgroup), "\n") {
		fields := strings.SplitN(line, ":", 3)
		if len(fields) != 3 || fields[2] == "" {
			return false
		}
		if fields[2] != "/" {
			nonRoot = true
		}
	}
	return nonRoot
}

// Shared non-root cgroups are not sufficient to identify a container. Require
// its runtime ID as a complete path component or a recognized systemd scope.
func cgroupNamesContainer(cgroup, containerID string) bool {
	if _, id, found := strings.Cut(containerID, "://"); found {
		containerID = id
	}
	if len(containerID) < 12 {
		return false
	}
	for _, c := range containerID {
		if !(c >= '0' && c <= '9') && !(c >= 'a' && c <= 'f') {
			return false
		}
	}
	for _, line := range strings.Split(strings.TrimSpace(cgroup), "\n") {
		fields := strings.SplitN(line, ":", 3)
		if len(fields) != 3 {
			return false
		}
		for _, component := range strings.Split(fields[2], "/") {
			if component == containerID {
				return true
			}
			for _, prefix := range []string{"docker-", "cri-containerd-", "crio-"} {
				if component == prefix+containerID+".scope" {
					return true
				}
			}
		}
	}
	return false
}

func enumerateTimeProcesses(procRoot string, leader int, cgroup, containerID string) ([]processIdentity, error) {
	entries, err := os.ReadDir(procRoot)
	if err != nil {
		return nil, err
	}
	children := make(map[int][]int)
	identities := make(map[int]processIdentity)
	members := map[int]bool{leader: true}
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid <= 0 {
			continue
		}
		data, err := os.ReadFile(filepath.Join(procRoot, entry.Name(), "stat"))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		stat, err := parseTimeProcessStat(pid, data)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		children[stat.Parent] = append(children[stat.Parent], pid)
		identities[pid] = stat.Identity
		if isContainerCgroup(cgroup) && cgroupNamesContainer(cgroup, containerID) {
			data, err := os.ReadFile(filepath.Join(procRoot, entry.Name(), "cgroup"))
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return nil, err
			}
			if strings.TrimSpace(string(data)) == strings.TrimSpace(cgroup) {
				members[pid] = true
			}
		}
	}
	queue := make([]int, 0, len(members))
	for member := range members {
		queue = append(queue, member)
	}
	visited := make(map[int]bool)
	for len(queue) > 0 {
		parent := queue[0]
		queue = queue[1:]
		if visited[parent] {
			continue
		}
		visited[parent] = true
		for _, child := range children[parent] {
			members[child] = true
			queue = append(queue, child)
		}
	}
	result := make([]processIdentity, 0, len(members))
	for pid := range members {
		if pid != leader {
			result = append(result, identities[pid])
		}
	}
	return result, nil
}
