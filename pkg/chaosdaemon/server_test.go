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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients/test"
	"github.com/chaos-mesh/chaos-mesh/pkg/log"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

var _ = Describe("netem server", func() {
	logger, err := log.NewDefaultZapLogger()
	Expect(err).To(BeNil())

	Context("newDaemonServer", func() {
		It("should work without socket path", func() {
			_, err := newDaemonServer(&crclients.CrClientConfig{Runtime: crclients.ContainerRuntimeDocker}, nil, logger)
			Expect(err).To(BeNil())
			defer mock.With("MockContainerdClient", &test.MockClient{})()
			_, err = newDaemonServer(&crclients.CrClientConfig{Runtime: crclients.ContainerRuntimeContainerd}, nil, logger)
			Expect(err).To(BeNil())
		})

		It("should work with socket path", func() {
			_, err := newDaemonServer(&crclients.CrClientConfig{
				Runtime:    crclients.ContainerRuntimeDocker,
				SocketPath: "/foo/bar/docker.socket"}, nil, logger)
			Expect(err).To(BeNil())
		})

		It("should work with socket path and ns", func() {
			defer mock.With("MockContainerdClient", &test.MockClient{})()
			_, err := newDaemonServer(&crclients.CrClientConfig{Runtime: crclients.ContainerRuntimeContainerd}, nil, logger)
			Expect(err).To(BeNil())
			_, err = newDaemonServer(&crclients.CrClientConfig{
				Runtime:      crclients.ContainerRuntimeContainerd,
				SocketPath:   "/foo/bar/containerd.socket",
				ContainerdNS: "chaos-mesh.org"}, nil, logger)
			Expect(err).To(BeNil())
		})

		It("should fail on CreateContainerRuntimeInfoClient", func() {
			_, err := newDaemonServer(&crclients.CrClientConfig{Runtime: "invalid-runtime"}, nil, logger)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("newGRPCServer", func() {
		It("should work", func() {
			defer mock.With("MockContainerdClient", &test.MockClient{})()
			daemonServer, err := newDaemonServer(&crclients.CrClientConfig{Runtime: crclients.ContainerRuntimeContainerd}, nil, logger)
			Expect(err).To(BeNil())
			_, err = newGRPCServer(daemonServer, &MockRegisterer{}, tlsConfig{})
			Expect(err).To(BeNil())
		})

		It("should panic", func() {
			Ω(func() {
				defer mock.With("MockContainerdClient", &test.MockClient{})()
				defer mock.With("PanicOnMustRegister", "mock panic")()
				daemonServer, err := newDaemonServer(&crclients.CrClientConfig{Runtime: crclients.ContainerRuntimeContainerd}, nil, logger)
				Expect(err).To(BeNil())
				_, err = newGRPCServer(daemonServer, &MockRegisterer{}, tlsConfig{})
				Expect(err).To(BeNil())
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
