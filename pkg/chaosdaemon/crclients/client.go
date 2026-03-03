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

package crclients

import (
	"context"

	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients/containerd"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients/crio"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients/docker"
)

const (
	ContainerRuntimeDocker     = "docker"
	ContainerRuntimeContainerd = "containerd"
	ContainerRuntimeCrio       = "crio"

	defaultDockerSocket     = "unix:///var/run/docker.sock"
	defaultContainerdSocket = "/run/containerd/containerd.sock"
	defaultCrioSocket       = "/var/run/crio/crio.sock"
	containerdDefaultNS     = "k8s.io"
)

// CrClientConfig contains the basic cr client configuration.
type CrClientConfig struct {
	// Support docker, containerd, crio for now
	Runtime      string
	SocketPath   string
	ContainerdNS string
}

// ContainerRuntimeInfoClient represents a struct which can give you information about container runtime
type ContainerRuntimeInfoClient interface {
	GetPidFromContainerID(ctx context.Context, containerID string) (uint32, error)
	ContainerKillByContainerID(ctx context.Context, containerID string) error
	FormatContainerID(ctx context.Context, containerID string) (string, error)
	ListContainerIDs(ctx context.Context) ([]string, error)
	GetLabelsFromContainerID(ctx context.Context, containerID string) (map[string]string, error)
}

// CreateContainerRuntimeInfoClient creates a container runtime information client.
func CreateContainerRuntimeInfoClient(clientConfig *CrClientConfig) (ContainerRuntimeInfoClient, error) {
	// TODO: support more container runtime

	var cli ContainerRuntimeInfoClient
	var err error
	socketPath := clientConfig.SocketPath
	switch clientConfig.Runtime {
	case ContainerRuntimeDocker:
		if socketPath == "" {
			socketPath = defaultDockerSocket
		} else {
			socketPath = "unix://" + socketPath
		}
		cli, err = docker.New(socketPath, "", nil, nil)
		if err != nil {
			return nil, err
		}
	case ContainerRuntimeContainerd:
		// TODO(yeya24): add more options?
		if socketPath == "" {
			socketPath = defaultContainerdSocket
		}
		containerdNS := containerdDefaultNS
		if clientConfig.ContainerdNS != "" {
			containerdNS = clientConfig.ContainerdNS
		}
		cli, err = containerd.New(socketPath, containerd.WithDefaultNamespace(containerdNS))
		if err != nil {
			return nil, err
		}
	case ContainerRuntimeCrio:
		if socketPath == "" {
			socketPath = defaultCrioSocket
		}
		cli, err = crio.New(socketPath)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.Errorf("only docker/containerd/crio is supported, but got %s", clientConfig.Runtime)
	}

	return cli, nil
}
