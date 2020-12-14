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
