package tcdaemon

import (
	"github.com/vishvananda/netlink"

	pb "github.com/pingcap/chaos-operator/pkg/tcdaemon/pb"
)

func ToNetlinkNetemAttrs(netem *pb.Netem) netlink.NetemQdiscAttrs {
	return netlink.NetemQdiscAttrs{
		Latency:       netem.Time,
		DelayCorr:     netem.DelayCorr,
		Limit:         netem.Limit,
		Loss:          netem.Loss,
		LossCorr:      netem.LossCorr,
		Gap:           netem.Gap,
		Duplicate:     netem.Duplicate,
		DuplicateCorr: netem.DuplicateCorr,
		Jitter:        netem.Jitter,
		ReorderProb:   netem.Reorder,
		ReorderCorr:   netem.ReorderCorr,
		CorruptProb:   netem.Corrupt,
		CorruptCorr:   netem.CorruptCorr,
	}
}
