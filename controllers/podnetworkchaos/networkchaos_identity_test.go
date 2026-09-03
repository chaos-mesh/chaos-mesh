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

package podnetworkchaos

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-logr/logr"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/networkchaos/partition"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/networkchaos/podnetworkchaosmanager"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/networkchaos/trafficcontrol"
	impltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/netutils"
	daemonclient "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/client"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

// Exercise the real Apply/Recover transactions and the pod controller's RPC
// conversion. The fake daemon records requests without running network commands.
func TestNetworkChaosRuleIdentity(t *testing.T) {
	for _, action := range []v1alpha1.NetworkChaosAction{v1alpha1.PartitionAction, v1alpha1.DelayAction} {
		for _, legacy := range []bool{false, true} {
			name := string(action) + "/new-rules"
			if legacy {
				name = string(action) + "/persisted-legacy-rules"
			}
			t.Run(name, func(t *testing.T) {
				ctx := context.Background()
				scheme := runtime.NewScheme()
				if err := corev1.AddToScheme(scheme); err != nil {
					t.Fatal(err)
				}
				if err := v1alpha1.AddToScheme(scheme); err != nil {
					t.Fatal(err)
				}
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{Namespace: "workload", Name: "pod"},
					Status: corev1.PodStatus{
						Phase:             corev1.PodRunning,
						ContainerStatuses: []corev1.ContainerStatus{{ContainerID: "containerd://test"}},
					},
				}
				c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod).Build()
				builder := podnetworkchaosmanager.NewBuilder(podnetworkchaosmanager.Params{
					Client: c, Reader: c, Scheme: scheme, Logger: logr.Discard(),
				})
				var impl impltypes.ChaosImpl = partition.NewImpl(c, builder, logr.Discard())
				if action == v1alpha1.DelayAction {
					impl = trafficcontrol.NewImpl(c, builder, logr.Discard())
				}
				first := &v1alpha1.NetworkChaos{
					ObjectMeta: metav1.ObjectMeta{Namespace: "team-a", Name: "isolate", UID: "first"},
					Spec: v1alpha1.NetworkChaosSpec{
						Action: action, Direction: v1alpha1.To, ExternalTargets: []string{"192.0.2.1"},
						TcParameter: v1alpha1.TcParameter{Delay: &v1alpha1.DelaySpec{Latency: "10ms", Jitter: "0ms", Correlation: "0"}},
					},
				}
				second := first.DeepCopy()
				second.Namespace, second.UID = "team-b", "second"
				second.Spec.ExternalTargets = []string{"192.0.2.2"}
				records := []*v1alpha1.Record{{Id: "workload/pod", SelectorKey: ".", Phase: v1alpha1.NotInjected}}
				read := func() *v1alpha1.PodNetworkChaos {
					t.Helper()
					obj := &v1alpha1.PodNetworkChaos{}
					if err := c.Get(ctx, client.ObjectKeyFromObject(pod), obj); err != nil {
						t.Fatal(err)
					}
					return obj
				}
				apply := func(chaos *v1alpha1.NetworkChaos) {
					t.Helper()
					phase, err := impl.Apply(ctx, 0, records, chaos)
					if err != nil || phase != "Not Injected/Wait" {
						t.Fatalf("Apply returned phase %q, error %v", phase, err)
					}
				}
				apply(first)
				obj := read()
				if legacy {
					useLegacyRuleNames(obj, first.Name, action)
					if err := c.Update(ctx, obj); err != nil {
						t.Fatal(err)
					}
				}
				firstSpec := rulesFromSource(obj.Spec, "team-a/isolate")
				apply(second)
				obj = read()
				if !reflect.DeepEqual(rulesFromSource(obj.Spec, "team-a/isolate"), firstSpec) {
					t.Fatal("adding the second experiment changed the first experiment's stored rules")
				}
				if len(obj.Spec.IPSets) != 6 || len(obj.Spec.Iptables)+len(obj.Spec.TrafficControls) != 2 {
					t.Fatalf("expected two independent experiments, got %#v", obj.Spec)
				}
				assertIndependentNetworkRules(t, obj.Spec)
				assertNetworkRuleRequests(t, pod, obj)

				// Retrying an injection replaces only that source's rules.
				beforeRetry := obj.Spec.DeepCopy()
				apply(second)
				obj = read()
				if !reflect.DeepEqual(&obj.Spec, beforeRetry) {
					t.Fatal("retrying Apply changed the stored rules")
				}

				// Recovery identifies ownership by Source, including legacy names.
				secondSpec := rulesFromSource(obj.Spec, "team-b/isolate")
				records[0].Phase = v1alpha1.Injected
				phase, err := impl.Recover(ctx, 0, records, first)
				if err != nil || phase != "Injected/Wait" {
					t.Fatalf("Recover returned phase %q, error %v", phase, err)
				}
				obj = read()
				if !reflect.DeepEqual(obj.Spec, secondSpec) {
					t.Fatalf("recovery did not preserve only the second experiment: %#v", obj.Spec)
				}
				assertNetworkRuleRequests(t, pod, obj)
			})
		}
	}
}

func rulesFromSource(spec v1alpha1.PodNetworkChaosSpec, source string) v1alpha1.PodNetworkChaosSpec {
	result := v1alpha1.PodNetworkChaosSpec{}
	for _, set := range spec.IPSets {
		if set.Source == source {
			result.IPSets = append(result.IPSets, set)
		}
	}
	for _, chain := range spec.Iptables {
		if chain.Source == source {
			result.Iptables = append(result.Iptables, chain)
		}
	}
	for _, tc := range spec.TrafficControls {
		if tc.Source == source {
			result.TrafficControls = append(result.TrafficControls, tc)
		}
	}
	return result
}

