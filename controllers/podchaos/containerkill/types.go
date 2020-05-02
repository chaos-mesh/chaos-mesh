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

package containerkill

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/controllers/twophase"
	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/pingcap/chaos-mesh/pkg/utils"
)

const (
	containerKillActionMsg = "delete container %s"
)

func newReconciler(c client.Client, log logr.Logger, req ctrl.Request,
	recorder record.EventRecorder) *Reconciler {
	return &Reconciler{
		Client:        c,
		EventRecorder: recorder,
		Log:           log,
	}
}

type Reconciler struct {
	client.Client
	record.EventRecorder
	Log logr.Logger
}

// NewTwoPhaseReconciler would create Reconciler for twophase package
func NewTwoPhaseReconciler(c client.Client, log logr.Logger, req ctrl.Request,
	recorder record.EventRecorder) *twophase.Reconciler {
	r := newReconciler(c, log, req, recorder)
	return twophase.NewReconciler(r, r.Client, r.Log)
}

// Apply implements the reconciler.InnerReconciler.Apply
func (r *Reconciler) Apply(ctx context.Context, req ctrl.Request, obj v1alpha1.InnerObject) error {
	var err error
	now := time.Now()

	podchaos, ok := obj.(*v1alpha1.PodChaos)
	if !ok {
		err = errors.New("chaos is not PodChaos")
		r.Log.Error(err, "chaos is not PodChaos", "chaos", obj)
		return err
	}

	if podchaos.Spec.ContainerName == "" {
		r.Log.Error(nil, "the name of container is empty", "name", req.Name, "namespace", req.Namespace)
		return fmt.Errorf("podchaos[%s/%s] the name of container is empty", podchaos.Namespace, podchaos.Name)
	}

	pods, err := utils.SelectAndFilterPods(ctx, r.Client, &podchaos.Spec)
	if err != nil {
		r.Log.Error(err, "fail to select and filter pods")
		return err
	}

	g := errgroup.Group{}
	for podIndex := range pods {
		pod := &pods[podIndex]
		haveContainer := false

		for containerIndex := range pod.Status.ContainerStatuses {
			containerName := pod.Status.ContainerStatuses[containerIndex].Name
			containerID := pod.Status.ContainerStatuses[containerIndex].ContainerID

			if containerName == podchaos.Spec.ContainerName {
				haveContainer = true
				g.Go(func() error {
					err = r.KillContainer(ctx, pod, containerID)
					if err != nil {
						r.Log.Error(err, "failed to kill container")
					}
					return err
				})
			}
		}

		if haveContainer == false {
			r.Log.Error(nil, fmt.Sprintf("the pod %s doesn't have container %s", pod.Name, podchaos.Spec.ContainerName))
		}
	}

	if err := g.Wait(); err != nil {
		return err
	}
	if err = r.updatePodchaos(ctx, *podchaos, pods, now); err != nil {
		return err
	}
	r.Event(obj, v1.EventTypeNormal, utils.EventChaosInjected, "")
	return nil
}

// Recover implements the reconciler.InnerReconciler.Recover
func (r *Reconciler) Recover(ctx context.Context, req ctrl.Request, obj v1alpha1.InnerObject) error {
	return nil
}

// Object implements the reconciler.InnerReconciler.Object
func (r *Reconciler) Object() v1alpha1.InnerObject {
	return &v1alpha1.PodChaos{}
}

// KillContainer kills container according to containerID
// Use client in chaos-daemon
func (r *Reconciler) KillContainer(ctx context.Context, pod *v1.Pod, containerID string) error {
	r.Log.Info("Try to kill container", "namespace", pod.Namespace, "podName", pod.Name, "containerID", containerID)

	pbClient, err := utils.NewChaosDaemonClient(ctx, r.Client, pod, os.Getenv("CHAOS_DAEMON_PORT"))
	if err != nil {
		return err
	}
	defer pbClient.Close()

	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
	}

	if _, err = pbClient.ContainerKill(ctx, &pb.ContainerRequest{
		Action: &pb.ContainerAction{
			Action: pb.ContainerAction_KILL,
		},
		ContainerId: containerID,
	}); err != nil {
		r.Log.Error(err, "kill container error", "namespace", pod.Namespace, "podName", pod.Name, "containerID", containerID)
		return err
	}

	return nil
}

func (r *Reconciler) updatePodchaos(ctx context.Context, podchaos v1alpha1.PodChaos, pods []v1.Pod, now time.Time) error {
	next, err := utils.NextTime(*podchaos.Spec.Scheduler, now)
	if err != nil {
		r.Log.Error(err, "failed to get next time")
		return err
	}

	podchaos.SetNextStart(*next)

	podchaos.Status.Experiment.StartTime = &metav1.Time{
		Time: now,
	}
	podchaos.Status.Experiment.EndTime = &metav1.Time{
		Time: now,
	}

	podchaos.Status.Experiment.Pods = []v1alpha1.PodStatus{}
	for _, pod := range pods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Action:    string(podchaos.Spec.Action),
			Message:   fmt.Sprintf(containerKillActionMsg, podchaos.Spec.ContainerName),
		}

		podchaos.Status.Experiment.Pods = append(podchaos.Status.Experiment.Pods, ps)
	}

	return nil
}
