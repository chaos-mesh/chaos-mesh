package chaosdaemon

import (
	"context"
	"errors"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/docker/docker/api/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pingcap/chaos-mesh/pkg/mock"
)

type MockClient struct {
	MockPid int
}

func (m *MockClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	if err := mock.On("ContainerInspectError"); err != nil {
		return types.ContainerJSON{}, err.(error)
	}

	return types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			State: &types.ContainerState{
				Pid: m.MockPid,
			},
		},
	}, nil
}

func (m *MockClient) ContainerKill(ctx context.Context, containerID, signal string) error {
	if err := mock.On("containerKillError"); err != nil {
		return err.(error)
	}
	return nil
}

func (m *MockClient) LoadContainer(ctx context.Context, id string) (containerd.Container, error) {
	if err := mock.On("LoadContainerError"); err != nil {
		return nil, err.(error)
	}

	return &MockContainer{MockPid: m.MockPid}, nil
}

type MockContainer struct {
	containerd.Container
	MockPid int
	Errors  map[string]error
}

func (m *MockContainer) Task(context.Context, cio.Attach) (containerd.Task, error) {
	if err := mock.On("TaskError"); err != nil {
		return nil, err.(error)
	}

	return &MockTask{MockPid: m.MockPid}, nil
}

type MockTask struct {
	containerd.Task
	MockPid int
}

func (m *MockTask) Pid() uint32 {
	return uint32(m.MockPid)
}

var _ = Describe("chaosdaemon util", func() {
	Context("DockerClient GetPidFromContainerID", func() {
		It("should return the magic number 9527", func() {
			m := &MockClient{MockPid: 9527}
			c := DockerClient{client: m}
			pid, err := c.GetPidFromContainerID(context.TODO(), "docker://valid-container-id")
			Expect(err).To(BeNil())
			Expect(pid).To(Equal(uint32(9527)))
		})

		It("should error with wrong protocol", func() {
			m := &MockClient{}
			c := DockerClient{client: m}
			_, err := c.GetPidFromContainerID(context.TODO(), "containerd://this-is-a-wrong-protocol")
			Expect(err).NotTo(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring(fmt.Sprintf("expected %s but got", dockerProtocolPrefix)))
		})

		It("should error with specified string", func() {
			errorStr := "this is a mocked error"
			defer mock.With("ContainerInspectError", errors.New(errorStr))()
			m := &MockClient{}
			c := DockerClient{client: m}
			_, err := c.GetPidFromContainerID(context.TODO(), "docker://valid-container-id")
			Expect(err).NotTo(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(Equal(errorStr))
		})
	})

	Context("ContainerdClient GetPidFromContainerID", func() {
		It("should return the magic number 9527", func() {
			m := &MockClient{MockPid: 9527}
			c := ContainerdClient{client: m}
			pid, err := c.GetPidFromContainerID(context.TODO(), "containerd://valid-container-id")
			Expect(err).To(BeNil())
			Expect(pid).To(Equal(uint32(9527)))
		})

		It("should error with wrong protocol", func() {
			m := &MockClient{}
			c := ContainerdClient{client: m}
			_, err := c.GetPidFromContainerID(context.TODO(), "docker://this-is-a-wrong-protocol")
			Expect(err).NotTo(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring(fmt.Sprintf("expected %s but got", containerdProtocolPrefix)))
		})

		It("should error with specified string", func() {
			errorStr := "this is a mocked error"
			mock.With("LoadContainerError", errors.New(errorStr))
			m := &MockClient{}
			c := ContainerdClient{client: m}
			_, err := c.GetPidFromContainerID(context.TODO(), "containerd://valid-container-id")
			Expect(err).NotTo(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(Equal(errorStr))
			mock.Reset("LoadContainerError")

			mock.With("TaskError", errors.New(errorStr))
			m = &MockClient{}
			c = ContainerdClient{client: m}
			_, err = c.GetPidFromContainerID(context.TODO(), "containerd://valid-container-id")
			Expect(err).NotTo(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(Equal(errorStr))
			mock.Reset("TaskError")
		})
	})

	Context("CreateContainerRuntimeInfoClient", func() {
		It("should work", func() {
			_, err := CreateContainerRuntimeInfoClient(containerRuntimeDocker)
			Expect(err).To(BeNil())
			errorStr := "this is a mocked error"
			defer mock.With("CreateContainerRuntimeInfoClientError", errors.New(errorStr))()
			_, err = CreateContainerRuntimeInfoClient(containerRuntimeContainerd)
			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(Equal(errorStr))
		})
	})
})
