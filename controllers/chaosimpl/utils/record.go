// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package utils

import (
	"context"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/controller"
	chaosdaemonclient "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/client"
)

type ContainerRecordDecoder struct {
	client.Client
	*chaosdaemon.ChaosDaemonClientBuilder
}

func NewContainerRecordDecoder(c client.Client, builder *chaosdaemon.ChaosDaemonClientBuilder) *ContainerRecordDecoder {
	return &ContainerRecordDecoder{
		Client:                   c,
		ChaosDaemonClientBuilder: builder,
	}
}

type DecodedContainerRecord struct {
	PbClient      chaosdaemonclient.ChaosDaemonClientInterface
	ContainerId   string
	ContainerName string
	Pod           *v1.Pod
}

func (d *ContainerRecordDecoder) DecodeContainerRecord(ctx context.Context, record *v1alpha1.Record, obj v1alpha1.InnerObject) (decoded DecodedContainerRecord, err error) {
	var pod v1.Pod
	podId, containerName, err := controller.ParseNamespacedNameContainer(record.Id)
	if err != nil {
		err = errors.Wrapf(ErrContainerNotFound, "container with id %s not found", record.Id)
		return
	}
	err = d.Client.Get(ctx, podId, &pod)
	if err != nil {
		err = errors.Wrapf(ErrContainerNotFound, "container with id %s not found", record.Id)
		return
	}
	decoded.Pod = &pod
	if len(pod.Status.ContainerStatuses) == 0 {
		err = errors.Wrapf(ErrContainerNotFound, "container with id %s not found", record.Id)
		return
	}

	for _, container := range pod.Status.ContainerStatuses {
		if container.Name == containerName {
			decoded.ContainerId = container.ContainerID
			decoded.ContainerName = containerName
			break
		}
	}
	if len(decoded.ContainerId) == 0 {
		err = errors.Wrapf(ErrContainerNotFound, "container with id %s not found", record.Id)
		return
	}

	decoded.PbClient, err = d.ChaosDaemonClientBuilder.Build(ctx, &pod, &types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	})
	if err != nil {
		return
	}

	return
}
