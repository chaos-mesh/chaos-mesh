// Copyright 2020 PingCAP, Inc.
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
	"os"

	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/controllers/common"
	"github.com/pingcap/chaos-mesh/controllers/reconciler"
	"github.com/pingcap/chaos-mesh/controllers/twophase"
	"github.com/pingcap/chaos-mesh/pkg/utils"

	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
	pb_ "github.com/pingcap/chaos-mesh/pkg/chaoskernel/pb"
)

const kernelChaosMsg = "kernel is injected with %v"

// Reconciler is KernelChaos reconciler
type Reconciler struct {
	client.Client
	Log logr.Logger
}

// Reconcile reconciles a request from controller
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("reconciling kernelchaos")
	ctx := context.Background()

	var kernelchaos v1alpha1.KernelChaos
	if err := r.Get(ctx, req.NamespacedName, &kernelchaos); err != nil {
		r.Log.Error(err, "unable to get kernelchaos")
		return ctrl.Result{}, nil
	}

	scheduler := kernelchaos.GetScheduler()
	duration, err := kernelchaos.GetDuration()
	if err != nil {
		r.Log.Error(err, fmt.Sprintf("unable to get kernelchaos[%s/%s]'s duration", kernelchaos.Namespace, kernelchaos.Name))
		return ctrl.Result{}, nil
	}
	if scheduler == nil && duration == nil {
		return r.commonKernelChaos(&kernelchaos, req)
	} else if scheduler != nil && duration != nil {
		return r.scheduleKernelChaos(&kernelchaos, req)
	}

	// This should be ensured by admission webhook in the future
	r.Log.Error(fmt.Errorf("kernelchaos[%s/%s] spec invalid", kernelchaos.Namespace, kernelchaos.Name), "scheduler and duration should be omitted or defined at the same time")
	return ctrl.Result{}, nil
}

func (r *Reconciler) commonKernelChaos(kernelchaos *v1alpha1.KernelChaos, req ctrl.Request) (ctrl.Result, error) {
	cr := common.NewReconciler(r, r.Client, r.Log)
	return cr.Reconcile(req)
}

func (r *Reconciler) scheduleKernelChaos(kernelchaos *v1alpha1.KernelChaos, req ctrl.Request) (ctrl.Result, error) {
	sr := twophase.NewReconciler(r, r.Client, r.Log)
	return sr.Reconcile(req)
}

// Apply applies KernelChaos
func (r *Reconciler) Apply(ctx context.Context, req ctrl.Request, chaos reconciler.InnerObject) error {
	kernelchaos, ok := chaos.(*v1alpha1.KernelChaos)
	if !ok {
		err := errors.New("chaos is not kernelchaos")
		r.Log.Error(err, "chaos is not KernelChaos", "chaos", chaos)
		return err
	}

	pods, err := utils.SelectAndGeneratePods(ctx, r.Client, &kernelchaos.Spec)

	if err != nil {
		r.Log.Error(err, "failed to select and generate pods")
		return err
	}

	err = r.applyAllPods(ctx, pods, kernelchaos)
	if err != nil {
		r.Log.Error(err, "failed to apply chaos on all pods")
		return err
	}

	kernelchaos.Status.Experiment.Pods = []v1alpha1.PodStatus{}

	for _, pod := range pods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Message:   fmt.Sprintf(kernelChaosMsg, kernelchaos.Spec.FailKernRequest),
		}

		kernelchaos.Status.Experiment.Pods = append(kernelchaos.Status.Experiment.Pods, ps)
	}

	return nil
}

// Recover means the reconciler recovers the chaos action
func (r *Reconciler) Recover(ctx context.Context, req ctrl.Request, chaos reconciler.InnerObject) error {
	kernelchaos, ok := chaos.(*v1alpha1.KernelChaos)
	if !ok {
		err := errors.New("chaos is not KernelChaos")
		r.Log.Error(err, "chaos is not KernelChaos", "chaos", chaos)
		return err
	}

	err := r.cleanFinalizersAndRecover(ctx, kernelchaos)
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) cleanFinalizersAndRecover(ctx context.Context, chaos *v1alpha1.KernelChaos) error {
	if len(chaos.Finalizers) == 0 {
		return nil
	}

	for _, key := range chaos.Finalizers {
		ns, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			return err
		}

		var pod v1.Pod
		err = r.Get(ctx, types.NamespacedName{
			Namespace: ns,
			Name:      name,
		}, &pod)

		if err != nil {
			if !k8serror.IsNotFound(err) {
				return err
			}

			r.Log.Info("Pod not found", "namespace", ns, "name", name)
			chaos.Finalizers = utils.RemoveFromFinalizer(chaos.Finalizers, key)
			continue
		}

		err = r.recoverPod(ctx, &pod, chaos)
		if err != nil {
			return err
		}

		chaos.Finalizers = utils.RemoveFromFinalizer(chaos.Finalizers, key)
	}

	return nil
}

