// Copyright 2020 Chaos Mesh Authors.
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
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
	"net"

	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

var _ = Describe("netem server", func() {
	Context("newDaemonServer", func() {
		It("should work", func() {
			defer mock.With("MockContainerdClient", &test.MockClient{})()
			_, err := newDaemonServer(containerRuntimeContainerd)
			Expect(err).To(BeNil())
		})

		It("should fail on CreateContainerRuntimeInfoClient", func() {
			_, err := newDaemonServer("invalid-runtime")
			Expect(err).ToNot(BeNil())
		})
	})

	Context("newGRPCServer", func() {
		It("should work", func() {
			defer mock.With("MockContainerdClient", &test.MockClient{})()
			_, err := newGRPCServer(containerRuntimeContainerd, &MockRegisterer{})
			Expect(err).To(BeNil())
		})

		It("should panic", func() {
			Ω(func() {
				defer mock.With("MockContainerdClient", &test.MockClient{})()
				defer mock.With("PanicOnMustRegister", "mock panic")()
				newGRPCServer(containerRuntimeContainerd, &MockRegisterer{})
			}).Should(Panic())
		})
	})
})

type MockRegisterer struct {
	RegisterGatherer
}

func (*MockRegisterer) MustRegister(...prometheus.Collector) {
	if err := mock.On("PanicOnMustRegister"); err != nil {
		panic(err)
	}
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

	grpcServer, err := newGRPCServer(conf.Runtime, reg)
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
