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
// limitations under the License

package main

import "fmt"

// struct scheduleCodeGenerator will render content of one file contains code blocks that required by workflow
type scheduleCodeGenerator struct {
	// name of each Kind of chaos, for example: PodChaos, IoChaos, DNSChaos
	chaosTypes []string
}

func newScheduleCodeGenerator(types []string) *scheduleCodeGenerator {
	return &scheduleCodeGenerator{chaosTypes: types}
}

func (it *scheduleCodeGenerator) AppendTypes(typeName string) {
	it.chaosTypes = append(it.chaosTypes, typeName)
}

func (it *scheduleCodeGenerator) Render() string {
	return fmt.Sprintf(`

	`)
}
