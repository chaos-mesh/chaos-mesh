package chaosdaemon

import (
	"context"

	"github.com/pingcap/chaos-mesh/pkg/time"

	"github.com/golang/protobuf/ptypes/empty"

	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
)

func (s *Server) SetTimeOffset(ctx context.Context, req *pb.TimeRequest) (*empty.Empty, error) {
	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "error while getting PID")
		return nil, err
	}

	err = time.ModifyTime(int(pid), int64(req.Sec), int64(req.Nsec))
	if err != nil {
		log.Error(err, "error while modifying time", "pid", pid)
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *Server) RecoverTimeOffset(ctx context.Context, req *pb.TimeRequest) (*empty.Empty, error) {
	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "error while getting PID")
		return nil, err
	}

	err = time.ModifyTime(int(pid), int64(0), int64(0))
	if err != nil {
		log.Error(err, "error while recovering", "pid", pid)
		return nil, err
	}
	return &empty.Empty{}, nil
}
