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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

var _ = Describe("container kill", func() {
	defer mock.With("MockContainerdClient", &MockClient{})()
	s, _ := newDaemonServer(containerRuntimeContainerd)

	Context("ContainerKill", func() {
		It("should work", func() {
			_, err := s.ContainerKill(context.TODO(), &pb.ContainerRequest{
				Action: &pb.ContainerAction{
					Action: pb.ContainerAction_KILL,
				},
				ContainerId: "containerd://container-id",
			})
			Expect(err).To(BeNil())
		})

		It("should fail on wrong action type", func() {
			const wrongActionType = 9527
			_, err := s.ContainerKill(context.TODO(), &pb.ContainerRequest{
				Action: &pb.ContainerAction{
					Action: pb.ContainerAction_Action(wrongActionType),
				},
				ContainerId: "containerd://container-id",
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("not kill"))
		})

		It("should fail on container kill", func() {
			const errorStr = "mock error on container kill"
			defer mock.With("KillError", errors.New(errorStr))()
			_, err := s.ContainerKill(context.TODO(), &pb.ContainerRequest{
				Action: &pb.ContainerAction{
					Action: pb.ContainerAction_KILL,
				},
				ContainerId: "containerd://container-id",
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(errorStr))
		})
	})
})
