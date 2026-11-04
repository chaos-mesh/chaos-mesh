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
	"os"

	"github.com/containerd/cgroups/v3"
	"github.com/containerd/cgroups/v3/cgroup1"
)

// defaults returns all known groups
func defaults(root string) ([]cgroup1.Subsystem, error) {
	h, err := cgroup1.NewHugetlb(root)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	s := []cgroup1.Subsystem{
		cgroup1.NewNamed(root, "systemd"),
		cgroup1.NewFreezer(root),
		cgroup1.NewPids(root),
		cgroup1.NewNetCls(root),
		cgroup1.NewNetPrio(root),
		cgroup1.NewPerfEvent(root),
		cgroup1.NewCpuset(root),
		cgroup1.NewCpu(root),
		cgroup1.NewCpuacct(root),
		cgroup1.NewMemory(root),
		cgroup1.NewBlkio(root),
		cgroup1.NewRdma(root),
	}
	// only add the devices cgroup if we are not in a user namespace
	// because modifications are not allowed
	if !cgroups.RunningInUserNS() {
		s = append(s, cgroup1.NewDevices(root))
	}
	// add the hugetlb cgroup if error wasn't due to missing hugetlb
	// cgroup support on the host
	if err == nil {
		s = append(s, h)
	}
	return s, nil
}

type pather interface {
	cgroup1.Subsystem
	Path(path string) string
}

func pathers(subystems []cgroup1.Subsystem) []pather {
	var out []pather
	for _, s := range subystems {
		if p, ok := s.(pather); ok {
			out = append(out, p)
		}
	}
	return out
}
