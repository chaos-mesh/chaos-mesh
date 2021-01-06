// Copyright 2020 Chaos Mesh Authors.
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
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	kubeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/recover"
	"github.com/chaos-mesh/chaos-mesh/pkg/events"
	"github.com/chaos-mesh/chaos-mesh/pkg/finalizer"
	"github.com/chaos-mesh/chaos-mesh/pkg/grpc"
	"github.com/chaos-mesh/chaos-mesh/pkg/router"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	end "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/client"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	pb_ "github.com/chaos-mesh/chaos-mesh/pkg/chaoskernel/pb"
)

const kernelChaosMsg = "kernel is injected with %v"

// endpoint is KernelChaos reconciler
type endpoint struct {
	ctx.Context
}

type recoverer struct {
	kubeclient.Client
	Log logr.Logger
}

// Apply applies KernelChaos
func (r *endpoint) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	kernelChaos, ok := chaos.(*v1alpha1.KernelChaos)
	if !ok {
		err := errors.New("chaos is not kernelChaos")
		r.Log.Error(err, "chaos is not KernelChaos", "chaos", chaos)
		return err
	}

	pods, err := selector.SelectAndFilterPods(ctx, r.Client, r.Reader, &kernelChaos.Spec, config.ControllerCfg.ClusterScoped, config.ControllerCfg.TargetNamespace, config.ControllerCfg.AllowedNamespaces, config.ControllerCfg.IgnoredNamespaces)
	if err != nil {
		r.Log.Error(err, "failed to select and filter pods")
		return err
	}

	if err = r.applyAllPods(ctx, pods, kernelChaos); err != nil {
		r.Log.Error(err, "failed to apply chaos on all pods")
		return err
	}

	kernelChaos.Status.Experiment.PodRecords = make([]v1alpha1.PodStatus, 0, len(pods))
	for _, pod := range pods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Message:   fmt.Sprintf(kernelChaosMsg, kernelChaos.Spec.FailKernRequest),
		}

		kernelChaos.Status.Experiment.PodRecords = append(kernelChaos.Status.Experiment.PodRecords, ps)
	}
	r.Event(kernelChaos, v1.EventTypeNormal, events.ChaosInjected, "")

	return nil
}

// Recover means the reconciler recovers the chaos action
func (r *endpoint) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	kernelchaos, ok := chaos.(*v1alpha1.KernelChaos)
	if !ok {
		err := errors.New("chaos is not KernelChaos")
		r.Log.Error(err, "chaos is not KernelChaos", "chaos", chaos)
		return err
	}

	rd := recover.Delegate{Client: r.Client, Log: r.Log, RecoverIntf: &recoverer{r.Client, r.Log}}

	finalizers, err := rd.CleanFinalizersAndRecover(ctx, chaos, kernelchaos.Finalizers, kernelchaos.Annotations)
	if err != nil {
		return err
	}
	kernelchaos.Finalizers = finalizers
	r.Event(kernelchaos, v1.EventTypeNormal, events.ChaosRecovered, "")

	return nil
}

func (r *recoverer) RecoverPod(ctx context.Context, pod *v1.Pod, somechaos v1alpha1.InnerObject) error {
	// judged type in `Recover` already so no need to judge again
	chaos, _ := somechaos.(*v1alpha1.KernelChaos)
	r.Log.Info("try to recover pod", "namespace", pod.Namespace, "name", pod.Name)

	pbClient, err := client.NewChaosDaemonClient(ctx, r.Client, pod, config.ControllerCfg.ChaosDaemonPort)
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
		r.Log.Error(err, "Get container pid error", "namespace", pod.Namespace, "name", pod.Name)
		return err
	}

	r.Log.Info("Get container pid", "namespace", pod.Namespace, "name", pod.Name)
	conn, err := grpc.CreateGrpcConnection(ctx, r.Client, pod, config.ControllerCfg.BPFKIPort)
	if err != nil {
		return err
	}
	defer conn.Close()

	var callchain []*pb_.FailKernRequestFrame
	for _, frame := range chaos.Spec.FailKernRequest.Callchain {
		callchain = append(callchain, &pb_.FailKernRequestFrame{
			Funcname:   frame.Funcname,
			Parameters: frame.Parameters,
			Predicate:  frame.Predicate,
		})
	}

	bpfClient := pb_.NewBPFKIServiceClient(conn)
	_, err = bpfClient.RecoverMMOrBIO(ctx, &pb_.FailKernRequest{
		Pid:       containerResponse.Pid,
		Callchain: callchain,
	})

	return err
}

// Object would return the instance of chaos
func (r *endpoint) Object() v1alpha1.InnerObject {
	return &v1alpha1.KernelChaos{}
}

func (r *endpoint) applyAllPods(ctx context.Context, pods []v1.Pod, chaos *v1alpha1.KernelChaos) error {
	g := errgroup.Group{}
	for index := range pods {
		pod := &pods[index]

		key, err := cache.MetaNamespaceKeyFunc(pod)
		if err != nil {
			return err
		}
		chaos.Finalizers = finalizer.InsertFinalizer(chaos.Finalizers, key)

		g.Go(func() error {
			return r.applyPod(ctx, pod, chaos)
		})
	}

	return g.Wait()
}

func (r *endpoint) applyPod(ctx context.Context, pod *v1.Pod, chaos *v1alpha1.KernelChaos) error {
	r.Log.Info("Try to inject kernel on pod", "namespace", pod.Namespace, "name", pod.Name)

	pbClient, err := client.NewChaosDaemonClient(ctx, r.Client, pod, config.ControllerCfg.ChaosDaemonPort)
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
		r.Log.Error(err, "Get container pid error", "namespace", pod.Namespace, "name", pod.Name)
		return err
	}

	r.Log.Info("Get container pid", "namespace", pod.Namespace, "name", pod.Name)
	conn, err := grpc.CreateGrpcConnection(ctx, r.Client, pod, config.ControllerCfg.BPFKIPort)
	if err != nil {
		return err
	}
	defer conn.Close()

	var callchain []*pb_.FailKernRequestFrame
	for _, frame := range chaos.Spec.FailKernRequest.Callchain {
		callchain = append(callchain, &pb_.FailKernRequestFrame{
			Funcname:   frame.Funcname,
			Parameters: frame.Parameters,
			Predicate:  frame.Predicate,
		})
	}

	bpfClient := pb_.NewBPFKIServiceClient(conn)
	_, err = bpfClient.FailMMOrBIO(ctx, &pb_.FailKernRequest{
		Pid:         containerResponse.Pid,
		Ftype:       pb_.FailKernRequest_FAILTYPE(chaos.Spec.FailKernRequest.FailType),
		Headers:     chaos.Spec.FailKernRequest.Headers,
		Callchain:   callchain,
		Probability: float32(chaos.Spec.FailKernRequest.Probability) / 100,
		Times:       chaos.Spec.FailKernRequest.Times,
	})

	return err
}

func init() {
	router.Register("kernelchaos", &v1alpha1.KernelChaos{}, func(obj runtime.Object) bool {
		return true
	}, func(ctx ctx.Context) end.Endpoint {
		return &endpoint{
			Context: ctx,
		}
	})
}
