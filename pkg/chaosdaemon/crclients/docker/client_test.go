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

package docker

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

func TestDockerClient(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Docker Container Client Test Suit")
}

var _ = Describe("docker client", func() {
	Context("DockerClient GetPidFromContainerID", func() {
		It("should return the magic number 9527", func() {
			defer mock.With("pid", int(9527))()

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
			defer mock.With("ContainerInspectError", errors.New(errorStr))()
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
			defer mock.With("ContainerKillError", errors.New(errorStr))()
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

	Context("DockerClient ListContainerIDs", func() {
		It("should work", func() {
			containerID := "valid-container-id"
			containerIDWithPrefix := fmt.Sprintf("%s%s", dockerProtocolPrefix, containerID)
			defer mock.With("containerID", containerID)()

			m := &test.MockClient{}
			c := DockerClient{client: m}
			containerIDs, err := c.ListContainerIDs(context.Background())

			Expect(err).To(BeNil())
			Expect(containerIDs).To(Equal([]string{containerIDWithPrefix}))
		})
	})

	Context("DockerClient GetLabelsFromContainerID", func() {
		It("should work", func() {
			sampleLabels := map[string]string{
				"io.kubernetes.pod.namespace":  "default",
				"io.kubernetes.pod.name":       "busybox-5f8dd756dd-6rjzw",
				"io.kubernetes.container.name": "busybox",
			}
			defer mock.With("labels", sampleLabels)()

			m := &test.MockClient{}
			c := DockerClient{client: m}
			labels, err := c.GetLabelsFromContainerID(context.Background(), "docker://valid-container-id")

			Expect(err).To(BeNil())
			Expect(labels).To(Equal(sampleLabels))
		})

		It("should error on wrong protocol", func() {
			sampleLabels := map[string]string{
				"io.kubernetes.pod.namespace":  "default",
				"io.kubernetes.pod.name":       "busybox-5f8dd756dd-6rjzw",
				"io.kubernetes.container.name": "busybox",
			}
			defer mock.With("labels", sampleLabels)()

			m := &test.MockClient{}
			c := DockerClient{client: m}
			_, err := c.GetLabelsFromContainerID(context.Background(), "containerd://this-is-a-wrong-protocol")

			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring(fmt.Sprintf("expected %s but got", dockerProtocolPrefix)))
		})

		It("should error on short protocol", func() {
			m := &test.MockClient{}
			c := DockerClient{client: m}
			_, err := c.GetLabelsFromContainerID(context.TODO(), "dock:")

			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring("is not a docker container id"))
		})
	})
})
