// Copyright 2020 Chaos Mesh Authors.
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
	"syscall"
	"testing"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/docker/docker/api/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/chaos-mesh/chaos-mesh/pkg/mock"

	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"chaosdaemon Suite",
		[]Reporter{envtest.NewlineReporter{}})
}

type MockClient struct{}

func (m *MockClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	if err := mock.On("ContainerInspectError"); err != nil {
		return types.ContainerJSON{}, err.(error)
	}

	var pid int
	if p := mock.On("pid"); p != nil {
		pid = p.(int)
	}
	return types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			State: &types.ContainerState{
				Pid: pid,
			},
		},
	}, nil
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

type MockContainer struct {
	containerd.Container
}

func (m *MockContainer) Task(context.Context, cio.Attach) (containerd.Task, error) {
	if err := mock.On("TaskError"); err != nil {
		return nil, err.(error)
	}

	return &MockTask{}, nil
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

type MockRegisterer struct {
	RegisterGatherer
}

func (*MockRegisterer) MustRegister(...prometheus.Collector) {
	if err := mock.On("PanicOnMustRegister"); err != nil {
		panic(err)
	}
}
