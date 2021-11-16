package utils

import (
	"bufio"
	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
)

func CollectTcRulesMetric(pid uint32) (int, error) {
	cmd := bpm.DefaultProcessBuilder("tc", "qdisc").
		SetNS(pid, bpm.NetNS).
		Build()

	stdout := bpm.NewBlockingBuffer()
	cmd.Stdout = stdout
	manager := bpm.NewBackgroundProcessManager()
	_, err := manager.StartProcess(cmd)
	if err != nil {
		return 0, err
	}

	var lines int
	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		lines++
	}

	return lines, nil
}
