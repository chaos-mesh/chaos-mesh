// Copyright 2026 Chaos Mesh Authors.
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

package steps

import (
	"bytes"
	"context"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	restfake "k8s.io/client-go/rest/fake"
)

func TestSteps(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gherkin steps")
}

func requestClient(handler func(*http.Request) (*http.Response, error)) kubernetes.Interface {
	cli, err := kubernetes.NewForConfigAndClient(&rest.Config{Host: "http://steps.invalid"}, restfake.CreateHTTPClient(handler))
	Expect(err).NotTo(HaveOccurred())
	return cli
}

func textResponse(body string) *http.Response {
	return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"text/plain"}}, Body: io.NopCloser(bytes.NewBufferString(body))}
}

func jsonResponse(value any) *http.Response {
	if status, ok := value.(*metav1.Status); ok {
		status.TypeMeta = metav1.TypeMeta{APIVersion: "v1", Kind: "Status"}
	}
	body, err := json.Marshal(value)
	Expect(err).NotTo(HaveOccurred())
	response := textResponse(string(body))
	response.Header.Set("Content-Type", "application/json")
	return response
}

func shortContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Millisecond)
}

var _ = Describe("Feature bindings", func() {
	It("binds every checked-in feature step exactly once", func() {
		var output bytes.Buffer
		tc := &TestContext{}
		suite := godog.TestSuite{Name: "bindings", ScenarioInitializer: func(sc *godog.ScenarioContext) {
			tc.RegisterSteps(sc)
			tc.RegisterPodChaosSteps(sc)
			tc.RegisterJVMChaosSteps(sc)
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) { return ctx, godog.ErrPending })
		}, Options: &godog.Options{Format: "progress", NoColors: true, Output: &output, Paths: []string{"../features"}, Strict: true}}
		Expect(suite.Run()).To(Equal(1), output.String())
		Expect(output.String()).NotTo(ContainSubstring("undefined"))
		Expect(output.String()).NotTo(ContainSubstring("ambiguous"))
		sources, err := filepath.Glob("*_steps.go")
		Expect(err).NotTo(HaveOccurred())
		var patterns []*regexp.Regexp
		for _, source := range sources {
			parsed, err := parser.ParseFile(token.NewFileSet(), source, nil, 0)
			Expect(err).NotTo(HaveOccurred())
			ast.Inspect(parsed, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				method, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || method.Sel.Name != "Step" {
					return true
				}
				literal, ok := call.Args[0].(*ast.BasicLit)
				Expect(ok).To(BeTrue())
				expression, err := strconv.Unquote(literal.Value)
				Expect(err).NotTo(HaveOccurred())
				patterns = append(patterns, regexp.MustCompile(expression))
				return true
			})
		}
		features, err := filepath.Glob("../features/*.feature")
		Expect(err).NotTo(HaveOccurred())
		checked := 0
		for _, feature := range features {
			contents, err := os.ReadFile(feature)
			Expect(err).NotTo(HaveOccurred())
			for _, line := range strings.Split(string(contents), "\n") {
				keyword, text, ok := strings.Cut(strings.TrimSpace(line), " ")
				if !ok || (keyword != "Given" && keyword != "When" && keyword != "Then" && keyword != "And" && keyword != "But") {
					continue
				}
				matches := 0
				for _, pattern := range patterns {
					if pattern.MatchString(text) {
						matches++
					}
				}
				Expect(matches).To(Equal(1), text)
				checked++
			}
		}
		Expect(checked).To(BeNumerically(">", 0))
	})
})