func useLegacyRuleNames(obj *v1alpha1.PodNetworkChaos, name string, action v1alpha1.NetworkChaosAction) {
	postfix := "tgt"
	if action == v1alpha1.DelayAction {
		postfix = "netgt"
	}
	names := map[string]string{}
	for i := range obj.Spec.IPSets {
		set := &obj.Spec.IPSets[i]
		prefix := map[v1alpha1.IPSetType]string{
			v1alpha1.NetIPSet: "net_", v1alpha1.NetPortIPSet: "netport_", v1alpha1.SetIPSet: "set_",
		}[set.IPSetType]
		legacy := netutils.CompressName(name, 27, prefix+postfix)
		names[set.Name], set.Name = legacy, legacy
	}
	for i := range obj.Spec.IPSets {
		for j, name := range obj.Spec.IPSets[i].SetNames {
			obj.Spec.IPSets[i].SetNames[j] = names[name]
		}
	}
	for i := range obj.Spec.Iptables {
		obj.Spec.Iptables[i].Name = "OUTPUT/" + netutils.CompressName(name, 20, "")
		for j, name := range obj.Spec.Iptables[i].IPSets {
			obj.Spec.Iptables[i].IPSets[j] = names[name]
		}
	}
	for i := range obj.Spec.TrafficControls {
		obj.Spec.TrafficControls[i].IPSet = names[obj.Spec.TrafficControls[i].IPSet]
	}
}

func assertIndependentNetworkRules(t *testing.T, spec v1alpha1.PodNetworkChaosSpec) {
	t.Helper()
	owners := map[string]string{}
	for _, set := range spec.IPSets {
		if previous, ok := owners[set.Name]; ok {
			t.Fatalf("IP set %q is shared by %s and %s", set.Name, previous, set.Source)
		}
		owners[set.Name] = set.Source
		if set.IPSetType == v1alpha1.NetIPSet {
			want := map[string][]string{"team-a/isolate": {"192.0.2.1/32"}, "team-b/isolate": {"192.0.2.2/32"}}[set.Source]
			if !reflect.DeepEqual(set.Cidrs, want) {
				t.Fatalf("unexpected destination for %s: %v", set.Source, set.Cidrs)
			}
		}
	}
	checkOwner := func(name, source string) {
		t.Helper()
		if owners[name] != source {
			t.Fatalf("%s references IP set %q owned by %q", source, name, owners[name])
		}
	}
	for _, set := range spec.IPSets {
		for _, name := range set.SetNames {
			checkOwner(name, set.Source)
		}
	}
	chains := map[string]bool{}
	for _, chain := range spec.Iptables {
		if chains[chain.Name] {
			t.Fatalf("duplicate chain name %q", chain.Name)
		}
		chains[chain.Name] = true
		for _, name := range chain.IPSets {
			checkOwner(name, chain.Source)
		}
	}
	for _, tc := range spec.TrafficControls {
		checkOwner(tc.IPSet, tc.Source)
	}
}

func assertNetworkRuleRequests(t *testing.T, pod *corev1.Pod, obj *v1alpha1.PodNetworkChaos) {
	t.Helper()
	d := &networkRuleClient{}
	r := &Reconciler{Log: logr.Discard()}
	ctx := context.Background()
	if err := r.SetIPSets(ctx, pod, obj, d); err != nil {
		t.Fatal(err)
	}
	if err := r.SetIptables(ctx, pod, obj, d); err != nil {
		t.Fatal(err)
	}
	if err := r.SetTcs(ctx, pod, obj, d); err != nil {
		t.Fatal(err)
	}
	if len(d.sets.Ipsets) != len(obj.Spec.IPSets) || len(d.chains.Chains) != len(obj.Spec.Iptables) || len(d.tcs.Tcs) != len(obj.Spec.TrafficControls) {
		t.Fatal("daemon requests omitted stored rules")
	}
	for i, set := range d.sets.Ipsets {
		stored := obj.Spec.IPSets[i]
		if set.Name != stored.Name || !reflect.DeepEqual(set.SetNames, stored.SetNames) || !reflect.DeepEqual(set.Cidrs, stored.Cidrs) {
			t.Fatalf("daemon IP set differs from persisted rule: %v", set)
		}
	}
	for i, chain := range d.chains.Chains {
		stored := obj.Spec.Iptables[i]
		if chain.Name != stored.Name || !reflect.DeepEqual(chain.Ipsets, stored.IPSets) {
			t.Fatalf("daemon chain differs from persisted rule: %v", chain)
		}
	}
	for i, tc := range d.tcs.Tcs {
		if tc.Ipset != obj.Spec.TrafficControls[i].IPSet {
			t.Fatalf("daemon TC differs from persisted rule: %v", tc)
		}
	}
}

type networkRuleClient struct {
	daemonclient.ChaosDaemonClientInterface
	sets   *pb.IPSetsRequest
	chains *pb.IptablesChainsRequest
	tcs    *pb.TcsRequest
}

func (c *networkRuleClient) FlushIPSets(_ context.Context, request *pb.IPSetsRequest, _ ...grpc.CallOption) (*empty.Empty, error) {
	c.sets = request
	return &empty.Empty{}, nil
}

func (c *networkRuleClient) SetIptablesChains(_ context.Context, request *pb.IptablesChainsRequest, _ ...grpc.CallOption) (*empty.Empty, error) {
	c.chains = request
	return &empty.Empty{}, nil
}

func (c *networkRuleClient) SetTcs(_ context.Context, request *pb.TcsRequest, _ ...grpc.CallOption) (*empty.Empty, error) {
	c.tcs = request
	return &empty.Empty{}, nil
}
