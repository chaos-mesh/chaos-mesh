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

// Package action introduces a multiplexer for actions.
package action

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	impltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
)

// TODO: refactor this as a map[name]impltypes.ChaosImpl style, remove reflect usage.

var _ impltypes.ChaosImpl = (*Multiplexer)(nil)

// Multiplexer could combine ChaosImpl implementations into one, and route them by Action in the ChaosSpec.
// Field impl should be a struct which contains several fields with struct tag "action", each field should be an implementation of ChaosImpl.
// For example:
//
//	type tempStruct struct {
//	  Impl1 impltypes.ChaosImpl `action:"action1"`
//	  Impl2 impltypes.ChaosImpl `action:"action2"`
//	}
//
// is valid to be the field in Multiplexer.
//
// Because we use reflect fo iterate fields in tempStruct, so fields in tempStruct should be public/exported.
//
// When some Chaos like:
//
//	type SomeChaos struct {
//	  ***
//	  Spec SomeChaosSpec `json:"spec"`
//	  ***
//	}
//	type SomeChaosSpec struct {
//	  ***
//	  // available actions: action1, action2
//	  Action string `json:"action"`
//	  ***
//	}
//
// is created, the corresponding ChaosImpl(s) for each action will be invoked by struct tag.
type Multiplexer struct {
	impl interface{}
}

func (i *Multiplexer) callAccordingToAction(action, methodName string, defaultPhase v1alpha1.Phase, args ...interface{}) (v1alpha1.Phase, error) {
	implType := reflect.TypeOf(i.impl).Elem()
	implVal := reflect.ValueOf(i.impl)

	reflectArgs := []reflect.Value{}
	for _, arg := range args {
		reflectArgs = append(reflectArgs, reflect.ValueOf(arg))
	}
	for i := 0; i < implType.NumField(); i++ {
		field := implType.Field(i)

		actions := strings.Split(field.Tag.Get("action"), ",")
		for i := range actions {
			if actions[i] == action {
				rets := implVal.Elem().FieldByIndex(field.Index).MethodByName(methodName).Call(reflectArgs)

				// nil.(error) will panic :(
				err := rets[1].Interface()
				if err == nil {
					return rets[0].Interface().(v1alpha1.Phase), nil
				}

				return rets[0].Interface().(v1alpha1.Phase), err.(error)
			}
		}
	}
	var gvk = ""
	if obj, ok := args[len(args)-1].(runtime.Object); ok {
		gvk = obj.GetObjectKind().GroupVersionKind().String()
	}
	return defaultPhase, NewErrorUnknownAction(gvk, action)
}

type ErrorUnknownAction struct {
	GroupVersionKind string
	Action           string
}

func NewErrorUnknownAction(GVK string, action string) ErrorUnknownAction {
	return ErrorUnknownAction{GroupVersionKind: GVK, Action: action}
}

func (it ErrorUnknownAction) Error() string {
	return fmt.Sprintf("unknown action: Action: %s, GroupVersionKind: %s", it.Action, it.GroupVersionKind)
}

// TODO: refactor this by introduce a new interface called ContainsAction
func (i *Multiplexer) getAction(obj v1alpha1.InnerObject) string {
	return reflect.ValueOf(obj).Elem().FieldByName("Spec").FieldByName("Action").String()
}

func (i *Multiplexer) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	return i.callAccordingToAction(i.getAction(obj), "Apply", v1alpha1.NotInjected, ctx, index, records, obj)
}

func (i *Multiplexer) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	return i.callAccordingToAction(i.getAction(obj), "Recover", v1alpha1.Injected, ctx, index, records, obj)
}

// NewMultiplexer is a constructor of Multiplexer.
// For the detail of the parameter "impl", see the comment of type Multiplexer.
func NewMultiplexer(impl interface{}) Multiplexer {
	return Multiplexer{
		impl,
	}
}
