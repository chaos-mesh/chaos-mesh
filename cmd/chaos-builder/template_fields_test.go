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

package main

import "testing"

func TestGeneratedChaosFields(t *testing.T) {
	want := "\t// +optional\n\tHTTPChaos *HTTPChaosSpec `json:\"httpChaos,omitempty\"`\n"
	for _, tc := range []struct {
		name     string
		generate func(string) string
	}{
		{name: "schedule", generate: generateScheduleItem},
		{name: "workflow", generate: generateEmbedChaos},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.generate("HTTPChaos"); got != want {
				t.Fatalf("generated field = %q, want %q", got, want)
			}
		})
	}
}
