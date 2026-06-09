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
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("jvm server", func() {
	var (
		containerRoot string
		sourceFile    string
	)

	BeforeEach(func() {
		containerRoot = GinkgoT().TempDir()
		originalContainerRootPath := containerRootPath
		containerRootPath = func(uint32) string {
			return containerRoot
		}
		DeferCleanup(func() {
			containerRootPath = originalContainerRootPath
		})

		sourceFile = filepath.Join(GinkgoT().TempDir(), "source.jar")
		Expect(os.WriteFile(sourceFile, []byte("jar content"), 0o644)).To(Succeed())
	})

	Context("mkdirInContainer", func() {
		It("creates directories through the container root", func() {
			Expect(mkdirInContainer(123, "/usr/local/byteman/lib")).To(Succeed())

			info, err := os.Stat(filepath.Join(containerRoot, "usr/local/byteman/lib"))
			Expect(err).ToNot(HaveOccurred())
			Expect(info.IsDir()).To(BeTrue())
		})
	})

	Context("copyFileAcrossNS", func() {
		It("copies a host file into the container root", func() {
			Expect(copyFileAcrossNS(context.Background(), sourceFile, "/usr/local/byteman/lib/byteman.jar", 123)).To(Succeed())

			content, err := os.ReadFile(filepath.Join(containerRoot, "usr/local/byteman/lib/byteman.jar"))
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(Equal("jar content"))
		})

		It("returns an error when the source file cannot be opened", func() {
			err := copyFileAcrossNS(context.Background(), filepath.Join(GinkgoT().TempDir(), "missing.jar"), "/usr/local/byteman/lib/byteman.jar", 123)

			Expect(err).To(HaveOccurred())
		})

		It("returns an error when the destination directory cannot be created", func() {
			Expect(os.WriteFile(filepath.Join(containerRoot, "usr"), []byte("not a directory"), 0o644)).To(Succeed())

			err := copyFileAcrossNS(context.Background(), sourceFile, "/usr/local/byteman/lib/byteman.jar", 123)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("create dest dir in container"))
		})

		It("returns an error when the destination file cannot be opened", func() {
			destDir := filepath.Join(containerRoot, "usr/local/byteman/lib/byteman.jar")
			Expect(os.MkdirAll(destDir, 0o755)).To(Succeed())

			err := copyFileAcrossNS(context.Background(), sourceFile, "/usr/local/byteman/lib/byteman.jar", 123)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("open dest in container"))
		})
	})

	Context("copyBytemanJarsIntoContainer", func() {
		It("copies byteman jars without using target-rootfs shell commands", func() {
			bytemanHome := GinkgoT().TempDir()
			Expect(os.MkdirAll(filepath.Join(bytemanHome, "lib"), 0o755)).To(Succeed())
			for _, jar := range []string{"byteman.jar", "byteman-helper.jar", "chaos-agent.jar"} {
				Expect(os.WriteFile(filepath.Join(bytemanHome, "lib", jar), []byte(jar), 0o644)).To(Succeed())
			}

			Expect(copyBytemanJarsIntoContainer(context.Background(), bytemanHome, 123)).To(Succeed())

			for _, jar := range []string{"byteman.jar", "byteman-helper.jar", "chaos-agent.jar"} {
				content, err := os.ReadFile(filepath.Join(containerRoot, "usr/local/byteman/lib", jar))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(content)).To(Equal(jar))
			}
			_, err := os.Stat(filepath.Join(containerRoot, bytemanHome, "lib"))
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
