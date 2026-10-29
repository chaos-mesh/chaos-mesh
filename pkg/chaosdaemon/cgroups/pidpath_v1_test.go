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

func writeCgroupFixture(path, content string) {
	GinkgoHelper()
	Expect(os.MkdirAll(filepath.Dir(path), 0755)).To(Succeed())
	Expect(os.WriteFile(path, []byte(content), 0644)).To(Succeed())
}

var _ = Describe("Cgroup v1 PID paths", func() {
	DescribeTable("resolves controller membership", func(entries, systemd string) {
		file := filepath.Join(GinkgoT().TempDir(), "cgroup")
		writeCgroupFixture(file, "7:cpu,cpuacct:/host/cpu\n6:memory:/host/memory\n"+entries)
		path := pidPathFromFile(file)
		for controller, want := range map[cgroup1.Name]string{
			cgroup1.Cpu: "/host/cpu", cgroup1.Cpuacct: "/host/cpu",
			cgroup1.Memory: "/host/memory", "systemd": systemd,
		} {
			got, err := path(controller)
			Expect(err).NotTo(HaveOccurred(), "controller %s", controller)
			Expect(got).To(Equal(want), "controller %s", controller)
		}
		got, err := path(cgroup1.Blkio)
		Expect(err).To(HaveOccurred())
		Expect(got).To(BeEmpty())
	},
		Entry("named fallback", "1:name=systemd:/host/session\n", "/host/session"),
		Entry("direct precedence", "1:name=systemd:/host/session\n2:systemd:/host/direct\n", "/host/direct"),
	)

	DescribeTable("captures proc errors before the file is replaced", func(malformed bool) {
		file := filepath.Join(GinkgoT().TempDir(), "cgroup")
		if malformed {
			writeCgroupFixture(file, "invalid cgroup entry\n")
		}
		path := pidPathFromFile(file)
		writeCgroupFixture(file, "7:cpu:/replaced\n6:memory:/replaced\n")
		for _, controller := range []cgroup1.Name{cgroup1.Cpu, cgroup1.Memory} {
			got, err := path(controller)
			Expect(err).To(HaveOccurred(), "controller %s", controller)
			Expect(got).To(BeEmpty(), "controller %s", controller)
			Expect(err.Error()).To(ContainSubstring("parse cgroup file " + file))
			if !malformed {
				Expect(errors.Is(err, os.ErrNotExist)).To(BeTrue())
			}
		}
	},
		Entry("missing", false),
		Entry("malformed", true),
	)

	It("uses only existing host controllers in the v1 hierarchy", func() {
		root := GinkgoT().TempDir()
		for _, name := range []string{"cpu", "memory", "systemd", "unrelated"} {
			Expect(os.Mkdir(filepath.Join(root, name), 0755)).To(Succeed())
		}
		subsystems, err := v1Hierarchy(root)
		Expect(err).NotTo(HaveOccurred())
		got := map[cgroup1.Name]string{}
		for _, subsystem := range subsystems {
			p, ok := subsystem.(interface{ Path(string) string })
			Expect(ok).To(BeTrue(), "%s has no filesystem path", subsystem.Name())
			got[subsystem.Name()] = p.Path("/work")
		}
		Expect(got).To(Equal(map[cgroup1.Name]string{
			cgroup1.Cpu:    filepath.Join(root, "cpu/work"),
			cgroup1.Memory: filepath.Join(root, "memory/work"),
			"systemd":      filepath.Join(root, "systemd/work"),
		}))
	})
})
