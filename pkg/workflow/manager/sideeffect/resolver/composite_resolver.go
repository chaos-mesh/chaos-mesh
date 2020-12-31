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
)

type CompositeResolver struct {
	backends     map[sideeffect.SideEffectType]SideEffectsResolver
	couldResolve []sideeffect.SideEffectType
}

func NewCompositeResolver(backends map[sideeffect.SideEffectType]SideEffectsResolver, couldResolve []sideeffect.SideEffectType) *CompositeResolver {
	return &CompositeResolver{backends: backends, couldResolve: couldResolve}
}

// Notice that overlapped resolve type is not allowed.
func NewCompositeResolverWith(backends ...SideEffectsResolver) (*CompositeResolver, error) {
	var backendsMap = make(map[sideeffect.SideEffectType]SideEffectsResolver)
	var couldResolve []sideeffect.SideEffectType

	for _, backend := range backends {
		effectTypes := backend.CouldResolve()
		for _, effectType := range effectTypes {
			if alreadyExistsBackend, exists := backendsMap[effectType]; exists {
				return nil, fmt.Errorf("can not add resovler %s for type %s, already have another resolver %s", backend.GetName(), effectType, alreadyExistsBackend.GetName())
			}

			backendsMap[effectType] = backend
			couldResolve = append(couldResolve, effectType)
		}
	}

	return NewCompositeResolver(backendsMap, couldResolve), nil
}

func (it *CompositeResolver) GetName() string {
	return "CompositeResolver"
}

func (it *CompositeResolver) ResolveSideEffect(sideEffect sideeffect.SideEffect) error {
	if resolver, ok := it.backends[sideEffect.GetSideEffectType()]; ok {
		return resolver.ResolveSideEffect(sideEffect)
	}

	return fmt.Errorf("composite resolver could not resolve side effect %+v", sideEffect)
}

func (it *CompositeResolver) CouldResolve() []sideeffect.SideEffectType {
	return it.couldResolve
}
