package utils

import (
	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	"github.com/retailnext/iptables_exporter/iptables"
)

func CollectIptablesMetrics(pid uint32) (chains uint64, rules uint64, packets uint64, packetBytes uint64, err error) {
	cmd := bpm.DefaultProcessBuilder("iptables-save", "-c").
		SetNS(pid, bpm.NetNS).
		Build()

	stdout := bpm.NewBlockingBuffer()
	cmd.Stdout = stdout
	manager := bpm.NewBackgroundProcessManager()
	_, err = manager.StartProcess(cmd)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	tables, err := iptables.ParseIptablesSave(stdout)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	for _, table := range tables {
		for _, chain := range table {
			chains++
			for _, r := range chain.Rules {
				rules++
				packets += r.Packets
				packetBytes += r.Bytes
			}
		}
	}

	return chains, rules, packets, packetBytes, nil
}
