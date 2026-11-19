// Copyright 2026 Chaos Mesh Authors.
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
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/containerd/cgroups/v3"
	"github.com/containerd/cgroups/v3/cgroup1"
	"github.com/pkg/errors"
)

// UnlimitedMemory indicates that the cgroup has no hard memory limit.
const UnlimitedMemory uint64 = math.MaxUint64

// MemoryLimit returns the effective hard memory limit in bytes, or UnlimitedMemory.
func (c CGroupInfo) MemoryLimit() (uint64, error) {
	return c.memoryLimit(hostCgroupRoot)
}

func (c CGroupInfo) memoryLimit(root string) (uint64, error) {
	if c.CGMode == cgroups.Unified {
		return v2MemoryLimit(root, c.V2CGroupPath)
	}

	group, err := c.V1Path(cgroup1.Memory)
	if errors.Is(err, cgroup1.ErrControllerNotActive) {
		return UnlimitedMemory, nil
	}
	if err != nil {
		return 0, err
	}
	path := filepath.Join(cgroup1.NewMemory(root).Path(group), "memory.stat")
	stat, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	for line := range strings.SplitSeq(string(stat), "\n") {
		if value, ok := strings.CutPrefix(line, "hierarchical_memory_limit "); ok {
			limit, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return 0, errors.Wrap(err, path)
			}
			// The v1 kernel reports LONG_MAX rounded down to a page boundary.
			if limit == uint64(math.MaxInt64)&^uint64(os.Getpagesize()-1) {
				return UnlimitedMemory, nil
			}
			return limit, nil
		}
	}
	return 0, errors.Errorf("hierarchical_memory_limit missing from %s", path)
}

func v2MemoryLimit(root, group string) (uint64, error) {
	limit := UnlimitedMemory
	for group = filepath.Clean("/" + group); group != "/"; group = filepath.Dir(group) {
		dir := filepath.Join(root, group)
		path := filepath.Join(dir, "memory.max")
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			// A cgroup can exist without the memory controller enabled at this level.
			if _, err = os.Stat(dir); err == nil {
				continue
			}
		}
		if err != nil {
			return 0, err
		}
		value := strings.TrimSpace(string(data))
		if value == "max" {
			continue
		}
		n, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return 0, errors.Wrap(err, path)
		}
		limit = min(limit, n)
	}
	return limit, nil
}
