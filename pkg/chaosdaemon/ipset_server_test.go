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
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients/test"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/log"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

var _ = Describe("ipset server", func() {
	defer mock.With("MockContainerdClient", &test.MockClient{})()
	logger, err := log.NewDefaultZapLogger()
	Expect(err).To(BeNil())
	s, _ := newDaemonServer(&crclients.CrClientConfig{
		Runtime: crclients.ContainerRuntimeContainerd}, 2000, nil, logger)

	Context("createIPSet", func() {
		It("should work", func() {
			defer mock.With("MockProcessBuild", func(ctx context.Context, cmd string, args ...string) *exec.Cmd {
				Expect(cmd).To(Equal("/usr/local/bin/nsexec"))
				Expect(args[0]).To(Equal("-n"))
				Expect(args[1]).To(Equal("/proc/1/ns/net"))
				Expect(args[2]).To(Equal("--"))
				Expect(args[3]).To(Equal("ipset"))
				Expect(args[4]).To(Equal("create"))
				Expect(args[5]).To(Equal("name"))
				Expect(args[6]).To(Equal("hash:net"))
				return exec.Command("echo", "mock command")
			})()
			err := createIPSet(context.TODO(), logger, true, 1, "name", v1alpha1.NetIPSet)
			Expect(err).To(BeNil())
		})

		It("should work since ipset exist", func() {
			// The mockfail.sh will fail only once
			err := os.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
echo $1
cat > /tmp/mockfail.sh << EOF
#! /bin/sh
exit 0
EOF
exit 1
			`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/mockfail.sh")
			defer mock.With("MockProcessBuild", func(ctx context.Context, cmd string, args ...string) *exec.Cmd {
				return exec.Command("/tmp/mockfail.sh", ipsetExistErr)
			})()
			err = createIPSet(context.TODO(), logger, true, 1, "name", v1alpha1.NetIPSet)
			Expect(err).To(BeNil())
		})

		It("shoud fail on the first command", func() {
			// The mockfail.sh will fail
			err := os.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
echo $1
exit 1
			`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/mockfail.sh")
			defer mock.With("MockProcessBuild", func(context.Context, string, ...string) *exec.Cmd {
				return exec.Command("/tmp/mockfail.sh", "fail msg")
			})()
			err = createIPSet(context.TODO(), logger, true, 1, "name", v1alpha1.NetIPSet)
			Expect(err).ToNot(BeNil())
		})

		It("shoud fail on the second command", func() {
			// The mockfail.sh will fail
			err := os.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
echo $1
exit 1
			`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/mockfail.sh")
			defer mock.With("MockProcessBuild", func(context.Context, string, ...string) *exec.Cmd {
				return exec.Command("/tmp/mockfail.sh", ipsetExistErr)
			})()
			err = createIPSet(context.TODO(), logger, true, 1, "name", v1alpha1.NetIPSet)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("addToIPSet", func() {
		It("should work", func() {
			defer mock.With("MockProcessBuild", func(context.Context, string, ...string) *exec.Cmd {
				return exec.Command("echo", "mock command")
			})()
			err := addToIPSet(context.TODO(), logger, true, 1, "name", "1.1.1.1")
			Expect(err).To(BeNil())
		})

		It("should work if ipset exists", func() {
			// The mockfail.sh will fail
			err := os.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
echo $1
exit 1
			`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/mockfail.sh")
			defer mock.With("MockProcessBuild", func(context.Context, string, ...string) *exec.Cmd {
				return exec.Command("/tmp/mockfail.sh", ipExistErr)
			})()
			err = addToIPSet(context.TODO(), logger, true, 1, "name", "1.1.1.1")
			Expect(err).To(BeNil())
		})

		It("should fail", func() {
			// The mockfail.sh will fail
			err := os.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
echo $1
exit 1
			`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/mockfail.sh")
			defer mock.With("MockProcessBuild", func(context.Context, string, ...string) *exec.Cmd {
				return exec.Command("/tmp/mockfail.sh", "fail msg")
			})()
			err = addToIPSet(context.TODO(), logger, true, 1, "name", "1.1.1.1")
			Expect(err).ToNot(BeNil())
		})
	})

	Context("renameIPSet", func() {
		It("should work", func() {
			defer mock.With("MockProcessBuild", func(context.Context, string, ...string) *exec.Cmd {
				return exec.Command("echo", "mock command")
			})()
			err := renameIPSet(context.TODO(), logger, true, 1, "name", "newname")
			Expect(err).To(BeNil())
		})

		It("should work since ipset exist", func() {
			// The mockfail.sh will fail only once
			err := os.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
echo $1
cat > /tmp/mockfail.sh << EOF
#! /bin/sh
exit 0
EOF
exit 1
			`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/mockfail.sh")
			defer mock.With("MockProcessBuild", func(context.Context, string, ...string) *exec.Cmd {
				return exec.Command("/tmp/mockfail.sh", ipsetNewNameExistErr)
			})()
			err = renameIPSet(context.TODO(), logger, true, 1, "name", "newname")
			Expect(err).To(BeNil())
		})

		It("shoud fail on the first command", func() {
			// The mockfail.sh will fail
			err := os.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
echo $1
exit 1
			`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/mockfail.sh")
			defer mock.With("MockProcessBuild", func(context.Context, string, ...string) *exec.Cmd {
				return exec.Command("/tmp/mockfail.sh", "fail msg")
			})()
			err = renameIPSet(context.TODO(), logger, true, 1, "name", "newname")
			Expect(err).ToNot(BeNil())
		})

		It("shoud fail on the second command", func() {
			// The mockfail.sh will fail
			err := os.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
echo $1
exit 1
			`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/mockfail.sh")
			defer mock.With("MockProcessBuild", func(context.Context, string, ...string) *exec.Cmd {
				return exec.Command("/tmp/mockfail.sh", ipsetExistErr)
			})()
			err = renameIPSet(context.TODO(), logger, true, 1, "name", "newname")
			Expect(err).ToNot(BeNil())
		})
	})

	Context("FlushIPSets", func() {
		It("should work", func() {
			defer mock.With("MockProcessBuild", func(context.Context, string, ...string) *exec.Cmd {
				return exec.Command("echo", "mock command")
			})()
			_, err := s.FlushIPSets(context.TODO(), &pb.IPSetsRequest{
				Ipsets: []*pb.IPSet{
					{
						Name:     "ipset-set-name",
						Type:     "list:set",
						SetNames: []string{"set-1", "set-2"},
					},
					{
						Name:  "ipset-net-name",
						Type:  "hash:net",
						Cidrs: []string{"0.0.0.0/24"},
					},
					{
						Name: "ipset-net-port-name",
						Type: "hash:net,port",
						CidrAndPorts: []*pb.CidrAndPort{{
							Cidr: "1.1.1.1/32",
							Port: 80,
						}},
					},
				},
				ContainerId: "containerd://container-id",
				EnterNS:     true,
			})
			Expect(err).To(BeNil())
		})

		It("should fail on get pid", func() {
			const errorStr = "mock get pid error"
			defer mock.With("TaskError", errors.New(errorStr))()
			_, err := s.FlushIPSets(context.TODO(), &pb.IPSetsRequest{
				Ipsets: []*pb.IPSet{{
					Name:     "ipset-name",
					SetNames: []string{"set-1", "set-2"},
				}},
				ContainerId: "containerd://container-id",
				EnterNS:     true,
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(errorStr))
		})

		It("should fail on unknown type", func() {
			_, err := s.FlushIPSets(context.TODO(), &pb.IPSetsRequest{
				Ipsets: []*pb.IPSet{{
					Name: "ipset-name",
					Type: "foo:bar",
				}},
				ContainerId: "containerd://container-id",
				EnterNS:     true,
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("unexpected IP set type: foo:bar"))
		})
	})
})
