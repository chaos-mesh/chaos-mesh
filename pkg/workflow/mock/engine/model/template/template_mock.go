package template

import (
	"fmt"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/template"
	"time"
)

type mockTemplate struct {
	name         string
	templateType template.TemplateType
	duration     time.Duration
	deadline     time.Duration
}

func NewMockTemplate() *mockTemplate {
	return &mockTemplate{}
}

func (it *mockTemplate) SetDeadline(deadline time.Duration) {
	it.deadline = deadline
}

func (it *mockTemplate) SetDuration(duration time.Duration) {
	it.duration = duration
}

func (it *mockTemplate) SetTemplateType(templateType template.TemplateType) {
	it.templateType = templateType
}

func (it *mockTemplate) SetName(name string) {
	it.name = name
}

func (it *mockTemplate) GetName() string {
	return it.name
}

func (it *mockTemplate) GetTemplateType() template.TemplateType {
	return it.templateType
}

func (it *mockTemplate) GetDuration() time.Duration {
	return it.duration
}

func (it *mockTemplate) GetDeadline() time.Duration {
	return it.deadline
}

type mockedTemplates struct {
	origin []template.Template
}

func NewMockedTemplates(origin []template.Template) *mockedTemplates {
	return &mockedTemplates{origin: origin}
}

func (it *mockedTemplates) FetchAllTemplates() []template.Template {
	return it.origin
}

func (it *mockedTemplates) FetchTemplateMap() map[string]template.Template {
	result := make(map[string]template.Template)
	for _, item := range it.origin {
		if _, exists := result[item.GetName()]; exists {
			panic(fmt.Sprintf("template %s already exist", item.GetName()))
		} else {
			result[item.GetName()] = item
		}
	}
	return result
}

func (it *mockedTemplates) GetByTemplateName(templateName string) template.Template {
	return it.FetchTemplateMap()[templateName]
}
