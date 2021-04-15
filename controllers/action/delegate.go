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

package action

import (
	"context"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type Delegate struct {
	impl interface{}
}

func (i *Delegate) callAccordingToAction(action, methodName string, defaultPhase v1alpha1.Phase, args ...interface{}) (v1alpha1.Phase, error) {
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

	return defaultPhase, errors.Errorf("unknown action %s", action)
}

func (i *Delegate) getAction(obj v1alpha1.InnerObject) string {
	return reflect.ValueOf(obj).Elem().FieldByName("Spec").FieldByName("Action").String()
}

func (i *Delegate) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	return i.callAccordingToAction(i.getAction(obj), "Apply", v1alpha1.NotInjected, ctx, index, records, obj)
}

func (i *Delegate) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	return i.callAccordingToAction(i.getAction(obj), "Recover", v1alpha1.Injected, ctx, index, records, obj)
}

func New(impl interface{}) Delegate {
	return Delegate{
		impl,
	}
}
