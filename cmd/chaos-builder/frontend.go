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

import "fmt"

type FrontendCodeGenerator struct {
	// name of each Kind of chaos, for example: PodChaos, IOChaos, DNSChaos
	chaosTypes []string
}

func newFrontendCodeGenerator(chaosTypes []string) *FrontendCodeGenerator {
	return &FrontendCodeGenerator{chaosTypes}
}

func (it *FrontendCodeGenerator) AppendTypes(typeName string) {
	it.chaosTypes = append(it.chaosTypes, typeName)
}

const typesTemplate = `import { ExperimentKind } from '@/components/NewExperiment/types'

const mapping = new Map<ExperimentKind, string>([
%s])

export function templateTypeToFieldName(templateType: ExperimentKind): string {
  return mapping.get(templateType)!
}
`

func (it *FrontendCodeGenerator) Render() string {
	return fmt.Sprintf(typesTemplate, it.mapEntries())
}

func (it *FrontendCodeGenerator) mapEntries() string {
	entries := ""
	for _, chaosType := range it.chaosTypes {
		entries += fmt.Sprintf(`  ['%s', '%s'],
`, chaosType, lowercaseCamelCase(chaosType))
	}
	return entries
}
