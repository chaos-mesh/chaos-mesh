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
	"net/http"
	"os/exec"
	"syscall"

	"github.com/containerd/containerd"
	"github.com/docker/docker/api/types"
	dockerclient "github.com/docker/docker/client"

	"github.com/pingcap/chaos-mesh/pkg/mock"
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

	defaultProcPrefix = "/proc"
)

// ContainerRuntimeInfoClient represents a struct which can give you information about container runtime
type ContainerRuntimeInfoClient interface {
	GetPidFromContainerID(ctx context.Context, containerID string) (uint32, error)
	ContainerKillByContainerID(ctx context.Context, containerID string) error
}

// DockerClientInterface represents the DockerClient, it's used to simply unit test
type DockerClientInterface interface {
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
	ContainerKill(ctx context.Context, containerID, signal string) error
}

// DockerClient can get information from docker
type DockerClient struct {
	client DockerClientInterface
}

// GetPidFromContainerID fetches PID according to container id
func (c DockerClient) GetPidFromContainerID(ctx context.Context, containerID string) (uint32, error) {
	if len(containerID) < len(dockerProtocolPrefix) {
		return 0, fmt.Errorf("container id %s is not a docker container id", containerID)
	}
	if containerID[0:len(dockerProtocolPrefix)] != dockerProtocolPrefix {
		return 0, fmt.Errorf("expected %s but got %s", dockerProtocolPrefix, containerID[0:len(dockerProtocolPrefix)])
	}
	container, err := c.client.ContainerInspect(ctx, containerID[len(dockerProtocolPrefix):])
	if err != nil {
		return 0, err
	}

	return uint32(container.State.Pid), nil
}

// ContainerdClientInterface represents the ContainerClient, it's used to simply unit test
type ContainerdClientInterface interface {
	LoadContainer(ctx context.Context, id string) (containerd.Container, error)
}

// ContainerdClient can get information from containerd
type ContainerdClient struct {
	client ContainerdClientInterface
}

// GetPidFromContainerID fetches PID according to container id
func (c ContainerdClient) GetPidFromContainerID(ctx context.Context, containerID string) (uint32, error) {
	if len(containerID) < len(containerdProtocolPrefix) {
		return 0, fmt.Errorf("container id %s is not a containerd container id", containerID)
	}
	if containerID[0:len(containerdProtocolPrefix)] != containerdProtocolPrefix {
		return 0, fmt.Errorf("expected %s but got %s", containerdProtocolPrefix, containerID[0:len(containerdProtocolPrefix)])
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

// newDockerclient returns a dockerclient.NewClient with mock points
func newDockerClient(host string, version string, client *http.Client, httpHeaders map[string]string) (DockerClientInterface, error) {
	// Mock point to return error or mock client in unit test
	if err := mock.On("NewDockerClientError"); err != nil {
		return nil, err.(error)
	}
	if client := mock.On("MockDockerClient"); client != nil {
		return client.(DockerClientInterface), nil
	}

	// The real logic
	return dockerclient.NewClient(host, version, client, httpHeaders)
}

// newContainerdClient returns a containerd.New with mock points
func newContainerdClient(address string, opts ...containerd.ClientOpt) (ContainerdClientInterface, error) {
	// Mock point to return error in unit test
	if err := mock.On("NewContainerdClientError"); err != nil {
		return nil, err.(error)
	}
	if client := mock.On("MockContainerdClient"); client != nil {
		return client.(ContainerdClientInterface), nil
	}

	// The real logic
	return containerd.New(address, opts...)
}

// CreateContainerRuntimeInfoClient creates a container runtime information client.
func CreateContainerRuntimeInfoClient(containerRuntime string) (ContainerRuntimeInfoClient, error) {
	// TODO: support more container runtime

	var cli ContainerRuntimeInfoClient
	switch containerRuntime {
	case containerRuntimeDocker:
		client, err := newDockerClient(defaultDockerSocket, "", nil, nil)
		if err != nil {
			return nil, err
		}
		cli = DockerClient{client}

	case containerRuntimeContainerd:
		// TODO(yeya24): add more options?
		client, err := newContainerdClient(defaultContainerdSocket, containerd.WithDefaultNamespace(containerdDefaultNS))
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

func f(x interface{}) *exec.Cmd {
	return x.(func(...interface{}) *exec.Cmd)(1, "")
}

func withNetNS(ctx context.Context, nsPath string, cmd string, args ...string) *exec.Cmd {
	// Mock point to return mock Cmd in unit test
	if c := mock.On("MockWithNetNs"); c != nil {
		f := c.(func(context.Context, string, string, ...string) *exec.Cmd)
		return f(ctx, nsPath, cmd, args...)
	}

	// BusyBox's nsenter is very confusing. This usage is found by several attempts
	args = append([]string{"-n" + nsPath, "--", cmd}, args...)

	return exec.CommandContext(ctx, "nsenter", args...)
}

// ContainerKillByContainerID kills container according to container id
func (c DockerClient) ContainerKillByContainerID(ctx context.Context, containerID string) error {
	if len(containerID) < len(dockerProtocolPrefix) {
		return fmt.Errorf("container id %s is not a docker container id", containerID)
	}
	if containerID[0:len(dockerProtocolPrefix)] != dockerProtocolPrefix {
		return fmt.Errorf("expected %s but got %s", dockerProtocolPrefix, containerID[0:len(dockerProtocolPrefix)])
	}
	err := c.client.ContainerKill(ctx, containerID[len(dockerProtocolPrefix):], "SIGKILL")

	return err
}

// ContainerKillByContainerID kills container according to container id
func (c ContainerdClient) ContainerKillByContainerID(ctx context.Context, containerID string) error {
	if len(containerID) < len(containerdProtocolPrefix) {
		return fmt.Errorf("container id %s is not a containerd container id", containerID)
	}
	if containerID[0:len(containerdProtocolPrefix)] != containerdProtocolPrefix {
		return fmt.Errorf("expected %s but got %s", containerdProtocolPrefix, containerID[0:len(containerdProtocolPrefix)])
	}
	containerID = containerID[len(containerdProtocolPrefix):]
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
