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
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"

	"github.com/containerd/cgroups/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cgroup memory limit", func() {
	DescribeTable("reads the v1 hierarchy limit of the target memory controller", func(value string, want uint64) {
		root := GinkgoT().TempDir()
		proc := filepath.Join(root, "proc-cgroup")
		writeCgroupFixture(proc, "7:cpu:/target/cpu\n6:memory:/target/memory\n")
		writeCgroupFixture(filepath.Join(root, "memory/target/memory/memory.stat"),
			fmt.Sprintf("rss 4096\nhierarchical_memory_limit %s\nhierarchical_memsw_limit 8192\n", value))
		info := CGroupInfo{CGMode: cgroups.Legacy, V1Path: pidPathFromFile(proc)}
		Expect(info.memoryLimit(root)).To(Equal(want))
	},
		Entry("finite", "268435456", uint64(268435456)),
		Entry("zero", "0", uint64(0)),
		Entry("unlimited", strconv.FormatInt(math.MaxInt64/int64(os.Getpagesize())*int64(os.Getpagesize()), 10), UnlimitedMemory),
	)

	It("accepts an inactive v1 memory controller but propagates missing membership", func() {
		root := GinkgoT().TempDir()
		proc := filepath.Join(root, "proc-cgroup")
		writeCgroupFixture(proc, "7:cpu:/target/cpu\n")
		info := CGroupInfo{CGMode: cgroups.Hybrid, V1Path: pidPathFromFile(proc)}
		Expect(info.memoryLimit(root)).To(Equal(UnlimitedMemory))
		info.V1Path = pidPathFromFile(filepath.Join(root, "missing"))
		_, err := info.memoryLimit(root)
		Expect(err).To(MatchError(os.ErrNotExist))
	})

	DescribeTable("finds the lowest v2 hard limit across the target and its ancestors", func(parent, leaf string, want uint64) {
		root := GinkgoT().TempDir()
		writeCgroupFixture(filepath.Join(root, "parent/memory.max"), parent)
		dir := filepath.Join(root, "parent/child")
		Expect(os.Mkdir(dir, 0755)).To(Succeed())
		if leaf != "" {
			writeCgroupFixture(filepath.Join(dir, "memory.max"), leaf)
		}
		info := CGroupInfo{CGMode: cgroups.Unified, V2CGroupPath: "/parent/child"}
		Expect(info.memoryLimit(root)).To(Equal(want))
	},
		Entry("stricter parent", "8192\n", "16384\n", uint64(8192)),
		Entry("stricter child", "16384\n", "8192\n", uint64(8192)),
		Entry("unlimited child", "16384\n", "max\n", uint64(16384)),
		Entry("disabled controller in child", "16384\n", "", uint64(16384)),
		Entry("no hard limits", "max\n", "max\n", UnlimitedMemory),
	)

	It("rejects missing v2 cgroups and malformed limits", func() {
		root := GinkgoT().TempDir()
		info := CGroupInfo{CGMode: cgroups.Unified, V2CGroupPath: "/target"}
		_, err := info.memoryLimit(root)
		Expect(err).To(MatchError(os.ErrNotExist))
		writeCgroupFixture(filepath.Join(root, "target/memory.max"), "invalid\n")
		_, err = info.memoryLimit(root)
		Expect(err).To(HaveOccurred())
	})
})
