package chaosdaemon

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/pingcap/chaos-mesh/pkg/mock"
)

var _ = Describe("netem server", func() {
	defer mock.With("MockContainerdClient", &MockClient{})()
	c, _ := CreateContainerRuntimeInfoClient(containerRuntimeContainerd)
	s := &daemonServer{c}

	Context("SetTbf", func() {
		It("should work", func() {
			const ignore = true
			defer mock.With("TbfApplyError", ignore)()
			_, err := s.SetTbf(context.TODO(), &pb.TbfRequest{
				ContainerId: "containerd://container-id",
			})
			Expect(err).To(BeNil())
		})

		It("should fail on get pid", func() {
			const errorStr = "mock error on Task()"
			defer mock.With("TaskError", errors.New(errorStr))()
			_, err := s.SetTbf(context.TODO(), &pb.TbfRequest{
				ContainerId: "containerd://container-id",
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring(errorStr))
		})

		It("should fail on Apply", func() {
			const errorStr = "mock error on Apply()"
			defer mock.With("TbfApplyError", errors.New(errorStr))()
			_, err := s.SetTbf(context.TODO(), &pb.TbfRequest{
				ContainerId: "containerd://container-id",
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring(errorStr))
		})
	})

	Context("DeleteTbf", func() {
		It("should work", func() {
			const ignore = true
			defer mock.With("TbfDeleteError", ignore)()
			_, err := s.DeleteTbf(context.TODO(), &pb.TbfRequest{
				ContainerId: "containerd://container-id",
			})
			Expect(err).To(BeNil())
		})

		It("should fail on get pid", func() {
			const errorStr = "mock error on Task()"
			defer mock.With("TaskError", errors.New(errorStr))()
			_, err := s.DeleteTbf(context.TODO(), &pb.TbfRequest{
				ContainerId: "containerd://container-id",
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring(errorStr))
		})

		It("should fail on Apply", func() {
			const errorStr = "mock error on Apply()"
			defer mock.With("TbfDeleteError", errors.New(errorStr))()
			_, err := s.DeleteTbf(context.TODO(), &pb.TbfRequest{
				ContainerId: "containerd://container-id",
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring(errorStr))
		})
	})
})
