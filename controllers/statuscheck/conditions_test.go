package statuscheck

import (
	"testing"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func Test_isThresholdExceed(t *testing.T) {
	type args struct {
		records   []v1alpha1.StatusCheckRecord
		want      v1alpha1.StatusCheckOutcome
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
				want:      v1alpha1.StatusCheckOutcomeFailure,
				threshold: 2,
			},
			want: false,
		},
		{
			name: "exceed",
			args: args{
				records: []v1alpha1.StatusCheckRecord{
					{Outcome: v1alpha1.StatusCheckOutcomeFailure},
					{Outcome: v1alpha1.StatusCheckOutcomeFailure},
					{Outcome: v1alpha1.StatusCheckOutcomeFailure},
				},
				want:      v1alpha1.StatusCheckOutcomeFailure,
				threshold: 3,
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
				want:      v1alpha1.StatusCheckOutcomeFailure,
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
				want:      v1alpha1.StatusCheckOutcomeFailure,
				threshold: 3,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isThresholdExceed(tt.args.records, tt.args.want, tt.args.threshold); got != tt.want {
				t.Errorf("isThresholdExceed() = %v, want %v", got, tt.want)
			}
		})
	}
}
