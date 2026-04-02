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

package collector

import (
	"context"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
	"github.com/chaos-mesh/chaos-mesh/pkg/testutils"
)

func TestSetUnarchivedExperiment(t *testing.T) {
	tests := []struct {
		name               string
		newGeneration      int64
		existingGeneration int64
		expectSetCalled    bool
		description        string
	}{
		{
			name:               "should skip update when generation is unchanged",
			newGeneration:      5,
			existingGeneration: 5,
			expectSetCalled:    false,
			description:        "ObservedGeneration matches, should return early without database write",
		},
		{
			name:               "should update when generation is different",
			newGeneration:      6,
			existingGeneration: 5,
			expectSetCalled:    true,
			description:        "ObservedGeneration differs, should update database",
		},
		{
			name:               "should update for new experiment",
			newGeneration:      1,
			existingGeneration: 0,
			expectSetCalled:    true,
			description:        "No existing record, should create new entry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockArchive := new(testutils.MockExperimentStore)
			mockEvent := new(testutils.MockEventStore)

			mockArchive.On("FindByUID", context.Background(), "test-uid").Return(&core.Experiment{
				ExperimentMeta: core.ExperimentMeta{
					UID:                "test-uid",
					Kind:               "PodChaos",
					Name:               "test-name",
					Namespace:          "test-ns",
					ObservedGeneration: tt.existingGeneration,
				},
			}, nil)

			if tt.expectSetCalled {
				mockArchive.On("Set", context.Background(), mock.Anything).Return(nil)
			}

			collector := &ChaosCollector{
				Log:     logr.Logger{},
				archive: mockArchive,
				event:   mockEvent,
			}

			podChaos := &v1alpha1.PodChaos{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "chaos-mesh.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test-name",
					Namespace:         "test-ns",
					UID:               "test-uid",
					Generation:        tt.newGeneration,
					CreationTimestamp: metav1.Time{Time: time.Now()},
				},
				Spec: v1alpha1.PodChaosSpec{
					Action: v1alpha1.PodKillAction,
				},
			}

			err := collector.setUnarchivedExperiment(ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-name",
					Namespace: "test-ns",
				},
			}, podChaos)
			assert.NoError(t, err)

			if tt.expectSetCalled {
				mockArchive.AssertCalled(t, "Set", context.Background(), mock.Anything)
			} else {
				mockArchive.AssertNotCalled(t, "Set")
			}
		})
	}
}

func TestConvertInnerObjectToExperiment(t *testing.T) {
	generation := int64(42)

	podChaos := &v1alpha1.PodChaos{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "chaos-mesh.org/v1alpha1",
			Kind:       "PodChaos",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-podchaos",
			Namespace:         "default",
			UID:               "test-uid-123",
			Generation:        generation,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.PodChaosSpec{
			Action: v1alpha1.PodKillAction,
		},
	}

	archive, err := convertInnerObjectToExperiment(podChaos)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if archive.ObservedGeneration != generation {
		t.Errorf("expected ObservedGeneration=%d, got %d", generation, archive.ObservedGeneration)
	}

	if archive.UID != "test-uid-123" {
		t.Errorf("expected UID=test-uid-123, got %s", archive.UID)
	}

	if archive.Kind != "PodChaos" {
		t.Errorf("expected Kind=PodChaos, got %s", archive.Kind)
	}

	if archive.Action != string(v1alpha1.PodKillAction) {
		t.Errorf("expected Action=pod-kill, got %s", archive.Action)
	}
}
