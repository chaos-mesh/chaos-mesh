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

	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/actor"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/manager/sideeffect"
)

type ActorResolver struct {
	pg actor.Playground
}

func NewActorResolver(pg actor.Playground) *ActorResolver {
	return &ActorResolver{pg: pg}
}

func (it *ActorResolver) GetName() string {
	return "ActorResolver"
}

func (it *ActorResolver) ResolveSideEffect(sideEffect sideeffect.SideEffect) error {
	if createActorSideEffect, ok := sideEffect.(*sideeffect.CreateActorEventSideEffect); ok {
		return createActorSideEffect.GetActor().PlayOn(it.pg)
	}
	return fmt.Errorf("can not parse NotifyNewEventSideEffect, side effect %+v", sideEffect)
}

func (it *ActorResolver) CouldResolve() []sideeffect.SideEffectType {
	return []sideeffect.SideEffectType{sideeffect.CreateActor}
}
