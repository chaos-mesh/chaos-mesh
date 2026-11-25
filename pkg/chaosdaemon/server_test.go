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
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/metadata"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients/test"
	"github.com/chaos-mesh/chaos-mesh/pkg/log"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

var _ = Describe("netem server", func() {
	logger := log.NewZapLoggerWithWriter(GinkgoWriter)

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

func TestDaemonHTTPRoutes(t *testing.T) {
	g := NewWithT(t)
	reg := prometheus.NewRegistry()
	reg.MustRegister(prometheus.NewGauge(prometheus.GaugeOpts{Name: "daemon_fixture_value"}))
	for _, profiling := range []bool{false, true} {
		server := newHTTPServerBuilder().Addr("127.0.0.1:0").Metrics(reg).Profiling(profiling).Build()
		g.Expect(server.Addr).To(Equal("127.0.0.1:0"))
		metrics := httptest.NewRecorder()
		server.Handler.ServeHTTP(metrics, httptest.NewRequest(http.MethodGet, "/metrics", nil))
		g.Expect(metrics.Code).To(Equal(http.StatusOK))
		g.Expect(metrics.Body.String()).To(ContainSubstring("daemon_fixture_value 0"))

		profile := httptest.NewRecorder()
		server.Handler.ServeHTTP(profile, httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil))
		if profiling {
			g.Expect(profile.Code).To(Equal(http.StatusOK))
			g.Expect(profile.Body.String()).To(ContainSubstring("Types of profiles available"))
		} else {
			g.Expect(profile.Code).To(Equal(http.StatusNotFound))
		}
	}
}

func TestDaemonServerReportsRuntimeAndBindErrors(t *testing.T) {
	g := NewWithT(t)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	g.Expect(err).NotTo(HaveOccurred())
	t.Cleanup(func() { _ = listener.Close() })
	conf := &Config{
		Host: "127.0.0.1", GRPCPort: listener.Addr().(*net.TCPAddr).Port,
		CrClientConfig: &crclients.CrClientConfig{Runtime: "invalid"},
	}
	g.Expect(BuildServer(conf, prometheus.NewRegistry(), logr.Discard())).Error().To(MatchError(ContainSubstring("create daemon server")))

	conf.CrClientConfig.Runtime = crclients.ContainerRuntimeDocker
	conf.CrClientConfig.SocketPath = filepath.Join(t.TempDir(), "fixture.sock")
	server, err := BuildServer(conf, prometheus.NewRegistry(), logr.Discard())
	g.Expect(err).NotTo(HaveOccurred())
	t.Cleanup(func() { g.Expect(server.Shutdown()).To(Succeed()) })
	g.Expect(server.Start()).To(MatchError(ContainSubstring("listen grpc address")))
}

func TestDaemonGRPCReportsMissingTLSFiles(t *testing.T) {
	g := NewWithT(t)
	daemon := &DaemonServer{crClient: &httpChaosRuntimeClientStub{}}
	dir := t.TempDir()
	caPath := filepath.Join(dir, "ca.crt")
	opts := tlsConfig{CaCert: caPath, Cert: filepath.Join(dir, "server.crt"), Key: filepath.Join(dir, "server.key")}
	g.Expect(newGRPCServer(daemon, prometheus.NewRegistry(), opts)).Error().To(MatchError(ContainSubstring("ca.crt")))
	g.Expect(os.WriteFile(caPath, []byte("fixture CA"), 0o600)).To(Succeed())
	g.Expect(newGRPCServer(daemon, prometheus.NewRegistry(), opts)).Error().To(MatchError(ContainSubstring("server.crt")))
}

func TestMetadataExtractorCarriesExperimentIdentity(t *testing.T) {
	g := NewWithT(t)
	extract := MetadataExtractor(log.MetaNamespacedName)
	request := &struct{}{}
	handler := func(ctx context.Context, actual any) (any, error) {
		g.Expect(actual).To(BeIdenticalTo(request))
		return ctx.Value(log.MetaNamespacedName), nil
	}
	g.Expect(extract(t.Context(), request, nil, handler)).Error().To(HaveOccurred())
	ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs(string(log.MetaNamespacedName), "work/experiment", string(log.MetaNamespacedName), "ignored"))
	g.Expect(extract(ctx, request, nil, handler)).To(Equal("work/experiment"))
}
