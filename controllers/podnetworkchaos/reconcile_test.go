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
	"errors"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
	daemonclient "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/client"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

type recordingNetworkClient struct {
	daemonclient.ChaosDaemonClientInterface
	fail   string
	calls  []string
	ipsets *pb.IPSetsRequest
	chains *pb.IptablesChainsRequest
	tcs    *pb.TcsRequest
}

func (d *recordingNetworkClient) response(operation string) (*emptypb.Empty, error) {
	d.calls = append(d.calls, operation)
	if d.fail == operation {
		return nil, errors.New("daemon could not apply " + operation)
	}
	return &emptypb.Empty{}, nil
}

func (d *recordingNetworkClient) FlushIPSets(_ context.Context, req *pb.IPSetsRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	d.ipsets = req
	return d.response("ipsets")
}

func (d *recordingNetworkClient) SetIptablesChains(_ context.Context, req *pb.IptablesChainsRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	d.chains = req
	return d.response("iptables")
}

func (d *recordingNetworkClient) SetTcs(_ context.Context, req *pb.TcsRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	d.tcs = req
	return d.response("tc")
}

func (d *recordingNetworkClient) Close() error { return nil }

func TestPodNetworkReconcileRetriesAndClearsFailure(t *testing.T) {
	g := NewWithT(t)
	ctx := t.Context()
	scheme := runtime.NewScheme()
	g.Expect(corev1.AddToScheme(scheme)).To(Succeed())
	g.Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
	meta := metav1.ObjectMeta{Namespace: "fixtures", Name: "target"}
	pod := &corev1.Pod{
		ObjectMeta: meta,
		Status:     corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{Name: "app", ContainerID: "containerd://target"}}},
	}
	pod.UID = "fixture-uid"
	chaos := &v1alpha1.PodNetworkChaos{
		ObjectMeta: meta,
		Spec: v1alpha1.PodNetworkChaosSpec{
			IPSets:   []v1alpha1.RawIPSet{{Name: "peers", IPSetType: v1alpha1.NetPortIPSet, CidrAndPorts: []v1alpha1.CidrAndPort{{Cidr: "192.0.2.0/24", Port: 443}}}},
			Iptables: []v1alpha1.RawIptables{{Name: "block", Direction: v1alpha1.Input, IPSets: []string{"peers"}, Device: "eth1"}},
			TrafficControls: []v1alpha1.RawTrafficControl{
				{Type: v1alpha1.Netem, Device: "eth1", IPSet: "peers", TcParameter: v1alpha1.TcParameter{
					Duplicate: &v1alpha1.DuplicateSpec{Duplicate: "8", Correlation: "2"},
					Corrupt:   &v1alpha1.CorruptSpec{Corrupt: "3", Correlation: "1"},
					Rate:      &v1alpha1.RateSpec{Rate: "1mbit"},
				}},
				{Type: v1alpha1.Bandwidth, Device: "eth2", TcParameter: v1alpha1.TcParameter{
					Bandwidth: &v1alpha1.BandwidthSpec{Rate: "2mbit", Buffer: 128, Limit: 256},
				}},
			},
		},
	}
	chaos.Generation = 2
	updates := 0
	kube := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(chaos).WithObjects(chaos, pod).
		WithInterceptorFuncs(interceptor.Funcs{SubResourceUpdate: func(ctx context.Context, c client.Client, sub string, obj client.Object, opts ...client.SubResourceUpdateOption) error {
			updates++
			if updates == 1 {
				return apierrors.NewConflict(schema.GroupResource{Group: "chaos-mesh.org", Resource: "podnetworkchaos"}, obj.GetName(), errors.New("concurrent status update"))
			}
			return c.SubResource(sub).Update(ctx, obj, opts...)
		}}).Build()
	daemon := &recordingNetworkClient{}
	cleanup := mock.With("MockChaosDaemonClient", daemon)
	t.Cleanup(func() { g.Expect(cleanup()).To(Succeed()) })
	r := &Reconciler{Client: kube, Recorder: recorder.NewDebugRecorder(), Log: logr.Discard(), ChaosDaemonClientBuilder: &chaosdaemon.ClientBuilder{Reader: kube}}
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(chaos)}
	for _, scenario := range []struct {
		failure, status string
		calls           []string
	}{
		{"ipsets", "unable to flush ip sets", []string{"ipsets"}},
		{"iptables", "unable to set ip tables chains", []string{"ipsets", "iptables"}},
		{"tc", "unable to set tcs", []string{"ipsets", "iptables", "tc"}},
		{"", "", []string{"ipsets", "iptables", "tc"}},
	} {
		t.Run(scenario.failure+"/attempt", func(t *testing.T) {
			g := NewWithT(t)
			daemon.calls, daemon.fail = nil, scenario.failure
			g.Expect(r.Reconcile(ctx, req)).To(Equal(ctrl.Result{Requeue: scenario.failure != ""}))
			g.Expect(daemon.calls).To(Equal(scenario.calls))
			g.Expect(kube.Get(ctx, req.NamespacedName, chaos)).To(Succeed())
			g.Expect(chaos.Status.ObservedGeneration).To(Equal(chaos.Generation))
			if scenario.failure == "" {
				g.Expect(chaos.Status.FailedMessage).To(BeEmpty())
			} else {
				g.Expect(chaos.Status.FailedMessage).To(ContainSubstring(scenario.status))
			}
		})
	}
	g.Expect(daemon.ipsets.GetContainerId()).To(Equal("containerd://target"))
	g.Expect(daemon.ipsets.GetPodUid()).To(Equal("fixture-uid"))
	g.Expect(daemon.ipsets.GetIpsets()).To(HaveLen(1))
	g.Expect(daemon.ipsets.GetIpsets()[0].GetCidrAndPorts()[0].GetPort()).To(Equal(uint32(443)))
	g.Expect(daemon.chains.GetChains()).To(HaveLen(1))
	g.Expect(daemon.chains.GetChains()[0].GetDirection()).To(Equal(pb.Chain_INPUT))
	g.Expect(daemon.chains.GetChains()[0].GetIpsets()).To(Equal([]string{"peers"}))
	g.Expect(daemon.tcs.GetTcs()).To(HaveLen(2))
	g.Expect(daemon.tcs.GetTcs()[0].GetDevice()).To(Equal("eth1"))
	g.Expect(daemon.tcs.GetTcs()[0].GetNetem().GetDuplicate()).To(Equal(float32(8)))
	g.Expect(daemon.tcs.GetTcs()[1].GetTbf().GetRate()).To(Equal("2mbit"))
	g.Expect(updates).To(Equal(5))

	daemon.calls = nil
	g.Expect(r.Reconcile(ctx, req)).To(Equal(ctrl.Result{}))
	g.Expect(daemon.calls).To(BeEmpty())
	g.Expect(updates).To(Equal(5))
}
