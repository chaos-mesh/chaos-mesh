package chaosdaemon

import (
	"strings"

	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/pingcap/chaos-mesh/pkg/mock"
	"github.com/vishvananda/netlink"
)

func applyTbf(tbf *pb.Tbf, pid uint32) error {
	log.Info("apply tbf on PID", "pid", pid)

	// Mock point to return error in unit test
	if err := mock.On("TbfApplyError"); err != nil {
		if e, ok := err.(error); ok {
			return e
		}
		if ignore, ok := err.(bool); ok && ignore {
			return nil
		}
	}

	ns, handle, link, err := newNetlinkHandle(pid)

	if err != nil {
		return err
	}
	defer ns.Close()

	tbfQdisc := &netlink.Tbf{
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

	if err = handle.QdiscAdd(tbfQdisc); err != nil {
		if !strings.Contains(err.Error(), "file exists") {
			log.Error(err, "failed to add Qdisc", "qdisc", tbfQdisc)
			return err
		}
	}

	return nil
}

func deleteTbf(tbf *pb.Tbf, pid uint32) error {
	return nil
}
