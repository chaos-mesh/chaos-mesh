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

package containerd

import (
	"context"
	"fmt"
	"syscall"

	"github.com/containerd/containerd"
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

const (
	containerdProtocolPrefix = "containerd://"

	// containerKindLabel is a label key intending to filter sandbox container
	// ref: https://github.com/containerd/containerd/blob/main/pkg/cri/server/helpers.go#L74-L80
	containerKindLabel     = "io.cri-containerd.kind"
	containerKindContainer = "container"
)

// ContainerdClientInterface represents the ContainerClient, it's used to simply unit test
type ContainerdClientInterface interface {
	LoadContainer(ctx context.Context, id string) (containerd.Container, error)
	Containers(ctx context.Context, filters ...string) ([]containerd.Container, error)
}

// ContainerdClient can get information from containerd
type ContainerdClient struct {
	client ContainerdClientInterface
}

// FormatContainerID strips protocol prefix from the container ID
func (c ContainerdClient) FormatContainerID(ctx context.Context, containerID string) (string, error) {
	if len(containerID) < len(containerdProtocolPrefix) {
		return "", errors.Errorf("container id %s is not a containerd container id", containerID)
	}
	if containerID[0:len(containerdProtocolPrefix)] != containerdProtocolPrefix {
		return "", errors.Errorf("expected %s but got %s", containerdProtocolPrefix, containerID[0:len(containerdProtocolPrefix)])
	}
	return containerID[len(containerdProtocolPrefix):], nil
}

// GetPidFromContainerID fetches PID according to container id
func (c ContainerdClient) GetPidFromContainerID(ctx context.Context, containerID string) (uint32, error) {
	id, err := c.FormatContainerID(ctx, containerID)
	if err != nil {
		return 0, err
	}
	container, err := c.client.LoadContainer(ctx, id)
	if err != nil {
		return 0, err
	}
	task, err := container.Task(ctx, nil)
	if err != nil {
		return 0, err
	}
	return task.Pid(), nil
}

// ContainerKillByContainerID kills container according to container id
func (c ContainerdClient) ContainerKillByContainerID(ctx context.Context, containerID string) error {
	containerID, err := c.FormatContainerID(ctx, containerID)
	if err != nil {
		return err
	}

	container, err := c.client.LoadContainer(ctx, containerID)
	if err != nil {
		return err
	}
	task, err := container.Task(ctx, nil)
	if err != nil {
		return err
	}

	err = task.Kill(ctx, syscall.SIGKILL)

	return err
}

// ListContainerIDs lists all container IDs
func (c ContainerdClient) ListContainerIDs(ctx context.Context) ([]string, error) {
	// filter sandbox containers
	// ref: https://github.com/containerd/containerd/blob/main/pkg/cri/server/helpers.go#L281-L285
	filter := fmt.Sprintf("labels.%q==%q", containerKindLabel, containerKindContainer)
	containers, err := c.client.Containers(ctx, filter)
	if err != nil {
		return nil, err
	}

	var ids []string
	for _, container := range containers {
		id := fmt.Sprintf("%s%s", containerdProtocolPrefix, container.ID())
		ids = append(ids, id)
	}
	return ids, nil
}

// GetLabelsFromContainerID returns the labels according to container ID
func (c ContainerdClient) GetLabelsFromContainerID(ctx context.Context, containerID string) (map[string]string, error) {
	id, err := c.FormatContainerID(ctx, containerID)
	if err != nil {
		return nil, err
	}

	container, err := c.client.LoadContainer(ctx, id)
	if err != nil {
		return nil, err
	}

	labels, err := container.Labels(ctx)
	if err != nil {
		return nil, err
	}

	return labels, nil
}

func New(address string, opts ...containerd.ClientOpt) (*ContainerdClient, error) {
	// Mock point to return error in unit test
	if err := mock.On("NewContainerdClientError"); err != nil {
		return nil, err.(error)
	}
	if client := mock.On("MockContainerdClient"); client != nil {
		return &ContainerdClient{
			client.(ContainerdClientInterface),
		}, nil
	}

	c, err := containerd.New(address, opts...)
	if err != nil {
		return nil, err
	}
	// The real logic
	return &ContainerdClient{
		client: c,
	}, nil
}

// WithDefaultNamespace is an alias for the function in containerd with the same name
var WithDefaultNamespace = containerd.WithDefaultNamespace
