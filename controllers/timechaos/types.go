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

package timechaos

import (
	"context"
	"errors"
	"fmt"

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
)

const timeChaosMsg = "time is shifted with %v"

// Reconciler is time-chaos reconciler
type Reconciler struct {
	client.Client
	Log logr.Logger
}

// Reconcile reconciles a request from controller
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("reconciling timechaos")
	ctx := context.Background()

	var timechaos v1alpha1.TimeChaos
	if err := r.Get(ctx, req.NamespacedName, &timechaos); err != nil {
		r.Log.Error(err, "unable to get timechaos")
		return ctrl.Result{}, nil
	}

	scheduler := timechaos.GetScheduler()
	duration, err := timechaos.GetDuration()
	if err != nil {
		r.Log.Error(err, fmt.Sprintf("unable to get timechaos[%s/%s]'s duration", timechaos.Namespace, timechaos.Name))
		return ctrl.Result{}, nil
	}
	if scheduler == nil && duration == nil {
		return r.commonTimeChaos(&timechaos, req)
	} else if scheduler != nil && duration != nil {
		return r.scheduleTimeChaos(&timechaos, req)
	}

	// This should be ensured by admission webhook in the future
	r.Log.Error(fmt.Errorf("timechaos[%s/%s] spec invalid", timechaos.Namespace, timechaos.Name), "scheduler and duration should be omitted or defined at the same time")
	return ctrl.Result{}, nil
}

func (r *Reconciler) commonTimeChaos(timechaos *v1alpha1.TimeChaos, req ctrl.Request) (ctrl.Result, error) {
	cr := common.NewReconciler(r, r.Client, r.Log)
	return cr.Reconcile(req)
}

func (r *Reconciler) scheduleTimeChaos(timechaos *v1alpha1.TimeChaos, req ctrl.Request) (ctrl.Result, error) {
	sr := twophase.NewReconciler(r, r.Client, r.Log)
	return sr.Reconcile(req)
}

// Apply applies time-chaos
func (r *Reconciler) Apply(ctx context.Context, req ctrl.Request, chaos reconciler.InnerObject) error {
	timechaos, ok := chaos.(*v1alpha1.TimeChaos)
	if !ok {
		err := errors.New("chaos is not timechaos")
		r.Log.Error(err, "chaos is not TimeChaos", "chaos", chaos)
		return err
	}

	timechaos.SetDefaultValue()

	pods, err := utils.SelectAndGeneratePods(ctx, r.Client, &timechaos.Spec)

	if err != nil {
		r.Log.Error(err, "failed to select and generate pods")
		return err
	}

	err = r.applyAllPods(ctx, pods, timechaos)
	if err != nil {
		r.Log.Error(err, "failed to apply chaos on all pods")
		return err
	}

	timechaos.Status.Experiment.Pods = []v1alpha1.PodStatus{}

	for _, pod := range pods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Message:   fmt.Sprintf(timeChaosMsg, timechaos.Spec.TimeOffset),
		}

		timechaos.Status.Experiment.Pods = append(timechaos.Status.Experiment.Pods, ps)
	}

	return nil
}

// Recover means the reonciler recover the chaos action
func (r *Reconciler) Recover(ctx context.Context, req ctrl.Request, chaos reconciler.InnerObject) error {
	timechaos, ok := chaos.(*v1alpha1.TimeChaos)
	if !ok {
		err := errors.New("chaos is not TimeChaos")
		r.Log.Error(err, "chaos is not TimeChaos", "chaos", chaos)
		return err
	}

	err := r.cleanFinalizersAndRecover(ctx, timechaos)
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) cleanFinalizersAndRecover(ctx context.Context, chaos *v1alpha1.TimeChaos) error {
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

func (r *Reconciler) recoverPod(ctx context.Context, pod *v1.Pod, chaos *v1alpha1.TimeChaos) error {
	r.Log.Info("try to recover pod", "namespace", pod.Namespace, "name", pod.Name)

	c, err := utils.CreateGrpcConnection(ctx, r.Client, pod)
	if err != nil {
		return err
	}
	defer c.Close()

	pbClient := pb.NewChaosDaemonClient(c)

	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
	}

	containerID := pod.Status.ContainerStatuses[0].ContainerID

	_, err = pbClient.RecoverTimeOffset(ctx, &pb.TimeRequest{
		ContainerId: containerID,
	})

	if err != nil {
		r.Log.Error(err, "recover pod error", "namespace", pod.Namespace, "name", pod.Name)
	} else {
		r.Log.Info("recover pod finished", "namespace", pod.Namespace, "name", pod.Name)
	}

	return err
}

// Object would return the instance of chaos
func (r *Reconciler) Object() reconciler.InnerObject {
	return &v1alpha1.TimeChaos{}
}

func (r *Reconciler) applyAllPods(ctx context.Context, pods []v1.Pod, chaos *v1alpha1.TimeChaos) error {
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

func (r *Reconciler) applyPod(ctx context.Context, pod *v1.Pod, chaos *v1alpha1.TimeChaos) error {
	r.Log.Info("Try to shift time on pod", "namespace", pod.Namespace, "name", pod.Name)

	c, err := utils.CreateGrpcConnection(ctx, r.Client, pod)
	if err != nil {
		return err
	}
	defer c.Close()

	pbClient := pb.NewChaosDaemonClient(c)

	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
	}

	containerID := pod.Status.ContainerStatuses[0].ContainerID

	mask, err := utils.EncodeClkIds(chaos.Spec.ClockIds)
	if err != nil {
		return err
	}

	r.Log.Info("setting time shift", "mask", mask, "sec", chaos.Spec.TimeOffset.Sec, "nsec", chaos.Spec.TimeOffset.NSec)
	_, err = pbClient.SetTimeOffset(ctx, &pb.TimeRequest{
		ContainerId: containerID,
		Sec:         chaos.Spec.TimeOffset.Sec,
		Nsec:        chaos.Spec.TimeOffset.NSec,
		ClkIdsMask:  mask,
	})

	return err
}
