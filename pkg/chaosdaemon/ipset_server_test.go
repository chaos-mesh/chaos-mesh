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
	"errors"
	"io/ioutil"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

var _ = Describe("ipset server", func() {
	defer mock.With("MockContainerdClient", &MockClient{})()
	c, _ := CreateContainerRuntimeInfoClient(containerRuntimeContainerd)
	m := bpm.NewBackgroundProcessManager()
	s := &DaemonServer{c, m}

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
			err := createIPSet(context.TODO(), 1, "name", false)
			Expect(err).To(BeNil())
		})

		It("should work since ipset exist", func() {
			// The mockfail.sh will fail only once
			err := ioutil.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
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
			err = createIPSet(context.TODO(), 1, "name", false)
			Expect(err).To(BeNil())
		})

		It("shoud fail on the first command", func() {
			// The mockfail.sh will fail
			err := ioutil.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
echo $1
exit 1
			`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/mockfail.sh")
			defer mock.With("MockProcessBuild", func(context.Context, string, ...string) *exec.Cmd {
				return exec.Command("/tmp/mockfail.sh", "fail msg")
			})()
			err = createIPSet(context.TODO(), 1, "name", false)
			Expect(err).ToNot(BeNil())
		})

		It("shoud fail on the second command", func() {
			// The mockfail.sh will fail
			err := ioutil.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
echo $1
exit 1
			`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/mockfail.sh")
			defer mock.With("MockProcessBuild", func(context.Context, string, ...string) *exec.Cmd {
				return exec.Command("/tmp/mockfail.sh", ipsetExistErr)
			})()
			err = createIPSet(context.TODO(), 1, "name", false)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("addIpsToIPSet", func() {
		It("should work", func() {
			defer mock.With("MockProcessBuild", func(context.Context, string, ...string) *exec.Cmd {
				return exec.Command("echo", "mock command")
			})()
			err := addCIDRsToIPSet(context.TODO(), 1, "name", []string{"1.1.1.1"}, false)
			Expect(err).To(BeNil())
		})

		It("should work since ip exist", func() {
			// The mockfail.sh will fail
			err := ioutil.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
echo $1
exit 1
			`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/mockfail.sh")
			defer mock.With("MockProcessBuild", func(context.Context, string, ...string) *exec.Cmd {
				return exec.Command("/tmp/mockfail.sh", ipExistErr)
			})()
			err = addCIDRsToIPSet(context.TODO(), 1, "name", []string{"1.1.1.1"}, false)
			Expect(err).To(BeNil())
		})

		It("should fail", func() {
			// The mockfail.sh will fail
			err := ioutil.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
echo $1
exit 1
			`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/mockfail.sh")
			defer mock.With("MockProcessBuild", func(context.Context, string, ...string) *exec.Cmd {
				return exec.Command("/tmp/mockfail.sh", "fail msg")
			})()
			err = addCIDRsToIPSet(context.TODO(), 1, "name", []string{"1.1.1.1"}, false)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("renameIPSet", func() {
		It("should work", func() {
			defer mock.With("MockProcessBuild", func(context.Context, string, ...string) *exec.Cmd {
				return exec.Command("echo", "mock command")
			})()
			err := renameIPSet(context.TODO(), 1, "name", "newname", false)
			Expect(err).To(BeNil())
		})

		It("should work since ipset exist", func() {
			// The mockfail.sh will fail only once
			err := ioutil.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
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
			err = renameIPSet(context.TODO(), 1, "name", "newname", false)
			Expect(err).To(BeNil())
		})

		It("shoud fail on the first command", func() {
			// The mockfail.sh will fail
			err := ioutil.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
echo $1
exit 1
			`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/mockfail.sh")
			defer mock.With("MockProcessBuild", func(context.Context, string, ...string) *exec.Cmd {
				return exec.Command("/tmp/mockfail.sh", "fail msg")
			})()
			err = renameIPSet(context.TODO(), 1, "name", "newname", false)
			Expect(err).ToNot(BeNil())
		})

		It("shoud fail on the second command", func() {
			// The mockfail.sh will fail
			err := ioutil.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
echo $1
exit 1
			`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/mockfail.sh")
			defer mock.With("MockProcessBuild", func(context.Context, string, ...string) *exec.Cmd {
				return exec.Command("/tmp/mockfail.sh", ipsetExistErr)
			})()
			err = renameIPSet(context.TODO(), 1, "name", "newname", false)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("FlushIPSets", func() {
		It("should work", func() {
			defer mock.With("MockProcessBuild", func(context.Context, string, ...string) *exec.Cmd {
				return exec.Command("echo", "mock command")
			})()
			_, err := s.FlushIPSets(context.TODO(), &pb.IPSetsRequest{
				Ipsets: []*pb.IPSet{{
					Name:  "ipset-name",
					Cidrs: []string{"1.1.1.1/32"},
				}},
				ContainerId: "containerd://container-id",
			})
			Expect(err).To(BeNil())
		})

		It("should fail on get pid", func() {
			const errorStr = "mock get pid error"
			defer mock.With("TaskError", errors.New(errorStr))()
			_, err := s.FlushIPSets(context.TODO(), &pb.IPSetsRequest{
				Ipsets: []*pb.IPSet{{
					Name:  "ipset-name",
					Cidrs: []string{"1.1.1.1/32"},
				}},
				ContainerId: "containerd://container-id",
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(errorStr))
		})
	})
})
