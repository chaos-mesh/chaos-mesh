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

	"github.com/containerd/cgroups/v3/cgroup1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Loading a process's cgroup controllers", func() {
	var root, procFile string
	var hierarchy cgroup1.Hierarchy

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
		hierarchy = func() ([]cgroup1.Subsystem, error) {
			return []cgroup1.Subsystem{
				cgroup1.NewNetPrio(root),
				cgroup1.NewCpu(root),
				cgroup1.NewMemory(root),
			}, nil
		}
	})

	It("adds the process to CPU and memory when net_prio is absent from its membership", func() {
		for _, controller := range []string{"cpu", "memory"} {
			writeFile(filepath.Join(root, controller, "workload", "cgroup.procs"), "")
		}

		group, err := cgroup1.Load(pidPathFromFile(procFile), cgroup1.WithHierarchy(hierarchy))
		Expect(err).NotTo(HaveOccurred())
		Expect(group.Add(cgroup1.Process{Pid: 4242})).To(Succeed())

		for _, controller := range []string{"cpu", "memory"} {
			expectContents(filepath.Join(root, controller, "workload", "cgroup.procs"), "4242")
		}
		expectContents(filepath.Join(root, "net_prio", "cgroup.procs"), "root membership\n")
	})

	It("reports a deleted cgroup when no target controller is usable", func() {
		group, err := cgroup1.Load(pidPathFromFile(procFile), cgroup1.WithHierarchy(hierarchy))
		Expect(errors.Is(err, cgroup1.ErrCgroupDeleted)).To(BeTrue())
		Expect(group).To(BeNil())

		for _, controller := range []string{"cpu", "memory", "net_prio"} {
			expectContents(filepath.Join(root, controller, "cgroup.procs"), "root membership\n")
		}
	})
})
