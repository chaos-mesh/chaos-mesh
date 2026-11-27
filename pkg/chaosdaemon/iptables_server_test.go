// Copyright 2021 Chaos Mesh Authors.
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
	"os/exec"
	"runtime"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients/test"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/log"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

func TestBuildIptablesRulesRejectsTCPThenRemainingTraffic(t *testing.T) {
	tests := []struct {
		name   string
		ipsets []string
		want   []string
	}{
		{
			name: "without an ipset",
			want: []string{
				"-A TEST -o eth0 --protocol tcp -j REJECT --reject-with tcp-reset -w 5",
				"-A TEST -o eth0 -j REJECT -w 5",
			},
		},
		{
			name:   "with an ipset",
			ipsets: []string{"TEST-IPSET"},
			want: []string{
				"-A TEST -o eth0 -m set --match-set TEST-IPSET dst,dst --protocol tcp -j REJECT --reject-with tcp-reset -w 5",
				"-A TEST -o eth0 -m set --match-set TEST-IPSET dst,dst -j REJECT -w 5",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules, err := buildIptablesRules(&pb.Chain{
				Name:           "TEST",
				Direction:      pb.Chain_OUTPUT,
				Ipsets:         tt.ipsets,
				Target:         "REJECT --reject-with tcp-reset",
				Protocol:       "tcp",
				FallbackTarget: "REJECT",
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(rules) != len(tt.want) {
				t.Fatalf("got %d rules, want %d: %q", len(rules), len(tt.want), rules)
			}
			for i := range tt.want {
				if rules[i] != tt.want[i] {
					t.Errorf("rule %d = %q, want %q", i, rules[i], tt.want[i])
				}
			}
		})
	}
}

func TestBuildIptablesRulesWithoutProtocol(t *testing.T) {
	tests := []struct {
		name   string
		ipsets []string
		target string
		want   string
	}{
		{
			name:   "drop without an ipset",
			target: "DROP",
			want:   "-A TEST -o eth0 -j DROP -w 5",
		},
		{
			name:   "traffic control with an ipset",
			ipsets: []string{"TEST-IPSET"},
			target: "CLASSIFY --set-class 1:4",
			want:   "-A TEST -o eth0 -m set --match-set TEST-IPSET dst,dst -j CLASSIFY --set-class 1:4 -w 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules, err := buildIptablesRules(&pb.Chain{
				Name:      "TEST",
				Direction: pb.Chain_OUTPUT,
				Ipsets:    tt.ipsets,
				Target:    tt.target,
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(rules) != 1 {
				t.Fatalf("got %d rules, want 1: %q", len(rules), rules)
			}
			if rules[0] != tt.want {
				t.Fatalf("rule = %q, want %q", rules[0], tt.want)
			}
		})
	}
}

var _ = Describe("iptables server", func() {
	BeforeEach(func() {
		if runtime.GOOS != "linux" {
			Skip("network namespaces require Linux")
		}
	})

	defer mock.With("MockContainerdClient", &test.MockClient{})()
	logger := log.NewZapLoggerWithWriter(GinkgoWriter)
	s, _ := newDaemonServer(&crclients.CrClientConfig{
		Runtime: crclients.ContainerRuntimeContainerd}, nil, logger)

	Context("FlushIptables", func() {
		It("installs IPv4 rules without routes", func() {
			defer mock.With("pid", 9527)()
			var installed bool
			defer mock.With("MockProcessBuild", func(_ context.Context, cmd string, args ...string) *exec.Cmd {
				Expect(cmd).To(Equal("/usr/local/bin/nsexec"))
				Expect(args[0]).To(Equal("-n"))
				Expect(args[1]).To(Equal("/proc/9527/ns/net"))
				Expect(args[2]).To(Equal("--"))
				Expect(args[3]).To(BeElementOf("iptables", "ip6tables"))
				if args[3] == "iptables" && args[5] == "-A" && args[6] == "TEST" {
					installed = true
				}
				return exec.Command("true")
			})()
			_, err := s.SetIptablesChains(context.TODO(), &pb.IptablesChainsRequest{
				Chains: []*pb.Chain{{
					Name:      "TEST",
					Direction: pb.Chain_INPUT,
					Ipsets:    []string{},
					IpVersion: pb.IpVersion_IPv4,
					Target:    "DROP",
				}},
				ContainerId: "containerd://container-id",
				EnterNS:     true,
			})
			Expect(err).To(BeNil())
			Expect(installed).To(BeTrue())
		})

		It("should reject IPv6 traffic", func() {
			defer mock.With("pid", 9527)()
			var rules []string
			defer mock.With("MockProcessBuild", func(_ context.Context, cmd string, args ...string) *exec.Cmd {
				Expect(cmd).To(Equal("/usr/local/bin/nsexec"))
				Expect(args[0]).To(Equal("-n"))
				Expect(args[1]).To(Equal("/proc/9527/ns/net"))
				Expect(args[2]).To(Equal("--"))
				Expect(args[3]).To(BeElementOf("iptables", "ip6tables"))
				if args[3] == "ip6tables" && args[5] == "-A" && args[6] == "TEST" {
					rules = append(rules, strings.Join(args[5:], " "))
				}
				return exec.Command("echo", "-n")
			})()
			_, err := s.SetIptablesChains(context.TODO(), &pb.IptablesChainsRequest{
				Chains: []*pb.Chain{{
					Name:           "TEST",
					Direction:      pb.Chain_INPUT,
					Ipsets:         []string{},
					IpVersion:      pb.IpVersion_IPv6,
					Protocol:       "tcp",
					Target:         "REJECT --reject-with tcp-reset",
					FallbackTarget: "REJECT",
				}},
				ContainerId: "containerd://container-id",
				EnterNS:     true,
			})
			Expect(err).To(BeNil())
			Expect(rules).To(Equal([]string{
				"-A TEST -i eth0 --protocol tcp -j REJECT --reject-with tcp-reset -w 5",
				"-A TEST -i eth0 -j REJECT -w 5",
			}))
		})

		It("should fallback to sandbox when container lookup fails", func() {
			defer mock.With("LoadContainerError", errors.New("container not found"))()
			defer mock.With("pid", int(9527))()
			defer mock.With("labels", map[string]string{
				"io.kubernetes.pod.uid": "test-pod-uid-12345",
			})()
			defer mock.With("MockProcessBuild", func(_ context.Context, _ string, _ ...string) *exec.Cmd {
				return exec.Command("true")
			})()

			_, err := s.SetIptablesChains(context.TODO(), &pb.IptablesChainsRequest{
				Chains: []*pb.Chain{{
					Name:      "TEST",
					Direction: pb.Chain_INPUT,
					Ipsets:    []string{},
					IpVersion: pb.IpVersion_IPv4,
				}},
				ContainerId: "containerd://container-id",
				PodUid:      "test-pod-uid-12345",
				EnterNS:     true,
			})
			Expect(err).To(BeNil())
		})

		It("should fail when both container and sandbox lookup fail", func() {
			defer mock.With("LoadContainerError", errors.New("container not found"))()
			defer mock.With("emptyContainers", true)()

			_, err := s.SetIptablesChains(context.TODO(), &pb.IptablesChainsRequest{
				Chains: []*pb.Chain{{
					Name:      "TEST",
					Direction: pb.Chain_INPUT,
					Ipsets:    []string{},
					IpVersion: pb.IpVersion_IPv4,
				}},
				ContainerId: "containerd://container-id",
				PodUid:      "test-pod-uid-12345",
				EnterNS:     true,
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("sandbox container not found"))
		})

		It("should fail on unknown chain direction", func() {
			defer mock.With("pid", 9527)()
			defer mock.With("MockProcessBuild", func(_ context.Context, cmd string, args ...string) *exec.Cmd {
				Expect(cmd).To(Equal("/usr/local/bin/nsexec"))
				Expect(args[0]).To(Equal("-n"))
				Expect(args[1]).To(Equal("/proc/9527/ns/net"))
				Expect(args[2]).To(Equal("--"))
				Expect(args[3]).To(BeElementOf("iptables", "ip6tables"))
				return exec.Command("true")
			})()

			_, err := s.SetIptablesChains(context.TODO(), &pb.IptablesChainsRequest{
				Chains: []*pb.Chain{{
					Name:      "TEST",
					Direction: pb.Chain_Direction(233),
					Ipsets:    []string{},
					IpVersion: pb.IpVersion_IPv4,
				}},
				ContainerId: "containerd://container-id",
				EnterNS:     true,
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("unknown chain direction 233"))
		})

		It("skips an optional family only when its filter table is unavailable", func() {
			defer mock.With("pid", 9527)()
			defer mock.With("MockProcessBuild", func(_ context.Context, _ string, args ...string) *exec.Cmd {
				if args[3] == "ip6tables" {
					Expect(args[5:]).To(Equal([]string{"-t", "filter", "-S"}))
					return exec.Command("sh", "-c", `echo "ip6tables: can't initialize ip6tables table 'filter': Table does not exist (do you need to insmod?)" >&2; exit 3`)
				}
				return exec.Command("true")
			})()
			chain := &pb.Chain{Direction: pb.Chain_INPUT, IpVersion: pb.IpVersion_IPv6, Optional: true}
			req := &pb.IptablesChainsRequest{
				Chains:      []*pb.Chain{chain},
				ContainerId: "containerd://container-id",
				EnterNS:     true,
			}

			_, err := s.SetIptablesChains(context.TODO(), req)
			Expect(err).To(Succeed())

			chain.Optional = false
			_, err = s.SetIptablesChains(context.TODO(), req)
			Expect(err).To(MatchError(ContainSubstring("Table does not exist")))
		})

		It("reports a firewall probe failure even for an optional family", func() {
			defer mock.With("pid", 9527)()
			defer mock.With("MockProcessBuild", func(_ context.Context, _ string, args ...string) *exec.Cmd {
				if args[3] == "ip6tables" {
					return exec.Command("sh", "-c", "echo 'Permission denied' >&2; exit 4")
				}
				return exec.Command("true")
			})()
			_, err := s.SetIptablesChains(context.TODO(), &pb.IptablesChainsRequest{
				Chains: []*pb.Chain{{
					Direction: pb.Chain_INPUT,
					IpVersion: pb.IpVersion_IPv6,
					Optional:  true,
				}},
				ContainerId: "containerd://container-id",
				EnterNS:     true,
			})
			Expect(err).To(MatchError(ContainSubstring("Permission denied")))
		})

		It("treats EAFNOSUPPORT and EPROTONOSUPPORT as unavailable IPv6 for IPv4 and recovery", func() {
			for _, stderr := range []string{
				"ip6tables v1.8.9 (legacy): can't initialize ip6tables table 'filter': Address family not supported by protocol",
				"ip6tables v1.8.9 (nf_tables): Could not fetch rule set generation id: Protocol not supported",
			} {
				func() {
					defer mock.With("pid", 9527)()
					defer mock.With("MockProcessBuild", func(_ context.Context, _ string, args ...string) *exec.Cmd {
						if args[3] == "ip6tables" {
							return exec.Command("sh", "-c", `printf '%s\n' "$1" >&2; exit 3`, "sh", stderr)
						}
						return exec.Command("true")
					})()

					_, err := s.SetIptablesChains(context.TODO(), &pb.IptablesChainsRequest{
						Chains:      nil,
						ContainerId: "containerd://container-id",
						EnterNS:     true,
					})
					Expect(err).To(Succeed())

					_, err = s.SetIptablesChains(context.TODO(), &pb.IptablesChainsRequest{
						Chains: []*pb.Chain{
							{Name: "V4", Direction: pb.Chain_INPUT, IpVersion: pb.IpVersion_IPv4, Target: "DROP"},
							{Name: "V6", Direction: pb.Chain_INPUT, IpVersion: pb.IpVersion_IPv6, Target: "DROP", Optional: true},
						},
						ContainerId: "containerd://container-id",
						EnterNS:     true,
					})
					Expect(err).To(Succeed())
				}()
			}
		})

		It("returns an error when neither IPv4 nor IPv6 filter table is available for non-empty optional chains", func() {
			defer mock.With("pid", 9527)()
			defer mock.With("MockProcessBuild", func(_ context.Context, _ string, _ ...string) *exec.Cmd {
				return exec.Command("sh", "-c", `echo "Table does not exist" >&2; exit 3`)
			})()
			_, err := s.SetIptablesChains(context.TODO(), &pb.IptablesChainsRequest{
				Chains: []*pb.Chain{
					{Name: "V4", Direction: pb.Chain_INPUT, IpVersion: pb.IpVersion_IPv4, Optional: true},
					{Name: "V6", Direction: pb.Chain_INPUT, IpVersion: pb.IpVersion_IPv6, Optional: true},
				},
				ContainerId: "containerd://container-id",
				EnterNS:     true,
			})
			Expect(err).To(MatchError(ContainSubstring("Table does not exist")))
		})

		It("rejects an unknown ip version before invoking iptables", func() {
			defer mock.With("pid", 9527)()
			var called bool
			defer mock.With("MockProcessBuild", func(_ context.Context, _ string, _ ...string) *exec.Cmd {
				called = true
				return exec.Command("true")
			})()
			_, err := s.SetIptablesChains(context.TODO(), &pb.IptablesChainsRequest{
				Chains: []*pb.Chain{{
					Name:      "TEST",
					Direction: pb.Chain_INPUT,
					IpVersion: pb.IpVersion(99),
				}},
				ContainerId: "containerd://container-id",
				EnterNS:     true,
			})
			Expect(err).To(MatchError("unknown ip version 99"))
			Expect(called).To(BeFalse())
		})
	})
})
