package chaosdaemon

import (
	"context"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/golang/protobuf/ptypes/empty"
)

func (s *DaemonServer) DiskFill(ctx context.Context, req *pb.DiskFillRequest) (*empty.Empty, error) {

	return nil, nil
}

func (s *DaemonServer) RecoverDiskFill(ctx context.Context, req *pb.DiskFillRecoverRequest) (*empty.Empty, error) {

	return nil, nil
}

func (s *DaemonServer) DiskPayload(ctx context.Context, req *pb.DiskPayloadRequest) (*empty.Empty, error) {
	return nil, nil
}

func (s *DaemonServer) RecoverDiskPayload(ctx context.Context, req *pb.DiskPayloadRecoverRequest) (*empty.Empty, error) {

	return nil, nil
}
