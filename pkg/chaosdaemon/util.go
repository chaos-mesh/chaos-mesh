package chaosdaemon

import (
	"context"
	"fmt"
	"os/exec"

	dockerclient "github.com/docker/docker/client"
	"github.com/juju/errors"
)

const (
	defaultDockerSocket  = "unix:///var/run/docker.sock"
	dockerProtocolPrefix = "docker://"

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
		return 0, errors.Errorf("only docker protocol is supported but got %s", containerID[0:len(dockerProtocolPrefix)])
	}
	container, err := c.client.ContainerInspect(ctx, containerID[len(dockerProtocolPrefix):])
	if err != nil {
		return 0, errors.Trace(err)
	}

	return uint32(container.State.Pid), nil
}

// CreateContainerRuntimeInfoClient will create container runtime information getter
func CreateContainerRuntimeInfoClient() (ContainerRuntimeInfoClient, error) {
	// TODO: support more container runtime

	client, err := dockerclient.NewClient(defaultDockerSocket, "", nil, nil)

	if err != nil {
		return nil, errors.Trace(err)
	}

	return DockerClient{
		client: client,
	}, nil
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
