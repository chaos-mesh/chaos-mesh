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

package main

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"
	"strings"
)

// struct workflowCodeGenerator will render content of one file contains code blocks that required by workflow
type workflowCodeGenerator struct {
	// name of each Kind of chaos, for example: PodChaos, IoChaos, DNSChaos
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

	spawnMethod := ""
	for _, item := range it.chaosTypes {
		spawnMethod += generateSpawnMethodItem(item)
	}
	allChaosTemplateTypeEntries := ""
	for _, item := range it.chaosTypes {
		allChaosTemplateTypeEntries += fmt.Sprintf(`	Type%s,
`, item)
	}

	imports := `import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)
`

	workflowTemplateTypesCodes := fmt.Sprintf(`%s

%s

const (
%s
)

var allChaosTemplateType = []TemplateType{
%s
}

type EmbedChaos struct {
%s
}

func (it *EmbedChaos) SpawnNewObject(templateType TemplateType) (runtime.Object, metav1.Object, error) {

	switch templateType {
%s
	default:
		return nil, nil, fmt.Errorf("unsupported template type %%s", templateType)
	}

	return nil, &metav1.ObjectMeta{}, nil
}
`,
		codeHeader,
		imports,
		workflowTemplateTypesEntries,
		allChaosTemplateTypeEntries,
		embedChaosEntries,
		spawnMethod,
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
		JsonField: camelCaseToSnakeCase(typeName),
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

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func camelCaseToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

const fillingEntryTemplate = `	case Type{{.Type}}:
		result := {{.Type}}{}
		result.Spec = *it.{{.Type}}
		return &result, result.GetObjectMeta(), nil
`

func generateSpawnMethodItem(typeName string) string {
	tmpl, err := template.New("fillingMethod").Parse(fillingEntryTemplate)
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
