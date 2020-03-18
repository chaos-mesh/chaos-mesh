package chaosdaemon

import (
	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/pingcap/chaos-mesh/pkg/mock"
	"github.com/vishvananda/netlink"
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

	return applyQdisc(pid, func(handle *netlink.Handle, link netlink.Link) netlink.Qdisc {
		return &netlink.Tbf{
			QdiscAttrs: netlink.QdiscAttrs{
				LinkIndex: link.Attrs().Index,
				Handle:    netlink.MakeHandle(1, 0),
				Parent:    netlink.HANDLE_ROOT,
			},
			Rate:     tbf.Rate,
			Limit:    tbf.Limit,
			Buffer:   tbf.Buffer,
			Peakrate: tbf.PeakRate,
			Minburst: tbf.MinBurst,
		}
	})
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

	return deleteQdisc(pid, func(handle *netlink.Handle, link netlink.Link) netlink.Qdisc {
		return &netlink.Tbf{
			QdiscAttrs: netlink.QdiscAttrs{
				LinkIndex: link.Attrs().Index,
				Handle:    netlink.MakeHandle(1, 0),
				Parent:    netlink.HANDLE_ROOT,
			},
			Rate:     tbf.Rate,
			Limit:    tbf.Limit,
			Buffer:   tbf.Buffer,
			Peakrate: tbf.PeakRate,
			Minburst: tbf.MinBurst,
		}
	})
}
