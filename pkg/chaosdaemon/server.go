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

package chaosdaemon

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/go-logr/logr"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/moby/locker"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/tasks"
	grpcUtils "github.com/chaos-mesh/chaos-mesh/pkg/grpc"
	"github.com/chaos-mesh/chaos-mesh/pkg/log"
	"github.com/chaos-mesh/chaos-mesh/pkg/metrics"
)

//go:generate protoc -I pb pb/chaosdaemon.proto --go_out=plugins=grpc:pb

// Config contains the basic chaos daemon configuration.
type Config struct {
	HTTPPort           int
	GRPCPort           int
	Host               string
	CrClientConfig     *crclients.CrClientConfig
	Profiling          bool
	TodaStartupTimeout int // Timeout in milliseconds for toda startup

	tlsConfig
}

// tlsConfig contains the config of TLS Server
type tlsConfig struct {
	CaCert string
	Cert   string
	Key    string
}

// Get the http address
func (c *Config) HttpAddr() string {
	return net.JoinHostPort(c.Host, fmt.Sprintf("%d", c.HTTPPort))
}

// Get the grpc address
func (c *Config) GrpcAddr() string {
	return net.JoinHostPort(c.Host, fmt.Sprintf("%d", c.GRPCPort))
}

// DaemonServer represents a grpc server for tc daemon
type DaemonServer struct {
	crClient                 crclients.ContainerRuntimeInfoClient
	backgroundProcessManager *bpm.BackgroundProcessManager
	rootLogger               logr.Logger
	todaStartupTimeout       int // Timeout in milliseconds for toda startup

	// tproxyLocker is a set of tproxy processes to lock stdin/stdout/stderr
	tproxyLocker *sync.Map

	IPSetLocker     *locker.Locker
	timeChaosServer TimeChaosServer
}

func (s *DaemonServer) getLoggerFromContext(ctx context.Context) logr.Logger {
	return log.EnrichLoggerWithContext(ctx, s.rootLogger)
}

func newDaemonServer(clientConfig *crclients.CrClientConfig, todaStartupTimeout int, reg prometheus.Registerer, log logr.Logger) (*DaemonServer, error) {
	crClient, err := crclients.CreateContainerRuntimeInfoClient(clientConfig)
	if err != nil {
		return nil, err
	}

	return NewDaemonServerWithCRClient(crClient, todaStartupTimeout, reg, log), nil
}

// NewDaemonServerWithCRClient returns DaemonServer with container runtime client
func NewDaemonServerWithCRClient(crClient crclients.ContainerRuntimeInfoClient, todaStartupTimeout int, reg prometheus.Registerer, log logr.Logger) *DaemonServer {
	return &DaemonServer{
		IPSetLocker:              locker.New(),
		crClient:                 crClient,
		backgroundProcessManager: bpm.StartBackgroundProcessManager(reg, log),
		tproxyLocker:             new(sync.Map),
		rootLogger:               log,
		todaStartupTimeout:       todaStartupTimeout,
		timeChaosServer: TimeChaosServer{
			podContainerNameProcessMap: tasks.NewPodProcessMap(),
			manager:                    tasks.NewTaskManager(logr.New(log.GetSink()).WithName("TimeChaos")),
			nameLocker:                 tasks.NewLockMap[tasks.PodContainerName](),
			logger:                     logr.New(log.GetSink()).WithName("TimeChaos"),
		},
	}
}

