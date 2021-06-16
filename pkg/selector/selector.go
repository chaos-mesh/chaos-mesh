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
	"reflect"

	"github.com/pkg/errors"
	"go.uber.org/fx"

	"github.com/chaos-mesh/chaos-mesh/pkg/selector/aws"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/container"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/gcp"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/pod"
)

type Selector struct {
	selectorMap map[reflect.Type]interface{}
}

type Target interface {
	Id() string
}

func (s *Selector) Select(ctx context.Context, spec interface{}) ([]Target, error) {
	if spec == nil {
		return []Target{}, nil
	}

	var targets []Target
	impl, ok := s.selectorMap[reflect.TypeOf(spec)]
	if ok {
		vals := reflect.ValueOf(impl).MethodByName("Select").Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(spec)})
		ret := vals[0]

		for i := 0; i < ret.Len(); i++ {
			targets = append(targets, ret.Index(i).Interface().(Target))
		}

		err := vals[1].Interface()
		if err == nil {
			return targets, nil
		}

		return targets, err.(error)
	}

	return nil, errors.Errorf("specification type not found: %s", reflect.TypeOf(spec))
}

type SelectorParams struct {
	fx.In

	PodSelector       *pod.SelectImpl
	ContainerSelector *container.SelectImpl
	AwsSelector       *aws.SelectImpl
	GcpSelector       *gcp.SelectImpl
}

func New(p SelectorParams) *Selector {
	selectorMap := make(map[reflect.Type]interface{})

	val := reflect.ValueOf(p)
	for i := 0; i < val.NumField(); i++ {
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

var Module = fx.Provide(
	New,

	pod.New,
	container.New,
	aws.New,
	gcp.New,
)
