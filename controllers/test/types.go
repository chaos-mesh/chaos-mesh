package test

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	chaosdaemon "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/pingcap/chaos-mesh/pkg/mock"
	"github.com/pingcap/chaos-mesh/pkg/utils"
	"google.golang.org/grpc"
)

// Assert *MockChaosDaemonClient implements chaosdaemon.ChaosDaemonClientInterface.
var _ utils.ChaosDaemonClientInterface = (*MockChaosDaemonClient)(nil)

type MockChaosDaemonClient struct{}

func mockError(name string) error {
	if err := mock.On(fmt.Sprintf("Mock%sError", name)); err != nil {
		return err.(error)
	}
	return nil
}

func (c *MockChaosDaemonClient) SetNetem(ctx context.Context, in *chaosdaemon.NetemRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("SetNetem")
}

func (c *MockChaosDaemonClient) DeleteNetem(ctx context.Context, in *chaosdaemon.NetemRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("DeleteNetem")
}

func (c *MockChaosDaemonClient) FlushIpSet(ctx context.Context, in *chaosdaemon.IpSetRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("FlushIpSet")
}

func (c *MockChaosDaemonClient) FlushIptables(ctx context.Context, in *chaosdaemon.IpTablesRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("FlushIptables")
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

func (c *MockChaosDaemonClient) Close() error {
	return mockError("CloseChaosDaemonClient")
}
