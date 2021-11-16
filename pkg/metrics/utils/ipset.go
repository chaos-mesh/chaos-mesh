package utils

import (
	"encoding/xml"

	"github.com/romana/ipset"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
)

func CollectIPSetMembersMetric(pid uint32) (int, error) {
	cmd := bpm.DefaultProcessBuilder("ipset", "save", "-o", "xml").
		SetNS(pid, bpm.NetNS).
		Build()

	stdout := bpm.NewBlockingBuffer()
	cmd.Stdout = stdout
	manager := bpm.NewBackgroundProcessManager()
	_, err := manager.StartProcess(cmd)
	if err != nil {
		return 0, err
	}

	var sets ipset.Ipset
	if err = xml.NewDecoder(stdout).Decode(&sets); err != nil {
		return 0, err
	}

	var members int
	for _, set := range sets.Sets {
		members += len(set.Members)
	}
	return members, nil
}
