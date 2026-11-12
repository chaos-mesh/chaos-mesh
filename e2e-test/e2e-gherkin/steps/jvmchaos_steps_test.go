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
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cucumber/godog"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("JVM recovery evidence", func() {
	DescribeTable("uses the kubelet cursor and requests all subsequent logs", func(window string, success bool) {
		calls := 0
		tc := &TestContext{Namespace: "test"}
		tc.KubeCli = requestClient(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path == "/api/v1/namespaces/test/pods/java" {
				return jsonResponse(&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "java", UID: "pod-1"}}), nil
			}
			Expect(req.URL.Path).To(Equal("/api/v1/namespaces/test/pods/java/log"))
			Expect(req.URL.Query().Get("timestamps")).To(Equal("true"))
			calls++
			if calls == 1 {
				Expect(req.URL.Query().Get("tailLines")).To(Equal("1"))
				return textResponse("2026-01-01T00:00:00.123456789Z BOOM\n"), nil
			}
			Expect(req.URL.Query().Get("tailLines")).To(BeEmpty())
			// Kubernetes serializes SinceTime to second precision; the local cursor retains nanoseconds.
			Expect(req.URL.Query().Get("sinceTime")).To(Equal("2026-01-01T00:00:00Z"))
			return textResponse(window), nil
		})
		ctx, cancel := shortContext()
		defer cancel()
		err := tc.thePodShouldStopPrintingLogLinesContaining(ctx, "java", "BOOM")
		if success {
			Expect(err).NotTo(HaveOccurred())
		} else {
			Expect(err).To(HaveOccurred())
		}
		Expect(calls).To(Equal(2))
	},
		Entry("new healthy logs", "2026-01-01T00:00:01Z 1. Hello World\n2026-01-01T00:00:02Z 2. Hello World\n2026-01-01T00:00:03Z 3. Hello World\n", true),
		Entry("stale healthy logs", "2026-01-01T00:00:00.1Z 1. Hello World\n2026-01-01T00:00:00.11Z 2. Hello World\n2026-01-01T00:00:00.123456789Z 3. Hello World\n", false),
		Entry("diagnostics", "2026-01-01T00:00:01Z byteman diagnostic\n2026-01-01T00:00:02Z byteman diagnostic\n2026-01-01T00:00:03Z byteman diagnostic\n", false),
		Entry("late failure", "2026-01-01T00:00:01Z 1. Hello World\n2026-01-01T00:00:02Z 2. Hello World\n2026-01-01T00:00:03Z 3. Hello World\n2026-01-01T00:00:04Z BOOM\n", false))
})

