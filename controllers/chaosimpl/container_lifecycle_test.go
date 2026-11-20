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

package chaosimpl_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/blockchaos"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/stresschaos"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/timechaos"
	impltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	daemonclient "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/client"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

func TestContainerExperiments(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Container experiment lifecycles")
}

type containerDaemon struct {
	daemonclient.ChaosDaemonClientInterface
	requests []proto.Message
	err      error
}

func (d *containerDaemon) Close() error { return nil }

func (d *containerDaemon) ApplyBlockChaos(_ context.Context, req *pb.ApplyBlockChaosRequest, _ ...grpc.CallOption) (*pb.ApplyBlockChaosResponse, error) {
	d.requests = append(d.requests, req)
	return &pb.ApplyBlockChaosResponse{InjectionId: 71}, d.err
}

func (d *containerDaemon) RecoverBlockChaos(_ context.Context, req *pb.RecoverBlockChaosRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	d.requests = append(d.requests, req)
	return &emptypb.Empty{}, d.err
}

func (d *containerDaemon) ExecStressors(_ context.Context, req *pb.ExecStressRequest, _ ...grpc.CallOption) (*pb.ExecStressResponse, error) {
	d.requests = append(d.requests, req)
	return &pb.ExecStressResponse{CpuInstance: "52", CpuStartTime: 123456, MemoryInstance: "74", MemoryStartTime: 234567}, d.err
}

func (d *containerDaemon) CancelStressors(_ context.Context, req *pb.CancelStressRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	d.requests = append(d.requests, req)
	return &emptypb.Empty{}, d.err
}

func (d *containerDaemon) SetTimeOffset(_ context.Context, req *pb.TimeRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	d.requests = append(d.requests, req)
	return &emptypb.Empty{}, d.err
}

func (d *containerDaemon) RecoverTimeOffset(_ context.Context, req *pb.TimeRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	d.requests = append(d.requests, req)
	return &emptypb.Empty{}, d.err
}

