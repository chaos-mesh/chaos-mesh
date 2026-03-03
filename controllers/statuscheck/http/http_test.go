// Copyright Chaos Mesh Authors.
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

package http

import "testing"

func Test_validateStatusCode(t *testing.T) {
	tcs := []struct {
		name     string
		criteria string
		result   int
		expect   bool
	}{
		{
			name:     "single code, correct result",
			criteria: "200",
			result:   200,
			expect:   true,
		}, {
			name:     "single code, wrong result",
			criteria: "200",
			result:   201,
			expect:   false,
		}, {
			name:     "code range, correct result",
			criteria: "200-200",
			result:   200,
			expect:   true,
		}, {
			name:     "code range, correct result",
			criteria: "200-400",
			result:   400,
			expect:   true,
		}, {
			name:     "code range, wrong result",
			criteria: "200-400",
			result:   500,
			expect:   false,
		}, {
			name:     "illegal criteria",
			criteria: "200.400",
			result:   500,
			expect:   false,
		}, {
			name:     "illegal criteria",
			criteria: "200-x",
			result:   500,
			expect:   false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ok := validateStatusCode(tc.criteria, response{statusCode: tc.result})
			if ok != tc.expect {
				t.Errorf("criteria: %s result: %d expect: %t", tc.criteria, tc.result, tc.expect)
			}
		})
	}
}
