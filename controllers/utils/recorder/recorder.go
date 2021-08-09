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
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	"github.com/iancoleman/strcase"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record/util"
	ref "k8s.io/client-go/tools/reference"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ChaosRecorder interface {
	Event(object runtime.Object, ev ChaosEvent)
}

type chaosRecorder struct {
	log    logr.Logger
	source v1.EventSource
	client client.Client
	scheme *runtime.Scheme
}

func (r *chaosRecorder) Event(object runtime.Object, ev ChaosEvent) {
	eventtype := ev.Type()
	reason := ev.Reason()
	message := ev.Message()

	annotations, err := generateAnnotations(ev)
	if err != nil {
		r.log.Error(err, "failed to generate annotations for event", "event", ev)
	}

	ref, err := ref.GetReference(r.scheme, object)
	if err != nil {
		r.log.Error(err, "fail to construct reference", "object", object)
		return
	}

	if !util.ValidateEventType(eventtype) {
		klog.Errorf("Unsupported event type: '%v'", eventtype)
		return
	}

	event := r.makeEvent(ref, annotations, eventtype, reason, message)
	event.Source = r.source
	go func() {
		err := r.client.Create(context.TODO(), event)
		if err != nil {
			r.log.Error(err, "fail to submit event", "event", event)
		}
	}()
}

func (r *chaosRecorder) makeEvent(ref *v1.ObjectReference, annotations map[string]string, eventtype, reason, message string) *v1.Event {
	t := metav1.Time{Time: time.Now()}
	namespace := ref.Namespace
	if namespace == "" {
		namespace = metav1.NamespaceDefault
	}
	return &v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%v.%x", ref.Name, t.UnixNano()),
			Namespace:   namespace,
			Annotations: annotations,
		},
		InvolvedObject: *ref,
		Reason:         reason,
		Message:        message,
		FirstTimestamp: t,
		LastTimestamp:  t,
		Count:          1,
		Type:           eventtype,
	}
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

type RecorderBuilder struct {
	c      client.Client
	logger logr.Logger
	scheme *runtime.Scheme
}

func (b *RecorderBuilder) Build(name string) ChaosRecorder {
	return &chaosRecorder{
		log: b.logger.WithName("event-recorder-" + name),
		source: v1.EventSource{
			Component: name,
		},
		client: b.c,
		scheme: b.scheme,
	}
}

func NewRecorderBuilder(c client.Client, logger logr.Logger, scheme *runtime.Scheme) *RecorderBuilder {
	return &RecorderBuilder{
		c,
		logger,
		scheme,
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
