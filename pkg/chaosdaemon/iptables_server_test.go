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

	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

var _ = Describe("iptables server", func() {
	defer mock.With("MockContainerdClient", &MockClient{})()
	s, _ := newDaemonServer(containerRuntimeContainerd)

	Context("FlushIptables", func() {
		It("should work", func() {
			defer mock.With("pid", 9527)()
			defer mock.With("MockProcessBuild", func(ctx context.Context, cmd string, args ...string) *exec.Cmd {
				Expect(cmd).To(Equal("/usr/local/bin/nsexec"))
				Expect(args[0]).To(Equal("-n"))
				Expect(args[1]).To(Equal("/proc/9527/ns/net"))
				Expect(args[2]).To(Equal("--"))
				Expect(args[3]).To(Equal(iptablesCmd))
				return exec.Command("echo", "-n")
			})()
			_, err := s.SetIptablesChains(context.TODO(), &pb.IptablesChainsRequest{
				Chains: []*pb.Chain{{
					Name:      "TEST",
					Direction: pb.Chain_INPUT,
					Ipsets:    []string{},
				}},
				ContainerId: "containerd://container-id",
				EnterNS:     true,
			})
			Expect(err).To(BeNil())
		})

		It("should fail on get pid", func() {
			const errorStr = "mock error on Task()"
			defer mock.With("TaskError", errors.New(errorStr))()
			_, err := s.SetIptablesChains(context.TODO(), &pb.IptablesChainsRequest{
				Chains: []*pb.Chain{{
					Name:      "TEST",
					Direction: pb.Chain_INPUT,
					Ipsets:    []string{},
				}},
				ContainerId: "containerd://container-id",
				EnterNS:     true,
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(errorStr))
		})

		It("should fail on unknown chain direction", func() {
			defer mock.With("pid", 9527)()
			defer mock.With("MockProcessBuild", func(ctx context.Context, cmd string, args ...string) *exec.Cmd {
				Expect(cmd).To(Equal("/usr/local/bin/nsexec"))
				Expect(args[0]).To(Equal("-n"))
				Expect(args[1]).To(Equal("/proc/9527/ns/net"))
				Expect(args[2]).To(Equal("--"))
				Expect(args[3]).To(Equal(iptablesCmd))
				return exec.Command("echo", "-n")
			})()

			_, err := s.SetIptablesChains(context.TODO(), &pb.IptablesChainsRequest{
				Chains: []*pb.Chain{{
					Name:      "TEST",
					Direction: pb.Chain_Direction(233),
					Ipsets:    []string{},
				}},
				ContainerId: "containerd://container-id",
				EnterNS:     true,
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("unknown chain direction 233"))
		})

		It("should fail on command error", func() {
			// The mockfail.sh will fail
			err := ioutil.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
exit 1
			`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/mockfail.sh")
			defer mock.With("MockProcessBuild", func(ctx context.Context, cmd string, args ...string) *exec.Cmd {
				return exec.Command("mockfail.sh")
			})()
			_, err = s.SetIptablesChains(context.TODO(), &pb.IptablesChainsRequest{
				Chains: []*pb.Chain{{
					Name:      "TEST",
					Direction: pb.Chain_INPUT,
					Ipsets:    []string{},
				}},
				ContainerId: "containerd://container-id",
				EnterNS:     true,
			})
			Expect(err).ToNot(BeNil())
		})
	})
})
