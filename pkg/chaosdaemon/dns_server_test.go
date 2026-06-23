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
	"os"
	"path/filepath"
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

	// Create a temp directory for testing
	tmpDir := t.TempDir()
	tempResolvConf := filepath.Join(tmpDir, "resolv.conf")
	initialContent := "nameserver 1.1.1.1\nnameserver 8.8.8.8\noptions ndots:5\n"
	err := os.WriteFile(tempResolvConf, []byte(initialContent), 0644)
	g.Expect(err).NotTo(HaveOccurred())

	// Override the configuration file path
	originalConfFile := chaosdaemon.DNSServerConfFile
	chaosdaemon.DNSServerConfFile = tempResolvConf
	defer func() { chaosdaemon.DNSServerConfFile = originalConfFile }()

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

	// Verify target resolv.conf contents
	modifiedContent, err := os.ReadFile(tempResolvConf)
	g.Expect(err).NotTo(HaveOccurred())
	expectedContent := "nameserver 8.6.4.2\nnameserver 8.6.4.2\noptions ndots:5\n"
	g.Expect(string(modifiedContent)).To(Equal(expectedContent))

	// Verify backup file exists and has original contents
	backupContent, err := os.ReadFile(tempResolvConf + ".chaos.bak")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(backupContent)).To(Equal(initialContent))
}

func Test_SetDNSServer_Enable_InvalidIP(t *testing.T) {
	g := NewWithT(t)

	cases := []string{"", "127.0.0.b", " 127.0.0.1", "127.0.0.1 ", ":g:1", "127.0.0.1;"}

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

	// Create a temp directory for testing
	tmpDir := t.TempDir()
	tempResolvConf := filepath.Join(tmpDir, "resolv.conf")
	initialContent := "nameserver 1.1.1.1\nnameserver 8.8.8.8\noptions ndots:5\n"
	err := os.WriteFile(tempResolvConf, []byte("nameserver 8.6.4.2\noptions ndots:5\n"), 0644)
	g.Expect(err).NotTo(HaveOccurred())

	// Create the backup file
	backupResolvConf := tempResolvConf + ".chaos.bak"
	err = os.WriteFile(backupResolvConf, []byte(initialContent), 0644)
	g.Expect(err).NotTo(HaveOccurred())

	// Override the configuration file path
	originalConfFile := chaosdaemon.DNSServerConfFile
	chaosdaemon.DNSServerConfFile = tempResolvConf
	defer func() { chaosdaemon.DNSServerConfFile = originalConfFile }()

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

	// Verify target resolv.conf contents are restored
	restoredContent, err := os.ReadFile(tempResolvConf)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(restoredContent)).To(Equal(initialContent))

	// Verify backup file has been deleted
	_, err = os.Stat(backupResolvConf)
	g.Expect(os.IsNotExist(err)).To(BeTrue())
}

func Test_SetDNSServer_Enable_EnterNS(t *testing.T) {
	g := NewWithT(t)

	// Create a temp directory for testing
	tmpDir := t.TempDir()
	tempResolvConf := filepath.Join(tmpDir, "resolv.conf")
	initialContent := "nameserver 1.1.1.1\nnameserver 8.8.8.8\noptions ndots:5\n"
	err := os.WriteFile(tempResolvConf, []byte(initialContent), 0644)
	g.Expect(err).NotTo(HaveOccurred())

	// Override the configuration file path
	originalConfFile := chaosdaemon.DNSServerConfFile
	chaosdaemon.DNSServerConfFile = tempResolvConf
	defer func() { chaosdaemon.DNSServerConfFile = originalConfFile }()

	mock.With("MockContainerdClient", &test.MockClient{})

	// Mock the PID returned by MockContainerdClient to be the current process PID
	myPid := os.Getpid()
	mock.With("pid", myPid)

	crc, err := crclients.CreateContainerRuntimeInfoClient(&crclients.CrClientConfig{
		Runtime: crclients.ContainerRuntimeContainerd,
	})
	g.Expect(err).NotTo(HaveOccurred())

	server := chaosdaemon.NewDaemonServerWithCRClient(crc, nil, logr.Discard())

	res, err := server.SetDNSServer(context.TODO(), &pb.SetDNSServerRequest{
		ContainerId: "containerd://foo",
		DnsServer:   "8.6.4.2",
		Enable:      true,
		EnterNS:     true, // Test the new EnterNS logic using procfs path interpolation
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res).NotTo(BeNil())

	// Verify target resolv.conf contents
	modifiedContent, err := os.ReadFile(tempResolvConf)
	g.Expect(err).NotTo(HaveOccurred())
	expectedContent := "nameserver 8.6.4.2\nnameserver 8.6.4.2\noptions ndots:5\n"
	g.Expect(string(modifiedContent)).To(Equal(expectedContent))

	// Verify backup file exists and has original contents
	backupContent, err := os.ReadFile(tempResolvConf + ".chaos.bak")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(backupContent)).To(Equal(initialContent))
}

