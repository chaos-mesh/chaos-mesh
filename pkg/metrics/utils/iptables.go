// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package utils

import (
	"github.com/retailnext/iptables_exporter/iptables"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
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
