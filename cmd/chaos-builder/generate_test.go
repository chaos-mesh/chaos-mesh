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
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

func TestGenerateBuildsAPIAndTemplateArtifacts(t *testing.T) {
	g := NewWithT(t)
	root := t.TempDir()
	apiDir := filepath.Join(root, "api", "v1alpha1")
	uiDir := filepath.Join(root, "ui", "app", "src", "api")
	g.Expect(os.MkdirAll(apiDir, 0o755)).To(Succeed())
	g.Expect(os.MkdirAll(uiDir, 0o755)).To(Succeed())

	fixtures := map[string]string{
		"pod.go": `package v1alpha1
// +chaos-mesh:experiment
// +chaos-mesh:oneshot=in.Spec.Action == PodKillAction
type PodChaos struct{}`,
		"http.go": `package v1alpha1
// +chaos-mesh:experiment
// +chaos-mesh:webhook:enableUpdate
type HTTPChaos struct{}`,
		"workflow.go": `package v1alpha1
// +chaos-mesh:experiment
type Workflow struct{}`,
		"statuscheck.go": `package v1alpha1
// +chaos-mesh:base
type StatusCheck struct{}`,
		"zz_generated.previous.go": "this invalid Go file must be ignored",
		"readme.txt":               "this file must be ignored",
	}
	for name, source := range fixtures {
		g.Expect(os.WriteFile(filepath.Join(apiDir, name), []byte(source), 0o600)).To(Succeed())
	}
	g.Expect(generate(root)).To(Succeed())

	readGo := func(name string) string {
		t.Helper()
		path := filepath.Join(apiDir, name)
		content, err := os.ReadFile(path)
		g.Expect(err).NotTo(HaveOccurred())
		_, err = parser.ParseFile(token.NewFileSet(), path, content, parser.AllErrors)
		g.Expect(err).NotTo(HaveOccurred())
		return string(content)
	}
	api := readGo("zz_generated.chaosmesh.go")
	g.Expect(api).To(ContainSubstring("func (in *PodChaos) IsOneShot()"))
	g.Expect(api).To(ContainSubstring("if in.Spec.Action == PodKillAction"))
	g.Expect(api).To(ContainSubstring("func (in *StatusCheck) ValidateCreate("))
	for kind, allowsUpdate := range map[string]bool{"HTTPChaos": true, "PodChaos": false} {
		_, function, found := strings.Cut(api, "func (in *"+kind+") ValidateUpdate(")
		g.Expect(found).To(BeTrue(), "expected validation method for %s", kind)
		body, _, found := strings.Cut(function, "\n}")
		g.Expect(found).To(BeTrue())
		g.Expect(strings.Contains(body, "ErrCanNotUpdateChaos")).To(Equal(!allowsUpdate), "unexpected update policy for %s", kind)
	}
	g.Expect(api).To(ContainSubstring("allScheduleItem.register(KindWorkflow"))
	g.Expect(api).NotTo(ContainSubstring("func (in *Workflow) IsOneShot()"))

	webhookTests := readGo("zz_generated.chaosmesh_test.go")
	g.Expect(webhookTests).To(ContainSubstring("func TestPodChaosGetDuration("))
	workflowTests := readGo("zz_generated.workflow.chaosmesh_test.go")
	g.Expect(workflowTests).To(ContainSubstring("func TestChaosKindMapShouldContainsHTTPChaos("))

	workflow := readGo("zz_generated.workflow.chaosmesh.go")
	g.Expect(workflow).To(ContainSubstring("TypePodChaos TemplateType"))
	g.Expect(workflow).To(ContainSubstring("json:\"httpChaos,omitempty\""))
	g.Expect(workflow).NotTo(ContainSubstring("TypeWorkflow TemplateType"))
	schedule := readGo("zz_generated.schedule.chaosmesh.go")
	g.Expect(schedule).To(ContainSubstring("ScheduleTypeWorkflow ScheduleTemplateType"))
	g.Expect(schedule).To(ContainSubstring("case ScheduleTypeWorkflow:"))
	g.Expect(schedule).To(ContainSubstring("*it.Workflow = chaos.Spec"))
	g.Expect(schedule).NotTo(ContainSubstring("ScheduleTypeStatusCheck"))

	frontend, err := os.ReadFile(filepath.Join(uiDir, "zz_generated.frontend.chaos-mesh.ts"))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(frontend)).To(ContainSubstring("['HTTPChaos', 'httpChaos']"))
	g.Expect(string(frontend)).To(ContainSubstring("['PodChaos', 'podChaos']"))
	g.Expect(string(frontend)).NotTo(ContainSubstring("['Workflow'"))
}

func TestGeneratePreservesOutputsWhenInputIsInvalid(t *testing.T) {
	for _, example := range []struct{ name, source, wantErr string }{
		{"invalid Go syntax", "package v1alpha1\ntype broken struct {", "broken.go"},
		{"marker on a function", "package v1alpha1\n// +chaos-mesh:experiment\nfunc f() {}", "node is not a *ast.GenDecl"},
		{"marker on a variable", "package v1alpha1\n// +chaos-mesh:base\nvar v = 1", "node.Tok is not token.TYPE"},
	} {
		t.Run(example.name, func(t *testing.T) {
			g := NewWithT(t)
			root := t.TempDir()
			apiDir := filepath.Join(root, "api", "v1alpha1")
			g.Expect(os.MkdirAll(apiDir, 0o755)).To(Succeed())
			existing := filepath.Join(apiDir, "zz_generated.chaosmesh.go")
			g.Expect(os.WriteFile(existing, []byte("previous API definitions"), 0o600)).To(Succeed())
			g.Expect(os.WriteFile(filepath.Join(apiDir, "broken.go"), []byte(example.source), 0o600)).To(Succeed())
			g.Expect(generate(root)).To(MatchError(ContainSubstring(example.wantErr)))
			content, err := os.ReadFile(existing)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(string(content)).To(Equal("previous API definitions"))
		})
	}
}

func TestGenerateReportsOutputWriteFailure(t *testing.T) {
	g := NewWithT(t)
	root := t.TempDir()
	apiDir := filepath.Join(root, "api", "v1alpha1")
	g.Expect(os.MkdirAll(filepath.Join(apiDir, "zz_generated.chaosmesh.go"), 0o755)).To(Succeed())
	g.Expect(generate(root)).To(MatchError(ContainSubstring("write")))
	g.Expect(generate(t.TempDir())).To(MatchError(ContainSubstring("read chaos API declarations")))
}
