// Copyright 2019 PingCAP, Inc.
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

package chaosdaemon

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/containerd/containerd"
	dockerclient "github.com/docker/docker/client"
)

const (
	containerRuntimeDocker     = "docker"
	containerRuntimeContainerd = "containerd"

	defaultDockerSocket  = "unix:///var/run/docker.sock"
	dockerProtocolPrefix = "docker://"

	// TODO(yeya24): make socket and ns configurable
	defaultContainerdSocket  = "/run/containerd/containerd.sock"
	containerdProtocolPrefix = "containerd://"
	containerdDefaultNS      = "k8s.io"

	defaultProcPrefix = "/mnt/proc"
)

// ContainerRuntimeInfoClient represents a struct which can give you information about container runtime
type ContainerRuntimeInfoClient interface {
	GetPidFromContainerID(ctx context.Context, containerID string) (uint32, error)
}

// DockerClient can get information from docker
type DockerClient struct {
	client *dockerclient.Client
}

// GetPidFromContainerID fetches PID according to container id
func (c DockerClient) GetPidFromContainerID(ctx context.Context, containerID string) (uint32, error) {
	if containerID[0:len(dockerProtocolPrefix)] != dockerProtocolPrefix {
		return 0, fmt.Errorf("expected %s but got %s", dockerProtocolPrefix, containerID[0:len(dockerProtocolPrefix)])
	}
	container, err := c.client.ContainerInspect(ctx, containerID[len(dockerProtocolPrefix):])
	if err != nil {
		return 0, err
	}

	return uint32(container.State.Pid), nil
}

// ContainerdClient can get information from containerd
type ContainerdClient struct {
	client *containerd.Client
}

// GetPidFromContainerID fetches PID according to container id
func (c ContainerdClient) GetPidFromContainerID(ctx context.Context, containerID string) (uint32, error) {
	if containerID[0:len(containerdProtocolPrefix)] != containerdProtocolPrefix {
		return 0, fmt.Errorf("expected %s but got %s", containerdProtocolPrefix, containerID[0:len(dockerProtocolPrefix)])
	}
	container, err := c.client.LoadContainer(ctx, containerID[len(containerdProtocolPrefix):])
	if err != nil {
		return 0, err
	}
	task, err := container.Task(ctx, nil)
	if err != nil {
		return 0, err
	}
	return task.Pid(), nil
}

// CreateContainerRuntimeInfoClient creates a container runtime information client.
func CreateContainerRuntimeInfoClient(containerRuntime string) (ContainerRuntimeInfoClient, error) {
	// TODO: support more container runtime

	var cli ContainerRuntimeInfoClient
	switch containerRuntime {
	case containerRuntimeDocker:
		client, err := dockerclient.NewClient(defaultDockerSocket, "", nil, nil)
		if err != nil {
			return nil, err
		}
		cli = DockerClient{client}

	case containerRuntimeContainerd:
		// TODO(yeya24): add more options?
		client, err := containerd.New(defaultContainerdSocket, containerd.WithDefaultNamespace(containerdDefaultNS))
		if err != nil {
			return nil, err
		}
		cli = ContainerdClient{client}

	default:
		return nil, fmt.Errorf("only docker and containerd is supported, but got %s", containerRuntime)
	}

	return cli, nil
}

// GetNetnsPath returns network namespace path
func GenNetnsPath(pid uint32) string {
	return fmt.Sprintf("%s/%d/ns/net", defaultProcPrefix, pid)
}

func withNetNS(ctx context.Context, nsPath string, cmd string, args ...string) *exec.Cmd {
	// BusyBox's nsenter is very confusing. This usage is found by several attempts
	args = append([]string{"-n" + nsPath, "--", cmd}, args...)

	return exec.CommandContext(ctx, "nsenter", args...)
}
