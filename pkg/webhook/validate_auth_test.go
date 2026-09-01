// Copyright 2026 Chaos Mesh Authors.
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

package webhook

import (
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
)

func TestSarVerb(t *testing.T) {
	cases := []struct {
		operation admissionv1.Operation
		want      string
	}{
		{admissionv1.Create, "create"},
		{admissionv1.Update, "update"},
		{admissionv1.Delete, "delete"},
	}

	for _, c := range cases {
		got := sarVerb(c.operation)
		if got != c.want {
			t.Errorf("sarVerb(%q) = %q, want %q", c.operation, got, c.want)
		}
	}
}
