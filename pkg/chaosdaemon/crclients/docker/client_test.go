// Copyright 2021 Chaos Mesh Authors.
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

package docker

import (
	"context"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients/test"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

var _ = Describe("docker client", func() {
	Context("DockerClient GetPidFromContainerID", func() {
		It("should return the magic number 9527", func() {
			defer func() {
				err := mock.With("pid", int(9527))()
				Expect(err).To(BeNil())
			}()

			m := &test.MockClient{}
			c := DockerClient{client: m}
			pid, err := c.GetPidFromContainerID(context.TODO(), "docker://valid-container-id")
			Expect(err).To(BeNil())
			Expect(pid).To(Equal(uint32(9527)))
		})

		It("should error with wrong protocol", func() {
			m := &test.MockClient{}
			c := DockerClient{client: m}
			_, err := c.GetPidFromContainerID(context.TODO(), "containerd://this-is-a-wrong-protocol")
			Expect(err).NotTo(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring(fmt.Sprintf("expected %s but got", dockerProtocolPrefix)))
		})

		It("should error on ContainerInspectError", func() {
			errorStr := "this is a mocked error"
			defer func() {
				err := mock.With("ContainerInspectError", errors.New(errorStr))()
				Expect(err).NotTo(BeNil())
			}()
			m := &test.MockClient{}
			c := DockerClient{client: m}
			_, err := c.GetPidFromContainerID(context.TODO(), "docker://valid-container-id")
			Expect(err).NotTo(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(Equal(errorStr))
		})
	})

	Context("DockerClient ContainerKillByContainerID", func() {
		It("should work", func() {
			m := &test.MockClient{}
			c := DockerClient{client: m}
			err := c.ContainerKillByContainerID(context.TODO(), "docker://valid-container-id")
			Expect(err).To(BeNil())
		})

		It("should error on ContainerKill", func() {
			errorStr := "this is a mocked error on ContainerKill"
			m := &test.MockClient{}
			c := DockerClient{client: m}
			defer func() {
				err := mock.With("ContainerKillError", errors.New(errorStr))()
				Expect(err).ToNot(BeNil())
			}()
			err := c.ContainerKillByContainerID(context.TODO(), "docker://valid-container-id")
			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(Equal(errorStr))
		})

		It("should error on wrong protocol", func() {
			m := &test.MockClient{}
			c := DockerClient{client: m}
			err := c.ContainerKillByContainerID(context.TODO(), "containerd://this-is-a-wrong-protocol")
			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring(fmt.Sprintf("expected %s but got", dockerProtocolPrefix)))
		})

		It("should error on short protocol", func() {
			m := &test.MockClient{}
			c := DockerClient{client: m}
			err := c.ContainerKillByContainerID(context.TODO(), "dock:")
			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring("is not a docker container id"))
		})
	})

})
