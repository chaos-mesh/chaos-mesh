// Copyright 2023 Chaos Mesh Authors.
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

func Test_SetDNSServer_Enable(t *testing.T) {
	g := NewWithT(t)

	type mockCmd struct {
		cmd  string
		args []string
	}
	var executedCommands []mockCmd

	mock.With("MockProcessBuild", func(ctx context.Context, cmd string, args ...string) *exec.Cmd {
		executedCommands = append(executedCommands, mockCmd{cmd, args})
		return exec.Command("echo", "mock command")
	})

	mock.With("MockContainerdClient", &test.MockClient{})

	crc, err := crclients.CreateContainerRuntimeInfoClient(&crclients.CrClientConfig{
		Runtime: crclients.ContainerRuntimeContainerd,
	})
	g.Expect(err).NotTo(HaveOccurred())

	server := chaosdaemon.NewDaemonServerWithCRClient(crc, nil, logr.Discard())

	res, err := server.SetDNSServer(context.TODO(), &pb.SetDNSServerRequest{
		ContainerId: "containerd://foo",
		DnsServer:   "8.6.4.2",
		Enable:      true,
		EnterNS:     false,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res).NotTo(BeNil())

	g.Expect(executedCommands).To(Equal([]mockCmd{
		{cmd: "sh", args: []string{"-c", "ls /etc/resolv.conf.chaos.bak || cp /etc/resolv.conf /etc/resolv.conf.chaos.bak"}},
		{cmd: "sh", args: []string{"-c", "cp /etc/resolv.conf /etc/resolv_conf_dnschaos_temp && sed -i 's/.*nameserver.*/nameserver 8.6.4.2/' /etc/resolv_conf_dnschaos_temp && cat /etc/resolv_conf_dnschaos_temp > /etc/resolv.conf && rm /etc/resolv_conf_dnschaos_temp"}},
	}))
}

func Test_SetDNSServer_Enable_InvalidIP(t *testing.T) {
	g := NewWithT(t)

	cases := []string{"", "127.0.0.b", " 127.0.0.1", "127.0.0.1 ", ":g:1", "127.0.0.1;"}

	mock.With("MockProcessBuild", func(ctx context.Context, cmd string, args ...string) *exec.Cmd {
		g.Fail("no process should be executed")
		return exec.Command("echo", "mock command")
	})

	mock.With("MockContainerdClient", &test.MockClient{})

	crc, err := crclients.CreateContainerRuntimeInfoClient(&crclients.CrClientConfig{
		Runtime: crclients.ContainerRuntimeContainerd,
	})
	g.Expect(err).NotTo(HaveOccurred())

	server := chaosdaemon.NewDaemonServerWithCRClient(crc, nil, logr.Discard())

	for _, tc := range cases {
		res, err := server.SetDNSServer(context.TODO(), &pb.SetDNSServerRequest{
			ContainerId: "containerd://foo",
			DnsServer:   tc,
			Enable:      true,
			EnterNS:     false,
		})
		g.Expect(err).To(Equal(chaosdaemon.ErrInvalidDNSServer))
		g.Expect(res).To(BeNil())
	}
}

func Test_SetDNSServer_Disable(t *testing.T) {
	g := NewWithT(t)

	type mockCmd struct {
		cmd  string
		args []string
	}
	var executedCommands []mockCmd

	mock.With("MockProcessBuild", func(ctx context.Context, cmd string, args ...string) *exec.Cmd {
		executedCommands = append(executedCommands, mockCmd{cmd, args})
		return exec.Command("echo", "mock command")
	})

	mock.With("MockContainerdClient", &test.MockClient{})

	crc, err := crclients.CreateContainerRuntimeInfoClient(&crclients.CrClientConfig{
		Runtime: crclients.ContainerRuntimeContainerd,
	})
	g.Expect(err).NotTo(HaveOccurred())

	server := chaosdaemon.NewDaemonServerWithCRClient(crc, nil, logr.Discard())

	res, err := server.SetDNSServer(context.TODO(), &pb.SetDNSServerRequest{
		ContainerId: "containerd://foo",
		DnsServer:   "",
		Enable:      false,
		EnterNS:     false,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res).NotTo(BeNil())

	g.Expect(executedCommands).To(Equal([]mockCmd{
		{cmd: "sh", args: []string{"-c", "ls /etc/resolv.conf.chaos.bak && cat /etc/resolv.conf.chaos.bak > /etc/resolv.conf || true"}},
	}))
}
