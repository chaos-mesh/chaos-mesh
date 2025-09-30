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
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
)

// struct workflowCodeGenerator will render content of one file contains code blocks that required by workflow
type workflowCodeGenerator struct {
	// name of each Kind of chaos, for example: PodChaos, IOChaos, DNSChaos
	chaosTypes []string
}

func newWorkflowCodeGenerator(types []string) *workflowCodeGenerator {
	return &workflowCodeGenerator{chaosTypes: types}
}

func (it *workflowCodeGenerator) AppendTypes(typeName string) {
	it.chaosTypes = append(it.chaosTypes, typeName)
}

func (it *workflowCodeGenerator) Render() string {

	workflowTemplateTypesEntries := ""
	for _, item := range it.chaosTypes {
		workflowTemplateTypesEntries += generateTemplateTypes(item)
	}

	embedChaosEntries := ""
	for _, item := range it.chaosTypes {
		embedChaosEntries += generateEmbedChaos(item)
	}

	spawnObjectMethod := ""
	for _, item := range it.chaosTypes {
		spawnObjectMethod += generateMethodItem(item, spawnObjectEntryTemplate)
	}
	restoreObjectMethod := ""
	for _, item := range it.chaosTypes {
		restoreObjectMethod += generateMethodItem(item, restoreObjectEntryTemplate)
	}
	spawnListMethod := ""
	for _, item := range it.chaosTypes {
		spawnListMethod += generateSpawnListMethodItem(item)
	}
	allChaosTemplateTypeEntries := ""
	for _, item := range it.chaosTypes {
		allChaosTemplateTypeEntries += fmt.Sprintf(`	Type%s,
`, item)
	}

	genericChaosListImplementations := ""
	for _, item := range it.chaosTypes {
		genericChaosListImplementations += generateGenericChaosList(item)
	}

	imports := `import (
	"github.com/pkg/errors"
)
`

	workflowTemplateTypesCodes := fmt.Sprintf(`%s

%s

const (
	TypeTask TemplateType = "Task"
	TypeSerial TemplateType = "Serial"
	TypeParallel TemplateType = "Parallel"
	TypeSuspend TemplateType = "Suspend"
	TypeStatusCheck TemplateType = "StatusCheck"
	TypeSchedule TemplateType = "Schedule"
%s
)

var allChaosTemplateType = []TemplateType{
	TypeSchedule,
%s
}

type EmbedChaos struct {
%s
}

func (it *EmbedChaos) SpawnNewObject(templateType TemplateType) (GenericChaos, error) {
	switch templateType {
%s
	default:
		return nil, errors.Wrapf(errInvalidValue, "unknown template type %%s", templateType)
	}
}

func (it *EmbedChaos) RestoreChaosSpec(root interface{}) error {
	switch chaos := root.(type) {
%s
	default:
		return errors.Wrapf(errInvalidValue, "unknown chaos %%#v", root)
	}
}

func (it *EmbedChaos) SpawnNewList(templateType TemplateType) (GenericChaosList, error) {
	switch templateType {
%s
	default:
		return nil, errors.Wrapf(errInvalidValue, "unknown template type %%s", templateType)
	}
}

%s
`,
		boilerplate,
		imports,
		workflowTemplateTypesEntries,
		allChaosTemplateTypeEntries,
		embedChaosEntries,
		spawnObjectMethod,
		restoreObjectMethod,
		spawnListMethod,
		genericChaosListImplementations,
	)

	return workflowTemplateTypesCodes
}

const workflowTemplateTypeEntryTemplate = `	Type{{.Type}} TemplateType = "{{.Type}}"
`

func generateTemplateTypes(typeName string) string {
	tmpl, err := template.New("workflowTemplates").Parse(workflowTemplateTypeEntryTemplate)
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

const embedChaosEntryTemplate = `	// +optional
	{{.Type}} *{{.Type}}Spec ` + "`" + `json:"{{.JsonField}},omitempty"` + "`" + `
`

func generateEmbedChaos(typeName string) string {
	value := struct {
		Type      string
		JsonField string
	}{
		Type:      typeName,
		JsonField: lowercaseCamelCase(typeName),
	}
	tmpl, err := template.New("workflowTemplates").Parse(embedChaosEntryTemplate)
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

func lowercaseCamelCase(str string) string {
	// here are some name thing issue about the acronyms, we used ALLCAP name in chaos kind, like DNSChaos or JVMChaos,
	// library could not resolve that well, so we just manually do it.
	if strings.Contains(str, "Chaos") {
		position := strings.Index(str, "Chaos")
		return strings.ToLower(str[:position]) + str[position:]
	}
	return strcase.ToLowerCamel(str)
}

const spawnObjectEntryTemplate = `	case Type{{.Type}}:
		result := {{.Type}}{}
		result.Spec = *it.{{.Type}}
		return &result, nil
`

const restoreObjectEntryTemplate = `	case *{{.Type}}:
		*it.{{.Type}} = chaos.Spec
		return nil
`

func generateMethodItem(typeName, methodTemplate string) string {
	tmpl, err := template.New("fillMethodEntry").Parse(methodTemplate)
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

const spawnListEntryTemplate = `	case Type{{.Type}}:
		result := {{.Type}}List{}
		return &result, nil
`

func generateSpawnListMethodItem(typeName string) string {
	tmpl, err := template.New("fillingMethod").Parse(spawnListEntryTemplate)
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

const genericChaosList = `func (in *{{.Type}}List) GetItems() []GenericChaos {
	var result []GenericChaos
	for _, item := range in.Items {
		item := item
		result = append(result, &item)
	}
	return result
}
`

func generateGenericChaosList(typeName string) string {
	tmpl, err := template.New("genericChaosList").Parse(genericChaosList)
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
