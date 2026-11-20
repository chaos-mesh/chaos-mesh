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
	"os/exec"
	"strings"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

var _ = Describe("Traffic control reconciliation", func() {
	const interfaces = `[{"ifname":"eth0"},{"ifname":"eth1"}]`
	var (
		server        *DaemonServer
		runtimeClient *httpChaosRuntimeClientStub
		request       *pb.TcsRequest
		commands      []string
		replies       map[string]string
		failures      map[string]string
	)

	BeforeEach(func() {
		runtimeClient = &httpChaosRuntimeClientStub{containerPID: 42}
		server = &DaemonServer{crClient: runtimeClient, rootLogger: logr.Discard()}
		request = &pb.TcsRequest{ContainerId: "runtime-id", PodUid: "workload-uid", EnterNS: true}
		commands = nil
		replies = map[string]string{"ip -j addr show": interfaces}
		failures = map[string]string{}
		DeferCleanup(mock.With("MockProcessBuild", func(ctx context.Context, cmd string, args ...string) *exec.Cmd {
			Expect(cmd).To(Equal("/usr/local/bin/nsexec"))
			Expect(args[:3]).To(Equal([]string{"-n", "/proc/42/ns/net", "--"}))
			command := strings.Join(args[3:], " ")
			commands = append(commands, command)
			if message, failed := failures[command]; failed {
				return daemonTestCommand(ctx, message, true)
			}
			return daemonTestCommand(ctx, replies[command], false)
		}))
		Expect(mock.On("MockProcessBuild")).NotTo(BeNil(), "run with the repository's failpoint instrumentation")
	})

	It("flushes each interface and installs global and filtered rules on the requested devices", func(ctx SpecContext) {
		failures["tc qdisc del dev eth1 root"] = ruleNotExistLowerVersion
		request.Tcs = []*pb.Tc{
			{Type: pb.Tc_NETEM, Netem: &pb.Netem{Time: "5ms"}},
			{Device: "eth0", Type: pb.Tc_NETEM, Netem: &pb.Netem{Time: "6ms"}},
			{Device: "eth0", Type: pb.Tc_NETEM, Netem: &pb.Netem{Time: "7ms"}, Ipset: "test-peers", Protocol: "tcp"},
			{Device: "eth0", Type: pb.Tc_NETEM, Netem: &pb.Netem{Time: "8ms"}, Ipset: "test-peers", Protocol: "tcp"},
			{Device: "eth1", Type: pb.Tc_BANDWIDTH, Tbf: &pb.Tbf{Rate: "1mbit", Buffer: 64, Limit: 128}},
		}
		Expect(server.SetTcs(ctx, request)).NotTo(BeNil())
		Expect(commands[:3]).To(Equal([]string{"ip -j addr show", "tc qdisc del dev eth0 root", "tc qdisc del dev eth1 root"}))
		Expect(commands[3:]).To(ConsistOf(
			"tc qdisc add dev eth0 root handle 1: netem delay 5ms",
			"tc qdisc add dev eth0 parent 1: handle 2: netem delay 6ms",
			"tc qdisc add dev eth0 parent 2: handle 3: prio bands 4 priomap 1 2 2 2 1 2 0 0 1 1 1 1 1 1 1 1",
			"tc qdisc add dev eth0 parent 3:1 handle 4: sfq",
			"tc qdisc add dev eth0 parent 3:2 handle 5: sfq",
			"tc qdisc add dev eth0 parent 3:3 handle 6: sfq",
			"tc qdisc add dev eth0 parent 3:4 handle 7: netem delay 7ms",
			"tc qdisc add dev eth0 parent 7: handle 8: netem delay 8ms",
			"tc qdisc add dev eth1 root handle 1: tbf rate 1mbit burst 64 limit 128",
			"iptables -w -N TC-TABLES-0", "iptables -w -F TC-TABLES-0", "iptables -w -S TC-TABLES-0",
			"iptables -w -A TC-TABLES-0 -o eth0 -m set --match-set test-peers dst,dst --protocol tcp -j CLASSIFY --set-class 3:4 -w 5",
			"iptables -w -S CHAOS-OUTPUT", "iptables -w -A CHAOS-OUTPUT -j TC-TABLES-0",
		))
	})

	It("reports an early flush failure after trying every interface and leaves new rules unapplied", func(ctx SpecContext) {
		failures["tc qdisc del dev eth0 root"] = "device busy"
		request.Tcs = []*pb.Tc{{Type: pb.Tc_NETEM, Netem: &pb.Netem{Time: "5ms"}}}
		_, err := server.SetTcs(ctx, request)
		Expect(err).To(MatchError(ContainSubstring("device busy")))
		Expect(commands).To(Equal([]string{"ip -j addr show", "tc qdisc del dev eth0 root", "tc qdisc del dev eth1 root"}))
	})

	It("returns an error when adding a rule fails and stops before the next rule", func(ctx SpecContext) {
		firstRule := "tc qdisc add dev eth0 root handle 1: netem delay 5ms"
		failures[firstRule] = "qdisc unavailable"
		request.Tcs = []*pb.Tc{
			{Type: pb.Tc_NETEM, Netem: &pb.Netem{Time: "5ms"}},
			{Type: pb.Tc_NETEM, Netem: &pb.Netem{Time: "6ms"}},
		}
		_, err := server.SetTcs(ctx, request)
		Expect(err).To(MatchError(ContainSubstring("qdisc unavailable")))
		Expect(commands).To(Equal([]string{"ip -j addr show", "tc qdisc del dev eth0 root", "tc qdisc del dev eth1 root", firstRule}))
	})

	It("uses the sandbox PID when the application container is unavailable", func(ctx SpecContext) {
		runtimeClient.containerErr = errors.New("container not running")
		runtimeClient.sandboxPID = 42
		failures["tc qdisc del dev eth0 root"] = ruleNotExist
		_, err := server.SetTcs(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Expect(runtimeClient.receivedPodUID).To(Equal(request.PodUid))
		Expect(commands).To(HaveLen(3))

		runtimeClient.sandboxErr = errors.New("sandbox not running")
		commands = nil
		_, err = server.SetTcs(ctx, request)
		Expect(status.Code(err)).To(Equal(codes.Internal))
		Expect(err).To(MatchError(ContainSubstring("sandbox not running")))
		Expect(commands).To(BeEmpty())
	})

	DescribeTable("rejects invalid interface discovery without using tc",
		func(ctx SpecContext, output string) {
			replies["ip -j addr show"] = output
			_, err := server.SetTcs(ctx, request)
			Expect(err).To(HaveOccurred())
			Expect(commands).To(Equal([]string{"ip -j addr show"}))
		},
		Entry("invalid JSON", "["),
		Entry("missing interface name", `[{"index":1}]`),
		Entry("non-string interface name", `[{"ifname":42}]`),
		Entry("empty interface name", `[{"ifname":""}]`),
	)
})

func daemonTestCommand(ctx context.Context, output string, failed bool) *exec.Cmd {
	if failed {
		return exec.CommandContext(ctx, "/bin/sh", "-c", `printf '%s' "$1" >&2; exit 1`, "sh", output)
	}
	return exec.CommandContext(ctx, "/bin/sh", "-c", `printf '%s' "$1"`, "sh", output)
}
