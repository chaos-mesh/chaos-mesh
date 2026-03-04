// Copyright 2022 Chaos Mesh Authors.
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

package remotecluster

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func TestHelmLifecycleManaged(t *testing.T) {
	RegisterTestingT(t)

	tests := []struct {
		name        string
		annotations map[string]string
		expected    bool
	}{
		{
			name:        "nil annotations defaults to managed",
			annotations: nil,
			expected:    true,
		},
		{
			name:        "empty annotations defaults to managed",
			annotations: map[string]string{},
			expected:    true,
		},
		{
			name: "annotation absent defaults to managed",
			annotations: map[string]string{
				"some-other-annotation": "value",
			},
			expected: true,
		},
		{
			name: "annotation set to true means managed",
			annotations: map[string]string{
				v1alpha1.AnnotationManagedHelmLifecycle: "true",
			},
			expected: true,
		},
		{
			name: "annotation set to false means not managed",
			annotations: map[string]string{
				v1alpha1.AnnotationManagedHelmLifecycle: "false",
			},
			expected: false,
		},
		{
			name: "annotation set to empty string means managed",
			annotations: map[string]string{
				v1alpha1.AnnotationManagedHelmLifecycle: "",
			},
			expected: true,
		},
		{
			name: "annotation set to arbitrary value means managed",
			annotations: map[string]string{
				v1alpha1.AnnotationManagedHelmLifecycle: "yes",
			},
			expected: true,
		},
		{
			name: "annotation set to False (capitalized) means managed",
			annotations: map[string]string{
				v1alpha1.AnnotationManagedHelmLifecycle: "False",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterTestingT(t)
			obj := &v1alpha1.RemoteCluster{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tt.annotations,
				},
			}
			Expect(helmLifecycleManaged(obj)).To(Equal(tt.expected))
		})
	}
}
