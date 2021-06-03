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

package test

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/client"
	chaosdaemon "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

// Assert *MockChaosDaemonClient implements chaosdaemon.ChaosDaemonClientInterface.
var _ client.ChaosDaemonClientInterface = (*MockChaosDaemonClient)(nil)

// MockChaosDaemonClient implements ChaosDaemonClientInterface for unit testing
type MockChaosDaemonClient struct{}

// ExecStressors mocks executing pod stressors on chaos-daemon
func (c *MockChaosDaemonClient) ExecStressors(ctx context.Context, in *chaosdaemon.ExecStressRequest, opts ...grpc.CallOption) (*chaosdaemon.ExecStressResponse, error) {
	return nil, mockError("ExecStressors")
}

// CancelStressors mocks canceling pod stressors on chaos-daemon
func (c *MockChaosDaemonClient) CancelStressors(ctx context.Context, in *chaosdaemon.CancelStressRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("CancelStressors")
}

func (c *MockChaosDaemonClient) ContainerGetPid(ctx context.Context, in *chaosdaemon.ContainerRequest, opts ...grpc.CallOption) (*chaosdaemon.ContainerResponse, error) {
	if resp := mock.On("MockContainerGetPidResponse"); resp != nil {
		return resp.(*chaosdaemon.ContainerResponse), nil
	}
	return nil, mockError("ContainerGetPid")
}

func mockError(name string) error {
	if err := mock.On(fmt.Sprintf("Mock%sError", name)); err != nil {
		return err.(error)
	}
	return nil
}

func (c *MockChaosDaemonClient) FlushIPSets(ctx context.Context, in *chaosdaemon.IPSetsRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("FlushIPSets")
}

func (c *MockChaosDaemonClient) SetIptablesChains(ctx context.Context, in *chaosdaemon.IptablesChainsRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("SetIptablesChains")
}

func (c *MockChaosDaemonClient) SetTimeOffset(ctx context.Context, in *chaosdaemon.TimeRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("SetTimeOffset")
}

func (c *MockChaosDaemonClient) RecoverTimeOffset(ctx context.Context, in *chaosdaemon.TimeRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("RecoverTimeOffset")
}

func (c *MockChaosDaemonClient) ContainerKill(ctx context.Context, in *chaosdaemon.ContainerRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("ContainerKill")
}

func (c *MockChaosDaemonClient) ApplyIOChaos(ctx context.Context, in *chaosdaemon.ApplyIOChaosRequest, opts ...grpc.CallOption) (*chaosdaemon.ApplyIOChaosResponse, error) {
	return nil, mockError("ApplyIOChaos")
}

func (c *MockChaosDaemonClient) ApplyHttpChaos(ctx context.Context, in *chaosdaemon.ApplyHttpChaosRequest, opts ...grpc.CallOption) (*chaosdaemon.ApplyHttpChaosResponse, error) {
	return nil, mockError("ApplyHttpChaos")
}

func (c *MockChaosDaemonClient) SetDNSServer(ctx context.Context, in *chaosdaemon.SetDNSServerRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("SetDNSServer")
}

func (c *MockChaosDaemonClient) SetTcs(ctx context.Context, in *chaosdaemon.TcsRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("SetTcs")
}

func (c *MockChaosDaemonClient) Close() error {
	return mockError("CloseChaosDaemonClient")
}
