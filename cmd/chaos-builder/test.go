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

const testImport = `
import (
	"reflect"
	"testing"

	"github.com/bxcodec/faker"
	. "github.com/onsi/gomega"
)
`

const testInit = `
func init() {
	faker.AddProvider("ioMethods", func(v reflect.Value) (interface{}, error) {
		return []IoMethod{LookUp}, nil
	})
}
`

const testTemplate = `
func Test{{.Type}}IsDeleted(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &{{.Type}}{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.IsDeleted()
}

func Test{{.Type}}IsIsPaused(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &{{.Type}}{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.IsPaused()
}

func Test{{.Type}}GetDuration(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &{{.Type}}{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.Spec.GetDuration()
}

func Test{{.Type}}GetChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &{{.Type}}{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetChaos()
}

func Test{{.Type}}GetStatus(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &{{.Type}}{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetStatus()
}

func Test{{.Type}}GetSpecAndMetaString(t *testing.T) {
	g := NewGomegaWithT(t)
	chaos := &{{.Type}}{}
	err := faker.FakeData(chaos)
	g.Expect(err).To(BeNil())
	chaos.GetSpecAndMetaString()
}

func Test{{.Type}}ListChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &{{.Type}}List{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.ListChaos()
}
`

func generateTest(name string) string {
	tmpl, err := template.New("test").Parse(testTemplate)
	if err != nil {
		log.Error(err, "fail to build template")
		return ""
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, &metadata{
		Type: name,
	})
	if err != nil {
		log.Error(err, "fail to execute template")
		return ""
	}

	return buf.String()
}
