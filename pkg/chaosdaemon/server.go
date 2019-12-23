package chaosdaemon

import (
	"fmt"
	"github.com/juju/errors"
	"net"

	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("chaos-daemon-server")

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

func StartServer(host string, port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Error(err, "failed to listen")
	}

	s := grpc.NewServer()
	chaosDaemonServer, err := newServer()
	if err != nil {
		log.Error(err, "failed to create server")
	}
	pb.RegisterChaosDaemonServer(s, chaosDaemonServer)

	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Error(err, "failed to serve")
	}
}