func newGRPCServer(daemonServer *DaemonServer, reg prometheus.Registerer, tlsConf tlsConfig) (*grpc.Server, error) {
	grpcMetrics := grpc_prometheus.NewServerMetrics()
	grpcMetrics.EnableHandlingTimeHistogram(
		grpc_prometheus.WithHistogramBuckets(metrics.ChaosDaemonGrpcServerBuckets),
		metrics.WithHistogramName("chaos_daemon_grpc_server_handling_seconds"),
	)
	reg.MustRegister(
		grpcMetrics,
		metrics.DefaultChaosDaemonMetricsCollector.InjectCrClient(daemonServer.crClient),
	)

	grpcOpts := []grpc.ServerOption{
		grpc_middleware.WithUnaryServerChain(
			grpcUtils.TimeoutServerInterceptor,
			grpcMetrics.UnaryServerInterceptor(),
			MetadataExtractor(log.MetaNamespacedName),
		),
	}

	if tlsConf != (tlsConfig{}) {
		caCert, err := os.ReadFile(tlsConf.CaCert)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		serverCert, err := tls.LoadX509KeyPair(tlsConf.Cert, tlsConf.Key)
		if err != nil {
			return nil, err
		}

		creds := credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{serverCert},
			ClientCAs:    caCertPool,
			ClientAuth:   tls.RequireAndVerifyClientCert,
			MinVersion:   tls.VersionTLS13,
		})

		grpcOpts = append(grpcOpts, grpc.Creds(creds))
	}

	s := grpc.NewServer(grpcOpts...)
	grpcMetrics.InitializeMetrics(s)

	pb.RegisterChaosDaemonServer(s, daemonServer)
	reflection.Register(s)

	return s, nil
}

func MetadataExtractor(keys ...log.Metadatkey) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// Get the metadata from the incoming context
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errors.New("couldn't parse incoming context metadata")
		}
		for _, key := range keys {
			values := md.Get(string(key))
			if len(values) > 0 {
				ctx = context.WithValue(ctx, key, values[0])
			}
		}

		return handler(ctx, req)
	}
}

// RegisterGatherer combine prometheus.Registerer and prometheus.Gatherer
type RegisterGatherer interface {
	prometheus.Registerer
	prometheus.Gatherer
}

// Server is the server for chaos daemon
type Server struct {
	daemonServer *DaemonServer
	httpServer   *http.Server
	grpcServer   *grpc.Server

	conf   *Config
	logger logr.Logger
}

// BuildServer builds a chaos daemon server
func BuildServer(conf *Config, reg RegisterGatherer, log logr.Logger) (*Server, error) {
	server := &Server{conf: conf, logger: log}
	var err error
	server.daemonServer, err = newDaemonServer(conf.CrClientConfig, conf.TodaStartupTimeout, reg, log)
	if err != nil {
		return nil, errors.Wrap(err, "create daemon server")
	}

	server.httpServer = newHTTPServerBuilder().Addr(conf.HttpAddr()).Metrics(reg).Profiling(conf.Profiling).Build()
	server.grpcServer, err = newGRPCServer(server.daemonServer, reg, conf.tlsConfig)
	if err != nil {
		return nil, errors.Wrap(err, "create grpc server")
	}

	return server, nil
}

// Start starts chaos-daemon.
func (s *Server) Start() error {
	grpcBindAddr := s.conf.GrpcAddr()
	grpcListener, err := net.Listen("tcp", grpcBindAddr)
	if err != nil {
		return errors.Wrapf(err, "listen grpc address %s", grpcBindAddr)
	}

	var eg errgroup.Group

	eg.Go(func() error {
		s.logger.Info("Starting http endpoint", "address", s.conf.HttpAddr())
		if err := s.httpServer.ListenAndServe(); err != nil {
			return errors.Wrap(err, "start http endpoint")
		}
		return nil
	})

	eg.Go(func() error {
		s.logger.Info("Starting grpc endpoint", "address", grpcBindAddr, "runtime", s.conf.CrClientConfig.Runtime)
		if err := s.grpcServer.Serve(grpcListener); err != nil {
			return errors.Wrap(err, "start grpc endpoint")
		}
		return nil
	})

	return eg.Wait()
}

func (s *Server) Shutdown() error {
	if err := s.httpServer.Shutdown(context.TODO()); err != nil {
		return errors.Wrap(err, "shut grpc endpoint down")
	}
	s.grpcServer.GracefulStop()
	s.daemonServer.backgroundProcessManager.Shutdown(context.TODO())
	return nil
}
