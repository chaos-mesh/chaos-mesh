package chaoscm

import (
	"context"
	"net"
	"os"

	pb "github.com/pingcap/chaos-mesh/pkg/chaoscm/pb"
	"github.com/pingcap/chaos-mesh/pkg/utils"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	ctrl "sigs.k8s.io/controller-runtime"
)

//go:generate protoc -I pb pb/chaoscm.proto --go_out=plugins=grpc:pb

var log = ctrl.Log.WithName("ChaosCM-server")

type server struct {
}

func (s *server) EatMemory(ctx context.Context, req *pb.MemoryRequest) (*empty.Empty, error) {
	log.Info("Eating memory of [%s]", req.Quota)
	return &empty.Empty{}, nil
}

func (s *server) RecoverMemory(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func (s *server) BurnCpu(ctx context.Context, req *pb.CpuRequest) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func (s *server) RecoverCpu(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func StartServer(addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error(err, "failed to listen tcp server", "address", addr)
		os.Exit(1)
	}
	s := grpc.NewServer(grpc.UnaryInterceptor(utils.TimeoutServerInterceptor))
	pb.RegisterChaosCMServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Error(err, "failed to start serve")
		os.Exit(1)
	}
}
