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

// Package docker adapts Docker for container ID normalization, inspection, listing, and termination.
package docker

import (
	"context"
	"fmt"
	"net/http"

	dockerclient "github.com/moby/moby/client"
	"github.com/pkg/errors"

	crerrors "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients/errors"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

const (
	dockerProtocolPrefix = "docker://"

	// containerKindLabel is a label key intending to filter sandbox container
	// ref: https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/dockershim/docker_service.go#L67-L75
	containerKindLabel     = "io.kubernetes.docker.type"
	containerKindContainer = "container"
)

// ClientInterface represents the Client, it's used to simply unit test
type ClientInterface interface {
	ContainerInspect(ctx context.Context, containerID string, options dockerclient.ContainerInspectOptions) (dockerclient.ContainerInspectResult, error)
	ContainerKill(ctx context.Context, containerID string, options dockerclient.ContainerKillOptions) (dockerclient.ContainerKillResult, error)
	ContainerList(ctx context.Context, options dockerclient.ContainerListOptions) (dockerclient.ContainerListResult, error)
}

// Client can get information from docker
type Client struct {
	client ClientInterface
}

// FormatContainerID strips protocol prefix from the container ID
func (c Client) FormatContainerID(_ context.Context, containerID string) (string, error) {
	if len(containerID) < len(dockerProtocolPrefix) {
		return "", errors.Errorf("container id %s is not a docker container id", containerID)
	}
	if containerID[0:len(dockerProtocolPrefix)] != dockerProtocolPrefix {
		return "", errors.Errorf("expected %s but got %s", dockerProtocolPrefix, containerID[0:len(dockerProtocolPrefix)])
	}
	return containerID[len(dockerProtocolPrefix):], nil
}

// GetPidFromContainerID fetches PID according to container id
func (c Client) GetPidFromContainerID(ctx context.Context, containerID string) (uint32, error) {
	id, err := c.FormatContainerID(ctx, containerID)
	if err != nil {
		return 0, err
	}
	container, err := c.client.ContainerInspect(ctx, id, dockerclient.ContainerInspectOptions{})
	if err != nil {
		return 0, err
	}

	if container.Container.State.Pid == 0 {
		return 0, errors.Errorf("container is not running, status: %s", container.Container.State.Status)
	}

	return uint32(container.Container.State.Pid), nil
}

// ContainerKillByContainerID kills container according to container id
func (c Client) ContainerKillByContainerID(ctx context.Context, containerID string) error {
	id, err := c.FormatContainerID(ctx, containerID)
	if err != nil {
		return err
	}
	_, err = c.client.ContainerKill(ctx, id, dockerclient.ContainerKillOptions{Signal: "SIGKILL"})

	return err
}

// ListContainerIDs lists all container IDs
func (c Client) ListContainerIDs(ctx context.Context) ([]string, error) {
	// filter sandbox containers
	containers, err := c.client.ContainerList(ctx, dockerclient.ContainerListOptions{
		Filters: dockerclient.Filters{
			"label": {fmt.Sprintf("%s=%s", containerKindLabel, containerKindContainer): true},
		},
	})
	if err != nil {
		return nil, err
	}

	var ids []string
	for _, container := range containers.Items {
		id := fmt.Sprintf("%s%s", dockerProtocolPrefix, container.ID)
		ids = append(ids, id)
	}
	return ids, nil
}

// GetLabelsFromContainerID returns the labels according to container ID
func (c Client) GetLabelsFromContainerID(ctx context.Context, containerID string) (map[string]string, error) {
	id, err := c.FormatContainerID(ctx, containerID)
	if err != nil {
		return nil, err
	}
	container, err := c.client.ContainerInspect(ctx, id, dockerclient.ContainerInspectOptions{})
	if err != nil {
		return nil, err
	}

	return container.Container.Config.Labels, nil
}

// GetSandboxPidFromPodUID is not supported for Docker runtime.
func (c Client) GetSandboxPidFromPodUID(_ context.Context, _ string) (uint32, error) {
	return 0, crerrors.ErrNotSupported
}

func New(host string, version string, client *http.Client, httpHeaders map[string]string) (*Client, error) {
	// Mock point to return error or mock client in unit test
	if err := mock.On("NewDockerClientError"); err != nil {
		return nil, err.(error)
	}
	if client := mock.On("MockDockerClient"); client != nil {
		return &Client{
			client: client.(ClientInterface),
		}, nil
	}

	opts := []dockerclient.Opt{
		dockerclient.WithTLSClientConfigFromEnv(),
		dockerclient.WithHostFromEnv(),
		dockerclient.WithHost(host),
		dockerclient.WithHTTPClient(client),
		dockerclient.WithHTTPHeaders(httpHeaders),
	}

	if version != "" {
		opts = append(opts, dockerclient.WithAPIVersion(version))
	} else {
		opts = append(opts, dockerclient.WithAPIVersionFromEnv())
	}

	c, err := dockerclient.New(opts...)
	if err != nil {
		return nil, err
	}
	// The real logic
	return &Client{
		client: c,
	}, nil
}
