// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package kernelchaos

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/controller"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	pb_kernel "github.com/chaos-mesh/chaos-mesh/pkg/chaoskernel/pb"
	grpcUtils "github.com/chaos-mesh/chaos-mesh/pkg/grpc"
)

type Impl struct {
	client.Client
	Log logr.Logger

	chaosDaemonClientBuilder *chaosdaemon.ChaosDaemonClientBuilder
}

// Apply applies KernelChaos
func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	kernelChaos := obj.(*v1alpha1.KernelChaos)
	record := records[index]

	log := impl.Log.WithValues("chaos", kernelChaos, "record", record)
	podId := controller.ParseNamespacedName(record.Id)
	var pod v1.Pod
	err := impl.Client.Get(ctx, podId, &pod)
	if err != nil {
		log.Error(err, "fail to get pod by record")
		// TODO: handle this error
		if k8sError.IsNotFound(err) {
			return v1alpha1.NotInjected, nil
		}
		return v1alpha1.NotInjected, err
	}

	log = log.WithValues("pod", pod)

	if err = impl.applyPod(ctx, &pod, kernelChaos); err != nil {
		log.Error(err, "failed to apply chaos on pod")
		return v1alpha1.NotInjected, err
	}

	return v1alpha1.Injected, nil
}

// Recover means the reconciler recovers the chaos action
func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	kernelChaos := obj.(*v1alpha1.KernelChaos)
	record := records[index]

	log := impl.Log.WithValues("chaos", kernelChaos, "record", record)
	podId := controller.ParseNamespacedName(record.Id)
	var pod v1.Pod
	err := impl.Client.Get(ctx, podId, &pod)
	if err != nil {
		log.Error(err, "fail to get pod by record")
		// TODO: handle this error
		if k8sError.IsNotFound(err) {
			return v1alpha1.NotInjected, nil
		}
		return v1alpha1.Injected, err
	}

	log = log.WithValues("pod", pod)

	if err = impl.recoverPod(ctx, &pod, kernelChaos); err != nil {
		log.Error(err, "failed to recover chaos on pod")
		return v1alpha1.Injected, err
	}

	return v1alpha1.NotInjected, nil
}

func (impl *Impl) recoverPod(ctx context.Context, pod *v1.Pod, somechaos v1alpha1.InnerObject) error {
	// judged type in `Recover` already so no need to judge again
	chaos, _ := somechaos.(*v1alpha1.KernelChaos)
	impl.Log.Info("try to recover pod", "namespace", pod.Namespace, "name", pod.Name)

	pbClient, err := impl.chaosDaemonClientBuilder.Build(ctx, pod)
	if err != nil {
		return err
	}
	defer pbClient.Close()

	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
	}

	containerResponse, err := pbClient.ContainerGetPid(ctx, &pb.ContainerRequest{
		Action: &pb.ContainerAction{
			Action: pb.ContainerAction_GETPID,
		},
		ContainerId: pod.Status.ContainerStatuses[0].ContainerID,
	})

	if err != nil {
		impl.Log.Error(err, "Get container pid error", "namespace", pod.Namespace, "name", pod.Name)
		return err
	}

	impl.Log.Info("Get container pid", "namespace", pod.Namespace, "name", pod.Name)
	conn, err := impl.CreateBPFKIConnection(ctx, impl.Client, pod)
	if err != nil {
		return err
	}
	defer conn.Close()

	var callchain []*pb_kernel.FailKernRequestFrame
	for _, frame := range chaos.Spec.FailKernRequest.Callchain {
		callchain = append(callchain, &pb_kernel.FailKernRequestFrame{
			Funcname:   frame.Funcname,
			Parameters: frame.Parameters,
			Predicate:  frame.Predicate,
		})
	}

	bpfClient := pb_kernel.NewBPFKIServiceClient(conn)
	_, err = bpfClient.RecoverMMOrBIO(ctx, &pb_kernel.FailKernRequest{
		Pid:       containerResponse.Pid,
		Callchain: callchain,
	})

	return err
}

func (impl *Impl) applyPod(ctx context.Context, pod *v1.Pod, chaos *v1alpha1.KernelChaos) error {
	impl.Log.Info("Try to inject kernel on pod", "namespace", pod.Namespace, "name", pod.Name)

	pbClient, err := impl.chaosDaemonClientBuilder.Build(ctx, pod)
	if err != nil {
		return err
	}
	defer pbClient.Close()

	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
	}

	containerResponse, err := pbClient.ContainerGetPid(ctx, &pb.ContainerRequest{
		Action: &pb.ContainerAction{
			Action: pb.ContainerAction_GETPID,
		},
		ContainerId: pod.Status.ContainerStatuses[0].ContainerID,
	})
	if err != nil {
		impl.Log.Error(err, "Get container pid error", "namespace", pod.Namespace, "name", pod.Name)
		return err
	}

	impl.Log.Info("Get container pid", "namespace", pod.Namespace, "name", pod.Name)
	conn, err := impl.CreateBPFKIConnection(ctx, impl.Client, pod)
	if err != nil {
		return err
	}
	defer conn.Close()

	var callchain []*pb_kernel.FailKernRequestFrame
	for _, frame := range chaos.Spec.FailKernRequest.Callchain {
		callchain = append(callchain, &pb_kernel.FailKernRequestFrame{
			Funcname:   frame.Funcname,
			Parameters: frame.Parameters,
			Predicate:  frame.Predicate,
		})
	}

	bpfClient := pb_kernel.NewBPFKIServiceClient(conn)
	_, err = bpfClient.FailMMOrBIO(ctx, &pb_kernel.FailKernRequest{
		Pid:         containerResponse.Pid,
		Ftype:       pb_kernel.FailKernRequest_FAILTYPE(chaos.Spec.FailKernRequest.FailType),
		Headers:     chaos.Spec.FailKernRequest.Headers,
		Callchain:   callchain,
		Probability: float32(chaos.Spec.FailKernRequest.Probability) / 100,
		Times:       chaos.Spec.FailKernRequest.Times,
	})

	return err
}

// CreateBPFKIConnection create a grpc connection with bpfki
func (impl *Impl) CreateBPFKIConnection(ctx context.Context, c client.Client, pod *v1.Pod) (*grpc.ClientConn, error) {
	daemonIP, err := impl.chaosDaemonClientBuilder.FindDaemonIP(ctx, pod)
	if err != nil {
		return nil, err
	}
	builder := grpcUtils.Builder(daemonIP, config.ControllerCfg.BPFKIPort).
		WithDefaultTimeout().
		Insecure()
	return builder.Build()
}

func NewImpl(c client.Client, log logr.Logger, builder *chaosdaemon.ChaosDaemonClientBuilder) *common.ChaosImplPair {
	return &common.ChaosImplPair{
		Name:   "kernelchaos",
		Object: &v1alpha1.KernelChaos{},
		Impl: &Impl{
			Client:                   c,
			Log:                      log.WithName("kernelchaos"),
			chaosDaemonClientBuilder: builder,
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
