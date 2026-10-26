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
	"errors"
	"os"
	"path/filepath"

	"github.com/containerd/cgroups"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Loading a process's cgroup controllers", func() {
	var root, procFile string
	var hierarchy cgroups.Hierarchy

	writeFile := func(path, contents string) {
		GinkgoHelper()
		Expect(os.MkdirAll(filepath.Dir(path), 0755)).To(Succeed())
		Expect(os.WriteFile(path, []byte(contents), 0644)).To(Succeed())
	}

	expectContents := func(path, contents string) {
		GinkgoHelper()
		actual, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(actual)).To(Equal(contents))
	}

	BeforeEach(func() {
		root = GinkgoT().TempDir()
		procFile = filepath.Join(root, "proc", "cgroup")
		writeFile(procFile, "2:cpu:/workload\n3:memory:/workload\n")
		for _, controller := range []string{"cpu", "memory", "net_prio"} {
			writeFile(filepath.Join(root, controller, "cgroup.procs"), "root membership\n")
		}
		hierarchy = func() ([]cgroups.Subsystem, error) {
			return []cgroups.Subsystem{
				cgroups.NewNetPrio(root),
				cgroups.NewCpu(root),
				cgroups.NewMemory(root),
			}, nil
		}
	})

	It("adds the process to CPU and memory when net_prio is absent from its membership", func() {
		for _, controller := range []string{"cpu", "memory"} {
			writeFile(filepath.Join(root, controller, "workload", "cgroup.procs"), "")
		}

		group, err := cgroups.Load(hierarchy, pidPathFromFile(procFile))
		Expect(err).NotTo(HaveOccurred())
		Expect(group.Add(cgroups.Process{Pid: 4242})).To(Succeed())

		for _, controller := range []string{"cpu", "memory"} {
			expectContents(filepath.Join(root, controller, "workload", "cgroup.procs"), "4242")
		}
		expectContents(filepath.Join(root, "net_prio", "cgroup.procs"), "root membership\n")
	})

	It("reports a deleted cgroup when no target controller is usable", func() {
		group, err := cgroups.Load(hierarchy, pidPathFromFile(procFile))
		Expect(errors.Is(err, cgroups.ErrCgroupDeleted)).To(BeTrue())
		Expect(group).To(BeNil())

		for _, controller := range []string{"cpu", "memory", "net_prio"} {
			expectContents(filepath.Join(root, controller, "cgroup.procs"), "root membership\n")
		}
	})
})
