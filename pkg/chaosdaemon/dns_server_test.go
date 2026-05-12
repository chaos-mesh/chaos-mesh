//go:build linux

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

package chaosdaemon_test

import (
	"context"
	"os/exec"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients/test"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

// buildTestDNSServer creates a DaemonServer wired to MockContainerdClient.
// MockContainerdClient is only consulted during client construction (containerd.New),
// so the failpoint can be cleared as soon as CreateContainerRuntimeInfoClient returns.
func buildTestDNSServer(t *testing.T) *chaosdaemon.DaemonServer {
	t.Helper()
	defer mock.With("MockContainerdClient", &test.MockClient{})()
	crc, err := crclients.CreateContainerRuntimeInfoClient(&crclients.CrClientConfig{
		Runtime: crclients.ContainerRuntimeContainerd,
	})
	if err != nil {
		t.Fatalf("create container runtime client: %v", err)
	}
	return chaosdaemon.NewDaemonServerWithCRClient(crc, nil, logr.Discard())
}

func TestSetDNSServer(t *testing.T) {
	cases := []struct {
		name      string
		dnsServer string
		enable    bool
		wantErr   bool
		wantErrIs error
	}{
		{
			name:      "valid IPv4 address",
			dnsServer: "8.8.8.8",
			enable:    true,
		},
		{
			name:      "valid IPv6 address",
			dnsServer: "::1",
			enable:    true,
		},
		{
			name:      "empty string returns ErrInvalidDNSServer",
			dnsServer: "",
			enable:    true,
			wantErr:   true,
			wantErrIs: chaosdaemon.ErrInvalidDNSServer,
		},
		{
			name:      "non-IP string returns ErrInvalidDNSServer",
			dnsServer: "notanip",
			enable:    true,
			wantErr:   true,
			wantErrIs: chaosdaemon.ErrInvalidDNSServer,
		},
		{
			// Enable=false runs the recovery shell command which contains "|| true",
			// so a missing backup file must not cause an error or panic.
			name:   "recover with no backup file does not error",
			enable: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			// MockProcessBuild is required on all platforms: bpm.Build panics on Darwin
			// without it and would call real system binaries on Linux.
			defer mock.With("MockProcessBuild", func(_ context.Context, _ string, _ ...string) *exec.Cmd {
				return exec.Command("echo", "mock")
			})()

			server := buildTestDNSServer(t)

			res, err := server.SetDNSServer(context.TODO(), &pb.SetDNSServerRequest{
				ContainerId: "containerd://foo",
				DnsServer:   tc.dnsServer,
				Enable:      tc.enable,
				EnterNS:     false,
			})

			if tc.wantErr {
				g.Expect(err).To(HaveOccurred())
				if tc.wantErrIs != nil {
					g.Expect(err).To(Equal(tc.wantErrIs))
				}
				g.Expect(res).To(BeNil())
			} else {
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(res).NotTo(BeNil())
			}
		})
	}
}
