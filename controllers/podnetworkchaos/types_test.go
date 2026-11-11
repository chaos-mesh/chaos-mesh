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

package podnetworkchaos

import (
	"context"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/cmd/chaos-controller-manager/provider"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/networkchaos/podnetworkchaosmanager"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/ipset"
	"github.com/chaos-mesh/chaos-mesh/controllers/test"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
	"github.com/chaos-mesh/chaos-mesh/pkg/testutils"
)

type tcRequestRecorder struct {
	test.MockChaosDaemonClient
	request *pb.TcsRequest
}

func (c *tcRequestRecorder) SetTcs(_ context.Context, request *pb.TcsRequest, _ ...grpc.CallOption) (*empty.Empty, error) {
	c.request = request
	return &empty.Empty{}, nil
}

func TestSetTcsSharesIdenticalFixedDelayFilters(t *testing.T) {
	const cidr = "198.18.0.2/32"
	for _, tt := range []struct {
		name       string
		target     v1alpha1.CidrAndPort
		device     string
		jitter     string
		wantShared bool
	}{
		{name: "same target and effective device", target: v1alpha1.CidrAndPort{Cidr: cidr}, device: "eth0", jitter: "0us", wantShared: true},
		{name: "different IP", target: v1alpha1.CidrAndPort{Cidr: "198.18.0.3/32"}, jitter: "0ms"},
		{name: "different port match", target: v1alpha1.CidrAndPort{Cidr: cidr, Port: 80}, jitter: "0ms"},
		{name: "different device", target: v1alpha1.CidrAndPort{Cidr: cidr}, device: "eth1", jitter: "0ms"},
		{name: "delay with jitter", target: v1alpha1.CidrAndPort{Cidr: cidr}, jitter: "1ms"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			chaos := &v1alpha1.PodNetworkChaos{}
			addDelay := func(name, device, jitter string, target v1alpha1.CidrAndPort) v1alpha1.RawTrafficControl {
				experiment := &v1alpha1.NetworkChaos{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: name}}
				source := "default/" + name
				sets := ipset.BuildIPSets(nil, []v1alpha1.CidrAndPort{target}, experiment, "ne_dst", source)
				filter := ipset.BuildSetIPSet(sets, experiment, "ne_dst", source)
				chaos.Spec.IPSets = append(chaos.Spec.IPSets, sets...)
				chaos.Spec.IPSets = append(chaos.Spec.IPSets, filter)
				tc := v1alpha1.RawTrafficControl{
					Type:   v1alpha1.Netem,
					IPSet:  filter.Name,
					Source: source,
					Device: device,
					TcParameter: v1alpha1.TcParameter{
						Delay: &v1alpha1.DelaySpec{Latency: "50ms", Jitter: jitter, Correlation: "0"},
					},
				}
				chaos.Spec.TrafficControls = append(chaos.Spec.TrafficControls, tc)
				return tc
			}
			a := addDelay("delay-a", "", "0ms", v1alpha1.CidrAndPort{Cidr: cidr})
			b := addDelay("delay-b", tt.device, tt.jitter, tt.target)
			g.Expect(a.IPSet).NotTo(Equal(b.IPSet))
			send := func(state *v1alpha1.PodNetworkChaos) []*pb.Tc {
				pod := &v1.Pod{Status: v1.PodStatus{ContainerStatuses: []v1.ContainerStatus{{ContainerID: "containerd://pod"}}}}
				daemon := &tcRequestRecorder{}
				g.Expect((&Reconciler{}).SetTcs(context.Background(), pod, state, daemon)).To(Succeed())
				g.Expect(daemon.request).NotTo(BeNil())
				return daemon.request.Tcs
			}
			tcs := send(chaos)
			g.Expect(tcs).To(HaveLen(2))
			if !tt.wantShared {
				g.Expect(tcs[0].Ipset).To(Equal(a.IPSet))
				g.Expect(tcs[1].Ipset).To(Equal(b.IPSet))
				return
			}
			g.Expect(tcs[0].Ipset).To(BeElementOf(a.IPSet, b.IPSet))
			g.Expect(tcs[1].Ipset).To(Equal(tcs[0].Ipset))
			for _, tc := range tcs {
				g.Expect(tc.Netem.Time).To(Equal("50ms"))
			}
			g.Expect(chaos.Spec.TrafficControls).To(Equal([]v1alpha1.RawTrafficControl{a, b}))
			for _, pair := range [][2]v1alpha1.RawTrafficControl{{a, b}, {b, a}} {
				state := chaos.DeepCopy()
				g.Expect((&podnetworkchaosmanager.Clear{Source: pair[0].Source}).Apply(state)).To(Succeed())
				g.Expect(state.Spec.IPSets).To(ContainElement(HaveField("Name", pair[1].IPSet)))
				tcs := send(state)
				g.Expect(tcs).To(HaveLen(1))
				g.Expect(tcs[0].Ipset).To(Equal(pair[1].IPSet))
				g.Expect(tcs[0].Netem.Time).To(Equal("50ms"))
				g.Expect((&podnetworkchaosmanager.Clear{Source: pair[1].Source}).Apply(state)).To(Succeed())
				g.Expect(send(state)).To(BeEmpty())
			}
		})
	}
}

func setHostNetwork(objs []runtime.Object) {
	for _, obj := range objs {
		if pod, ok := obj.(*v1.Pod); ok {
			pod.Spec.HostNetwork = true
		}
	}
}

