// Copyright 2019 PingCAP, Inc.
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

package netem

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/go-logr/logr"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/controllers/twophase"
	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/pingcap/chaos-mesh/pkg/utils"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
)

const (
	networkDelayActionMsg = "delay network for %s"
)

// NetemSpec defines the interface to convert to a Netem protobuf
type NetemSpec interface {
	ToNetem() (*pb.Netem, error)
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

type Reconciler struct {
	client.Client
	Log logr.Logger
}

func (r *Reconciler) Object() twophase.InnerObject {
	return &v1alpha1.NetworkChaos{}
}

func (r *Reconciler) Apply(ctx context.Context, req ctrl.Request, chaos twophase.InnerObject) error {
	networkchaos, ok := chaos.(*v1alpha1.NetworkChaos)
	if !ok {
		err := errors.New("chaos is not NetworkChaos")
		r.Log.Error(err, "chaos is not NetworkChaos", "chaos", chaos)
		return err
	}

	pods, err := utils.SelectAndGeneratePods(ctx, r.Client, &networkchaos.Spec)

	if err != nil {
		r.Log.Error(err, "fail to select and generate pods")
		return err
	}

	err = r.applyAllPods(ctx, pods, networkchaos)
	if err != nil {
		return err
	}

	networkchaos.Status.Experiment.Pods = []v1alpha1.PodStatus{}

	for _, pod := range pods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Action:    string(networkchaos.Spec.Action),
			Message:   fmt.Sprintf(networkDelayActionMsg, networkchaos.Spec.Duration),
		}

		networkchaos.Status.Experiment.Pods = append(networkchaos.Status.Experiment.Pods, ps)
	}

	return nil
}

func (r *Reconciler) Recover(ctx context.Context, req ctrl.Request, chaos twophase.InnerObject) error {
	networkchaos, ok := chaos.(*v1alpha1.NetworkChaos)
	if !ok {
		err := errors.New("chaos is not NetworkChaos")
		r.Log.Error(err, "chaos is not NetworkChaos", "chaos", chaos)
		return err
	}

	err := r.cleanFinalizersAndRecover(ctx, networkchaos)
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) cleanFinalizersAndRecover(ctx context.Context, networkchaos *v1alpha1.NetworkChaos) error {
	if len(networkchaos.Finalizers) == 0 {
		return nil
	}

	for _, key := range networkchaos.Finalizers {
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
			networkchaos.Finalizers = utils.RemoveFromFinalizer(networkchaos.Finalizers, key)
			continue
		}

		err = r.recoverPod(ctx, &pod, networkchaos)
		if err != nil {
			return err
		}

		networkchaos.Finalizers = utils.RemoveFromFinalizer(networkchaos.Finalizers, key)
	}

	return nil
}

func (r *Reconciler) recoverPod(ctx context.Context, pod *v1.Pod, networkchaos *v1alpha1.NetworkChaos) error {
	r.Log.Info("try to recover pod", "namespace", pod.Namespace, "name", pod.Name)

	c, err := utils.CreateGrpcConnection(ctx, r.Client, pod)
	if err != nil {
		return err
	}
	defer c.Close()

	pbClient := pb.NewChaosDaemonClient(c)

	containerID := pod.Status.ContainerStatuses[0].ContainerID
	_, err = pbClient.DeleteNetem(context.Background(), &pb.NetemRequest{
		ContainerId: containerID,
		Netem:       nil,
	})

	if err != nil {
		r.Log.Error(err, "recover pod error", "namespace", pod.Namespace, "name", pod.Name)
	} else {
		r.Log.Info("recover pod finished", "namespace", pod.Namespace, "name", pod.Name)
	}

	return err
}

func (r *Reconciler) applyAllPods(ctx context.Context, pods []v1.Pod, networkchaos *v1alpha1.NetworkChaos) error {
	g := errgroup.Group{}
	for index := range pods {
		pod := &pods[index]

		key, err := cache.MetaNamespaceKeyFunc(pod)
		if err != nil {
			return err
		}
		networkchaos.Finalizers = utils.InsertFinalizer(networkchaos.Finalizers, key)

		g.Go(func() error {
			return r.applyPod(ctx, pod, networkchaos)
		})
	}

	return g.Wait()
}

func (r *Reconciler) applyPod(ctx context.Context, pod *v1.Pod, networkchaos *v1alpha1.NetworkChaos) error {
	r.Log.Info("Try to apply netem on pod", "namespace", pod.Namespace, "name", pod.Name)

	action := string(networkchaos.Spec.Action)
	action = strings.Title(action)

	spec, ok := reflect.Indirect(reflect.ValueOf(networkchaos.Spec)).FieldByName(action).Interface().(NetemSpec)
	if !ok {
		return fmt.Errorf("spec %s is not a NetemSpec", action)
	}

	c, err := utils.CreateGrpcConnection(ctx, r.Client, pod)
	if err != nil {
		return err
	}
	defer c.Close()

	pbClient := pb.NewChaosDaemonClient(c)

	containerID := pod.Status.ContainerStatuses[0].ContainerID

	netem, err := spec.ToNetem()
	if err != nil {
		return err
	}

	_, err = pbClient.SetNetem(ctx, &pb.NetemRequest{
		ContainerId: containerID,
		Netem:       netem,
	})

	return err
}
