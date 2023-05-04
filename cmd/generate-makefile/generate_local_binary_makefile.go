// Copyright 2023 Chaos Mesh Authors.
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

// localBinaryGeneratedMkTemplate is the template for the file local-binary.generated.mk, use binaryGeneratedMkOptions as the context.
const localBinaryGeneratedMkTemplate = `# Generated by ./cmd/generate-makefile. DO NOT EDIT.

##@ Generated targets in local-binary.generated.mk

{{ .Content -}}

.PHONY: clean-local-binary
clean-local-binary:
{{- range .Recipes }}
	rm -f {{ .OutputPath }}
{{- end }}
`

// localBinaryRecipeTemplate is the template for one target, use localBinaryRecipeOptions as the context.
const localBinaryRecipeTemplate = `.PHONY: {{ .OutputPath }}
{{ .TargetName }}: {{ StringsJoin .DependencyTargets " " }} ## {{ .Comment }}
{{- if .UseCGO }}
	$(CGO) build -ldflags "$(LDFLAGS)" -tags "${BUILD_TAGS}" -o {{ .OutputPath }} {{ .SourcePath }}
{{- else }}
	$(GO) build -ldflags "$(LDFLAGS)" -tags "${BUILD_TAGS}" -o {{ .OutputPath }} {{ .SourcePath }}
{{- end }}

`

// localBinaryRecipes is the list of binaryRecipes to generate, edit here to build new binaries.
var localBinaryRecipes = []binaryRecipeOptions{
	{
		TargetName: "local/chaos-controller-manager",
		SourcePath: "cmd/chaos-controller-manager/main.go",
		OutputPath: "bin/chaos-controller-manager",
		UseCGO:     false,
		Comment:    "Build binary chaos-controller-manager in local environment",
	},
	{
		TargetName: "local/chaos-dashboard",
		SourcePath: "cmd/chaos-dashboard/main.go",
		OutputPath: "bin/chaos-dashboard",
		UseCGO:     true,
		Comment:    "Build binary chaos-dashboard in local environment",
	},
}
