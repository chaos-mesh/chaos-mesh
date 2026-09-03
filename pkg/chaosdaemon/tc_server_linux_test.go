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
	"fmt"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/go-logr/logr"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

type tcTestRuntimeClient struct {
	crclients.ContainerRuntimeInfoClient
}

func (tcTestRuntimeClient) GetPidFromContainerID(context.Context, string) (uint32, error) {
	return 123, nil
}

// recordTcCommands models chain replacement and intercepts every process, so the
// test never runs tc, iptables, or enters a network namespace.
func recordTcCommands(t *testing.T) (map[string][]string, *[]string) {
	t.Helper()
	chains := make(map[string][]string)
	var tcCommands []string
	cleanup := mock.With("MockProcessBuild", func(ctx context.Context, command string, args ...string) *exec.Cmd {
		if command != "/usr/local/bin/nsexec" || len(args) < 4 || args[0] != "-n" || args[1] != "/proc/123/ns/net" || args[2] != "--" {
			t.Fatalf("unexpected command: %s %v", command, args)
		}
		command, args = args[3], args[4:]
		switch command {
		case "ip":
			return exec.CommandContext(ctx, "/bin/echo", `[{"ifname":"eth0"},{"ifname":"net1"}]`)
		case "tc":
			tcCommands = append(tcCommands, strings.Join(args, " "))
		case "iptables":
			if len(args) < 3 || args[0] != "-w" {
				t.Fatalf("unexpected iptables arguments: %v", args)
			}
			args = args[1:]
			chain := args[1]
			switch args[0] {
			case "-N":
				if len(chain) > 28 {
					t.Fatalf("iptables chain name exceeds 28 characters: %s", chain)
				}
			case "-F":
				chains[chain] = nil
			case "-S":
				return exec.CommandContext(ctx, "/bin/echo", strings.Join(chains[chain], "\n"))
			case "-A":
				chains[chain] = append(chains[chain], strings.Join(args, " "))
			default:
				t.Fatalf("unexpected iptables operation: %v", args)
			}
		default:
			t.Fatalf("unexpected command: %s %v", command, args)
		}
		return exec.CommandContext(ctx, "/bin/true")
	})
	t.Cleanup(func() {
		if err := cleanup(); err != nil {
			t.Fatal(err)
		}
	})
	if mock.On("MockProcessBuild") == nil {
		t.Skip("requires failpoint instrumentation; refusing to execute network commands")
	}
	return chains, &tcCommands
}

