// Copyright 2020 Chaos Mesh Authors.
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

package template

import "github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/template"

type mockSerialTemplate struct {
	mockTemplate
	children []template.Template
}

func NewMockSerialTemplate() *mockSerialTemplate {
	return &mockSerialTemplate{
		mockTemplate: mockTemplate{
			templateType: template.Serial,
		},
	}
}

func (it *mockSerialTemplate) SetSerialChildrenList(children []template.Template) {
	it.children = children
}

func (it *mockSerialTemplate) GetSerialChildrenList() []template.Template {
	return it.children
}
