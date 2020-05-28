// Copyright 2020 PingCAP, Inc.
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
		expectedQuery string
		expectedArgs  []string
	}{
		{
			kind:          "",
			ns:            "",
			name:          "",
			expectedQuery: "",
			expectedArgs:  []string{},
		},
		{
			kind:          "PodChaos",
			ns:            "",
			name:          "",
			expectedQuery: "kind = ?",
			expectedArgs:  []string{"PodChaos"},
		},
		{
			kind:          "",
			ns:            "test-ns",
			name:          "",
			expectedQuery: "namespace = ?",
			expectedArgs:  []string{"test-ns"},
		},
		{
			kind:          "",
			ns:            "",
			name:          "test-name",
			expectedQuery: "name = ?",
			expectedArgs:  []string{"test-name"},
		},
		{
			kind:          "PodChaos",
			ns:            "test-ns",
			name:          "",
			expectedQuery: "kind = ? AND namespace = ?",
			expectedArgs:  []string{"PodChaos", "test-ns"},
		},
		{
			kind:          "PodChaos",
			ns:            "test-ns",
			name:          "test-name",
			expectedQuery: "kind = ? AND namespace = ? AND name = ?",
			expectedArgs:  []string{"PodChaos", "test-ns", "test-name"},
		},
	}

	for _, c := range cases {
		query, args := constructQueryArgs(c.kind, c.ns, c.name)
		if query != c.expectedQuery {
			t.Errorf("expected query %s but got %s", c.expectedQuery, query)
		}
		if !reflect.DeepEqual(c.expectedArgs, args) {
			t.Errorf("expected args %v but got %v", c.expectedArgs, args)
		}
	}
}