var _ = Describe("Fault and recovery requests for selected containers", func() {
	const target = "containerd://application"
	const record = "apps/pod/app"
	const blockRecord = record + "//mnt/data"
	var (
		kube    client.Client
		pod     *corev1.Pod
		daemon  *containerDaemon
		decoder *utils.ContainerRecordDecoder
		block   impltypes.ChaosImpl
		stress  impltypes.ChaosImpl
		clock   impltypes.ChaosImpl
	)
	ctx := context.Background()
	logger := logr.Discard()
	unavailable := errors.New("daemon unavailable")

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		pod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod", Namespace: "apps", UID: "pod-uid"},
			Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{
				{Name: "sidecar", ContainerID: "containerd://sidecar"},
				{Name: "app", ContainerID: target},
			}},
		}
		kube = fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod).WithStatusSubresource(pod).Build()
		daemon = &containerDaemon{}
		DeferCleanup(mock.With("MockChaosDaemonClient", daemon))
		Expect(mock.On("MockChaosDaemonClient")).To(BeIdenticalTo(daemon), "run the suite with failpoint instrumentation")
		decoder = utils.NewContainerRecordDecoder(kube, &chaosdaemon.ClientBuilder{Reader: kube})
		block = blockchaos.NewImpl(kube, logger, decoder).Impl
		stress = stresschaos.NewImpl(kube, logger, decoder).Impl
		clock = timechaos.NewImpl(kube, logger, decoder).Impl
	})

	It("tracks block injection IDs until recovery succeeds and passes the exact mount path", func() {
		records := []*v1alpha1.Record{{Id: blockRecord}}
		experiment := &v1alpha1.BlockChaos{Spec: v1alpha1.BlockChaosSpec{
			Action: v1alpha1.BlockDelay,
			Delay:  &v1alpha1.BlockDelaySpec{Latency: "2ms", Correlation: "0.5", Jitter: "50us"},
		}}
		Expect(block.Recover(ctx, 0, records, experiment)).To(Equal(v1alpha1.NotInjected))
		daemon.err = unavailable
		phase, err := block.Apply(ctx, 0, records, experiment)
		Expect(phase).To(Equal(v1alpha1.NotInjected))
		Expect(err).To(MatchError(unavailable))
		Expect(experiment.Status.InjectionIds).To(BeEmpty())
		daemon.err = nil

		Expect(block.Apply(ctx, 0, records, experiment)).To(Equal(v1alpha1.Injected))
		Expect(block.Apply(ctx, 0, records, experiment)).To(Equal(v1alpha1.Injected))
		Expect(daemon.requests).To(HaveLen(2))
		Expect(proto.Equal(daemon.requests[1], &pb.ApplyBlockChaosRequest{
			ContainerId: target, VolumePath: "/mnt/data", Action: pb.ApplyBlockChaosRequest_Delay, EnterNS: true,
			Delay: &pb.BlockDelaySpec{Delay: 2_000_000, Correlation: 0.5, Jitter: 50_000},
		})).To(BeTrue())
		Expect(experiment.Status.InjectionIds).To(HaveKeyWithValue(blockRecord, 71))

		daemon.err = unavailable
		phase, err = block.Recover(ctx, 0, records, experiment)
		Expect(phase).To(Equal(v1alpha1.Injected))
		Expect(err).To(MatchError(unavailable))
		Expect(experiment.Status.InjectionIds).To(HaveKeyWithValue(blockRecord, 71))
		daemon.err = nil
		Expect(block.Recover(ctx, 0, records, experiment)).To(Equal(v1alpha1.NotInjected))
		Expect(block.Recover(ctx, 0, records, experiment)).To(Equal(v1alpha1.NotInjected))
		Expect(daemon.requests).To(HaveLen(4))
		for _, request := range daemon.requests[2:] {
			Expect(proto.Equal(request, &pb.RecoverBlockChaosRequest{InjectionId: 71})).To(BeTrue())
		}
		Expect(experiment.Status.InjectionIds).To(BeEmpty())
	})

	DescribeTable("rejects invalid block instructions before calling the daemon",
		func(id string, action v1alpha1.BlockChaosAction, delay v1alpha1.BlockDelaySpec, message string) {
			experiment := &v1alpha1.BlockChaos{Spec: v1alpha1.BlockChaosSpec{Action: action, Delay: &delay}}
			phase, err := block.Apply(ctx, 0, []*v1alpha1.Record{{Id: id}}, experiment)
			Expect(phase).To(Equal(v1alpha1.NotInjected))
			Expect(err).To(MatchError(ContainSubstring(message)))
			Expect(daemon.requests).To(BeEmpty())
		},
		Entry("missing mount path", record, v1alpha1.BlockDelay, v1alpha1.BlockDelaySpec{}, "parse container and volumePath"),
		Entry("latency", blockRecord, v1alpha1.BlockDelay, v1alpha1.BlockDelaySpec{Latency: "invalid"}, "parse latency"),
		Entry("correlation", blockRecord, v1alpha1.BlockDelay, v1alpha1.BlockDelaySpec{Latency: "2ms", Correlation: "invalid"}, "parse corr"),
		Entry("jitter", blockRecord, v1alpha1.BlockDelay, v1alpha1.BlockDelaySpec{Latency: "2ms", Correlation: "0.5", Jitter: "invalid"}, "parse jitter"),
		Entry("unsupported action", blockRecord, v1alpha1.BlockChaosAction("unsupported"), v1alpha1.BlockDelaySpec{}, utils.ErrUnknownAction.Error()),
	)

	It("retries stress recovery without losing process identity or truncating start times", func() {
		load := 50
		records := []*v1alpha1.Record{{Id: record}}
		experiment := &v1alpha1.StressChaos{Spec: v1alpha1.StressChaosSpec{Stressors: &v1alpha1.Stressors{
			CPUStressor: &v1alpha1.CPUStressor{Stressor: v1alpha1.Stressor{Workers: 2}, Load: &load, Options: []string{"--metrics-brief"}},
			MemoryStressor: &v1alpha1.MemoryStressor{
				Stressor: v1alpha1.Stressor{Workers: 1}, Size: "8MiB", Time: "250ms", OOMScoreAdj: 100, Options: []string{"--size 16MiB"},
			},
		}}}
		Expect(stress.Recover(ctx, 0, records, experiment)).To(Equal(v1alpha1.NotInjected))
		daemon.err = unavailable
		phase, err := stress.Apply(ctx, 0, records, experiment)
		Expect(phase).To(Equal(v1alpha1.NotInjected))
		Expect(err).To(MatchError(unavailable))
		Expect(experiment.Status.Instances).To(BeEmpty())
		daemon.err = nil

		Expect(stress.Apply(ctx, 0, records, experiment)).To(Equal(v1alpha1.Injected))
		Expect(stress.Apply(ctx, 0, records, experiment)).To(Equal(v1alpha1.Injected))
		Expect(daemon.requests).To(HaveLen(2))
		Expect(proto.Equal(daemon.requests[1], &pb.ExecStressRequest{
			Scope: pb.ExecStressRequest_CONTAINER, Target: target, EnterNS: true, OomScoreAdj: 100,
			CpuStressors:    " --cpu-load-slice 10 --cpu-method sqrt --cpu 2 --cpu-load 50 --metrics-brief ",
			MemoryStressors: "--workers 1 --size 8MiB --time 250ms --size 16MiB",
		})).To(BeTrue())
		instance := experiment.Status.Instances[record]
		Expect(instance.UID).To(Equal("52"))
		Expect(instance.MemoryUID).To(Equal("74"))
		Expect(instance.StartTime.Time).To(BeTemporally("==", time.UnixMilli(123456)))
		Expect(instance.MemoryStartTime.Time).To(BeTemporally("==", time.UnixMilli(234567)))

		daemon.err = unavailable
		phase, err = stress.Recover(ctx, 0, records, experiment)
		Expect(phase).To(Equal(v1alpha1.Injected))
		Expect(err).To(MatchError(unavailable))
		Expect(experiment.Status.Instances).To(HaveKey(record))
		daemon.err = nil
		Expect(stress.Recover(ctx, 0, records, experiment)).To(Equal(v1alpha1.NotInjected))
		Expect(stress.Recover(ctx, 0, records, experiment)).To(Equal(v1alpha1.NotInjected))
		Expect(daemon.requests).To(HaveLen(4))
		for _, request := range daemon.requests[2:] {
			Expect(proto.Equal(request, &pb.CancelStressRequest{
				CpuInstance: "52", CpuStartTime: 123456, MemoryInstance: "74", MemoryStartTime: 234567,
			})).To(BeTrue())
		}
		Expect(experiment.Status.Instances).To(BeEmpty())
	})

	It("uses the same container identity to apply and recover negative clock offsets", func() {
		records := []*v1alpha1.Record{{Id: record}}
		experiment := &v1alpha1.TimeChaos{ObjectMeta: metav1.ObjectMeta{UID: "experiment-uid"}, Spec: v1alpha1.TimeChaosSpec{
			TimeOffset: "-1500ms", ClockIds: []string{"CLOCK_REALTIME", "CLOCK_MONOTONIC"},
		}}
		daemon.err = unavailable
		phase, err := clock.Apply(ctx, 0, records, experiment)
		Expect(phase).To(Equal(v1alpha1.NotInjected))
		Expect(err).To(MatchError(unavailable))
		daemon.err = nil
		Expect(clock.Apply(ctx, 0, records, experiment)).To(Equal(v1alpha1.Injected))
		Expect(proto.Equal(daemon.requests[1], &pb.TimeRequest{
			ContainerId: target, Uid: "experiment-uidpod-uid", PodContainerName: "pod-uid:app",
			Sec: -1, Nsec: -500_000_000, ClkIdsMask: 3,
		})).To(BeTrue())
		daemon.err = unavailable
		phase, err = clock.Recover(ctx, 0, records, experiment)
		Expect(phase).To(Equal(v1alpha1.Injected))
		Expect(err).To(MatchError(unavailable))
		daemon.err = nil
		Expect(clock.Recover(ctx, 0, records, experiment)).To(Equal(v1alpha1.NotInjected))
		Expect(proto.Equal(daemon.requests[3], &pb.TimeRequest{
			ContainerId: target, Uid: "experiment-uidpod-uid", PodContainerName: "pod-uid:app",
		})).To(BeTrue())
	})

	DescribeTable("does not send invalid time specifications to the daemon",
		func(offset string, ids []string, message string) {
			experiment := &v1alpha1.TimeChaos{Spec: v1alpha1.TimeChaosSpec{TimeOffset: offset, ClockIds: ids}}
			phase, err := clock.Apply(ctx, 0, []*v1alpha1.Record{{Id: record}}, experiment)
			Expect(phase).To(Equal(v1alpha1.NotInjected))
			Expect(err).To(MatchError(ContainSubstring(message)))
			Expect(daemon.requests).To(BeEmpty())
		},
		Entry("invalid clock", "-1s", []string{"invalid"}, "invalid"),
		Entry("invalid duration", "invalid", []string{"CLOCK_REALTIME"}, "invalid duration"),
	)

	It("treats disappeared containers as already recovered for each experiment type", func() {
		Expect(kube.Delete(ctx, pod)).To(Succeed())
		for _, experiment := range []struct {
			impl   impltypes.ChaosImpl
			object v1alpha1.InnerObject
			id     string
		}{
			{block, &v1alpha1.BlockChaos{}, blockRecord},
			{stress, &v1alpha1.StressChaos{}, record},
			{clock, &v1alpha1.TimeChaos{}, record},
		} {
			records := []*v1alpha1.Record{{Id: experiment.id}}
			phase, err := experiment.impl.Apply(ctx, 0, records, experiment.object)
			Expect(phase).To(Equal(v1alpha1.NotInjected))
			Expect(err).To(MatchError(utils.ErrContainerNotFound))
			Expect(experiment.impl.Recover(ctx, 0, records, experiment.object)).To(Equal(v1alpha1.NotInjected))
		}
		Expect(daemon.requests).To(BeEmpty())
	})
})
