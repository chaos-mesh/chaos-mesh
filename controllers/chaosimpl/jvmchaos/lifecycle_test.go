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

package jvmchaos

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	daemonclient "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/client"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

type jvmDaemon struct {
	daemonclient.ChaosDaemonClientInterface
	installed   []*pb.InstallJVMRulesRequest
	uninstalled []*pb.UninstallJVMRulesRequest
	rpcErr      error
	closed      int
}

func (d *jvmDaemon) InstallJVMRules(_ context.Context, in *pb.InstallJVMRulesRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	d.installed = append(d.installed, in)
	return &emptypb.Empty{}, d.rpcErr
}

func (d *jvmDaemon) UninstallJVMRules(_ context.Context, in *pb.UninstallJVMRulesRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	d.uninstalled = append(d.uninstalled, in)
	return &emptypb.Empty{}, d.rpcErr
}

func (d *jvmDaemon) Close() error {
	d.closed++
	return errors.New("fixture connection already closed")
}

func TestJVMRuleLifecycle(t *testing.T) {
	g := NewWithT(t)
	ctx := t.Context()
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: "workload", Name: "jvm"},
		Status:     corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{Name: "app", ContainerID: "containerd://jvm"}}},
	}
	kube := fake.NewClientBuilder().WithObjects(pod).Build()
	daemon := &jvmDaemon{}
	cleanup := mock.With("MockChaosDaemonClient", daemon)
	t.Cleanup(func() { g.Expect(cleanup()).To(Succeed()) })
	decoder := utils.NewContainerRecordDecoder(kube, &chaosdaemon.ClientBuilder{Reader: kube})
	impl := NewImpl(kube, decoder, logr.Discard()).Impl
	chaos := &v1alpha1.JVMChaos{
		ObjectMeta: metav1.ObjectMeta{Namespace: "workload", Name: "experiment"},
		Spec:       v1alpha1.JVMChaosSpec{Action: v1alpha1.JVMLatencyAction, Name: "rule", Class: "Example", Method: "execute", LatencyDuration: 25, Port: 9277},
	}
	records := []*v1alpha1.Record{{Id: "workload/jvm/app"}}
	g.Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(v1alpha1.Injected))
	g.Expect(chaos.Spec.RuleData).To(ContainSubstring("Thread.sleep(25L)"))
	g.Expect(impl.Recover(ctx, 0, records, chaos)).To(Equal(v1alpha1.NotInjected))
	g.Expect(daemon.installed).To(HaveLen(1))
	g.Expect(daemon.uninstalled).To(HaveLen(1))
	g.Expect(proto.Equal(daemon.installed[0], &pb.InstallJVMRulesRequest{
		ContainerId: "containerd://jvm", Port: 9277, EnterNS: true, Rule: chaos.Spec.RuleData,
	})).To(BeTrue())
	g.Expect(proto.Equal(daemon.uninstalled[0], &pb.UninstallJVMRulesRequest{
		ContainerId: "containerd://jvm", Port: 9277, EnterNS: true, Rule: chaos.Spec.RuleData,
	})).To(BeTrue())

	daemon.rpcErr = errors.New("agent unavailable")
	phase, err := impl.Apply(ctx, 0, records, chaos)
	g.Expect(err).To(MatchError(daemon.rpcErr))
	g.Expect(phase).To(Equal(v1alpha1.NotInjected))
	phase, err = impl.Recover(ctx, 0, records, chaos)
	g.Expect(err).To(MatchError(daemon.rpcErr))
	g.Expect(phase).To(Equal(v1alpha1.Injected))
	daemon.rpcErr = errors.New("Connection refused")
	g.Expect(impl.Recover(ctx, 0, records, chaos)).To(Equal(v1alpha1.NotInjected))

	invalid := chaos.DeepCopy()
	invalid.Spec = v1alpha1.JVMChaosSpec{Action: v1alpha1.JVMMySQLAction, MySQLConnectorVersion: "9"}
	phase, err = impl.Apply(ctx, 0, records, invalid)
	g.Expect(err).To(MatchError(ContainSubstring("mysql connector version 9 is not supported")))
	g.Expect(phase).To(Equal(v1alpha1.NotInjected))
	phase, err = impl.Recover(ctx, 0, records, invalid)
	g.Expect(err).To(HaveOccurred())
	g.Expect(phase).To(Equal(v1alpha1.Injected))
	g.Expect(daemon.installed).To(HaveLen(2))
	g.Expect(daemon.uninstalled).To(HaveLen(3))
	g.Expect(daemon.closed).To(Equal(7))

	vanished := []*v1alpha1.Record{{Id: "workload/jvm/gone"}}
	phase, err = impl.Apply(ctx, 0, vanished, chaos)
	g.Expect(err).To(MatchError(utils.ErrContainerNotFound))
	g.Expect(phase).To(Equal(v1alpha1.NotInjected))
	g.Expect(impl.Recover(ctx, 0, vanished, chaos)).To(Equal(v1alpha1.NotInjected))

	withoutDecoder := NewImpl(kube, nil, logr.Discard()).Impl
	phase, err = withoutDecoder.Apply(ctx, 0, records, chaos)
	g.Expect(err).To(MatchError(errNilDecoder.Error()))
	g.Expect(phase).To(Equal(v1alpha1.NotInjected))
	phase, err = withoutDecoder.Recover(ctx, 0, records, chaos)
	g.Expect(err).To(MatchError(errNilDecoder.Error()))
	g.Expect(phase).To(Equal(v1alpha1.Injected))
}
