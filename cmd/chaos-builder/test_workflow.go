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
	"text/template"
)

// struct workflowTestCodeGenerator will render content of one file for testing the coupling with chaosKindMap
type workflowTestCodeGenerator struct {
	// name of each Kind of chaos, for example: PodChaos, IOChaos, DNSChaos
	chaosTypes []string
}

func newWorkflowTestCodeGenerator(types []string) *workflowTestCodeGenerator {
	return &workflowTestCodeGenerator{chaosTypes: types}
}

func (it *workflowTestCodeGenerator) AppendTypes(typeName string) {
	it.chaosTypes = append(it.chaosTypes, typeName)
}

func (it *workflowTestCodeGenerator) Render() string {
	imports := `import (
	. "github.com/onsi/gomega"
	"testing"
)
`
	testMethods := ""
	for _, typeName := range it.chaosTypes {
		testMethods += generateTestTemplateEntry(typeName)
	}
	return fmt.Sprintf(`%s
// this file tests the coupling with all kinds map and each TemplateType
%s
%s
`,
		boilerplate,
		imports,
		testMethods,
	)
}

const testEntryTemplate = `func TestChaosKindMapShouldContains{{.Type}}(t *testing.T) {
	g := NewGomegaWithT(t)
	var requiredType TemplateType
	requiredType = Type{{.Type}}

	_, ok := all.kinds[string(requiredType)]
	g.Expect(ok).To(Equal(true), "all kinds map should contains this type", requiredType)
}
`

func generateTestTemplateEntry(typeName string) string {
	tmpl, err := template.New("testEntryTemplate").Parse(testEntryTemplate)
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
