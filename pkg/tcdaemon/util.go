package tcdaemon

import (
	"context"

	dockerclient "github.com/docker/docker/client"
	"github.com/juju/errors"
)

const (
	defaultDockerSocket = "unix:///var/run/docker.sock"
)

type ContainerRuntimeInfoClient interface {
	GetPidFromContainerId(ctx context.Context, containerId string) (int, error)
}

type DockerClient struct {
	client *dockerclient.Client
}

func (c DockerClient) GetPidFromContainerId(ctx context.Context, containerId string) (int, error) {
	container, err := c.client.ContainerInspect(ctx, containerId)
	if err != nil {
		return 0, errors.Trace(err)
	}

	return container.State.Pid, nil
}

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
