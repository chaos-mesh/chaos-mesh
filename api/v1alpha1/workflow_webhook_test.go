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

package v1alpha1

import (
	"fmt"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func Test_entryMustExists(t *testing.T) {
	entryPath := field.NewPath("spec", "entry")

	type args struct {
		path      *field.Path
		entry     string
		templates []Template
	}
	tests := []struct {
		name string
		args args
		want field.ErrorList
	}{
		{
			name: "entry is empty",
			args: args{
				path:      entryPath,
				entry:     "",
				templates: nil,
			},
			want: field.ErrorList{
				field.Required(entryPath, "the entry of workflow is required"),
				field.Invalid(entryPath, "", fmt.Sprintf("can not find a template with name %s", "")),
			},
		}, {
			name: "entry does not exist in templates",
			args: args{
				path:  entryPath,
				entry: "entry",
				templates: []Template{
					{
						Name: "whatever is not entry",
						Type: TypeSuspend,
					},
				},
			},
			want: field.ErrorList{
				field.Invalid(entryPath, "entry", fmt.Sprintf("can not find a template with name %s", "entry")),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := entryMustExists(tt.args.path, tt.args.entry, tt.args.templates); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("entryMustExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateTemplates(t *testing.T) {
	templatesPath := field.NewPath("spec", "templates")
	var nilTemplates []Template
	type args struct {
		path      *field.Path
		templates []Template
	}
	tests := []struct {
		name string
		args args
		want field.ErrorList
	}{
		{
			name: "templates is nil",
			args: args{
				path:      templatesPath,
				templates: nil,
			},
			want: field.ErrorList{
				field.Invalid(templatesPath, nilTemplates, "templates in workflow could not be empty"),
			},
		}, {
			name: "templates is empty",
			args: args{
				path:      templatesPath,
				templates: []Template{},
			},
			want: field.ErrorList{
				field.Invalid(templatesPath, []Template{}, "templates in workflow could not be empty"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateTemplates(tt.args.path, tt.args.templates); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("validateTemplates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_shouldBeNoTask(t *testing.T) {
	templatePath := field.NewPath("spec", "templates").Index(0)
	mockTask := Task{
		Container: &corev1.Container{Name: "fake-container"},
	}
	type args struct {
		path     *field.Path
		template Template
	}
	tests := []struct {
		name string
		args args
		want field.ErrorList
	}{
		{
			name: "contains unexpected task",
			args: args{
				path: templatePath,
				template: Template{
					Task: &mockTask,
				},
			},
			want: field.ErrorList{
				field.Invalid(templatePath, &mockTask, "this template should not contain Task"),
			},
		}, {
			name: "does not contain task",
			args: args{
				path:     templatePath,
				template: Template{},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldBeNoTask(tt.args.path, tt.args.template); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("shouldBeNoTask() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_shouldBeNoChildren(t *testing.T) {
	templatePath := field.NewPath("spec", "templates").Index(0)
	mockChildren := []string{"child-a", "child-b"}
	type args struct {
		path     *field.Path
		template Template
	}
	tests := []struct {
		name string
		args args
		want field.ErrorList
	}{
		{
			name: "contains unexpected children",
			args: args{
				path: templatePath,
				template: Template{
					Children: mockChildren,
				},
			},
			want: field.ErrorList{
				field.Invalid(templatePath, mockChildren, "this template should not contain Children"),
			},
		}, {
			name: "does not contain children",
			args: args{
				path:     templatePath,
				template: Template{},
			},
			want: nil,
		}, {
			name: "empty array is also valid",
			args: args{
				path: templatePath,
				template: Template{
					Children: []string{},
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldBeNoChildren(tt.args.path, tt.args.template); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("shouldBeNoChildren() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_shouldBeNoConditionalBranches(t *testing.T) {
	templatePath := field.NewPath("spec", "templates").Index(0)
	mockConditionalBranches := []ConditionalBranch{
		{Target: "", Expression: ""},
	}
	type args struct {
		path     *field.Path
		template Template
	}
	tests := []struct {
		name string
		args args
		want field.ErrorList
	}{
		{
			name: "contains unexpected conditional branches",
			args: args{
				path: templatePath,
				template: Template{
					ConditionalBranches: mockConditionalBranches,
				},
			},
			want: field.ErrorList{
				field.Invalid(templatePath, mockConditionalBranches, "this template should not contain ConditionalBranches"),
			},
		}, {
			name: "does not contain conditional branches",
			args: args{
				path:     templatePath,
				template: Template{},
			},
			want: nil,
		}, {
			name: "empty array is also valid",
			args: args{
				path: templatePath,
				template: Template{
					ConditionalBranches: []ConditionalBranch{},
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldBeNoConditionalBranches(tt.args.path, tt.args.template); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("shouldBeNoConditionalBranches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_shouldBeNoEmbedChaos(t *testing.T) {
	templatePath := field.NewPath("spec", "templates").Index(0)
	type args struct {
		path     *field.Path
		template Template
	}
	mockedEmbedChaos := &EmbedChaos{
		PodChaos: &PodChaosSpec{
			ContainerSelector: ContainerSelector{
				PodSelector: PodSelector{
					Selector: PodSelectorSpec{
						Namespaces: []string{"default"},
					},
				},
			},
			Action: PodKillAction,
		},
	}
	tests := []struct {
		name string
		args args
		want field.ErrorList
	}{
		{
			name: "unexpected embedded chaos",
			args: args{
				path: templatePath,
				template: Template{
					EmbedChaos: mockedEmbedChaos,
				},
			},
			want: field.ErrorList{
				field.Invalid(templatePath, mockedEmbedChaos, "this template should not contain any Chaos"),
			},
		}, {
			name: "only nil embedded chaos is valid",
			args: args{
				path: templatePath,
				template: Template{
					EmbedChaos: nil,
				},
			},
			want: nil,
		}, {
			name: "an embedded chaos with all the nil fields is also INVALID",
			args: args{
				path: templatePath,
				template: Template{
					EmbedChaos: &EmbedChaos{},
				},
			},
			want: field.ErrorList{
				field.Invalid(templatePath, &EmbedChaos{}, "this template should not contain any Chaos"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldBeNoEmbedChaos(tt.args.path, tt.args.template); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("shouldBeNoEmbedChaos() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_shouldBeNoSchedule(t *testing.T) {
	templatePath := field.NewPath("spec", "templates").Index(0)
	type args struct {
		path     *field.Path
		template Template
	}
	mockedSchedule := &ChaosOnlyScheduleSpec{
		Type: ScheduleTypePodChaos,
	}
	tests := []struct {
		name string
		args args
		want field.ErrorList
	}{
		{
			name: "unexpected schedule",
			args: args{
				path: templatePath,
				template: Template{
					Schedule: mockedSchedule,
				},
			},
			want: field.ErrorList{
				field.Invalid(templatePath, mockedSchedule, "this template should not contain Schedule"),
			},
		}, {
			name: "no schedule",
			args: args{
				path: templatePath,
				template: Template{
					Schedule: nil,
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldBeNoSchedule(tt.args.path, tt.args.template); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("shouldBeNoSchedule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_namesCouldNotBeDuplicated(t *testing.T) {
	templatesPath := field.NewPath("spec", "templates")
	type args struct {
		templatesPath *field.Path
		names         []string
	}
	tests := []struct {
		name string
		args args
		want field.ErrorList
	}{
		{
			name: "names could not be duplicated",
			args: args{
				templatesPath: templatesPath,
				names:         []string{"template-a", "template-b", "template-c", "template-a", "template-b", "template-d"},
			},
			want: field.ErrorList{
				field.Invalid(templatesPath, "", fmt.Sprintf("template name must be unique, duplicated names: %s", []string{"template-a", "template-b"})),
			},
		}, {
			name: "names could not be duplicated",
			args: args{
				templatesPath: templatesPath,
				names:         []string{"template-a", "template-b", "template-c", "template-d"},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := namesCouldNotBeDuplicated(tt.args.templatesPath, tt.args.names); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("namesCouldNotBeDuplicated() = %v, want %v", got, tt.want)
			}
		})
	}
}
