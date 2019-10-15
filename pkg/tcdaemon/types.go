package tcdaemon

import "github.com/vishvananda/netlink"

// Netem represents a netem qdisc
type Netem struct {
	Time             uint32  `json:"time"`
	Jitter           uint32  `json:"jitter"`
	DelayCorrelation float32 `json:"delayCorr"`
	Limit            uint32  `json:"limit"`
	Loss             float32 `json:"loss"`
	LossCorr         float32 `json:"lossCorr"`
	Gap              uint32  `json:"gap"`
	Duplicate        float32 `json:"duplicate"`
	DuplicateCorr    float32 `json:"duplicateCorr"`
	ReorderProb      float32 `json:"reorder"`
	ReorderCorr      float32 `json:"reorderCorr"`
	CorruptProb      float32 `json:"corrupt"`
	CorruptCorr      float32 `json:"corruptCorr"`
}

func (netem *Netem) getNetlinkNetemAttrs() netlink.NetemQdiscAttrs {
	return netlink.NetemQdiscAttrs{
		Latency:       netem.Time,
		DelayCorr:     netem.DelayCorrelation,
		Limit:         netem.Limit,
		Loss:          netem.Loss,
		LossCorr:      netem.LossCorr,
		Gap:           netem.Gap,
		Duplicate:     netem.Duplicate,
		DuplicateCorr: netem.DuplicateCorr,
		Jitter:        netem.Jitter,
		ReorderProb:   netem.ReorderProb,
		ReorderCorr:   netem.ReorderCorr,
		CorruptProb:   netem.CorruptProb,
		CorruptCorr:   netem.CorruptCorr,
	}
}
