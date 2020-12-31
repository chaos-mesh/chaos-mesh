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

	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/trigger"
)

type NotifyNewEventResolver struct {
	operableTrigger trigger.OperableTrigger
}

func NewNotifyNewEventResolver(operableTrigger trigger.OperableTrigger) *NotifyNewEventResolver {
	return &NotifyNewEventResolver{operableTrigger: operableTrigger}
}

func (it *NotifyNewEventResolver) GetName() string {
	return "NotifyNewEventResolver"
}

func (it *NotifyNewEventResolver) ResolveSideEffect(sideEffect sideeffect.SideEffect) error {
	if notifyNewEventSideEffect, ok := sideEffect.(*sideeffect.NotifyNewEventSideEffect); ok {
		if notifyNewEventSideEffect.Delay > 0 {
			return it.operableTrigger.NotifyDelay(notifyNewEventSideEffect.NewEvent, notifyNewEventSideEffect.Delay)
		}
		return it.operableTrigger.Notify(notifyNewEventSideEffect.NewEvent)
	}
	return fmt.Errorf("can not parse NotifyNewEventSideEffect, side effect %+v", sideEffect)

}

func (it *NotifyNewEventResolver) CouldResolve() []sideeffect.SideEffectType {
	return []sideeffect.SideEffectType{sideeffect.NotifyNewEvent}
}
