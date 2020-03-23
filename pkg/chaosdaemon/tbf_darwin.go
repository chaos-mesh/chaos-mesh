package chaosdaemon

import (
	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/pingcap/chaos-mesh/pkg/mock"
)

func applyTbf(tbf *pb.Tbf, pid uint32) error {
	// Mock point to return error in unit test
	if err := mock.On("TbfApplyError"); err != nil {
		if e, ok := err.(error); ok {
			return e
		}
		if ignore, ok := err.(bool); ok && ignore {
			return nil
		}
	}

	panic("unimplemented")
}

func deleteTbf(tbf *pb.Tbf, pid uint32) error {
	// Mock point to return error in unit test
	if err := mock.On("TbfDeleteError"); err != nil {
		if e, ok := err.(error); ok {
			return e
		}
		if ignore, ok := err.(bool); ok && ignore {
			return nil
		}
	}

	panic("unimplemented")
}
