package chaosdaemon

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
)

func (s *daemonServer) SetTbf(ctx context.Context, in *pb.TbfRequest) (*empty.Empty, error) {
	log.Info("Set Tbf", "Request", in)

	pid, err := s.crClient.GetPidFromContainerID(ctx, in.ContainerId)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "get pid from containerID error: %v", err)
	}

	if err := applyTbf(in.Tbf, pid); err != nil {
		return nil, status.Errorf(codes.Internal, "tbf apply error: %v", err)
	}

	return &empty.Empty{}, nil
}

func (s *daemonServer) DeleteTbf(ctx context.Context, in *pb.TbfRequest) (*empty.Empty, error) {
	log.Info("Delete Tbf", "Request", in)

	pid, err := s.crClient.GetPidFromContainerID(ctx, in.ContainerId)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "get pid from containerID error: %v", err)
	}

	if err := deleteTbf(in.Tbf, pid); err != nil {
		return nil, status.Errorf(codes.Internal, "tbf apply error: %v", err)
	}

	return &empty.Empty{}, nil
}
