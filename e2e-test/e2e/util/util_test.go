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

package util

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func TestChaosConditionsTrue(t *testing.T) {
	testCases := []struct {
		name     string
		status   *v1alpha1.ChaosStatus
		expected []v1alpha1.ChaosConditionType
		want     bool
	}{
		{
			name: "all expected conditions are true",
			status: &v1alpha1.ChaosStatus{Conditions: []v1alpha1.ChaosCondition{
				{Type: v1alpha1.ConditionSelected, Status: corev1.ConditionTrue},
				{Type: v1alpha1.ConditionAllInjected, Status: corev1.ConditionTrue},
			}},
			expected: []v1alpha1.ChaosConditionType{v1alpha1.ConditionSelected, v1alpha1.ConditionAllInjected},
			want:     true,
		},
		{
			name: "expected condition is absent",
			status: &v1alpha1.ChaosStatus{Conditions: []v1alpha1.ChaosCondition{
				{Type: v1alpha1.ConditionSelected, Status: corev1.ConditionTrue},
			}},
			expected: []v1alpha1.ChaosConditionType{v1alpha1.ConditionSelected, v1alpha1.ConditionAllInjected},
			want:     false,
		},
		{
			name: "expected condition is not true",
			status: &v1alpha1.ChaosStatus{Conditions: []v1alpha1.ChaosCondition{
				{Type: v1alpha1.ConditionSelected, Status: corev1.ConditionFalse},
			}},
			expected: []v1alpha1.ChaosConditionType{v1alpha1.ConditionSelected},
			want:     false,
		},
		{
			name:     "nil status",
			status:   nil,
			expected: []v1alpha1.ChaosConditionType{v1alpha1.ConditionSelected},
			want:     false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := ChaosConditionsTrue(testCase.status, testCase.expected...)
			if got != testCase.want {
				t.Errorf("ChaosConditionsTrue() = %t, want %t", got, testCase.want)
			}
		})
	}
}
