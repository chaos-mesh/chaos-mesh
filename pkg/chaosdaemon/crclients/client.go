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

package crclients

import (
	"context"
	"fmt"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients/containerd"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients/docker"
)

const (
	containerRuntimeDocker     = "docker"
	containerRuntimeContainerd = "containerd"

	// TODO(yeya24): make socket and ns configurable
	defaultDockerSocket  = "unix:///var/run/docker.sock"
	defaultContainerdSocket  = "/run/containerd/containerd.sock"
	containerdDefaultNS      = "k8s.io"
)

// ContainerRuntimeInfoClient represents a struct which can give you information about container runtime
type ContainerRuntimeInfoClient interface {
	GetPidFromContainerID(ctx context.Context, containerID string) (uint32, error)
	ContainerKillByContainerID(ctx context.Context, containerID string) error
	FormatContainerID(ctx context.Context, containerID string) (string, error)
}

// CreateContainerRuntimeInfoClient creates a container runtime information client.
func CreateContainerRuntimeInfoClient(containerRuntime string) (ContainerRuntimeInfoClient, error) {
	// TODO: support more container runtime

	var cli ContainerRuntimeInfoClient
	var err error
	switch containerRuntime {
	case containerRuntimeDocker:
		cli, err = docker.New(defaultDockerSocket, "", nil, nil)
		if err != nil {
			return nil, err
		}

	case containerRuntimeContainerd:
		// TODO(yeya24): add more options?
		cli, err = containerd.New(defaultContainerdSocket, containerd.WithDefaultNamespace(containerdDefaultNS))
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("only docker and containerd is supported, but got %s", containerRuntime)
	}

	return cli, nil
}
