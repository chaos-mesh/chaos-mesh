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

package containerkill

import (
	"context"
	"fmt"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/controller"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Impl struct {
	client.Client

	Log logr.Logger
}

func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	var pod v1.Pod
	podId, containerName := controller.ParseNamespacedNameContainer(records[index].Id)
	err := impl.Get(ctx, podId, &pod)
	if err != nil {
		// TODO: handle this error
		return v1alpha1.NotInjected, err
	}

	pbClient, err := chaosdaemon.NewChaosDaemonClient(ctx, impl.Client, &pod)
	defer pbClient.Close()
	if len(pod.Status.ContainerStatuses) == 0 {
		// TODO: organize the error in a better way
		return v1alpha1.NotInjected, fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
	}

	containerID := ""
	for _, container := range pod.Status.ContainerStatuses {
		if container.Name == containerName {
			containerID = container.ContainerID
			break
		}
	}
	if len(containerID) == 0 {
		// TODO: organize the error in a better way
		return v1alpha1.NotInjected, fmt.Errorf("cannot find container %s in %s", containerName, podId.String())
	}

	if _, err = pbClient.ContainerKill(ctx, &pb.ContainerRequest{
		Action: &pb.ContainerAction{
			Action: pb.ContainerAction_KILL,
		},
		ContainerId: containerID,
	}); err != nil {
		impl.Log.Error(err, "kill container error", "namespace", pod.Namespace, "podName", pod.Name, "containerID", containerID)
		return v1alpha1.NotInjected, err
	}

	return v1alpha1.Injected, nil
}

func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	return v1alpha1.NotInjected, nil
}

func NewImpl(c client.Client, log logr.Logger) *Impl {
	return &Impl{
		Client: c,
		Log: log.WithName("containerkill"),
	}
}
