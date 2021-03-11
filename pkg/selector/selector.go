// Copyright 2019 Chaos Mesh Authors.
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

package selector

import (
	"context"
	"errors"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/container"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/pod"
	"reflect"
)

type Selector struct {
	selectorMap map[reflect.Type]interface{}
}

func (s *Selector) Select(ctx context.Context, selector interface{}) ([]interface{}, error) {
	impl, ok := s.selectorMap[reflect.TypeOf(selector)]
	if ok {
		vals := reflect.ValueOf(impl).MethodByName("Select").Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(selector)})
		ret := vals[0].Interface().([]interface{})
		err := vals[1].Interface().(error)
		return ret, err
	}

	return nil, errors.New("specification type not found")
}

type SelectorParams struct {
	PodSelector *pod.SelectImpl
	ContainerSelector *container.SelectImpl
}

func NewSelector(p SelectorParams) *Selector {
	selectorMap := make(map[reflect.Type]interface{})

	val := reflect.ValueOf(p)
	for i:=0;i<=val.NumField();i++ {
		method := val.Field(i).MethodByName("Select")
		if method.IsValid() && method.Type().NumIn() > 1 {
			typ := method.Type().In(1)
			selectorMap[typ] = val.Field(i).Interface()
		}
	}

	return &Selector{
		selectorMap,
	}
}
