// Copyright 2019 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package chaosdaemon

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/moby/locker"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	grpcUtils "github.com/chaos-mesh/chaos-mesh/pkg/grpc"
)

var log = ctrl.Log.WithName("chaos-daemon-server")

//go:generate protoc -I pb pb/chaosdaemon.proto --go_out=plugins=grpc:pb

// Config contains the basic chaos daemon configuration.
type Config struct {
	HTTPPort  int
	GRPCPort  int
	Host      string
	Runtime   string
	Profiling bool

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
	backgroundProcessManager bpm.BackgroundProcessManager

	IPSetLocker *locker.Locker
}

func newDaemonServer(containerRuntime string) (*DaemonServer, error) {
	crClient, err := crclients.CreateContainerRuntimeInfoClient(containerRuntime)
	if err != nil {
		return nil, err
	}

	return NewDaemonServerWithCRClient(crClient), nil
}

// NewDaemonServerWithCRClient returns DaemonServer with container runtime client
func NewDaemonServerWithCRClient(crClient crclients.ContainerRuntimeInfoClient) *DaemonServer {
	return &DaemonServer{
		IPSetLocker:              locker.New(),
		crClient:                 crClient,
		backgroundProcessManager: bpm.NewBackgroundProcessManager(),
	}
}

func newGRPCServer(containerRuntime string, reg prometheus.Registerer, tlsConf tlsConfig) (*grpc.Server, error) {
	ds, err := newDaemonServer(containerRuntime)
	if err != nil {
		return nil, err
	}

	grpcMetrics := grpc_prometheus.NewServerMetrics()
	grpcMetrics.EnableHandlingTimeHistogram(
		grpc_prometheus.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 10}),
	)
	reg.MustRegister(grpcMetrics)

	grpcOpts := []grpc.ServerOption{
		grpc_middleware.WithUnaryServerChain(
			grpcUtils.TimeoutServerInterceptor,
			grpcMetrics.UnaryServerInterceptor(),
		),
	}

	if tlsConf != (tlsConfig{}) {
		caCert, err := ioutil.ReadFile(tlsConf.CaCert)
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
		})

		grpcOpts = append(grpcOpts, grpc.Creds(creds))
	}

	s := grpc.NewServer(grpcOpts...)
	grpcMetrics.InitializeMetrics(s)

	pb.RegisterChaosDaemonServer(s, ds)
	reflection.Register(s)

	return s, nil
}

// RegisterGatherer combine prometheus.Registerer and prometheus.Gatherer
type RegisterGatherer interface {
	prometheus.Registerer
	prometheus.Gatherer
}

// StartServer starts chaos-daemon.
func StartServer(conf *Config, reg RegisterGatherer) error {
	g := &errgroup.Group{}

	httpBindAddr := conf.HttpAddr()
	httpServer := newHTTPServerBuilder().Addr(httpBindAddr).Metrics(reg).Profiling(conf.Profiling).Build()

	grpcBindAddr := conf.GrpcAddr()
	grpcListener, err := net.Listen("tcp", grpcBindAddr)
	if err != nil {
		log.Error(err, "failed to listen grpc address", "grpcBindAddr", grpcBindAddr)
		return err
	}

	grpcServer, err := newGRPCServer(conf.Runtime, reg, conf.tlsConfig)
	if err != nil {
		log.Error(err, "failed to create grpc server")
		return err
	}

	g.Go(func() error {
		log.Info("Starting http endpoint", "address", httpBindAddr)
		if err := httpServer.ListenAndServe(); err != nil {
			log.Error(err, "failed to start http endpoint")
			httpServer.Shutdown(context.Background())
			return err
		}
		return nil
	})

	g.Go(func() error {
		log.Info("Starting grpc endpoint", "address", grpcBindAddr, "runtime", conf.Runtime)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Error(err, "failed to start grpc endpoint")
			grpcServer.Stop()
			return err
		}
		return nil
	})

	return g.Wait()
}
