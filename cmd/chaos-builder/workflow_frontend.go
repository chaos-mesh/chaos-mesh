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

import "fmt"

type workflowFrontendCodeGenerator struct {
	// name of each Kind of chaos, for example: PodChaos, IOChaos, DNSChaos
	chaosTypes []string
}

func newWorkflowFrontendCodeGenerator(chaosTypes []string) *workflowFrontendCodeGenerator {
	return &workflowFrontendCodeGenerator{chaosTypes: chaosTypes}
}

func (it *workflowFrontendCodeGenerator) AppendTypes(typeName string) {
	it.chaosTypes = append(it.chaosTypes, typeName)
}

const typescriptTemplate = `export const mapping = new Map<string, string>([
%s])
`

func (it *workflowFrontendCodeGenerator) Render() string {
	return fmt.Sprintf(typescriptTemplate,
		it.mapEntries(),
	)
}

func (it *workflowFrontendCodeGenerator) mapEntries() string {
	entries := ""
	for _, chaosType := range it.chaosTypes {
		entries += fmt.Sprintf(`  ['%s', '%s'],
`, chaosType, LowercaseCamelCase(chaosType))
	}
	return entries
}
