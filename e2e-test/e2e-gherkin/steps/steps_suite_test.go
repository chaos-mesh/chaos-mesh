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
	"fmt"
	"io"
	"net/http"
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

type bindingFormatter struct {
	*godog.BaseFmt
	output  io.Writer
	checked int
}

func init() {
	godog.Format("step-bindings", "Check feature step bindings", func(name string, out io.Writer) godog.Formatter {
		return &bindingFormatter{BaseFmt: godog.NewBaseFmt(name, out), output: out}
	})
}

func (f *bindingFormatter) Defined(_ *godog.Scenario, step *godog.Step, definition *godog.StepDefinition) {
	f.checked++
	if definition == nil {
		fmt.Fprintf(f.output, "step %q must match exactly one definition\n", step.Text)
	}
}

func (f *bindingFormatter) Summary() {
	if f.checked == 0 {
		fmt.Fprintln(f.output, "no feature steps checked")
	}
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
		}, Options: &godog.Options{Format: "step-bindings", NoColors: true, Output: &output, Paths: []string{"../features"}, Strict: true}}
		Expect(suite.Run()).To(Equal(1), output.String())
		Expect(output.String()).To(BeEmpty())
	})
})
