// Copyright 2022 Chaos Mesh Authors.
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

package migrate

import (
	"bytes"
	"go/parser"
	"go/printer"
	"go/token"
	"testing"

	"github.com/chaos-mesh/chaos-mesh/pkg/log"
)

func TestMigrateAst(t *testing.T) {
	type testCase struct {
		originalSource string
		newSource      string
		name           string

		from string
		to   string
	}

	cases := []testCase{
		{
			name: "should-migrate",
			originalSource: `package test

import (
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func main() {
	v1alpha1.CallTestFunc("test")
}
`,
			newSource: `package test

import (
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha2"
)

func main() {
	v1alpha2.CallTestFunc("test")
}
`,
			from: "v1alpha1",
			to:   "v1alpha2",
		},
		{
			name: "should-not-migrate",
			originalSource: `package test

import (
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha3"
)

func main() {
	v1alpha3.CallTestFunc("test")
}
`,
			newSource: `package test

import (
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha3"
)

func main() {
	v1alpha3.CallTestFunc("test")
}
`,
			from: "v1alpha1",
			to:   "v1alpha2",
		},
		{
			name: "should-migrate-with-alias",
			originalSource: `package test

import (
	v1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func main() {
	v1alpha1.CallTestFunc("test")
}
`,
			newSource: `package test

import (
	v1alpha2 "github.com/chaos-mesh/chaos-mesh/api/v1alpha2"
)

func main() {
	v1alpha2.CallTestFunc("test")
}
`,
			from: "v1alpha1",
			to:   "v1alpha2",
		},
		{
			name: "should-not-migrate-with-unknown-alias",
			originalSource: `package test

import (
	test "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func main() {
	test.CallTestFunc("test")
}
`,
			newSource: `package test

import (
	test "github.com/chaos-mesh/chaos-mesh/api/v1alpha2"
)

func main() {
	test.CallTestFunc("test")
}
`,
			from: "v1alpha1",
			to:   "v1alpha2",
		},
	}

	log, err := log.NewDefaultZapLogger()
	if err != nil {
		t.Fatal("fail to create logger")
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fileSet := token.NewFileSet()
			fileAst, err := parser.ParseFile(fileSet, "test", c.originalSource, parser.ParseComments)
			if err != nil {
				t.Fatal("fail to parse file")
			}

			migrateAst(log, fileAst, c.from, c.to)

			output := new(bytes.Buffer)
			printer.Fprint(output, fileSet, fileAst)

			expect := c.newSource
			result := output.String()
			if expect != result {
				t.Fatalf(`source not equal:
expect:
%s
got:
%s`, expect, result)
			}
		})
	}
}
