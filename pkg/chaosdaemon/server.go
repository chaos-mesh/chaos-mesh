package chaosdaemon

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/golang/glog"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/juju/errors"

	pb "github.com/pingcap/chaos-operator/pkg/chaosdaemon/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

//go:generate protoc -I pb pb/chaosdaemon.proto --go_out=plugins=grpc:pb

// Server represents an HTTP server for tc daemon
type Server struct {
	crClient ContainerRuntimeInfoClient
}

func newServer() (*Server, error) {
	crClient, err := CreateContainerRuntimeInfoClient()
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &Server{
		crClient: crClient,
	}, nil
}

func (s *Server) SetNetem(ctx context.Context, in *pb.NetemRequest) (*empty.Empty, error) {
	glog.Infof("Request : SetNetem %v", in)

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
	glog.Infof("Request : DeleteNetem %v", in)

	pid, err := s.crClient.GetPidFromContainerID(ctx, in.ContainerId)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "get pid from containerID error: %v", err)
	}

	if err := Cancel(in.Netem, pid); err != nil {
		return nil, status.Errorf(codes.Internal, "netem cancel error: %v", err)
	}

	return &empty.Empty{}, nil
}

func StartServer(host string, port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	chaosDaemonServer, err := newServer()
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}
	pb.RegisterChaosDaemonServer(s, chaosDaemonServer)

	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
