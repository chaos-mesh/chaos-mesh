// Copyright 2020 Chaos Mesh Authors.
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

package resolver

import (
	"fmt"

	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/manager/sideeffect"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/workflowrepo"
)

type UpdateNodePhaseResolver struct {
	repo workflowrepo.WorkflowRepo
}

func NewUpdateNodePhaseResolver(repo workflowrepo.WorkflowRepo) *UpdateNodePhaseResolver {
	return &UpdateNodePhaseResolver{repo: repo}
}

func (it *UpdateNodePhaseResolver) GetName() string {
	return "UpdateNodePhaseResolver"
}

func (it *UpdateNodePhaseResolver) ResolveSideEffect(sideEffect sideeffect.SideEffect) error {
	if updateNodePhaseSideEffect, ok := sideEffect.(*sideeffect.UpdateNodePhaseSideEffect); ok {
		return it.repo.UpdateNodePhase(
			updateNodePhaseSideEffect.Namespace,
			updateNodePhaseSideEffect.WorkflowName,
			updateNodePhaseSideEffect.NodeName,
			updateNodePhaseSideEffect.TargetPhase,
		)
	}
	return fmt.Errorf("can not parse NotifyNewEventSideEffect, side effect %+v", sideEffect)

}

func (it *UpdateNodePhaseResolver) CouldResolve() []sideeffect.SideEffectType {
	return []sideeffect.SideEffectType{sideeffect.UpdateNodePhase}
}
