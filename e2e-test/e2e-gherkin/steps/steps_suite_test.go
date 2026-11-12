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
	"io"
	"net/http"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
	body, err := json.Marshal(value)
	Expect(err).NotTo(HaveOccurred())
	response := textResponse(string(body))
	response.Header.Set("Content-Type", "application/json")
	return response
}

func shortContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Millisecond)
}
