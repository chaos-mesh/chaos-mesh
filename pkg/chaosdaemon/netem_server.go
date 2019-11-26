package chaosdaemon

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/pingcap/chaos-operator/pkg/chaosdaemon/pb"
)

func (s *Server) SetNetem(ctx context.Context, in *pb.NetemRequest) (*empty.Empty, error) {
	log.Info("Set netem", "Request", in)

	pid, err := s.crClient.GetPidFromContainerID(ctx, in.ContainerId)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "get pid from containerID error: %v", err)
	}

	if err := Apply(in.Netem, pid); err != nil {
		return nil, status.Errorf(codes.Internal, "netem apply error: %v", err)
	}

	return &empty.Empty{}, nil
}

func (s *Server) DeleteNetem(ctx context.Context, in *pb.NetemRequest) (*empty.Empty, error) {
	log.Info("Delete netem", "Request", in)

	pid, err := s.crClient.GetPidFromContainerID(ctx, in.ContainerId)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "get pid from containerID error: %v", err)
	}

	if err := Cancel(in.Netem, pid); err != nil {
		return nil, status.Errorf(codes.Internal, "netem cancel error: %v", err)
	}

	return &empty.Empty{}, nil
}
