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

var _ = Describe("time server", func() {
	defer mock.With("MockContainerdClient", &MockClient{})()
	s, _ := newDaemonServer(containerRuntimeContainerd)

	Context("SetTimeOffset", func() {
		It("should work", func() {
			// Inject nil error to ignore any error
			const ignore = true
			defer mock.With("ModifyTimeError", ignore)()

			_, err := s.SetTimeOffset(context.TODO(), &pb.TimeRequest{
				ContainerId: "containerd://container-id",
			})
			Expect(err).To(BeNil())
		})

		It("should fail on get pid", func() {
			const errorStr = "mock error on load container"
			defer mock.With("LoadContainerError", errors.New(errorStr))()

			_, err := s.SetTimeOffset(context.TODO(), &pb.TimeRequest{
				ContainerId: "containerd://container-id",
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(errorStr))
		})

		It("should fail on modify time", func() {
			const errorStr = "mock error on modify time"
			defer mock.With("ModifyTimeError", errors.New(errorStr))()

			_, err := s.SetTimeOffset(context.TODO(), &pb.TimeRequest{
				ContainerId: "containerd://container-id",
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(errorStr))
		})
	})

	Context("RecoverTimeOffset", func() {
		It("should work", func() {
			// Inject nil error to ignore any error
			const ignore = true
			defer mock.With("ModifyTimeError", ignore)()

			_, err := s.RecoverTimeOffset(context.TODO(), &pb.TimeRequest{
				ContainerId: "containerd://container-id",
			})
			Expect(err).To(BeNil())
		})

		It("should fail on get pid", func() {
			const errorStr = "mock error on load container"
			defer mock.With("LoadContainerError", errors.New(errorStr))()

			_, err := s.RecoverTimeOffset(context.TODO(), &pb.TimeRequest{
				ContainerId: "containerd://container-id",
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(errorStr))
		})

		It("should fail on modify time", func() {
			const errorStr = "mock error on modify time"
			defer mock.With("ModifyTimeError", errors.New(errorStr))()

			_, err := s.RecoverTimeOffset(context.TODO(), &pb.TimeRequest{
				ContainerId: "containerd://container-id",
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(errorStr))
		})
	})
})
