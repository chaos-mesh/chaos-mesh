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

package kernelchaos

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	impltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	chaosdaemonclient "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/client"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	pb_kernel "github.com/chaos-mesh/chaos-mesh/pkg/chaoskernel/pb"
	grpcUtils "github.com/chaos-mesh/chaos-mesh/pkg/grpc"
)

var _ impltypes.ChaosImpl = (*Impl)(nil)

type Impl struct {
	client.Client
	Log logr.Logger

	decoder *utils.ContainerRecordDecoder
}

// Apply applies KernelChaos
func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	impl.Log.Info("kernelchaos apply", "record", records[index])

	decodedContainer, err := impl.decoder.DecodeContainerRecord(ctx, records[index], obj)
	pbClient := decodedContainer.PbClient
	if pbClient != nil {
		defer pbClient.Close()
	}
	if err != nil {
		return v1alpha1.NotInjected, err
	}
	pod := decodedContainer.Pod
	containerId := decodedContainer.ContainerId

	impl.Log.Info("Try to apply KernelChaos", "namespace", pod.Namespace, "pod", pod, "containerName", decodedContainer.ContainerName)
	conn, pid, err := impl.prepareBPFKI(ctx, pbClient, pod, containerId)
	if err != nil {
		return v1alpha1.NotInjected, err
	}
	defer conn.Close()

	kernelChaos := obj.(*v1alpha1.KernelChaos)
	callchain := generateFailKernRequestFrame(kernelChaos.Spec.FailKernRequest.Callchain)

	bpfClient := pb_kernel.NewBPFKIServiceClient(conn)
	_, err = bpfClient.FailMMOrBIO(ctx, &pb_kernel.FailKernRequest{
		Pid:         pid,
		Ftype:       pb_kernel.FailKernRequest_FAILTYPE(kernelChaos.Spec.FailKernRequest.FailType),
		Headers:     kernelChaos.Spec.FailKernRequest.Headers,
		Callchain:   callchain,
		Probability: float32(kernelChaos.Spec.FailKernRequest.Probability) / 100,
		Times:       kernelChaos.Spec.FailKernRequest.Times,
	})
	if err != nil {
		err = errors.Wrapf(err, "Fail mm or bio error")
		return v1alpha1.NotInjected, err
	}

	return v1alpha1.Injected, nil
}

// Recover means the reconciler recovers the chaos action
func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	impl.Log.Info("kernelchaos recover", "record", records[index])

	decodedContainer, err := impl.decoder.DecodeContainerRecord(ctx, records[index], obj)
	pbClient := decodedContainer.PbClient
	if pbClient != nil {
		defer pbClient.Close()
	}
	if err != nil {
		return v1alpha1.Injected, err
	}
	var pod = decodedContainer.Pod
	containerId := decodedContainer.ContainerId

	impl.Log.Info("Try to recover KernelChaos", "namespace", pod.Namespace, "pod", pod, "containerName", decodedContainer.ContainerName)
	conn, pid, err := impl.prepareBPFKI(ctx, pbClient, pod, containerId)
	if err != nil {
		return v1alpha1.NotInjected, err
	}
	defer conn.Close()

	kernelChaos := obj.(*v1alpha1.KernelChaos)
	callchain := generateFailKernRequestFrame(kernelChaos.Spec.FailKernRequest.Callchain)

	bpfClient := pb_kernel.NewBPFKIServiceClient(conn)
	_, err = bpfClient.RecoverMMOrBIO(ctx, &pb_kernel.FailKernRequest{
		Pid:       pid,
		Callchain: callchain,
	})
	isMemoryFault := kernelChaos.Spec.FailKernRequest.FailType == 0 || kernelChaos.Spec.FailKernRequest.FailType == 1
	if err != nil && isMemoryFault && strings.Contains(err.Error(), "rpc error: code = Internal desc = Error removing value: No such file or directory") {
		// Inject memory fault in kernel may cause container restart and doesn't need to recover.
		// The container ID will change, but the pod name will not. TODO check the container ID
		impl.Log.Error(err, "KernelChaos target container probably restarted and doesn't need to recover")
		return v1alpha1.NotInjected, nil
	}
	if err != nil {
		err = errors.Wrapf(err, "Recover mm or bio error")
		return v1alpha1.Injected, err
	}

	return v1alpha1.NotInjected, nil
}

func generateFailKernRequestFrame(callChain []v1alpha1.Frame) (requestFrame []*pb_kernel.FailKernRequestFrame) {
	for _, frame := range callChain {
		requestFrame = append(requestFrame, &pb_kernel.FailKernRequestFrame{
			Funcname:   frame.Funcname,
			Parameters: frame.Parameters,
			Predicate:  frame.Predicate,
		})
	}
	return
}

func (impl *Impl) prepareBPFKI(ctx context.Context, pbClient chaosdaemonclient.ChaosDaemonClientInterface, pod *v1.Pod, containerId string) (*grpc.ClientConn, uint32, error) {
	containerResponse, err := pbClient.ContainerGetPid(ctx, &pb.ContainerRequest{
		Action: &pb.ContainerAction{
			Action: pb.ContainerAction_GETPID,
		},
		ContainerId: containerId,
	})
	if err != nil {
		err = errors.Wrapf(err, "Get container %s/%s %s pid error", pod.Namespace, pod.Name, containerId)
		return nil, 0, err
	}

	daemonIP, err := impl.decoder.FindDaemonIP(ctx, pod)
	if err != nil {
		return nil, 0, err
	}
	builder := grpcUtils.Builder(daemonIP, config.ControllerCfg.BPFKIPort).
		WithDefaultTimeout().
		Insecure()
	clientConn, err := builder.Build()
	if err != nil {
		return nil, 0, err
	}

	return clientConn, containerResponse.Pid, nil
}

func NewImpl(c client.Client, log logr.Logger, decoder *utils.ContainerRecordDecoder) *impltypes.ChaosImplPair {
	return &impltypes.ChaosImplPair{
		Name:   "kernelchaos",
		Object: &v1alpha1.KernelChaos{},
		Impl: &Impl{
			Client:  c,
			Log:     log.WithName("kernelchaos"),
			decoder: decoder,
		},
		ObjectList: &v1alpha1.KernelChaosList{},
	}
}

var Module = fx.Provide(
	fx.Annotated{
		Group:  "impl",
		Target: NewImpl,
	},
)
