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

package main

import (
	"bytes"
	"text/template"
)

const initTemplate = `
	SchemeBuilder.Register(&{{.Type}}{}, &{{.Type}}List{})
	all.register(Kind{{.Type}}, &ChaosKind{
		Chaos:     &{{.Type}}{},
		GenericChaosList: &{{.Type}}List{},
	})
`

func generateInit(name string) string {
	tmpl, err := template.New("ini").Parse(initTemplate)
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
