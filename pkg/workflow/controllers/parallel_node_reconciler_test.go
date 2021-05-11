// Copyright 2021 Chaos Mesh Authors.
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

package controllers

import (
	"reflect"
	"sort"
	"testing"
)

func Test_getTaskNameFromGeneratedName(t *testing.T) {
	type args struct {
		generatedNodeName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"common case",
			args{"name-1"},
			"name",
		}, {
			"common case",
			args{"name-1-2"},
			"name-1",
		}, {
			"common case",
			args{"name"},
			"name",
		}, {
			"common case",
			args{"name-"},
			"name",
		},
		{
			"common case",
			args{"-name"},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTaskNameFromGeneratedName(tt.args.generatedNodeName); got != tt.want {
				t.Errorf("getTaskNameFromGeneratedName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_relativeComplementSet(t *testing.T) {
	type args struct {
		former []string
		latter []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "common_case",
			args: args{
				former: []string{"a", "b", "c"},
				latter: []string{},
			},
			want: []string{"a", "b", "c"},
		}, {
			name: "common_case",
			args: args{
				former: []string{"a", "b", "c"},
				latter: []string{"b", "c"},
			},
			want: []string{"a"},
		}, {
			name: "common_case",
			args: args{
				former: []string{"a", "b", "c"},
				latter: []string{"c", "a"},
			},
			want: []string{"b"},
		}, {
			name: "common_case",
			args: args{
				former: []string{"a", "b", "c"},
				latter: []string{"c", "b", "d"},
			},
			want: []string{"a"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := relativeComplementSet(test.args.former, test.args.latter)
			sort.Strings(got)
			sort.Strings(test.want)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("getTaskNameFromGeneratedName() = %v, want %v", got, test.want)
			}
		})
	}
}