func TestBuildIptablesChain(t *testing.T) {
	tests := []struct {
		name           string
		behavior       v1alpha1.NetworkChaosPartitionBehavior
		target         string
		protocol       string
		fallbackTarget string
	}{
		{
			name:   "defaults to drop",
			target: "DROP",
		},
		{
			name:     "explicitly drops",
			behavior: v1alpha1.DropPartitionBehavior,
			target:   "DROP",
		},
		{
			name:           "rejects TCP with a reset",
			behavior:       v1alpha1.RejectPartitionBehavior,
			target:         "REJECT --reject-with tcp-reset",
			protocol:       "tcp",
			fallbackTarget: "REJECT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain, err := buildIptablesChain(v1alpha1.RawIptables{
				Name:              "OUTPUT/example",
				Direction:         v1alpha1.Output,
				PartitionBehavior: tt.behavior,
			})
			if err != nil {
				t.Fatal(err)
			}
			if chain.Target != tt.target || chain.Protocol != tt.protocol || chain.FallbackTarget != tt.fallbackTarget {
				t.Fatalf(
					"got target=%q protocol=%q fallback target=%q, want target=%q protocol=%q fallback target=%q",
					chain.Target,
					chain.Protocol,
					chain.FallbackTarget,
					tt.target,
					tt.protocol,
					tt.fallbackTarget,
				)
			}
		})
	}

	t.Run("rejects an unknown direction", func(t *testing.T) {
		_, err := buildIptablesChain(v1alpha1.RawIptables{Direction: "sideways"})
		if err == nil {
			t.Fatal("expected an error for an unknown direction")
		}
	})

	t.Run("rejects an unknown partition behavior", func(t *testing.T) {
		_, err := buildIptablesChain(v1alpha1.RawIptables{
			Direction:         v1alpha1.Output,
			PartitionBehavior: "unknown",
		})
		if err == nil {
			t.Fatal("expected an error for an unknown partition behavior")
		}
	})
}

func TestHostNetworkOption(t *testing.T) {
	defer mock.With("MockChaosDaemonClient", &test.MockChaosDaemonClient{})()
	RegisterTestingT(t)

	testCases := []struct {
		name                     string
		enableHostNetworkTesting bool
		errorEvaluation          func(err string)
	}{
		{
			name:                     "host networking testing disabled (default)",
			enableHostNetworkTesting: false,
			errorEvaluation: func(err string) {
				Expect(err).To(ContainSubstring("It's dangerous to inject network chaos on a pod"))
			},
		},
		{
			name:                     "host networking testing enabled",
			enableHostNetworkTesting: true,
			errorEvaluation: func(err string) {
				Expect(err).To(Equal(""))
			},
		},
	}

	for _, testCase := range testCases {

		objs, _ := testutils.GenerateNPods("p", 1, testutils.PodArg{})

		setHostNetwork(objs)

		chaos := &v1alpha1.PodNetworkChaos{
			Kind:       "PodNetworkChaos",
			APIVersion: "v1",
			Namespace:  metav1.NamespaceDefault,
			Name:       "p0",
			Generation: 1,
			Spec:       v1alpha1.PodNetworkChaosSpec{},
		}
		objs = append(objs, chaos)

		fakeClient := fake.NewClientBuilder().
			WithScheme(provider.NewScheme()).
			WithRuntimeObjects(objs...).
			WithStatusSubresource(&v1alpha1.PodNetworkChaos{}).
			Build()

		recorder := recorder.NewDebugRecorder()
		h := &Reconciler{
			Client:                  fakeClient,
			Recorder:                recorder,
			Log:                     zap.New(zap.UseDevMode(true)),
			AllowHostNetworkTesting: testCase.enableHostNetworkTesting,
		}

		_, err := h.Reconcile(
			context.TODO(),
			ctrl.Request{
				Namespace: metav1.NamespaceDefault,
				Name:      "p0",
			})
		Expect(err).To(BeNil())

		fakeClient.Get(context.Background(), types.NamespacedName{
			Namespace: metav1.NamespaceDefault,
			Name:      "p0",
		}, chaos)

		testCase.errorEvaluation(chaos.Status.FailedMessage)
	}
}

func TestMergenetem(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		spec := v1alpha1.TcParameter{}
		_, err := mergeNetem(spec)
		if err == nil {
			t.Errorf("expect invalid spec failed with message %s but got nil", invalidNetemSpecMsg)
		}
		if err != nil && err.Error() != invalidNetemSpecMsg {
			t.Errorf("expect merge failed with message %s but got %v", invalidNetemSpecMsg, err)
		}
	})

	t.Run("delay loss rate", func(t *testing.T) {
		g := NewGomegaWithT(t)

		spec := v1alpha1.TcParameter{
			Delay: &v1alpha1.DelaySpec{
				Latency:     "1s",
				Correlation: "25",
				Jitter:      "100ms",
			},
			Loss: &v1alpha1.LossSpec{
				Loss:        "25",
				Correlation: "25",
			},
			Rate: &v1alpha1.RateSpec{
				Rate: "25mbps",
			},
		}
		m, err := mergeNetem(spec)
		g.Expect(err).ShouldNot(HaveOccurred())
		em := &pb.Netem{
			Time:      "1s",
			Jitter:    "100ms",
			DelayCorr: 25,
			Loss:      25,
			LossCorr:  25,
			Rate:      "25mbps",
		}
		g.Expect(m).Should(Equal(em))

		// Latency should have units in duration
		spec = v1alpha1.TcParameter{
			Delay: &v1alpha1.DelaySpec{
				Latency: "1000",
			},
		}
		_, err = mergeNetem(spec)
		g.Expect(err).Should(HaveOccurred())

		// Jitter should have units in duration
		spec = v1alpha1.TcParameter{
			Delay: &v1alpha1.DelaySpec{
				Latency: "1s",
				Jitter:  "100",
			},
		}
		_, err = mergeNetem(spec)
		g.Expect(err).Should(HaveOccurred())
	})
}
