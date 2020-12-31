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

type CreateNewNodeResolver struct {
	repo workflowrepo.WorkflowRepo
}

func NewCreateNewNodeResolver(repo workflowrepo.WorkflowRepo) *CreateNewNodeResolver {
	return &CreateNewNodeResolver{repo: repo}
}

func (it *CreateNewNodeResolver) GetName() string {
	return "CreateNewNodeResolver"
}

func (it *CreateNewNodeResolver) ResolveSideEffect(sideEffect sideeffect.SideEffect) error {
	if createNewNodeSideEffect, ok := sideEffect.(*sideeffect.CreateNewNodeSideEffect); ok {
		return it.repo.CreateNodes(createNewNodeSideEffect.WorkflowName, createNewNodeSideEffect.ParentNodeName, createNewNodeSideEffect.NodeName, createNewNodeSideEffect.TemplateName)
	}
	return fmt.Errorf("can not parse NotifyNewEventSideEffect, side effect %+v", sideEffect)

}

func (it *CreateNewNodeResolver) CouldResolve() []sideeffect.SideEffectType {
	return []sideeffect.SideEffectType{sideeffect.CreateNewNode}
}
