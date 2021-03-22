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

package docker

import (
	"context"
	"fmt"
	"net/http"

	"github.com/docker/docker/api/types"
	dockerclient "github.com/docker/docker/client"

	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

const (
	dockerProtocolPrefix = "docker://"
)

// DockerClientInterface represents the DockerClient, it's used to simply unit test
type DockerClientInterface interface {
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
	ContainerKill(ctx context.Context, containerID, signal string) error
}

// DockerClient can get information from docker
type DockerClient struct {
	client DockerClientInterface
}

// FormatContainerID strips protocol prefix from the container ID
func (c DockerClient) FormatContainerID(ctx context.Context, containerID string) (string, error) {
	if len(containerID) < len(dockerProtocolPrefix) {
		return "", fmt.Errorf("container id %s is not a docker container id", containerID)
	}
	if containerID[0:len(dockerProtocolPrefix)] != dockerProtocolPrefix {
		return "", fmt.Errorf("expected %s but got %s", dockerProtocolPrefix, containerID[0:len(dockerProtocolPrefix)])
	}
	return containerID[len(dockerProtocolPrefix):], nil
}

// GetPidFromContainerID fetches PID according to container id
func (c DockerClient) GetPidFromContainerID(ctx context.Context, containerID string) (uint32, error) {
	id, err := c.FormatContainerID(ctx, containerID)
	if err != nil {
		return 0, err
	}
	container, err := c.client.ContainerInspect(ctx, id)
	if err != nil {
		return 0, err
	}

	if container.State.Pid == 0 {
		return 0, fmt.Errorf("container is not running, status: %s", container.State.Status)
	}

	return uint32(container.State.Pid), nil
}

// ContainerKillByContainerID kills container according to container id
func (c DockerClient) ContainerKillByContainerID(ctx context.Context, containerID string) error {
	id, err := c.FormatContainerID(ctx, containerID)
	if err != nil {
		return err
	}
	err = c.client.ContainerKill(ctx, id, "SIGKILL")

	return err
}

func New(host string, version string, client *http.Client, httpHeaders map[string]string) (*DockerClient, error) {
	// Mock point to return error or mock client in unit test
	if err := mock.On("NewDockerClientError"); err != nil {
		return nil, err.(error)
	}
	if client := mock.On("MockDockerClient"); client != nil {
		return &DockerClient{
			client: client.(DockerClientInterface),
		}, nil
	}

	c, err := dockerclient.NewClientWithOpts(
		dockerclient.WithHost(host),
		dockerclient.WithVersion(version),
		dockerclient.WithHTTPClient(client),
		dockerclient.WithHTTPHeaders(httpHeaders))
	if err != nil {
		return nil, err
	}
	// The real logic
	return &DockerClient{
		client: c,
	}, nil
}
