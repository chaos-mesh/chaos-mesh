// Copyright 2025 Chaos Mesh Authors.
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

package core

import (
	"reflect"
	"testing"
)

func TestFilter_ConstructQueryArgs(t *testing.T) {
	cases := []struct {
		name          string
		filter        Filter
		expectedQuery string
		expectedArgs  []string
	}{
		{
			name: "empty filter",
			filter: Filter{
				Start: zeroTime,
				End:   zeroTime,
			},
			expectedQuery: "",
			expectedArgs:  []string{},
		},
		{
			name: "single namespace",
			filter: Filter{
				Namespace: "test-ns",
				Start:     zeroTime,
				End:       zeroTime,
			},
			expectedQuery: "namespace IN ( ? )",
			expectedArgs:  []string{"test-ns"},
		},
		{
			name: "comma-separated namespaces",
			filter: Filter{
				Namespace: "test-ns,prod-ns,staging-ns",
				Start:     zeroTime,
				End:       zeroTime,
			},
			expectedQuery: "namespace IN ( ?,?,? )",
			expectedArgs:  []string{"test-ns", "prod-ns", "staging-ns"},
		},
		{
			name: "comma-separated namespaces with spaces",
			filter: Filter{
				Namespace: "test-ns, prod-ns, staging-ns",
				Start:     zeroTime,
				End:       zeroTime,
			},
			expectedQuery: "namespace IN ( ?,?,? )",
			expectedArgs:  []string{"test-ns", "prod-ns", "staging-ns"},
		},
		{
			name: "namespace with name",
			filter: Filter{
				Namespace: "test-ns",
				Name:      "test-name",
				Start:     zeroTime,
				End:       zeroTime,
			},
			expectedQuery: "namespace IN ( ? ) AND name = ?",
			expectedArgs:  []string{"test-ns", "test-name"},
		},
		{
			name: "multiple namespaces with name",
			filter: Filter{
				Namespace: "test-ns,prod-ns",
				Name:      "test-name",
				Start:     zeroTime,
				End:       zeroTime,
			},
			// Note: Order may vary due to map iteration, but should have 3 placeholders total (2 for namespaces, 1 for name)
			expectedQuery: "namespace IN ( ?,? ) AND name = ?",
			expectedArgs:  []string{"test-ns", "prod-ns", "test-name"},
		},
		{
			name: "namespace with kind",
			filter: Filter{
				Namespace: "test-ns",
				Kind:      "PodChaos",
				Start:     zeroTime,
				End:       zeroTime,
			},
			expectedQuery: "namespace IN ( ? ) AND kind = ?",
			expectedArgs:  []string{"test-ns", "PodChaos"},
		},
		{
			name: "namespace with object_id",
			filter: Filter{
				Namespace: "test-ns",
				ObjectID:  "test-uid",
				Start:     zeroTime,
				End:       zeroTime,
			},
			expectedQuery: "namespace IN ( ? ) AND object_id = ?",
			expectedArgs:  []string{"test-ns", "test-uid"},
		},
		{
			name: "multiple namespaces with all fields",
			filter: Filter{
				Namespace: "test-ns,prod-ns,staging-ns",
				Name:      "test-name",
				Kind:      "PodChaos",
				ObjectID:  "test-uid",
				Start:     zeroTime,
				End:       zeroTime,
			},
			// Note: Order may vary due to map iteration, but should have 6 placeholders total (3 for namespaces, 1 for name, 1 for kind, 1 for object_id)
			expectedQuery: "namespace IN ( ?,?,? ) AND name = ? AND kind = ? AND object_id = ?",
			expectedArgs:  []string{"test-ns", "prod-ns", "staging-ns", "test-name", "PodChaos", "test-uid"},
		},
		{
			name: "empty namespace filtered out",
			filter: Filter{
				Namespace: "test-ns,,",
				Start:     zeroTime,
				End:       zeroTime,
			},
			expectedQuery: "namespace IN ( ? )",
			expectedArgs:  []string{"test-ns"},
		},
		{
			name: "whitespace-only namespaces filtered out",
			filter: Filter{
				Namespace: "test-ns,  ,prod-ns",
				Start:     zeroTime,
				End:       zeroTime,
			},
			expectedQuery: "namespace IN ( ?,? )",
			expectedArgs:  []string{"test-ns", "prod-ns"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			query, args := c.filter.ConstructQueryArgs()

			// Check that query contains the expected namespace IN clause
			if c.expectedQuery != "" {
				if !contains(query, "namespace IN (") {
					t.Errorf("expected query to contain 'namespace IN (' but got %s", query)
				}
				// Verify the number of placeholders matches expected
				expectedPlaceholders := countPlaceholders(c.expectedQuery)
				actualPlaceholders := countPlaceholders(query)
				if expectedPlaceholders != actualPlaceholders {
					t.Errorf("expected %d placeholders but got %d in query: %s", expectedPlaceholders, actualPlaceholders, query)
				}
			} else {
				if query != "" {
					t.Errorf("expected empty query but got %s", query)
				}
			}

			// Verify args contain all expected values (order independent)
			argsMap := make(map[interface{}]int)
			for _, arg := range args {
				argsMap[arg]++
			}
			expectedMap := make(map[interface{}]int)
			for _, arg := range c.expectedArgs {
				expectedMap[arg]++
			}
			if !reflect.DeepEqual(argsMap, expectedMap) {
				t.Errorf("expected args %v but got %v", c.expectedArgs, args)
			}
		})
	}
}

func countPlaceholders(s string) int {
	count := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '?' {
			count++
		}
	}
	return count
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
