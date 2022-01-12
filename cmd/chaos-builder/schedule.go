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

package main

import (
	"bytes"
	"fmt"
	"text/template"
)

const scheduleTemplate = `
	allScheduleItem.register(Kind{{.Type}}, &ChaosKind{
		chaos: &{{.Type}}{},
		list:  &{{.Type}}List{},
	})
`

func generateScheduleRegister(name string) string {
	tmpl, err := template.New("ini").Parse(scheduleTemplate)
	if err != nil {
		log.Error(err, "fail to build template")
		panic(err)
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, &metadata{
		Type: name,
	})
	if err != nil {
		log.Error(err, "fail to execute template")
		panic(err)
	}

	return buf.String()
}

type scheduleCodeGenerator struct {
	// name of each Kind of chaos, for example: PodChaos, IOChaos, DNSChaos
	chaosTypes []string
}

func newScheduleCodeGenerator(types []string) *scheduleCodeGenerator {
	return &scheduleCodeGenerator{chaosTypes: types}
}

func (it *scheduleCodeGenerator) AppendTypes(typeName string) {
	it.chaosTypes = append(it.chaosTypes, typeName)
}

func (it *scheduleCodeGenerator) Render() string {

	scheduleTemplateTypesEntries := ""
	for _, item := range it.chaosTypes {
		scheduleTemplateTypesEntries += generateScheduleTemplateTypes(item)
	}

	embedChaosEntries := ""
	for _, item := range it.chaosTypes {
		embedChaosEntries += generateScheduleItem(item)
	}

	scheduleTemplateTypeEntries := ""
	for _, item := range it.chaosTypes {
		scheduleTemplateTypeEntries += fmt.Sprintf(`	ScheduleType%s,
`, item)
	}

	spawnMethod := ""
	for _, item := range it.chaosTypes {
		spawnMethod += generateFillingMethodScheduleItem(item, scheduleFillingEntryTemplate)
	}

	restoreMethod := ""
	for _, item := range it.chaosTypes {
		restoreMethod += generateFillingMethodScheduleItem(item, scheduleRestoreEntryTemplate)
	}

	imports := `import (
	"github.com/pkg/errors"
)
`

	scheduleTemplateTypesCodes := fmt.Sprintf(`%s

%s

const (
%s
)

var allScheduleTemplateType = []ScheduleTemplateType{
%s
}

func (it *ScheduleItem) SpawnNewObject(templateType ScheduleTemplateType) (GenericChaos, error) {
	switch templateType {
%s
	default:
		return nil, errors.Wrapf(errInvalidValue, "unknown template type %%s", templateType)
	}
}

func (it *ScheduleItem) RestoreChaosSpec(root interface{}) error {
	switch chaos := root.(type) {
%s
	default:
		return errors.Wrapf(errInvalidValue, "unknown chaos %%#v", root)
	}
}
`,
		boilerplate,
		imports,
		scheduleTemplateTypesEntries,
		scheduleTemplateTypeEntries,
		spawnMethod,
		restoreMethod,
	)

	return scheduleTemplateTypesCodes
}

const scheduleTemplateTypeEntryTemplate = `	ScheduleType{{.Type}} ScheduleTemplateType = "{{.Type}}"
`

func generateScheduleTemplateTypes(typeName string) string {
	tmpl, err := template.New("scheduleTemplates").Parse(scheduleTemplateTypeEntryTemplate)
	if err != nil {
		log.Error(err, "fail to build template")
		return ""
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, &metadata{
		Type: typeName,
	})
	if err != nil {
		log.Error(err, "fail to execute template")
		return ""
	}

	return buf.String()
}

const scheduleItemTemplate = `	// +optional
	{{.Type}} *{{.Type}}Spec ` + "`" + `json:"{{.JsonField}},omitempty"` + "`" + `
`

func generateScheduleItem(typeName string) string {
	value := struct {
		Type      string
		JsonField string
	}{
		Type:      typeName,
		JsonField: lowercaseCamelCase(typeName),
	}
	tmpl, err := template.New("scheduleTemplates").Parse(scheduleItemTemplate)
	if err != nil {
		log.Error(err, "fail to build template")
		return ""
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, &value)
	if err != nil {
		log.Error(err, "fail to execute template")
		return ""
	}

	return buf.String()
}

const scheduleFillingEntryTemplate = `	case ScheduleType{{.Type}}:
		result := {{.Type}}{}
		result.Spec = *it.{{.Type}}
		return &result, nil
`

const scheduleRestoreEntryTemplate = `	case *{{.Type}}:
		*it.{{.Type}} = chaos.Spec
		return nil
`

func generateFillingMethodScheduleItem(typeName, methodTemplate string) string {
	tmpl, err := template.New("fillingMethod").Parse(methodTemplate)
	if err != nil {
		log.Error(err, "fail to build template")
		return ""
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, &metadata{
		Type: typeName,
	})
	if err != nil {
		log.Error(err, "fail to execute template")
		return ""
	}

	return buf.String()
}
