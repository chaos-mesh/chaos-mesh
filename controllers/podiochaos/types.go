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

// Handler applys podiochaos
type Handler struct {
	client.Client
	Log logr.Logger
}

// Apply flushes io configuration on pod
func (h *Handler) Apply(ctx context.Context, chaos *v1alpha1.PodIoChaos) error {
	h.Log.Info("updating io chaos", "pod", chaos.Namespace+"/"+chaos.Name, "spec", chaos.Spec)

	pod := &v1.Pod{}

	err := h.Client.Get(ctx, types.NamespacedName{
		Name:      chaos.Name,
		Namespace: chaos.Namespace,
	}, pod)
	if err != nil {
		h.Log.Error(err, "fail to find pod")
		return err
	}

	pbClient, err := chaosDaemonClient.NewChaosDaemonClient(ctx, h, pod, config.ControllerCfg.ChaosDaemonPort)
	if err != nil {
		return err
	}
	defer pbClient.Close()

	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
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
			return fmt.Errorf("cannot find container with name %s", *chaos.Spec.Container)
		}
	}

	actions, err := json.Marshal(chaos.Spec.Actions)
	if err != nil {
		return err
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
		return err
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

	return nil
}
