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
	"reflect"

	"github.com/go-logr/logr"
	"github.com/iancoleman/strcase"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
)

type ChaosRecorder interface {
	Event(object runtime.Object, ev ChaosEvent)
}

type chaosRecorder struct {
	recorder record.EventRecorder
	log      logr.Logger
}

func (r *chaosRecorder) Event(object runtime.Object, ev ChaosEvent) {
	annotations, err := generateAnnotations(ev)
	if err != nil {
		r.log.Error(err, "failed to generate annotations for event", "event", ev)
	}

	r.recorder.AnnotatedEventf(object, annotations, ev.Type(), ev.Reason(), ev.Message())
}

type ChaosEvent interface {
	Type() string
	Reason() string
	Message() string
}

var allEvents = make(map[string]ChaosEvent)

func register(ev ...ChaosEvent) {
	for _, ev := range ev {
		val := reflect.ValueOf(ev)
		val = reflect.Indirect(val)

		allEvents[strcase.ToKebab(val.Type().Name())] = ev
	}
}

func NewRecorder(mgr ctrl.Manager, name string, logger logr.Logger) ChaosRecorder {
	return &chaosRecorder{
		mgr.GetEventRecorderFor(name),
		logger.WithName("event-recorder" + name),
	}
}

type debugRecorder struct {
	Events map[types.NamespacedName][]ChaosEvent
}

func (d *debugRecorder) Event(object runtime.Object, ev ChaosEvent) {
	obj := object.(metav1.Object)
	id := types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}

	if d.Events[id] == nil {
		d.Events[id] = []ChaosEvent{}
	}

	d.Events[id] = append(d.Events[id], ev)
}

func NewDebugRecorder() *debugRecorder {
	return &debugRecorder{
		Events: make(map[types.NamespacedName][]ChaosEvent),
	}
}
