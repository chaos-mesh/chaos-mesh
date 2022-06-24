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
	"reflect"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

var (
	key = types.NamespacedName{
		Namespace: "default",
		Name:      "result-cache-test",
	}
	record1 = v1alpha1.StatusCheckRecord{
		StartTime: &metav1.Time{Time: time.Now()},
		Outcome:   v1alpha1.StatusCheckOutcomeFailure,
	}
	record2 = v1alpha1.StatusCheckRecord{
		StartTime: &metav1.Time{Time: time.Now().Add(5 * time.Second)},
		Outcome:   v1alpha1.StatusCheckOutcomeSuccess,
	}
)

func Test_limitRecords(t *testing.T) {
	type args struct {
		records []v1alpha1.StatusCheckRecord
		limit   uint
	}
	tests := []struct {
		name string
		args args
		want []v1alpha1.StatusCheckRecord
	}{
		{
			name: "not exceeded",
			args: args{
				records: []v1alpha1.StatusCheckRecord{record1},
				limit:   2,
			},
			want: []v1alpha1.StatusCheckRecord{record1},
		},
		{
			name: "exceeded",
			args: args{
				records: []v1alpha1.StatusCheckRecord{record1, record2},
				limit:   1,
			},
			want: []v1alpha1.StatusCheckRecord{record2},
		},
		{
			name: "empty",
			args: args{
				records: []v1alpha1.StatusCheckRecord{},
				limit:   1,
			},
			want: []v1alpha1.StatusCheckRecord{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := limitRecords(tt.args.records, tt.args.limit); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("limitRecords() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resultCache_get(t *testing.T) {
	type fields struct {
		results map[types.NamespacedName]Result
	}
	type args struct {
		key types.NamespacedName
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Result
		want1  bool
	}{
		{
			name: "found",
			fields: fields{
				results: map[types.NamespacedName]Result{
					key: {
						Records:             []v1alpha1.StatusCheckRecord{record1},
						Count:               1,
						recordsHistoryLimit: 1,
					},
				},
			},
			args: args{key: key},
			want: Result{
				Records:             []v1alpha1.StatusCheckRecord{record1},
				Count:               1,
				recordsHistoryLimit: 1,
			},
			want1: true,
		},
		{
			name: "not found",
			fields: fields{
				results: map[types.NamespacedName]Result{
					key: {
						Records:             []v1alpha1.StatusCheckRecord{record1},
						Count:               1,
						recordsHistoryLimit: 1,
					},
				},
			},
			args: args{key: types.NamespacedName{
				Namespace: "default",
				Name:      "cache-test",
			}},
			want:  Result{},
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &resultCache{
				results: tt.fields.results,
			}
			got, got1 := c.get(tt.args.key)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("get() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("get() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_resultCache_append(t *testing.T) {
	type fields struct {
		results map[types.NamespacedName]Result
	}
	type args struct {
		key types.NamespacedName
		obj v1alpha1.StatusCheckRecord
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Result
		want1  bool
	}{
		{
			name: "append",
			fields: fields{
				results: map[types.NamespacedName]Result{
					key: {
						Records:             []v1alpha1.StatusCheckRecord{record1},
						Count:               1,
						recordsHistoryLimit: 2,
					},
				},
			},
			args: args{
				key: key,
				obj: record2,
			},
			want: Result{
				Records:             []v1alpha1.StatusCheckRecord{record1, record2},
				Count:               2,
				recordsHistoryLimit: 2,
			},
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &resultCache{
				results: tt.fields.results,
			}
			c.append(tt.args.key, tt.args.obj)
			got, got1 := c.get(tt.args.key)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("append() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("append() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_resultCache_delete(t *testing.T) {
	type fields struct {
		results map[types.NamespacedName]Result
	}
	type args struct {
		key types.NamespacedName
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "delete",
			fields: fields{
				results: map[types.NamespacedName]Result{
					key: {
						Records:             []v1alpha1.StatusCheckRecord{record1},
						Count:               1,
						recordsHistoryLimit: 1,
					},
				},
			},
			args: args{key: key},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &resultCache{
				results: tt.fields.results,
			}
			c.delete(tt.args.key)
			_, ok := c.get(tt.args.key)
			if ok {
				t.Errorf("object exists after delete it")
			}
		})
	}
}
