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

package experiment

import (
	"reflect"
	"testing"
)

func TestConstructQueryArgs(t *testing.T) {
	cases := []struct {
		kind          string
		ns            string
		name          string
		uid           string
		expectedQuery string
		expectedArgs  []string
	}{
		{
			kind:          "",
			ns:            "",
			name:          "",
			uid:           "",
			expectedQuery: "",
			expectedArgs:  []string{},
		},
		{
			kind:          "PodChaos",
			ns:            "",
			name:          "",
			uid:           "",
			expectedQuery: "kind = ?",
			expectedArgs:  []string{"PodChaos"},
		},
		{
			kind:          "",
			ns:            "test-ns",
			name:          "",
			uid:           "",
			expectedQuery: "namespace = ?",
			expectedArgs:  []string{"test-ns"},
		},
		{
			kind:          "",
			ns:            "",
			name:          "test-name",
			uid:           "",
			expectedQuery: "name = ?",
			expectedArgs:  []string{"test-name"},
		},
		{
			kind:          "PodChaos",
			ns:            "test-ns",
			name:          "",
			uid:           "",
			expectedQuery: "kind = ? AND namespace = ?",
			expectedArgs:  []string{"PodChaos", "test-ns"},
		},
		{
			kind:          "PodChaos",
			ns:            "test-ns",
			name:          "test-name",
			uid:           "",
			expectedQuery: "kind = ? AND namespace = ? AND name = ?",
			expectedArgs:  []string{"PodChaos", "test-ns", "test-name"},
		},
		{
			kind:          "",
			ns:            "",
			name:          "",
			uid:           "test-uid",
			expectedQuery: "uid = ?",
			expectedArgs:  []string{"test-uid"},
		},
		{
			kind:          "PodChaos",
			ns:            "",
			name:          "",
			uid:           "test-uid",
			expectedQuery: "kind = ? AND uid = ?",
			expectedArgs:  []string{"PodChaos", "test-uid"},
		},
		{
			kind:          "",
			ns:            "test-ns",
			name:          "",
			uid:           "test-uid",
			expectedQuery: "namespace = ? AND uid = ?",
			expectedArgs:  []string{"test-ns", "test-uid"},
		},
		{
			kind:          "",
			ns:            "",
			name:          "test-name",
			uid:           "test-uid",
			expectedQuery: "name = ? AND uid = ?",
			expectedArgs:  []string{"test-name", "test-uid"},
		},
		{
			kind:          "PodChaos",
			ns:            "test-ns",
			name:          "",
			uid:           "test-uid",
			expectedQuery: "kind = ? AND namespace = ? AND uid = ?",
			expectedArgs:  []string{"PodChaos", "test-ns", "test-uid"},
		},
		{
			kind:          "PodChaos",
			ns:            "test-ns",
			name:          "test-name",
			uid:           "test-uid",
			expectedQuery: "kind = ? AND namespace = ? AND name = ? AND uid = ?",
			expectedArgs:  []string{"PodChaos", "test-ns", "test-name", "test-uid"},
		},
	}

	for _, c := range cases {
		query, args := constructQueryArgs(c.kind, c.ns, c.name, c.uid)
		if query != c.expectedQuery {
			t.Errorf("expected query %s but got %s", c.expectedQuery, query)
		}
		if !reflect.DeepEqual(c.expectedArgs, args) {
			t.Errorf("expected args %v but got %v", c.expectedArgs, args)
		}
	}
}
