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
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/containerd/cgroups/v3"
	"github.com/containerd/cgroups/v3/cgroup1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cgroup attachment", Serial, func() {
	It("writes v1 controller membership without changing sibling cgroups", func() {
		root := GinkgoT().TempDir()
		proc := filepath.Join(root, "proc-cgroup")
		writeCgroupFixture(proc, "7:cpu,cpuacct:/work/cpu\n6:memory:/work/memory\n")
		for _, file := range []string{"cpu/work/cpu/cgroup.procs", "memory/work/memory/cgroup.procs"} {
			writeCgroupFixture(filepath.Join(root, file), "")
		}
		for _, file := range []string{"cpu/work/sibling/cgroup.procs", "memory/work/sibling/cgroup.procs"} {
			writeCgroupFixture(filepath.Join(root, file), "unchanged")
		}
		attacher := &AttachCGroupV1{mode: cgroups.Legacy, path: pidPathFromFile(proc)}
		Expect(attacher.attachProcess(4321, func() ([]cgroup1.Subsystem, error) { return v1Hierarchy(root) })).To(Succeed())
		for file, want := range map[string]string{
			"cpu/work/cpu/cgroup.procs": "4321", "memory/work/memory/cgroup.procs": "4321",
			"cpu/work/sibling/cgroup.procs": "unchanged", "memory/work/sibling/cgroup.procs": "unchanged",
		} {
			data, err := os.ReadFile(filepath.Join(root, file))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal(want), "file %s", file)
		}
	})

	DescribeTable("preserves v1 load and write errors", func(failure string) {
		root := GinkgoT().TempDir()
		proc := filepath.Join(root, "proc-cgroup")
		writeCgroupFixture(proc, "7:cpu:/work/cpu\n6:memory:/work/memory\n")
		Expect(os.Mkdir(filepath.Join(root, "cpu"), 0755)).To(Succeed())
		target := filepath.Join(root, "cpu/work/cpu/cgroup.procs")
		if failure == "write" {
			Expect(os.MkdirAll(target, 0755)).To(Succeed())
		}
		attacher := &AttachCGroupV1{mode: cgroups.Legacy, path: pidPathFromFile(proc)}
		err := attacher.attachProcess(4321, func() ([]cgroup1.Subsystem, error) { return v1Hierarchy(root) })
		Expect(err).To(HaveOccurred())
		context := "load cgroup v1 manager"
		if failure == "load" {
			Expect(errors.Is(err, cgroup1.ErrCgroupDeleted)).To(BeTrue())
		} else {
			context = "add process to cgroup"
			var pathErr *os.PathError
			Expect(errors.As(err, &pathErr)).To(BeTrue())
			Expect(pathErr.Path).To(Equal(target))
		}
		for _, part := range []string{context, "pid 4321", "cpu path /work/cpu", "memory path /work/memory"} {
			Expect(err.Error()).To(ContainSubstring(part))
		}
	},
		Entry("load", "load"),
		Entry("write", "write"),
	)

	DescribeTable("uses v1 in legacy and hybrid modes", func(mode cgroups.CGMode) {
		GinkgoT().Setenv("PATH", GinkgoT().TempDir())
		attacher, err := getAttacherForPID(-1, mode)
		Expect(err).NotTo(HaveOccurred())
		info := attacher.TargetCGroup()
		Expect(info.CGMode).To(Equal(mode))
		Expect(info.V1Path).NotTo(BeNil())
		Expect(info.V2CGroupPath).To(BeEmpty())
		_, err = info.V1Path(cgroup1.Cpu)
		Expect(errors.Is(err, os.ErrNotExist)).To(BeTrue())
		Expect(err.Error()).To(ContainSubstring("/proc/-1/cgroup"))
	},
		Entry("legacy", cgroups.Legacy),
		Entry("hybrid", cgroups.Hybrid),
	)

	It("uses the host namespace for unified lookup and attachment", func() {
		assertArgs := fakeNsenter("5:memory:/ignored\n0::/work/target\n", 0)
		attacher, err := getAttacherForPID(2468, cgroups.Unified)
		Expect(err).NotTo(HaveOccurred())
		assertArgs("-C", "-t", "1", "cat", "/proc/2468/cgroup")
		info := attacher.TargetCGroup()
		Expect(info.CGMode).To(Equal(cgroups.Unified))
		Expect(info.V1Path).To(BeNil())
		Expect(info.V2CGroupPath).To(Equal("/work/target"))
		Expect(attacher.AttachProcess(4321)).To(Succeed())
		assertArgs("-C", "-t", "1", "--", "sh", "-c", "echo 4321 >> /host-sys/fs/cgroup/work/target/cgroup.procs")
	})

	It("preserves unified lookup command errors", func() {
		assertArgs := fakeNsenter("permission denied", 7)
		attacher, err := getAttacherForPID(2468, cgroups.Unified)
		Expect(attacher).To(BeNil())
		var exitErr *exec.ExitError
		Expect(errors.As(err, &exitErr)).To(BeTrue())
		Expect(exitErr.ExitCode()).To(Equal(7))
		Expect(err.Error()).To(ContainSubstring("pid 2468"))
		assertArgs("-C", "-t", "1", "cat", "/proc/2468/cgroup")
	})

	It("rejects malformed unified membership", func() {
		fakeNsenter("invalid membership\n", 0)
		attacher, err := getAttacherForPID(2468, cgroups.Unified)
		Expect(err).To(HaveOccurred())
		Expect(attacher).To(BeNil())
	})

	It("preserves unified attachment command errors", func() {
		assertArgs := fakeNsenter("permission denied", 7)
		attacher := &AttachCGroupV2{mode: cgroups.Unified, path: "/work/target"}
		err := attacher.AttachProcess(4321)
		var exitErr *exec.ExitError
		Expect(errors.As(err, &exitErr)).To(BeTrue())
		Expect(exitErr.ExitCode()).To(Equal(7))
		for _, part := range []string{"pid 4321", "/host-sys/fs/cgroup/work/target/cgroup.procs", "permission denied"} {
			Expect(err.Error()).To(ContainSubstring(part))
		}
		assertArgs("-C", "-t", "1", "--", "sh", "-c", "echo 4321 >> /host-sys/fs/cgroup/work/target/cgroup.procs")
	})
})

func fakeNsenter(output string, exitCode int) func(...string) {
	GinkgoHelper()
	dir := GinkgoT().TempDir()
	script := `#!/bin/sh
printf '%s\n' "$@" > "$CGROUP_TEST_ARGS"
if [ "$CGROUP_TEST_EXIT" -ne 0 ]; then
 printf '%s' "$CGROUP_TEST_OUTPUT" >&2
else
 printf '%s' "$CGROUP_TEST_OUTPUT"
fi
exit "$CGROUP_TEST_EXIT"
`
	Expect(os.WriteFile(filepath.Join(dir, "nsenter"), []byte(script), 0755)).To(Succeed())
	argsFile := filepath.Join(dir, "args")
	GinkgoT().Setenv("PATH", dir)
	GinkgoT().Setenv("CGROUP_TEST_ARGS", argsFile)
	GinkgoT().Setenv("CGROUP_TEST_OUTPUT", output)
	GinkgoT().Setenv("CGROUP_TEST_EXIT", strconv.Itoa(exitCode))
	return func(want ...string) {
		GinkgoHelper()
		data, err := os.ReadFile(argsFile)
		Expect(err).NotTo(HaveOccurred())
		got := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
		Expect(got).To(Equal(want))
	}
}
