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

package test

import (
	"context"
	"syscall"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/moby/moby/api/types/container"
	dockerclient "github.com/moby/moby/client"

	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

type MockClient struct{}

func (m *MockClient) ContainerInspect(_ context.Context, _ string, _ dockerclient.ContainerInspectOptions) (dockerclient.ContainerInspectResult, error) {
	if err := mock.On("ContainerInspectError"); err != nil {
		return dockerclient.ContainerInspectResult{}, err.(error)
	}

	containerJSON := container.InspectResponse{}
	if pid := mock.On("pid"); pid != nil {
		containerJSON.State = &container.State{
			Pid: pid.(int),
		}
	}

	if labels := mock.On("labels"); labels != nil {
		containerJSON.Config = &container.Config{
			Labels: labels.(map[string]string),
		}
	}

	return dockerclient.ContainerInspectResult{Container: containerJSON}, nil
}

func (m *MockClient) ContainerKill(_ context.Context, _ string, _ dockerclient.ContainerKillOptions) (dockerclient.ContainerKillResult, error) {
	if err := mock.On("ContainerKillError"); err != nil {
		return dockerclient.ContainerKillResult{}, err.(error)
	}
	return dockerclient.ContainerKillResult{}, nil
}

func (m *MockClient) LoadContainer(_ context.Context, _ string) (containerd.Container, error) {
	if err := mock.On("LoadContainerError"); err != nil {
		return nil, err.(error)
	}

	return &MockContainer{}, nil
}

func (m *MockClient) ContainerList(_ context.Context, _ dockerclient.ContainerListOptions) (dockerclient.ContainerListResult, error) {
	if err := mock.On("ContainerListError"); err != nil {
		return dockerclient.ContainerListResult{}, err.(error)
	}

	c := container.Summary{}
	if id := mock.On("containerID"); id != nil {
		c.ID = id.(string)
	}

	return dockerclient.ContainerListResult{Items: []container.Summary{c}}, nil
}

func (m *MockClient) Containers(_ context.Context, _ ...string) ([]containerd.Container, error) {
	if err := mock.On("ContainersError"); err != nil {
		return nil, err.(error)
	}

	if mock.On("emptyContainers") != nil {
		return []containerd.Container{}, nil
	}

	return []containerd.Container{&MockContainer{}}, nil
}

type MockContainer struct {
	containerd.Container
}

func (m *MockContainer) Task(context.Context, cio.Attach) (containerd.Task, error) {
	if err := mock.On("TaskError"); err != nil {
		return nil, err.(error)
	}

	return &MockTask{}, nil
}

func (m *MockContainer) ID() string {
	if err := mock.On("IDError"); err != nil {
		return ""
	}

	if id := mock.On("containerID"); id != nil {
		return id.(string)
	}
	return ""
}

func (m *MockContainer) Labels(_ context.Context) (map[string]string, error) {
	if err := mock.On("LabelsError"); err != nil {
		return nil, err.(error)
	}

	if labels := mock.On("labels"); labels != nil {
		return labels.(map[string]string), nil
	}
	return nil, nil
}

type MockTask struct {
	containerd.Task
}

func (m *MockTask) Pid() uint32 {
	var pid int
	if p := mock.On("pid"); p != nil {
		pid = p.(int)
	}
	return uint32(pid)
}

func (m *MockTask) Kill(context.Context, syscall.Signal, ...containerd.KillOpts) error {
	if err := mock.On("KillError"); err != nil {
		return err.(error)
	}
	return nil
}
