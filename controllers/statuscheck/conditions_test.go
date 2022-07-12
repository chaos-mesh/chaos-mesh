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

package statuscheck

import (
	"testing"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func Test_isThresholdExceed(t *testing.T) {
	type args struct {
		records   []v1alpha1.StatusCheckRecord
		outcome   v1alpha1.StatusCheckOutcome
		threshold int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "not exceed",
			args: args{
				records: []v1alpha1.StatusCheckRecord{
					{Outcome: v1alpha1.StatusCheckOutcomeFailure},
					{Outcome: v1alpha1.StatusCheckOutcomeSuccess},
					{Outcome: v1alpha1.StatusCheckOutcomeFailure},
				},
				outcome:   v1alpha1.StatusCheckOutcomeFailure,
				threshold: 2,
			},
			want: false,
		},
		{
			name: "exceed",
			args: args{
				records: []v1alpha1.StatusCheckRecord{
					{Outcome: v1alpha1.StatusCheckOutcomeFailure},
				},
				outcome:   v1alpha1.StatusCheckOutcomeFailure,
				threshold: 1,
			},
			want: true,
		},
		{
			name: "exceed",
			args: args{
				records: []v1alpha1.StatusCheckRecord{
					{Outcome: v1alpha1.StatusCheckOutcomeSuccess},
					{Outcome: v1alpha1.StatusCheckOutcomeFailure},
					{Outcome: v1alpha1.StatusCheckOutcomeFailure},
				},
				outcome:   v1alpha1.StatusCheckOutcomeFailure,
				threshold: 2,
			},
			want: true,
		},
		{
			name: "threshold is bigger than the length of record",
			args: args{
				records: []v1alpha1.StatusCheckRecord{
					{Outcome: v1alpha1.StatusCheckOutcomeFailure},
					{Outcome: v1alpha1.StatusCheckOutcomeFailure},
				},
				outcome:   v1alpha1.StatusCheckOutcomeFailure,
				threshold: 3,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isThresholdExceed(tt.args.records, tt.args.outcome, tt.args.threshold); got != tt.want {
				t.Errorf("isThresholdExceed() = %v, want %v", got, tt.want)
			}
		})
	}
}
