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
	"strconv"

	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/controllers/twophase"
	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/pingcap/chaos-mesh/pkg/utils"
)

type Reconciler struct {
	client.Client
	Log logr.Logger
}

func NewReconciler(c client.Client, log logr.Logger, req ctrl.Request) twophase.Reconciler {
	return twophase.Reconciler{
		InnerReconciler: &Reconciler{
			Client: c,
			Log:    log,
		},
		Client: c,
		Log:    log,
	}
}

func (r *Reconciler) Object() twophase.InnerObject {
	return &v1alpha1.TimeChaos{}
}

// Apply is a functions used to apply time chaos.
func (r *Reconciler) Apply(ctx context.Context, req ctrl.Request, chaos twophase.InnerObject) error {
	r.Log.Info("applying time")

	timechaos, ok := chaos.(*v1alpha1.TimeChaos)
	if !ok {
		err := errors.New("chaos is not TimeChaos")
		r.Log.Error(err, "chaos is not TimeChaos", "chaos", chaos)

		return err
	}

	allPods, err := utils.SelectAndGeneratePods(ctx, r.Client, &timechaos.Spec)

	if err != nil {
		r.Log.Error(err, "failed to select and generate pods")
		return err
	}

	// Set up ipset in every related pods
	g := errgroup.Group{}
	for index := range allPods {
		pod := &allPods[index]

		key, err := cache.MetaNamespaceKeyFunc(pod)
		if err != nil {
			return err
		}
		timechaos.Finalizers = utils.InsertFinalizer(timechaos.Finalizers, key)

		r.Log.Info("PODS", "name", pod.Name, "namespace", pod.Namespace)
		g.Go(func() error {
			err := r.setTimeOffset(ctx, pod, timechaos.Spec.GetValue())
			if err != nil {
				return err
			}
			return nil
		})
	}

	if err = g.Wait(); err != nil {
		r.Log.Error(err, "grpc error")
		return err
	}

	timechaos.Status.Experiment.Pods = []v1alpha1.PodStatus{}

	for _, pod := range allPods {
		message := "time chaos without duration"
		if timechaos.Spec.Duration != nil {
			message = fmt.Sprintf("time chaos for %s", *timechaos.Spec.Duration)
		}
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Message:   message,
		}

		timechaos.Status.Experiment.Pods = append(timechaos.Status.Experiment.Pods, ps)
	}

	return nil
}

func (r *Reconciler) setTimeOffset(ctx context.Context, pod *v1.Pod, value string) error {
	offset, err := strconv.ParseInt(value, 0, 64)
	if err != nil {
		return err
	}

	r.Log.Info("setTimeOffset", "value", offset)
	c, err := utils.CreateGrpcConnection(ctx, r.Client, pod)
	if err != nil {
		r.Log.Info("create grpc error", "error", err)
		return err
	}
	defer c.Close()

	pbClient := pb.NewChaosDaemonClient(c)

	containerID := pod.Status.ContainerStatuses[0].ContainerID

	_, err = pbClient.SetTimeOffset(ctx, &pb.TimeRequest{
		ContainerId: containerID,
		Sec:         int32(offset / 1000),
		Usec:        int32(offset % 1000),
	})

	r.Log.Info("setTimeOffset return", "err", err)
	return err
}

func (r *Reconciler) Recover(ctx context.Context, req ctrl.Request, chaos twophase.InnerObject) error {
	timechaos, ok := chaos.(*v1alpha1.TimeChaos)
	if !ok {
		err := errors.New("chaos is not TimeChaos")
		r.Log.Error(err, "chaos is not TimeChaos", "chaos", chaos)

		return err
	}

	for _, key := range timechaos.Finalizers {
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
			timechaos.Finalizers = utils.RemoveFromFinalizer(timechaos.Finalizers, key)
			continue
		}

		r.Log.Info("recover", "pod", pod.Name)
		if err := r.recoverTimeOffset(ctx, &pod); err != nil {
			return err
		}

		timechaos.Finalizers = utils.RemoveFromFinalizer(timechaos.Finalizers, key)
	}

	return nil
}

func (r *Reconciler) recoverTimeOffset(ctx context.Context, pod *v1.Pod) error {
	c, err := utils.CreateGrpcConnection(ctx, r.Client, pod)
	if err != nil {
		r.Log.Info("create grpc error", "error", err)
		return err
	}
	defer c.Close()

	pbClient := pb.NewChaosDaemonClient(c)

	containerID := pod.Status.ContainerStatuses[0].ContainerID

	_, err = pbClient.RecoverTimeOffset(ctx, &pb.TimeRequest{
		ContainerId: containerID,
	})

	r.Log.Info("recoverTimeOffset return", "err", err)
	return err
}
