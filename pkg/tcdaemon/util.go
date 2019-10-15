package tcdaemon

import (
	"context"

	dockerclient "github.com/docker/docker/client"
	"github.com/juju/errors"
)

const (
	defaultDockerSocket = "unix:///var/run/docker.sock"
)

// ContainerRuntimeInfoClient represents a struct which can give you information about container runtime
type ContainerRuntimeInfoClient interface {
	GetPidFromContainerID(ctx context.Context, containerID string) (int, error)
}

// DockerClient can get information from docker
type DockerClient struct {
	client *dockerclient.Client
}

// GetPidFromContainerID fetches PID according to container id
func (c DockerClient) GetPidFromContainerID(ctx context.Context, containerID string) (int, error) {
	container, err := c.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return 0, errors.Trace(err)
	}

	return container.State.Pid, nil
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
