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

	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/pingcap/chaos-mesh/pkg/mock"
)

var _ = Describe("iptables server", func() {
	defer mock.With("MockContainerdClient", &MockClient{})()
	c, _ := CreateContainerRuntimeInfoClient(containerRuntimeContainerd)
	s := &daemonServer{c}

	Context("addIptablesRule", func() {
		It("should work", func() {
			err := s.addIptablesRules(context.TODO(), exec.Command("echo", "mock command"))
			Expect(err).To(BeNil())
		})

		It("should fail on command", func() {
			// The mockfail.sh will fail
			err := ioutil.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
exit 1
			`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/mockfail.sh")
			err = s.addIptablesRules(context.TODO(), exec.Command("/tmp/mockfail.sh"))
			Expect(err).ToNot(BeNil())
		})
	})

	Context("deleteIptablesRules", func() {
		It("should work", func() {
			err := s.deleteIptablesRules(context.TODO(), exec.Command("echo", "mock command"))
			Expect(err).To(BeNil())
		})

		It("should fail on command", func() {
			// The mockfail.sh will fail
			err := ioutil.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
exit 1
			`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/mockfail.sh")
			err = s.deleteIptablesRules(context.TODO(), exec.Command("/tmp/mockfail.sh"))
			Expect(err).ToNot(BeNil())
		})

		It("should work since iptablesBadRuleErr or iptablesIPSetNotExistErr", func() {
			// The mockfail.sh will fail
			err := ioutil.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
echo $1
exit 1
			`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/mockfail.sh")
			err = s.deleteIptablesRules(context.TODO(), exec.Command("/tmp/mockfail.sh", iptablesBadRuleErr))
			Expect(err).To(BeNil())
			err = s.deleteIptablesRules(context.TODO(), exec.Command("/tmp/mockfail.sh", iptablesIPSetNotExistErr))
			Expect(err).To(BeNil())
		})
	})

	Context("FlushIptables", func() {
		It("should work", func() {
			defer mock.With("pid", 9527)()
			defer mock.With("MockWithNetNs", func(ctx context.Context, ns, cmd string, args ...string) *exec.Cmd {
				Expect(ns).To(Equal("/proc/9527/ns/net"))
				Expect(cmd).To(Equal(iptablesCmd))
				return exec.Command("echo", "mock command")
			})()
			_, err := s.FlushIptables(context.TODO(), &pb.IpTablesRequest{
				Rule: &pb.Rule{
					Direction: pb.Rule_INPUT,
					Action:    pb.Rule_ADD,
				},
				ContainerId: "containerd://container-id",
			})
			Expect(err).To(BeNil())
		})

		It("should fail on get pid", func() {
			const errorStr = "mock error on Task()"
			defer mock.With("TaskError", errors.New(errorStr))()
			_, err := s.FlushIptables(context.TODO(), &pb.IpTablesRequest{
				Rule: &pb.Rule{
					Direction: pb.Rule_INPUT,
					Action:    pb.Rule_ADD,
				},
				ContainerId: "containerd://container-id",
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(errorStr))
		})

		It("should fail on unknown rule direction", func() {
			_, err := s.FlushIptables(context.TODO(), &pb.IpTablesRequest{
				Rule: &pb.Rule{
					Action:    pb.Rule_ADD,
					Direction: pb.Rule_Direction(233),
				},
				ContainerId: "containerd://container-id",
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("unknown rule direction"))
		})

		It("should fail on unknow rule action", func() {
			_, err := s.FlushIptables(context.TODO(), &pb.IpTablesRequest{
				Rule: &pb.Rule{
					Direction: pb.Rule_OUTPUT,
					Action:    pb.Rule_Action(233),
				},
				ContainerId: "containerd://container-id",
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("unknown rule action"))
		})

		It("should fail on command error", func() {
			// The mockfail.sh will fail
			err := ioutil.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
exit 1
			`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/mockfail.sh")
			defer mock.With("MockWithNetNs", func(ctx context.Context, ns, cmd string, args ...string) *exec.Cmd {
				return exec.Command("mockfail.sh")
			})()
			_, err = s.FlushIptables(context.TODO(), &pb.IpTablesRequest{
				Rule: &pb.Rule{
					Direction: pb.Rule_INPUT,
					Action:    pb.Rule_DELETE,
				},
				ContainerId: "containerd://container-id",
			})
			Expect(err).ToNot(BeNil())
		})
	})
})
