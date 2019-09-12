// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package podchaos

import (
	"github.com/cwen0/chaos-operator/pkg/apis/pingcap.com/v1alpha1"
	"github.com/cwen0/chaos-operator/pkg/manager"
)

type podChaosManager struct {
	base manager.ManagerBaseInterface
}

func NewPodChaosManager() *podChaosManager {
	return &podChaosManager{}
}

// Sync syncs the PodChaos resource to manager.
func (m *podChaosManager) Sync(pc *v1alpha1.PodChaos) error {
	runner := m.newRunner(pc)
	m.base.AddRunner(runner)
	return nil
}

func (m *podChaosManager) newRunner(pc *v1alpha1.PodChaos) manager.Runner {
	return manager.Runner{
		Job: podKillJob{},
	}
}