var _ = Describe("Registered JVM recovery step", func() {
	run := func(tc *TestContext, ctx context.Context) (int, string) {
		var output bytes.Buffer
		suite := godog.TestSuite{
			Name: "jvm-recovery", ScenarioInitializer: tc.RegisterJVMChaosSteps,
			Options: &godog.Options{Format: "progress", NoColors: true, Output: &output, DefaultContext: ctx,
				FeatureContents: []godog.Feature{{Name: "recovery.feature", Contents: []byte("Feature: JVM recovery\nScenario: Recovery\nThen the pod named \"java\" should eventually stop printing log lines containing \"BOOM\"\n")}},
			},
		}
		code := suite.Run()
		return code, output.String()
	}
	baseline := "2026-01-01T00:00:00Z BOOM\n"
	healthy := "2026-01-01T00:00:01Z 1. Hello World\n2026-01-01T00:00:02Z 2. Hello World\n2026-01-01T00:00:03Z 3. Hello World\n"
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "java", UID: "pod-1"}}

	It("retries a temporary read failure without replacing the recovery baseline", func() {
		baselines := 0
		failed := false
		tc := &TestContext{Namespace: "test", KubeCli: requestClient(func(req *http.Request) (*http.Response, error) {
			if !strings.HasSuffix(req.URL.Path, "/log") {
				return jsonResponse(pod), nil
			}
			if req.URL.Query().Get("tailLines") == "1" {
				baselines++
				Expect(baselines).To(Equal(1), "successful retries must retain the original cursor")
				return textResponse(baseline), nil
			}
			if !failed {
				failed = true
				response := textResponse("temporary API failure")
				response.StatusCode = http.StatusServiceUnavailable
				return response, nil
			}
			Expect(req.URL.Query().Get("sinceTime")).To(Equal("2026-01-01T00:00:00Z"))
			return textResponse(healthy), nil
		})}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		code, output := run(tc, ctx)
		Expect(code).To(BeZero(), output)
		Expect(baselines).To(Equal(1))
		Expect(failed).To(BeTrue(), "the requested temporary failure must be reached")
	})

	It("rejects a Pod replaced while healthy logs are being read", func() {
		replaced := false
		tc := &TestContext{Namespace: "test", KubeCli: requestClient(func(req *http.Request) (*http.Response, error) {
			if !strings.HasSuffix(req.URL.Path, "/log") {
				current := pod.DeepCopy()
				if replaced {
					current.UID = "replacement"
				}
				return jsonResponse(current), nil
			}
			if req.URL.Query().Get("tailLines") == "1" {
				return textResponse(baseline), nil
			}
			replaced = true
			return textResponse(healthy), nil
		})}
		code, output := run(tc, context.Background())
		Expect(code).NotTo(BeZero(), output)
		Expect(output).To(ContainSubstring("was replaced during recovery"))
	})

	It("fails when Pod read errors persist after healthy logs until the caller deadline", func() {
		healthyRead := false
		tc := &TestContext{Namespace: "test", KubeCli: requestClient(func(req *http.Request) (*http.Response, error) {
			if !strings.HasSuffix(req.URL.Path, "/log") {
				if healthyRead {
					return nil, fmt.Errorf("API unavailable")
				}
				return jsonResponse(pod), nil
			}
			if req.URL.Query().Get("tailLines") == "1" {
				return textResponse(baseline), nil
			}
			healthyRead = true
			return textResponse(healthy), nil
		})}
		ctx, cancel := shortContext()
		defer cancel()
		code, output := run(tc, ctx)
		Expect(code).NotTo(BeZero(), output)
		Expect(ctx.Err()).To(MatchError(context.DeadlineExceeded))
	})

	It("does not accept recovery after cancellation during the final Pod read", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		healthyRead := false
		tc := &TestContext{Namespace: "test", KubeCli: requestClient(func(req *http.Request) (*http.Response, error) {
			if !strings.HasSuffix(req.URL.Path, "/log") {
				if healthyRead {
					cancel()
				}
				return jsonResponse(pod), nil
			}
			if req.URL.Query().Get("tailLines") == "1" {
				return textResponse(baseline), nil
			}
			healthyRead = true
			return textResponse(healthy), nil
		})}
		code, output := run(tc, ctx)
		Expect(code).NotTo(BeZero(), output)
		Expect(ctx.Err()).To(MatchError(context.Canceled))
	})
})

var _ = Describe("Missing JVM log cursor", func() {
	It("rejects an empty initial response before accepting old healthy logs", func() {
		calls := 0
		tc := &TestContext{Namespace: "test", KubeCli: requestClient(func(req *http.Request) (*http.Response, error) {
			if !strings.HasSuffix(req.URL.Path, "/log") {
				return jsonResponse(&corev1.Pod{ObjectMeta: metav1.ObjectMeta{UID: "pod-1"}}), nil
			}
			calls++
			if calls == 1 {
				return textResponse(""), nil
			}
			return textResponse("2020-01-01T00:00:01Z 1. Hello World\n2020-01-01T00:00:02Z 2. Hello World\n2020-01-01T00:00:03Z 3. Hello World\n"), nil
		})}
		Expect(tc.thePodShouldStopPrintingLogLinesContaining(context.Background(), "java", "BOOM")).To(HaveOccurred())
		Expect(calls).To(Equal(1))
	})
})
