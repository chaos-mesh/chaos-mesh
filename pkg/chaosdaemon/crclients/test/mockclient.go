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

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"

	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

type MockClient struct{}

func (m *MockClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	if err := mock.On("ContainerInspectError"); err != nil {
		return types.ContainerJSON{}, err.(error)
	}

	containerJSON := types.ContainerJSON{}
	if pid := mock.On("pid"); pid != nil {
		containerJSON.ContainerJSONBase = &types.ContainerJSONBase{
			State: &types.ContainerState{
				Pid: pid.(int),
			},
		}
	}

	if labels := mock.On("labels"); labels != nil {
		containerJSON.Config = &container.Config{
			Labels: labels.(map[string]string),
		}
	}

	return containerJSON, nil
}

func (m *MockClient) ContainerKill(ctx context.Context, containerID, signal string) error {
	if err := mock.On("ContainerKillError"); err != nil {
		return err.(error)
	}
	return nil
}

func (m *MockClient) LoadContainer(ctx context.Context, id string) (containerd.Container, error) {
	if err := mock.On("LoadContainerError"); err != nil {
		return nil, err.(error)
	}

	return &MockContainer{}, nil
}

func (m *MockClient) ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	if err := mock.On("ContainerListError"); err != nil {
		return nil, err.(error)
	}

	c := types.Container{}
	if id := mock.On("containerID"); id != nil {
		c.ID = id.(string)
	}

	return []types.Container{c}, nil
}

func (m *MockClient) Containers(ctx context.Context, filters ...string) ([]containerd.Container, error) {
	if err := mock.On("ContainersError"); err != nil {
		return nil, err.(error)
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

func (m *MockContainer) Labels(ctx context.Context) (map[string]string, error) {
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
