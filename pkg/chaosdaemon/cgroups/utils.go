// Copyright 2021 Chaos Mesh Authors.
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

package cgroups

import (
	"os"

	"github.com/containerd/cgroups"
)

// defaults returns all known groups
func defaults(root string) ([]cgroups.Subsystem, error) {
	h, err := cgroups.NewHugetlb(root)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	s := []cgroups.Subsystem{
		cgroups.NewNamed(root, "systemd"),
		cgroups.NewFreezer(root),
		cgroups.NewPids(root),
		cgroups.NewNetCls(root),
		cgroups.NewNetPrio(root),
		cgroups.NewPerfEvent(root),
		cgroups.NewCpuset(root),
		cgroups.NewCpu(root),
		cgroups.NewCpuacct(root),
		cgroups.NewMemory(root),
		cgroups.NewBlkio(root),
		cgroups.NewRdma(root),
	}
	// only add the devices cgroup if we are not in a user namespace
	// because modifications are not allowed
	if !cgroups.RunningInUserNS() {
		s = append(s, cgroups.NewDevices(root))
	}
	// add the hugetlb cgroup if error wasn't due to missing hugetlb
	// cgroup support on the host
	if err == nil {
		s = append(s, h)
	}
	return s, nil
}

type pather interface {
	cgroups.Subsystem
	Path(path string) string
}

func pathers(subystems []cgroups.Subsystem) []pather {
	var out []pather
	for _, s := range subystems {
		if p, ok := s.(pather); ok {
			out = append(out, p)
		}
	}
	return out
}
