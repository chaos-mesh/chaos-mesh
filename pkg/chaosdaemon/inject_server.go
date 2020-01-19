package chaosdaemon

import (
	"context"
	"os"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb/timechaos"
)

const (
	InjectServerAddress = "127.0.0.1:50051"
)

func (s *Server) SetTimeOffset(ctx context.Context, req *pb.TimeRequest) (*empty.Empty, error) {
	address := os.Getenv("INJECT_SERVER")
	if address == "" {
		address = InjectServerAddress
	}

	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "error while getting PID")
		return nil, err
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	c := timechaos.NewInjectClient(conn)

	return c.SetTimeval(ctx, &timechaos.Request{
		Pid:  pid,
		Tid:  false,
		Sec:  req.Sec,
		Usec: req.Usec,
	})
}

func (s *Server) RecoverTimeOffset(ctx context.Context, req *pb.TimeRequest) (*empty.Empty, error) {
	address := os.Getenv("INJECT_SERVER")
	if address == "" {
		address = InjectServerAddress
	}

	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "error while getting PID")
		return nil, err
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	c := timechaos.NewInjectClient(conn)

	return c.Recover(ctx, &timechaos.Request{
		Pid:  pid,
		Tid:  false,
		Sec:  req.Sec,
		Usec: req.Usec,
	})
}
