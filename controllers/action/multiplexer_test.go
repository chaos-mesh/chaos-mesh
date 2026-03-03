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
	"testing"

	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type mockApplyError struct {
}

func (it *mockApplyError) Error() string {
	return "mock apply error"
}

type mockRecoverError struct {
}

func (it mockRecoverError) Error() string {
	return "mock recover error"
}

type chaosImplMustFailed struct {
}

func (it *chaosImplMustFailed) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	return v1alpha1.NotInjected, &mockApplyError{}

}

func (it *chaosImplMustFailed) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	return v1alpha1.Injected, mockRecoverError{}
}

func TestMultiplexer_passthroughsError(t *testing.T) {
	type adHoc struct {
		Backend *chaosImplMustFailed `action:"must-failed"`
	}
	multiplexer := NewMultiplexer(&adHoc{
		Backend: &chaosImplMustFailed{},
	})

	// Just use PodChaos as example, you could use any struct that contains Spec.Action for it.
	chaos := v1alpha1.PodChaos{
		Spec: v1alpha1.PodChaosSpec{
			Action: "must-failed",
		},
	}
	_, err := multiplexer.Apply(context.Background(), 0, []*v1alpha1.Record{}, &chaos)
	applyError := &mockApplyError{}
	if !errors.As(err, &applyError) {
		t.Fatal("returned error is not mockApplyError")
	}

	_, err = multiplexer.Recover(context.Background(), 0, []*v1alpha1.Record{}, &chaos)
	recoverError := mockRecoverError{}
	if !errors.As(err, &recoverError) {
		t.Fatal("returned error is not recoverError")
	}
}

func TestMultiplexer_unhandledAction(t *testing.T) {
	type adHoc struct {
		Backend *chaosImplMustFailed `action:"must-failed"`
	}
	multiplexer := NewMultiplexer(&adHoc{
		// No fields here
	})
	chaos := v1alpha1.PodChaos{
		Spec: v1alpha1.PodChaosSpec{
			Action: "not-exist",
		},
	}
	_, err := multiplexer.Apply(context.Background(), 0, []*v1alpha1.Record{}, &chaos)
	unknownAction := ErrorUnknownAction{}
	if !errors.As(err, &unknownAction) {
		t.Fatal("should not return error")
	}

}
