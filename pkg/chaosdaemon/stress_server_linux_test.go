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

package chaosdaemon

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/cgroups"
)

var _ = Describe("Memory stress size", func() {
	DescribeTable("passes absolute bytes to memStress", func(input, want string) {
		args := strings.Fields(input)
		err := resolveMemorySize(args, func() (uint64, error) { return 256 << 20, nil })
		Expect(err).NotTo(HaveOccurred())
		Expect(args).To(Equal(strings.Fields(want)))
	},
		Entry("typed percentage with two workers", "--workers 2 --size 25% --time 5s", "--workers 2 --size 67108864B --time 5s"),
		Entry("last option wins", "--size 10% -size=12.5%", "--size 10% -size=33554432B"),
		Entry("split single-dash option", "-size +50%", "-size 134217728B"),
		Entry("ignores positional arguments", "--size=50% -- --size=1%", "--size=134217728B -- --size=1%"),
		Entry("zero is a valid percentage", "--size=0%", "--size=0B"),
	)

	DescribeTable("leaves other arguments alone without reading the cgroup", func(input string) {
		args := strings.Fields(input)
		err := resolveMemorySize(args, func() (uint64, error) { return 0, errors.New("unexpected cgroup read") })
		Expect(err).NotTo(HaveOccurred())
		Expect(args).To(Equal(strings.Fields(input)))
	},
		Entry("fixed size", "--workers 2 --size 64MiB --time 3s"),
		Entry("last fixed size overrides a percentage", "--size=10% -size 48MB"),
		Entry("omitted size", "--workers 1"),
		Entry("flags after positional arguments", "--client 1 --size 25%"),
	)

	It("uses node MemTotal only when the cgroup has no limit", func() {
		root := GinkgoT().TempDir()
		GinkgoT().Setenv("HOST_PROC", root)
		Expect(os.WriteFile(filepath.Join(root, "meminfo"), []byte("MemTotal:       131072 kB\nMemAvailable:     4096 kB\n"), 0644)).To(Succeed())
		for limit, want := range map[uint64]string{
			cgroups.UnlimitedMemory: "33554432B",
			512 << 20:               "134217728B",
			0:                       "0B",
		} {
			args := []string{"--size", "25%"}
			Expect(resolveMemorySize(args, func() (uint64, error) { return limit, nil })).To(Succeed())
			Expect(args[1]).To(Equal(want))
		}
	})

	It("rejects an unavailable node total", func() {
		GinkgoT().Setenv("HOST_PROC", GinkgoT().TempDir())
		args := []string{"--size", "25%"}
		err := resolveMemorySize(args, func() (uint64, error) { return cgroups.UnlimitedMemory, nil })
		Expect(err).To(HaveOccurred())
	})

	It("propagates a cgroup read error", func() {
		args := []string{"--size", "25%"}
		err := resolveMemorySize(args, func() (uint64, error) { return 0, os.ErrPermission })
		Expect(err).To(MatchError(os.ErrPermission))
	})
})
