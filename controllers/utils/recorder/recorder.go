// Copyright 2021 Chaos Mesh Authors.
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

package recorder

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
)

type ChaosRecorder struct {
	recorder record.EventRecorder
}

func (r *ChaosRecorder) Event(object runtime.Object, ev ChaosEvent) {
	r.recorder.Event(object, ev.Type(), ev.Reason(), ev.Message())
}

type ChaosEvent interface {
	Type() string
	Reason() string
	Message() string

	Parse(message string) ChaosEvent
}

var allEvents = []ChaosEvent{}

func register(ev ...ChaosEvent) {
	allEvents = append(allEvents, ev...)
}

func Parse(message string) ChaosEvent {
	for _, ev := range allEvents {
		ev := ev.Parse(message)
		if ev != nil {
			return ev
		}
	}
	return nil
}

func NewRecorder(mgr ctrl.Manager, name string) ChaosRecorder {
	return ChaosRecorder{
		mgr.GetEventRecorderFor(name),
	}
}
