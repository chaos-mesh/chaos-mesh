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

package delay

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-logr/logr"

	"golang.org/x/sync/errgroup"

	"google.golang.org/grpc"

	"github.com/pingcap/chaos-operator/api/v1alpha1"
	"github.com/pingcap/chaos-operator/controllers/twophase"
	pb "github.com/pingcap/chaos-operator/pkg/tcdaemon/pb"
	"github.com/pingcap/chaos-operator/pkg/utils"

	v1 "k8s.io/api/core/v1"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	networkDelayActionMsg = "delay network for %s"
)

func NewConciler(c client.Client, log logr.Logger, req ctrl.Request) twophase.Reconciler {
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
	}

	pods, err := utils.SelectAndGeneratePods(ctx, r.Client, &networkchaos.Spec)

	if err != nil {
		r.Log.Error(err, "fail to select and generate pods")
		return err
	}

	err = r.delayAllPods(ctx, pods, networkchaos)
	if err != nil {
		return err
	}

	networkchaos.Status.Experiment.StartTime = &metav1.Time{
		Time: time.Now(),
	}
	networkchaos.Status.Experiment.Pods = []v1alpha1.PodStatus{}
	networkchaos.Status.Experiment.Phase = v1alpha1.ExperimentPhaseRunning

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
	}

	err := r.cleanFinalizersAndRecover(ctx, networkchaos)
	if err != nil {
		return err
	}

	networkchaos.Status.Experiment.EndTime = &metav1.Time{
		Time: time.Now(),
	}
	networkchaos.Status.Experiment.Phase = v1alpha1.ExperimentPhaseFinished

	return nil
}

func (r *Reconciler) cleanFinalizersAndRecover(ctx context.Context, networkchaos *v1alpha1.NetworkChaos) error {
	if len(networkchaos.Finalizers) == 0 {
		return nil
	}

	for index, key := range networkchaos.Finalizers {
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
			if !k8sError.IsNotFound(err) {
				return err
			}

			r.Log.Info("Pod not found", "namespace", ns, "name", name)
			networkchaos.Finalizers = utils.RemoveFromFinalizer(networkchaos.Finalizers, index)
			continue
		}

		err = r.recoverPod(ctx, &pod, networkchaos)
		if err != nil {
			return err
		}

		networkchaos.Finalizers = utils.RemoveFromFinalizer(networkchaos.Finalizers, index)
	}

	return nil
}

func (r *Reconciler) recoverPod(ctx context.Context, pod *v1.Pod, networkchaos *v1alpha1.NetworkChaos) error {
	r.Log.Info("Try to resume pod", "namespace", pod.Namespace, "name", pod.Name)

	c, err := r.createGrpcConnection(ctx, pod)
	if err != nil {
		return err
	}
	defer c.Close()

	pbClient := pb.NewTcDaemonClient(c)

	containerId := pod.Status.ContainerStatuses[0].ContainerID
	_, err = pbClient.DeleteNetem(context.Background(), &pb.NetemRequest{
		ContainerId: containerId,
		Netem:       nil,
	})

	return err
}

func (r *Reconciler) delayAllPods(ctx context.Context, pods []v1.Pod, networkchaos *v1alpha1.NetworkChaos) error {
	g := errgroup.Group{}
	for _, pod := range pods {
		g.Go(func() error {
			key, err := cache.MetaNamespaceKeyFunc(&pod)
			if err != nil {
				return err
			}
			networkchaos.Finalizers = utils.InsertFinalizer(networkchaos.Finalizers, key)

			if err := r.Update(ctx, networkchaos); err != nil {
				r.Log.Error(err, "unable to update podchaos finalizers")
				return err
			}

			return r.delayPod(ctx, &pod, networkchaos)
		})
	}

	return g.Wait()
}

func (r *Reconciler) delayPod(ctx context.Context, pod *v1.Pod, networkchaos *v1alpha1.NetworkChaos) error {
	delay := networkchaos.Spec.Delay

	r.Log.Info("Try to delay pod", "namespace", pod.Namespace, "name", pod.Name)

	c, err := r.createGrpcConnection(ctx, pod)
	if err != nil {
		return err
	}
	defer c.Close()

	pbClient := pb.NewTcDaemonClient(c)

	containerId := pod.Status.ContainerStatuses[0].ContainerID

	delayTime, err := time.ParseDuration(delay.Latency)
	if err != nil {
		r.Log.Error(err, "fail to parse delay time")
	}
	jitter, err := time.ParseDuration(delay.Jitter)
	if err != nil {
		r.Log.Error(err, "fail to parse delay jitter")
	}

	delayCorr, err := strconv.ParseFloat(delay.Correlation, 32)
	if err != nil {
		r.Log.Error(err, "fail to parse delay correlation")
	}
	_, err = pbClient.SetNetem(context.Background(), &pb.NetemRequest{
		ContainerId: containerId,
		Netem: &pb.Netem{
			Time:      uint32(delayTime.Nanoseconds() / 1e3),
			DelayCorr: float32(delayCorr),
			Jitter:    uint32(jitter.Nanoseconds() / 1e3),
		},
	})

	return err
}

func (r *Reconciler) createGrpcConnection(ctx context.Context, pod *v1.Pod) (*grpc.ClientConn, error) {
	port := os.Getenv("TC_DAEMON_PORT")
	if port == "" {
		port = "8080"
	}

	nodeName := pod.Spec.NodeName
	r.Log.Info("Creating client to tcdaemon", "node", nodeName)

	var node v1.Node
	err := r.Get(ctx, types.NamespacedName{
		Name: nodeName,
	}, &node)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", node.Status.Addresses[0].Address, port), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return conn, nil
}
