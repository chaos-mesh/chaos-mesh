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
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/containerd/containerd"
	"github.com/docker/docker/api/types"
	dockerclient "github.com/docker/docker/client"

	"github.com/pingcap/chaos-mesh/pkg/mock"
	"github.com/pingcap/chaos-mesh/pkg/utils"
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
	FormatContainerID(ctx context.Context, containerID string) (string, error)
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

// FormatContainerID strips protocol prefix from the container ID
func (c ContainerdClient) FormatContainerID(ctx context.Context, containerID string) (string, error) {
	if len(containerID) < len(containerdProtocolPrefix) {
		return "", fmt.Errorf("container id %s is not a containerd container id", containerID)
	}
	if containerID[0:len(containerdProtocolPrefix)] != containerdProtocolPrefix {
		return "", fmt.Errorf("expected %s but got %s", containerdProtocolPrefix, containerID[0:len(containerdProtocolPrefix)])
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

type nsType string

const (
	mountNS nsType = "mnt"
	utsNS   nsType = "uts"
	ipcNS   nsType = "ipc"
	netNS   nsType = "net"
	pidNS   nsType = "pid"
	userNS  nsType = "user"
)

var nsArgMap = map[nsType]string{
	mountNS: "m",
	utsNS:   "u",
	ipcNS:   "i",
	netNS:   "n",
	pidNS:   "p",
	userNS:  "U",
}

// GetNsPath returns corresponding namespace path
func GetNsPath(pid uint32, typ nsType) string {
	return fmt.Sprintf("%s/%d/ns/%s", defaultProcPrefix, pid, string(typ))
}

func withNetNS(ctx context.Context, nsPath string, cmd string, args ...string) *exec.Cmd {
	// Mock point to return mock Cmd in unit test
	if c := mock.On("MockWithNetNs"); c != nil {
		f := c.(func(context.Context, string, string, ...string) *exec.Cmd)
		return f(ctx, nsPath, cmd, args...)
	}

	return withNS(ctx, []nsOption{{
		Typ:  netNS,
		Path: nsPath,
	}}, cmd, args...)
}

func withPidNS(ctx context.Context, nsPath string, cmd string, args ...string) *exec.Cmd {
	// Mock point to return mock Cmd in unit test
	if c := mock.On("MockWithPidNs"); c != nil {
		f := c.(func(context.Context, string, string, ...string) *exec.Cmd)
		return f(ctx, nsPath, cmd, args...)
	}

	return withNS(ctx, []nsOption{{
		Typ:  pidNS,
		Path: nsPath,
	}}, cmd, args...)
}

type nsOption struct {
	Typ  nsType
	Path string
}

func withNS(ctx context.Context, options []nsOption, cmd string, args ...string) *exec.Cmd {
	// Mock point to return mock Cmd in unit test
	if c := mock.On("MockWithNs"); c != nil {
		f := c.(func(context.Context, []nsOption, string, ...string) *exec.Cmd)
		return f(ctx, options, cmd, args...)
	}

	args = append([]string{"--", cmd}, args...)
	for _, option := range options {
		args = append([]string{"-" + nsArgMap[option.Typ] + option.Path}, args...)
	}

	log.Info("running command", "command", "nsenter "+strings.Join(args, " "))

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

// GetChildProcesses will return all child processes's pid. Include all generations.
func GetChildProcesses(ppid uint32) ([]uint32, error) {
	procs, err := ioutil.ReadDir(defaultProcPrefix)
	if err != nil {
		return nil, err
	}

	type processPair struct {
		Pid  uint32
		Ppid uint32
	}

	pairs := make(chan processPair)
	done := make(chan bool)

	go func() {
		var wg sync.WaitGroup

		for _, proc := range procs {
			_, err := strconv.ParseUint(proc.Name(), 10, 32)
			if err != nil {
				continue
			}

			statusPath := defaultProcPrefix + "/" + proc.Name() + "/stat"

			wg.Add(1)
			go func() {
				defer wg.Done()

				reader, err := os.Open(statusPath)
				if err != nil {
					log.Error(err, "read status file error", "path", statusPath)
					return
				}

				var (
					pid    uint32
					comm   string
					state  string
					parent uint32
				)
				// according to procfs's man page
				fmt.Fscanf(reader, "%d %s %s %d", &pid, &comm, &state, &parent)

				pairs <- processPair{
					Pid:  pid,
					Ppid: parent,
				}
			}()
		}

		wg.Wait()
		done <- true
	}()

	processGraph := utils.NewGraph()
	for {
		select {
		case pair := <-pairs:
			processGraph.Insert(pair.Ppid, pair.Pid)
		case <-done:
			return processGraph.Flatten(ppid), nil
		}
	}
}
