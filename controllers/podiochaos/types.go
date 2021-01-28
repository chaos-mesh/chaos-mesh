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

package podiochaos

import (
	"context"
	"encoding/json"
	"fmt"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	chaosDaemonClient "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/client"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

// Reconciler applys podiochaos
type Reconciler struct {
	client.Client
	Log logr.Logger
}

// Reconcile flushes io configuration on pod
func (h *Reconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	// TODO: set the error information in the chaos status
	ctx := context.TODO()

	chaos := &v1alpha1.PodIoChaos{}
	err := h.Client.Get(ctx, req.NamespacedName, chaos)
	if err != nil {
		h.Log.Error(err, "fail to find podiochaos")
		return reconcile.Result{}, nil
	}
	h.Log.Info("updating io chaos", "pod", chaos.Namespace+"/"+chaos.Name, "spec", chaos.Spec)

	pod := &v1.Pod{}

	err = h.Client.Get(ctx, types.NamespacedName{
		Name:      chaos.Name,
		Namespace: chaos.Namespace,
	}, pod)
	if err != nil {
		h.Log.Error(err, "fail to find pod")
		return reconcile.Result{}, nil
	}

	pbClient, err := chaosDaemonClient.NewChaosDaemonClient(ctx, h, pod, config.ControllerCfg.ChaosDaemonPort)
	if err != nil {
		return reconcile.Result{}, nil
	}
	defer pbClient.Close()

	if len(pod.Status.ContainerStatuses) == 0 {
		err :=  fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
		h.Log.Error(err, "")
		return reconcile.Result{}, nil
	}

	containerID := pod.Status.ContainerStatuses[0].ContainerID
	if chaos.Spec.Container != nil &&
		len(strings.TrimSpace(*chaos.Spec.Container)) != 0 {
		containerID = ""
		for _, container := range pod.Status.ContainerStatuses {
			if container.Name == *chaos.Spec.Container {
				containerID = container.ContainerID
				break
			}
		}
		if len(containerID) == 0 {
			err := fmt.Errorf("cannot find container with name %s", *chaos.Spec.Container)
			h.Log.Error(err, "")
			return reconcile.Result{}, nil
		}
	}

	actions, err := json.Marshal(chaos.Spec.Actions)
	if err != nil {
		h.Log.Error(err, "fail to marshal actions")
		return reconcile.Result{}, nil
	}
	input := string(actions)
	h.Log.Info("input with", "config", input)

	res, err := pbClient.ApplyIoChaos(ctx, &pb.ApplyIoChaosRequest{
		Actions:     input,
		Volume:      chaos.Spec.VolumeMountPath,
		ContainerId: containerID,

		Instance:  chaos.Spec.Pid,
		StartTime: chaos.Spec.StartTime,
		EnterNS:   true,
	})
	if err != nil {
		h.Log.Error(err, "fail to apply iochaos")
		return reconcile.Result{}, nil
	}

	chaos.Spec.Pid = res.Instance
	chaos.Spec.StartTime = res.StartTime
	chaos.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion: pod.APIVersion,
			Kind:       pod.Kind,
			Name:       pod.Name,
			UID:        pod.UID,
		},
	}

	return reconcile.Result{}, nil
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.PodIoChaos{}).
		Complete(r)

	return err
}
