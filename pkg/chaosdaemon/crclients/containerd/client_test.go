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

package containerd

import (
	"context"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients/test"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

var _ = Describe("containerd client", func() {
	Context("ContainerdClient GetPidFromContainerID", func() {
		It("should return the magic number 9527", func() {
			defer func() {
				err := mock.With("pid", int(9527))()
				Expect(err).To(BeNil())
			}()

			m := &test.MockClient{}
			c := ContainerdClient{client: m}
			pid, err := c.GetPidFromContainerID(context.TODO(), "containerd://valid-container-id")
			Expect(err).To(BeNil())
			Expect(pid).To(Equal(uint32(9527)))
		})

		It("should error with wrong protocol", func() {
			m := &test.MockClient{}
			c := ContainerdClient{client: m}
			_, err := c.GetPidFromContainerID(context.TODO(), "docker://this-is-a-wrong-protocol")
			Expect(err).NotTo(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring(fmt.Sprintf("expected %s but got", containerdProtocolPrefix)))
		})

		It("should error with specified string", func() {
			errorStr := "this is a mocked error"
			mock.With("LoadContainerError", errors.New(errorStr))
			m := &test.MockClient{}
			c := ContainerdClient{client: m}
			_, err := c.GetPidFromContainerID(context.TODO(), "containerd://valid-container-id")
			Expect(err).NotTo(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(Equal(errorStr))
			err = mock.Reset("LoadContainerError")
			Expect(err).NotTo(BeNil())

			mock.With("TaskError", errors.New(errorStr))
			m = &test.MockClient{}
			c = ContainerdClient{client: m}
			_, err = c.GetPidFromContainerID(context.TODO(), "containerd://valid-container-id")
			Expect(err).NotTo(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(Equal(errorStr))
			err = mock.Reset("TaskError")
			Expect(err).NotTo(BeNil())
		})
	})

	Context("ContainerdClient ContainerKillByContainerID", func() {
		It("should work", func() {
			m := &test.MockClient{}
			c := ContainerdClient{client: m}
			err := c.ContainerKillByContainerID(context.TODO(), "containerd://valid-container-id")
			Expect(err).To(BeNil())
		})

		errorPoints := []string{"LoadContainer", "Task", "Kill"}
		for _, e := range errorPoints {
			It(fmt.Sprintf("should error on %s", e), func() {
				errorStr := fmt.Sprintf("this is a mocked error on %s", e)
				m := &test.MockClient{}
				c := ContainerdClient{client: m}
				defer mock.With(e+"Error", errors.New(errorStr))()
				err := c.ContainerKillByContainerID(context.TODO(), "containerd://valid-container-id")
				Expect(err).ToNot(BeNil())
				Expect(fmt.Sprintf("%s", err)).To(Equal(errorStr))
			})
		}

		It("should error on wrong protocol", func() {
			m := &test.MockClient{}
			c := ContainerdClient{client: m}
			err := c.ContainerKillByContainerID(context.TODO(), "docker://this-is-a-wrong-protocol")
			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring(fmt.Sprintf("expected %s but got", containerdProtocolPrefix)))
		})

		It("should error on short protocol", func() {
			m := &test.MockClient{}
			c := ContainerdClient{client: m}
			err := c.ContainerKillByContainerID(context.TODO(), "dock:")
			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring("is not a containerd container id"))
		})
	})
})
