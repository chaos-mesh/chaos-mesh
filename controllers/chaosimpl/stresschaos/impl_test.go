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

package stresschaos

import (
	"context"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/common/records"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

func TestStressChaos(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "StressChaos Suite")
}

type cancelStressDaemon struct {
	pb.UnimplementedChaosDaemonServer
	attempts atomic.Int32
}

func (d *cancelStressDaemon) CancelStressors(context.Context, *pb.CancelStressRequest) (*emptypb.Empty, error) {
	if d.attempts.Add(1) == 1 {
		return nil, status.Error(codes.Unavailable, "temporary daemon failure")
	}
	return &emptypb.Empty{}, nil
}

var _ = Describe("StressChaos recovery", func() {
	It("requeues a failed cancellation and clears the instance when the retry succeeds", func(ctx SpecContext) {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		Expect(err).NotTo(HaveOccurred())
		daemon := &cancelStressDaemon{}
		server := grpc.NewServer()
		pb.RegisterChaosDaemonServer(server, daemon)
		go func() { _ = server.Serve(listener) }()
		DeferCleanup(server.Stop)

		oldConfig := config.ControllerCfg
		cfg := *oldConfig
		cfg.ChaosDaemonPort = listener.Addr().(*net.TCPAddr).Port
		cfg.ChaosMeshCACert = ""
		config.ControllerCfg = &cfg
		DeferCleanup(func() { config.ControllerCfg = oldConfig })

		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "target"},
			Spec:       v1.PodSpec{NodeName: "node"},
			Status:     v1.PodStatus{ContainerStatuses: []v1.ContainerStatus{{Name: "app", ContainerID: "containerd://app"}}},
		}
		endpoints := &discoveryv1.EndpointSlice{
			ObjectMeta:  metav1.ObjectMeta{Namespace: cfg.Namespace, Name: "chaos-daemon-test"},
			AddressType: discoveryv1.AddressTypeIPv4,
			Endpoints:   []discoveryv1.Endpoint{{NodeName: &pod.Spec.NodeName, Addresses: []string{"127.0.0.1"}}},
		}
		const id = "default/target/app"
		instance := v1alpha1.StressInstance{UID: "52"}
		chaos := &v1alpha1.StressChaos{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "stress"}}
		chaos.Status.Experiment.DesiredPhase = v1alpha1.StoppedPhase
		chaos.Status.Experiment.Records = []*v1alpha1.Record{{Id: id, Phase: v1alpha1.Injected}}
		chaos.Status.Instances = map[string]v1alpha1.StressInstance{id: instance}

		scheme := runtime.NewScheme()
		Expect(v1.AddToScheme(scheme)).To(Succeed())
		Expect(discoveryv1.AddToScheme(scheme)).To(Succeed())
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		cli := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod, endpoints, chaos).Build()
		log := logr.Discard()
		decoder := utils.NewContainerRecordDecoder(cli, &chaosdaemon.ClientBuilder{Reader: cli})
		impl := NewImpl(cli, log, decoder)
		r := &records.Reconciler{Impl: impl.Impl, Object: impl.Object, Client: cli, Recorder: recorder.NewDebugRecorder(), Log: log}
		req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(chaos)}

		result, err := r.Reconcile(ctx, req)
		Expect(err).NotTo(HaveOccurred())
		Expect(daemon.attempts.Load()).To(Equal(int32(1)))
		afterFailure := &v1alpha1.StressChaos{}
		Expect(cli.Get(ctx, req.NamespacedName, afterFailure)).To(Succeed())
		Expect(afterFailure.Status.Experiment.Records[0].Phase).To(Equal(v1alpha1.Injected))
		Expect(afterFailure.Status.Instances).To(HaveKeyWithValue(id, instance))
		Expect(result).To(Equal(ctrl.Result{Requeue: true}))

		result, err = r.Reconcile(ctx, req)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(ctrl.Result{}))
		Expect(daemon.attempts.Load()).To(Equal(int32(2)))
		afterRetry := &v1alpha1.StressChaos{}
		Expect(cli.Get(ctx, req.NamespacedName, afterRetry)).To(Succeed())
		Expect(afterRetry.Status.Experiment.Records[0].Phase).To(Equal(v1alpha1.NotInjected))
		Expect(afterRetry.Status.Instances).To(BeEmpty())
	}, SpecTimeout(10*time.Second))
})