func TestSetTcsPreservesFiltersAcrossDevices(t *testing.T) {
	for _, devices := range [][]string{{"eth0", "net1"}, {"net1", "eth0"}} {
		t.Run(strings.Join(devices, "_"), func(t *testing.T) {
			chains, commands := recordTcCommands(t)
			ctx := context.Background()
			server := &DaemonServer{crClient: tcTestRuntimeClient{}, rootLogger: logr.Discard()}
			iptables := buildIptablesClient(ctx, true, 123)
			if err := iptables.initializeEnv(); err != nil {
				t.Fatal(err)
			}

			rules := map[string][]*pb.Tc{
				"eth0": {
					{Type: pb.Tc_NETEM, Device: "eth0", Netem: &pb.Netem{Time: "10ms"}},
					{Type: pb.Tc_NETEM, Device: "eth0", Ipset: "target-b", Protocol: "udp", Netem: &pb.Netem{Time: "200ms"}},
					{Type: pb.Tc_NETEM, Device: "eth0", Ipset: "target-a", Protocol: "tcp", Netem: &pb.Netem{Time: "100ms"}},
					{Type: pb.Tc_NETEM, Device: "eth0", Ipset: "target-a", Protocol: "tcp", Netem: &pb.Netem{Loss: 10}},
				},
				"net1": {
					{Type: pb.Tc_NETEM, Device: "net1", Ipset: "target-a", Protocol: "tcp", Netem: &pb.Netem{Time: "300ms"}},
				},
			}
			request := &pb.TcsRequest{EnterNS: true}
			for _, device := range devices {
				request.Tcs = append(request.Tcs, rules[device]...)
			}

			wantRules := []string{
				"-A TC-TABLES-0 -o eth0 -m set --match-set target-a dst,dst --protocol tcp -j CLASSIFY --set-class 2:4 -w 5",
				"-A TC-TABLES-1 -o eth0 -m set --match-set target-b dst,dst --protocol udp -j CLASSIFY --set-class 2:5 -w 5",
				"-A TC-TABLES-2 -o net1 -m set --match-set target-a dst,dst --protocol tcp -j CLASSIFY --set-class 1:4 -w 5",
			}
			var firstCommands []string
			for attempt := 0; attempt < 2; attempt++ {
				*commands = nil
				if _, err := server.SetTcs(ctx, request); err != nil {
					t.Fatal(err)
				}
				for index, want := range wantRules {
					name := fmt.Sprintf("TC-TABLES-%d", index)
					if !reflect.DeepEqual(chains[name], []string{want}) {
						t.Errorf("attempt %d: %s = %v, want %s", attempt, name, chains[name], want)
					}
				}
				wantJumps := []string{"-A CHAOS-OUTPUT -j TC-TABLES-0", "-A CHAOS-OUTPUT -j TC-TABLES-1", "-A CHAOS-OUTPUT -j TC-TABLES-2"}
				if !reflect.DeepEqual(chains["CHAOS-OUTPUT"], wantJumps) {
					t.Errorf("attempt %d: classifier jumps = %v, want %v", attempt, chains["CHAOS-OUTPUT"], wantJumps)
				}
				if attempt == 0 {
					firstCommands = append([]string(nil), (*commands)...)
				} else if !reflect.DeepEqual(*commands, firstCommands) {
					t.Errorf("tc commands changed on retry: %v; first attempt: %v", *commands, firstCommands)
				}
			}

			// Qdisc handles remain local to each interface even though classifier
			// chain indices are shared across all interfaces.
			for _, want := range []string{
				"qdisc add dev eth0 root handle 1: netem delay 10ms",
				"qdisc add dev eth0 parent 2:4 handle 6: netem delay 100ms",
				"qdisc add dev eth0 parent 6: handle 7: netem loss 10.000000",
				"qdisc add dev eth0 parent 2:5 handle 8: netem delay 200ms",
				"qdisc add dev net1 parent 1:4 handle 5: netem delay 300ms",
			} {
				found := false
				for _, command := range *commands {
					if command == want {
						found = true
					}
				}
				if !found {
					t.Errorf("missing tc command %q in %v", want, *commands)
				}
			}

			// Reconciliation resets CHAOS-OUTPUT before SetTcs, so removing an
			// interface must leave no jumps to its previous classifier chains.
			if err := iptables.initializeEnv(); err != nil {
				t.Fatal(err)
			}
			if _, err := server.SetTcs(ctx, &pb.TcsRequest{EnterNS: true, Tcs: rules["net1"]}); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(chains["CHAOS-OUTPUT"], []string{"-A CHAOS-OUTPUT -j TC-TABLES-0"}) {
				t.Errorf("stale classifier jumps after removing eth0: %v", chains["CHAOS-OUTPUT"])
			}
			if len(chains["TC-TABLES-0"]) != 1 || !strings.Contains(chains["TC-TABLES-0"][0], "-o net1 ") {
				t.Errorf("remaining interface lost its classifier: %v", chains["TC-TABLES-0"])
			}
			if err := iptables.initializeEnv(); err != nil {
				t.Fatal(err)
			}
			if _, err := server.SetTcs(ctx, &pb.TcsRequest{EnterNS: true}); err != nil {
				t.Fatal(err)
			}
			if len(chains["CHAOS-OUTPUT"]) != 0 {
				t.Errorf("classifier jumps remain after recovery: %v", chains["CHAOS-OUTPUT"])
			}
		})
	}
}