func (r *Reconciler) recoverPod(ctx context.Context, pod *v1.Pod, chaos *v1alpha1.KernelChaos) error {
	r.Log.Info("try to recover pod", "namespace", pod.Namespace, "name", pod.Name)

	c1, err := utils.CreateGrpcConnection(ctx, r.Client, pod, os.Getenv("CHAOS_DAEMON_PORT"))
	if err != nil {
		return err
	}
	defer c1.Close()

	pbClient := pb.NewChaosDaemonClient(c1)

	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
	}

	containerID := pod.Status.ContainerStatuses[0].ContainerID

	containerResponse, err := pbClient.ContainerGetPid(ctx, &pb.ContainerRequest{
		Action: &pb.ContainerAction{
			Action: pb.ContainerAction_GETPID,
		},
		ContainerId: containerID,
	})

	if err != nil {
		r.Log.Error(err, "Get container pid error", "namespace", pod.Namespace, "name", pod.Name)
	} else {
		r.Log.Info("Get container pid", "namespace", pod.Namespace, "name", pod.Name)
	}

	c2, err := utils.CreateGrpcConnection(ctx, r.Client, pod, os.Getenv("BPFKI_PORT"))
	if err != nil {
		return err
	}
	defer c2.Close()

	var callchain []*pb_.FailKernRequestFrame
	for _, frame := range chaos.Spec.FailKernRequest.Callchain {
		callchain = append(callchain, &pb_.FailKernRequestFrame{
			Funcname:   frame.Funcname,
			Parameters: frame.Parameters,
			Predicate:  frame.Predicate,
		})
	}

	pbClient2 := pb_.NewBPFKIServiceClient(c2)
	_, err = pbClient2.RecoverMMOrBIO(ctx, &pb_.FailKernRequest{
		Pid:       containerResponse.Pid,
		Callchain: callchain,
	})

	return err
}

// Object would return the instance of chaos
func (r *Reconciler) Object() reconciler.InnerObject {
	return &v1alpha1.KernelChaos{}
}

func (r *Reconciler) applyAllPods(ctx context.Context, pods []v1.Pod, chaos *v1alpha1.KernelChaos) error {
	g := errgroup.Group{}
	for index := range pods {
		pod := &pods[index]

		key, err := cache.MetaNamespaceKeyFunc(pod)
		if err != nil {
			return err
		}
		chaos.Finalizers = utils.InsertFinalizer(chaos.Finalizers, key)

		g.Go(func() error {
			return r.applyPod(ctx, pod, chaos)
		})
	}

	return g.Wait()
}

func (r *Reconciler) applyPod(ctx context.Context, pod *v1.Pod, chaos *v1alpha1.KernelChaos) error {
	r.Log.Info("Try to inject kernel on pod", "namespace", pod.Namespace, "name", pod.Name)

	c1, err := utils.CreateGrpcConnection(ctx, r.Client, pod, os.Getenv("CHAOS_DAEMON_PORT"))
	if err != nil {
		return err
	}
	defer c1.Close()

	pbClient := pb.NewChaosDaemonClient(c1)

	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
	}

	containerID := pod.Status.ContainerStatuses[0].ContainerID

	containerResponse, err := pbClient.ContainerGetPid(ctx, &pb.ContainerRequest{
		Action: &pb.ContainerAction{
			Action: pb.ContainerAction_GETPID,
		},
		ContainerId: containerID,
	})

	if err != nil {
		r.Log.Error(err, "Get container pid error", "namespace", pod.Namespace, "name", pod.Name)
	} else {
		r.Log.Info("Get container pid", "namespace", pod.Namespace, "name", pod.Name)
	}

	c2, err := utils.CreateGrpcConnection(ctx, r.Client, pod, os.Getenv("BPFKI_PORT"))
	if err != nil {
		return err
	}
	defer c2.Close()

	var callchain []*pb_.FailKernRequestFrame
	for _, frame := range chaos.Spec.FailKernRequest.Callchain {
		callchain = append(callchain, &pb_.FailKernRequestFrame{
			Funcname:   frame.Funcname,
			Parameters: frame.Parameters,
			Predicate:  frame.Predicate,
		})
	}

	pbClient2 := pb_.NewBPFKIServiceClient(c2)
	_, err = pbClient2.FailMMOrBIO(ctx, &pb_.FailKernRequest{
		Pid:         containerResponse.Pid,
		Ftype:       pb_.FailKernRequest_FAILTYPE(chaos.Spec.FailKernRequest.FailType),
		Headers:     chaos.Spec.FailKernRequest.Headers,
		Callchain:   callchain,
		Probability: float32(chaos.Spec.FailKernRequest.Probability) / 100,
		Times:       chaos.Spec.FailKernRequest.Times,
	})

	return err
}
