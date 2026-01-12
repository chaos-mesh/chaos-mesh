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

package utils

import (
	"reflect"
	"testing"
)

func TestParseNamespaceQuery(t *testing.T) {
	cases := []struct {
		name           string
		namespaceParam string
		expected       []string
	}{
		{
			name:           "empty string",
			namespaceParam: "",
			expected:       []string{},
		},
		{
			name:           "single namespace",
			namespaceParam: "default",
			expected:       []string{"default"},
		},
		{
			name:           "comma-separated namespaces",
			namespaceParam: "default,production,staging",
			expected:       []string{"default", "production", "staging"},
		},
		{
			name:           "comma-separated with spaces",
			namespaceParam: "default, production, staging",
			expected:       []string{"default", "production", "staging"},
		},
		{
			name:           "comma-separated with extra spaces",
			namespaceParam: "  default  ,  production  ,  staging  ",
			expected:       []string{"default", "production", "staging"},
		},
		{
			name:           "empty namespaces filtered out",
			namespaceParam: "default,,production",
			expected:       []string{"default", "production"},
		},
		{
			name:           "only empty namespaces",
			namespaceParam: ",,",
			expected:       []string{},
		},
		{
			name:           "whitespace-only filtered out",
			namespaceParam: "default,  ,production",
			expected:       []string{"default", "production"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := ParseNamespaceQuery(c.namespaceParam)
			if !reflect.DeepEqual(result, c.expected) {
				t.Errorf("expected %v but got %v", c.expected, result)
			}
		})
	}
}
