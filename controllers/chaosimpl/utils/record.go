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

package utils

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/controller"
	chaosdaemonclient "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/client"
)

type ContianerRecordDecoder struct {
	client.Client
	*chaosdaemon.ChaosDaemonClientBuilder
}

func NewContainerRecordDecoder(c client.Client, builder *chaosdaemon.ChaosDaemonClientBuilder) *ContianerRecordDecoder {
	return &ContianerRecordDecoder{
		Client:                   c,
		ChaosDaemonClientBuilder: builder,
	}
}

type DecodedContainerRecord struct {
	PbClient    chaosdaemonclient.ChaosDaemonClientInterface
	ContainerId string

	Pod *v1.Pod
}

func (d *ContianerRecordDecoder) DecodeContainerRecord(ctx context.Context, record *v1alpha1.Record) (decoded DecodedContainerRecord, err error) {
	var pod v1.Pod
	podId, containerName := controller.ParseNamespacedNameContainer(record.Id)
	err = d.Client.Get(ctx, podId, &pod)
	if err != nil {
		// TODO: organize the error in a better way
		err = NewFailToFindContainer(pod.Namespace, pod.Name, containerName, err)
		return
	}
	decoded.Pod = &pod
	if len(pod.Status.ContainerStatuses) == 0 {
		// TODO: organize the error in a better way
		err = NewFailToFindContainer(pod.Namespace, pod.Name, containerName, nil)
		return
	}

	for _, container := range pod.Status.ContainerStatuses {
		if container.Name == containerName {
			decoded.ContainerId = container.ContainerID
			break
		}
	}
	if len(decoded.ContainerId) == 0 {
		// TODO: organize the error in a better way
		err = NewFailToFindContainer(pod.Namespace, pod.Name, containerName, nil)
		return
	}

	decoded.PbClient, err = d.ChaosDaemonClientBuilder.Build(ctx, &pod)
	if err != nil {
		return
	}

	return
}
