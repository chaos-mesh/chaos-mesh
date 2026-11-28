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

package generic

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func TestRandomFixedIndexes(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name              string
		start             uint
		end               uint
		count             uint
		expectedOutputLen int
	}

	tcs := []TestCase{
		{
			name:              "start 0, end 10, count 3",
			start:             0,
			end:               10,
			count:             3,
			expectedOutputLen: 3,
		},
		{
			name:              "start 0, end 10, count 12",
			start:             0,
			end:               10,
			count:             12,
			expectedOutputLen: 10,
		},
		{
			name:              "start 5, end 10, count 3",
			start:             5,
			end:               10,
			count:             3,
			expectedOutputLen: 3,
		},
	}

	for _, tc := range tcs {
		values := RandomFixedIndexes(tc.start, tc.end, tc.count)
		g.Expect(len(values)).To(Equal(tc.expectedOutputLen), tc.name)

		for _, v := range values {
			g.Expect(v).Should(BeNumerically(">=", tc.start), tc.name)
			g.Expect(v).Should(BeNumerically("<", tc.end), tc.name)
		}
	}
}

func TestFilterObjectsByMode(t *testing.T) {
	for _, example := range []struct {
		name      string
		mode      v1alpha1.SelectorMode
		value     string
		count     int
		min, max  int
		wantError string
	}{
		{name: "empty candidates", mode: v1alpha1.AllMode, wantError: "empty list"},
		{name: "one candidate", mode: v1alpha1.OneMode, count: 5, min: 1, max: 1},
		{name: "all candidates", mode: v1alpha1.AllMode, count: 5, min: 5, max: 5},
		{name: "fixed size capped at available candidates", mode: v1alpha1.FixedMode, value: "12", count: 5, min: 5, max: 5},
		{name: "invalid fixed size", mode: v1alpha1.FixedMode, value: "two", count: 5, wantError: "invalid syntax"},
		{name: "negative fixed size", mode: v1alpha1.FixedMode, value: "-1", count: 5, wantError: "below or equal 0"},
		{name: "fixed percent rounds up", mode: v1alpha1.FixedPercentMode, value: "25", count: 5, min: 2, max: 2},
		{name: "invalid fixed percent", mode: v1alpha1.FixedPercentMode, value: "25%", count: 5, wantError: "invalid syntax"},
		{name: "zero fixed percent", mode: v1alpha1.FixedPercentMode, value: "0", count: 5, wantError: "below or equal 0"},
		{name: "fixed percent exceeds 100", mode: v1alpha1.FixedPercentMode, value: "101", count: 5, wantError: "invalid"},
		{name: "random selection respects the maximum", mode: v1alpha1.RandomMaxPercentMode, value: "20", count: 20, max: 4},
		{name: "invalid random maximum", mode: v1alpha1.RandomMaxPercentMode, value: "20%", count: 5, wantError: "invalid syntax"},
		{name: "zero random maximum", mode: v1alpha1.RandomMaxPercentMode, value: "0", count: 5, wantError: "below or equal 0"},
		{name: "random maximum exceeds 100", mode: v1alpha1.RandomMaxPercentMode, value: "101", count: 5, wantError: "invalid"},
		{name: "unknown mode", mode: "unsupported", count: 5, wantError: "not supported"},
	} {
		t.Run(example.name, func(t *testing.T) {
			g := NewWithT(t)
			indexes, err := FilterObjectsByMode(example.mode, example.value, example.count)
			if example.wantError != "" {
				g.Expect(err).To(MatchError(ContainSubstring(example.wantError)))
				g.Expect(indexes).To(BeEmpty())
				return
			}
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(len(indexes)).To(BeNumerically(">=", example.min))
			g.Expect(len(indexes)).To(BeNumerically("<=", example.max))
			seen := make(map[uint]struct{}, len(indexes))
			for _, index := range indexes {
				g.Expect(index).To(BeNumerically("<", example.count))
				g.Expect(seen).NotTo(HaveKey(index))
				seen[index] = struct{}{}
			}
		})
	}
}
