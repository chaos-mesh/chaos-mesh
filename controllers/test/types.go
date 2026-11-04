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
func (c *MockChaosDaemonClient) ExecStressors(_ context.Context, _ *chaosdaemon.ExecStressRequest, _ ...grpc.CallOption) (*chaosdaemon.ExecStressResponse, error) {
	return nil, mockError("ExecStressors")
}

// CancelStressors mocks canceling pod stressors on chaos-daemon
func (c *MockChaosDaemonClient) CancelStressors(_ context.Context, _ *chaosdaemon.CancelStressRequest, _ ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("CancelStressors")
}

func (c *MockChaosDaemonClient) ContainerGetPid(_ context.Context, _ *chaosdaemon.ContainerRequest, _ ...grpc.CallOption) (*chaosdaemon.ContainerResponse, error) {
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

func (c *MockChaosDaemonClient) FlushIPSets(_ context.Context, _ *chaosdaemon.IPSetsRequest, _ ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("FlushIPSets")
}

func (c *MockChaosDaemonClient) SetIptablesChains(_ context.Context, _ *chaosdaemon.IptablesChainsRequest, _ ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("SetIptablesChains")
}

func (c *MockChaosDaemonClient) SetTimeOffset(_ context.Context, _ *chaosdaemon.TimeRequest, _ ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("SetTimeOffset")
}

func (c *MockChaosDaemonClient) RecoverTimeOffset(_ context.Context, _ *chaosdaemon.TimeRequest, _ ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("RecoverTimeOffset")
}

func (c *MockChaosDaemonClient) ContainerKill(_ context.Context, _ *chaosdaemon.ContainerRequest, _ ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("ContainerKill")
}

func (c *MockChaosDaemonClient) ApplyIOChaos(_ context.Context, _ *chaosdaemon.ApplyIOChaosRequest, _ ...grpc.CallOption) (*chaosdaemon.ApplyIOChaosResponse, error) {
	return nil, mockError("ApplyIOChaos")
}

//revive:disable-next-line:var-naming Preserve the generated protobuf RPC contract.
func (c *MockChaosDaemonClient) ApplyHttpChaos(_ context.Context, _ *chaosdaemon.ApplyHttpChaosRequest, _ ...grpc.CallOption) (*chaosdaemon.ApplyHttpChaosResponse, error) {
	return nil, mockError("ApplyHttpChaos")
}

func (c *MockChaosDaemonClient) SetDNSServer(_ context.Context, _ *chaosdaemon.SetDNSServerRequest, _ ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("SetDNSServer")
}

func (c *MockChaosDaemonClient) SetTcs(_ context.Context, _ *chaosdaemon.TcsRequest, _ ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("SetTcs")
}

func (c *MockChaosDaemonClient) Close() error {
	return mockError("CloseChaosDaemonClient")
}

func (c *MockChaosDaemonClient) InstallJVMRules(_ context.Context, _ *chaosdaemon.InstallJVMRulesRequest, _ ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("InstallJVMRules")
}

func (c *MockChaosDaemonClient) UninstallJVMRules(_ context.Context, _ *chaosdaemon.UninstallJVMRulesRequest, _ ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("UninstallJVMRules")
}

func (c *MockChaosDaemonClient) ApplyBlockChaos(_ context.Context, _ *chaosdaemon.ApplyBlockChaosRequest, _ ...grpc.CallOption) (*chaosdaemon.ApplyBlockChaosResponse, error) {
	return nil, mockError("ApplyBlockChaosRequest")
}

func (c *MockChaosDaemonClient) RecoverBlockChaos(_ context.Context, _ *chaosdaemon.RecoverBlockChaosRequest, _ ...grpc.CallOption) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}
