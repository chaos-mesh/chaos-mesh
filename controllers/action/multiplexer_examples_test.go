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

package action

import (
	"context"
	"fmt"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type chaosImplForAction1 struct {
}

func (it *chaosImplForAction1) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	fmt.Println("action1-apply")
	return v1alpha1.Injected, nil
}

func (it *chaosImplForAction1) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	fmt.Println("action1-recover")
	return v1alpha1.NotInjected, nil
}

type chaosImplForAction2 struct {
}

func (it *chaosImplForAction2) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	fmt.Println("action2-apply")
	return v1alpha1.Injected, nil
}

func (it *chaosImplForAction2) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	fmt.Println("action2-recover")
	return v1alpha1.NotInjected, nil
}

func ExampleMultiplexer() {
	type adHoc struct {
		AnyName1        *chaosImplForAction1 `action:"struct-tag"`
		WhateverTheName *chaosImplForAction2 `action:"is-important"`
	}
	multiplexer := NewMultiplexer(&adHoc{
		AnyName1:        &chaosImplForAction1{},
		WhateverTheName: &chaosImplForAction2{},
	})

	// Just use PodChaos as example, you could use any struct that contains Spec.Action for it.
	chaosA := v1alpha1.PodChaos{
		Spec: v1alpha1.PodChaosSpec{
			Action: "struct-tag",
		},
	}
	chaosB := v1alpha1.PodChaos{
		Spec: v1alpha1.PodChaosSpec{
			Action: "is-important",
		},
	}

	if _, err := multiplexer.Apply(context.Background(), 0, nil, &chaosA); err != nil {
		panic(err)
	}
	if _, err := multiplexer.Recover(context.Background(), 0, nil, &chaosA); err != nil {
		panic(err)
	}
	if _, err := multiplexer.Apply(context.Background(), 0, nil, &chaosB); err != nil {
		panic(err)
	}
	if _, err := multiplexer.Recover(context.Background(), 0, nil, &chaosB); err != nil {
		panic(err)
	}

	// Output: action1-apply
	// action1-recover
	// action2-apply
	// action2-recover
}
