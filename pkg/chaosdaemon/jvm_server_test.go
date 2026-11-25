// Copyright Chaos Mesh Authors.
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

//go:build linux

package chaosdaemon

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

var _ = Describe("JVM rule lifecycle", func() {
	const rule = "fixture rule body"
	const install = "bminstall.sh"
	const submitHelper = "bmsubmit.sh -p 9091 -b "
	const submitRule = "bmsubmit.sh -p 9091 -l "
	const removeRule = "bmsubmit.sh -p 9091 -u "
	var (
		server        *DaemonServer
		runtimeClient *httpChaosRuntimeClientStub
		tempDir       string
		bytemanHome   string
		commands      []string
		failAt        string
		message       string
		namespaced    []string
	)

	BeforeEach(func() {
		ginkgo := GinkgoT()
		tempDir = ginkgo.TempDir()
		bytemanHome = ginkgo.TempDir()
		ginkgo.Setenv("TMPDIR", tempDir)
		ginkgo.Setenv("BYTEMAN_HOME", bytemanHome)
		runtimeClient = &httpChaosRuntimeClientStub{containerPID: 90_000_000}
		server = &DaemonServer{crClient: runtimeClient, rootLogger: logr.Discard()}
		commands, namespaced, failAt, message = nil, nil, "", ""
		DeferCleanup(mock.With("MockProcessBuild", func(ctx context.Context, cmd string, args ...string) *exec.Cmd {
			if filepath.Base(cmd) == "nsexec" {
				Expect(args).To(ContainElement("--"))
				Expect(args[len(args)-3 : len(args)-1]).To(Equal([]string{"sh", "-c"}))
				namespaced = append(namespaced, strings.Join(args[:len(args)-4], " "))
			} else {
				Expect(cmd).To(Equal("sh"))
				Expect(args).To(HaveLen(2))
				Expect(args[0]).To(Equal("-c"))
			}
			command := args[len(args)-1]
			commands = append(commands, command)
			for _, prefix := range []string{submitRule, removeRule} {
				if path, ok := strings.CutPrefix(command, prefix); ok {
					Expect(filepath.Dir(path)).To(Equal(tempDir))
					Expect(os.ReadFile(path)).To(Equal([]byte(rule)))
				}
			}
			failed := failAt != "" && strings.HasPrefix(command, failAt)
			if failed {
				return daemonTestCommand(ctx, message, true)
			}
			return daemonTestCommand(ctx, "ok", false)
		}))
		Expect(mock.On("MockProcessBuild")).NotTo(BeNil(), "run with the repository's failpoint instrumentation")
	})

	It("submits the helper and rule after an already attached agent and cleans the rule file", func(ctx SpecContext) {
		failAt = install
		message = "Agent JAR loaded but agent failed to initialize"
		_, err := server.InstallJVMRules(ctx, &pb.InstallJVMRulesRequest{ContainerId: "runtime-id", Port: 9091, Rule: rule})
		Expect(err).NotTo(HaveOccurred())
		Expect(commands).To(HaveLen(3))
		Expect(commands[0]).To(HavePrefix(install))
		Expect(commands[1]).To(Equal(submitHelper + filepath.Join(bytemanHome, "lib", "byteman-helper.jar")))
		Expect(commands[2]).To(HavePrefix(submitRule))
		Expect(os.ReadDir(tempDir)).To(BeEmpty())
	})

	It("stops before submitting a rule when the helper could not be installed", func(ctx SpecContext) {
		failAt = submitHelper
		message = "helper upload failed"
		_, err := server.InstallJVMRules(ctx, &pb.InstallJVMRulesRequest{ContainerId: "runtime-id", Port: 9091, Rule: rule})
		Expect(err).To(HaveOccurred())
		Expect(commands).To(HaveLen(2))
		Expect(commands[1]).To(HavePrefix(submitHelper))
		Expect(os.ReadDir(tempDir)).To(BeEmpty())
	})

	It("cleans the rule file even if submitting it fails", func(ctx SpecContext) {
		failAt = submitRule
		message = "invalid rule data"
		_, err := server.InstallJVMRules(ctx, &pb.InstallJVMRulesRequest{ContainerId: "runtime-id", Port: 9091, Rule: rule})
		Expect(err).To(MatchError(ContainSubstring(message)))
		Expect(os.ReadDir(tempDir)).To(BeEmpty())
	})

	DescribeTable("cleans the temporary rule file after an uninstall",
		func(ctx SpecContext, output string, wantFailure bool) {
			if output != "" {
				failAt, message = removeRule, output
			}
			_, err := server.UninstallJVMRules(ctx, &pb.UninstallJVMRulesRequest{ContainerId: "runtime-id", Port: 9091, Rule: rule})
			if wantFailure {
				Expect(err).To(MatchError(ContainSubstring(output)))
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(commands).To(ConsistOf(HavePrefix(removeRule)))
			Expect(os.ReadDir(tempDir)).To(BeEmpty())
		},
		Entry("normal uninstall", "", false),
		Entry("already removed", "No rule scripts to remove", false),
		Entry("connection failure", "agent connection failed", true),
	)

	It("returns a runtime lookup failure before executing any command", func(ctx SpecContext) {
		runtimeClient.containerErr = errors.New("container unavailable")
		_, err := server.UninstallJVMRules(ctx, &pb.UninstallJVMRulesRequest{ContainerId: "runtime-id", Port: 9091, Rule: rule})
		Expect(err).To(MatchError(runtimeClient.containerErr))
		_, err = server.InstallJVMRules(ctx, &pb.InstallJVMRulesRequest{ContainerId: "runtime-id", Port: 9091, Rule: rule})
		Expect(err).To(MatchError(runtimeClient.containerErr))
		Expect(commands).To(BeEmpty())
	})

	It("copies Byteman jars before applying the rule in the network namespace", func(ctx SpecContext) {
		lib := filepath.Join(bytemanHome, "lib")
		Expect(os.Mkdir(lib, 0700)).To(Succeed())
		jars := []string{"byteman.jar", "byteman-helper.jar", "chaos-agent.jar"}
		for _, jar := range jars {
			Expect(os.WriteFile(filepath.Join(lib, jar), []byte(jar), 0600)).To(Succeed())
		}
		req := &pb.InstallJVMRulesRequest{ContainerId: "runtime-id", Port: 9091, Rule: rule, EnterNS: true}
		_, err := server.InstallJVMRules(ctx, req)
		Expect(err).NotTo(HaveOccurred())
		Expect(commands).To(HaveLen(7))
		Expect(commands[0]).To(Equal("mkdir -p " + lib + "/"))
		for i, jar := range jars {
			Expect(commands[i+1]).To(Equal("cat > /usr/local/byteman/lib/" + jar))
		}
		Expect(commands[4]).To(HavePrefix(install))
		Expect(commands[5]).To(Equal(submitHelper + filepath.Join(lib, "byteman-helper.jar")))
		Expect(commands[6]).To(HavePrefix(submitRule))
		Expect(namespaced).To(Equal([]string{
			"-p /proc/90000000/ns/pid -m /proc/90000000/ns/mnt",
			"-p /proc/90000000/ns/pid -m /proc/90000000/ns/mnt",
			"-p /proc/90000000/ns/pid -m /proc/90000000/ns/mnt",
			"-p /proc/90000000/ns/pid -m /proc/90000000/ns/mnt",
			"-n /proc/90000000/ns/net",
			"-n /proc/90000000/ns/net",
		}))

		commands = nil
		failAt, message = "cat > ", "copy failed"
		_, err = server.InstallJVMRules(ctx, req)
		Expect(err).To(MatchError(ContainSubstring(message)))
		Expect(commands).To(HaveLen(2))
		commands = nil
		failAt, message = "mkdir -p ", "mkdir failed"
		_, err = server.InstallJVMRules(ctx, req)
		Expect(err).To(MatchError(ContainSubstring(message)))
		Expect(commands).To(HaveLen(1))
	})

	It("rejects missing setup and propagates a fatal agent installation error", func(ctx SpecContext) {
		req := &pb.InstallJVMRulesRequest{ContainerId: "runtime-id", Port: 9091, Rule: rule}
		GinkgoT().Setenv("BYTEMAN_HOME", "")
		_, err := server.InstallJVMRules(ctx, req)
		Expect(err).To(MatchError(ContainSubstring("BYTEMAN_HOME not set")))
		Expect(commands).To(BeEmpty())
		GinkgoT().Setenv("BYTEMAN_HOME", bytemanHome)
		failAt, message = install, "agent installation failed"
		_, err = server.InstallJVMRules(ctx, req)
		Expect(err).To(MatchError(ContainSubstring(message)))
		Expect(commands).To(ConsistOf(HavePrefix(install)))
	})
})
