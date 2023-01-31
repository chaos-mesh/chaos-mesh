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

package containerd

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients/test"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

func TestContainerdClient(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Containerd Container Client Test Suit")
}

var _ = Describe("containerd client", func() {
	Context("ContainerdClient GetPidFromContainerID", func() {
		It("should return the magic number 9527", func() {
			defer mock.With("pid", int(9527))()

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
			mock.Reset("LoadContainerError")

			mock.With("TaskError", errors.New(errorStr))
			m = &test.MockClient{}
			c = ContainerdClient{client: m}
			_, err = c.GetPidFromContainerID(context.TODO(), "containerd://valid-container-id")
			Expect(err).NotTo(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(Equal(errorStr))
			mock.Reset("TaskError")
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

	Context("ContainerdClient ListContainerIDs", func() {
		It("should work", func() {
			containerID := "valid-container-id"
			containerIDWithPrefix := fmt.Sprintf("%s%s", containerdProtocolPrefix, containerID)
			defer mock.With("containerID", containerID)()

			m := &test.MockClient{}
			c := ContainerdClient{client: m}
			containerIDs, err := c.ListContainerIDs(context.Background())

			Expect(err).To(BeNil())
			Expect(containerIDs).To(Equal([]string{containerIDWithPrefix}))
		})
	})

	Context("ContainerdClient GetLabelsFromContainerID", func() {
		It("should work", func() {
			sampleLabels := map[string]string{
				"io.kubernetes.pod.namespace":  "default",
				"io.kubernetes.pod.name":       "busybox-5f8dd756dd-6rjzw",
				"io.kubernetes.container.name": "busybox",
			}
			defer mock.With("labels", sampleLabels)()

			m := &test.MockClient{}
			c := ContainerdClient{client: m}
			labels, err := c.GetLabelsFromContainerID(context.Background(), "containerd://valid-container-id")

			Expect(err).To(BeNil())
			Expect(labels).To(Equal(sampleLabels))
		})

		It("should error on wrong protocol", func() {
			m := &test.MockClient{}
			c := ContainerdClient{client: m}
			_, err := c.GetLabelsFromContainerID(context.Background(), "docker://this-is-a-wrong-protocol")

			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring(fmt.Sprintf("expected %s but got", containerdProtocolPrefix)))
		})

		It("should error on short protocol", func() {
			m := &test.MockClient{}
			c := ContainerdClient{client: m}
			_, err := c.GetLabelsFromContainerID(context.TODO(), "dock:")

			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring("is not a containerd container id"))
		})
	})
})
