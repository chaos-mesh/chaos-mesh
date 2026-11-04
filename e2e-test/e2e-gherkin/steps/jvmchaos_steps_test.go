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
	cursor := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	DescribeTable("evaluates the entire fresh application log window", func(messages []string, healthy bool) {
		lines := make([]podLogLine, 0, len(messages))
		for i, message := range messages {
			lines = append(lines, podLogLine{Timestamp: cursor.Add(time.Duration(i+1) * time.Second), Message: message})
		}
		Expect(recoveredJVMLogs(lines, cursor, "BOOM")).To(Equal(healthy))
	},
		Entry("fresh healthy lines", []string{"1. Hello World", "2. Hello World", "3. Hello World"}, true),
		Entry("empty logs", []string{}, false),
		Entry("diagnostics only", []string{"byteman diagnostic 1", "byteman diagnostic 2", "byteman diagnostic 3"}, false),
		Entry("healthy text inside diagnostics", []string{"println(1. Hello World)", "rule 2. Hello World", "3. Hello World trace"}, false),
		Entry("only two healthy lines", []string{"1. Hello World", "byteman diagnostic", "2. Hello World"}, false),
		Entry("diagnostics do not break healthy output", []string{"1. Hello World", "byteman diagnostic", "2. Hello World", "3. Hello World"}, true),
		Entry("later fault invalidates earlier health", []string{"1. Hello World", "2. Hello World", "3. Hello World", "BOOM"}, false),
		Entry("other application exceptions reset recovery", []string{"1. Hello World", "2. Hello World", "Got an exception! other failure", "3. Hello World"}, false),
		Entry("three healthy lines after final fault", []string{"BOOM", "1. Hello World", "2. Hello World", "3. Hello World"}, true))

	It("rejects old logs and repeated timestamps", func() {
		old := []podLogLine{{Timestamp: cursor.Add(-time.Second), Message: "1. Hello World"}, {Timestamp: cursor, Message: "2. Hello World"}, {Timestamp: cursor.Add(time.Second), Message: "3. Hello World"}}
		Expect(recoveredJVMLogs(old, cursor, "BOOM")).To(BeFalse())
		repeated := []podLogLine{{Timestamp: cursor.Add(time.Second), Message: "1. Hello World"}, {Timestamp: cursor.Add(time.Second), Message: "1. Hello World"}, {Timestamp: cursor.Add(time.Second), Message: "1. Hello World"}}
		Expect(recoveredJVMLogs(repeated, cursor, "BOOM")).To(BeFalse())
		Expect(recoveredJVMLogs(repeated, cursor, "BOOM")).To(BeFalse())
	})

	It("orders the window before checking the most recent fault", func() {
		lines := []podLogLine{{Timestamp: cursor.Add(4 * time.Second), Message: "BOOM"}, {Timestamp: cursor.Add(time.Second), Message: "1. Hello World"}, {Timestamp: cursor.Add(2 * time.Second), Message: "2. Hello World"}, {Timestamp: cursor.Add(3 * time.Second), Message: "3. Hello World"}}
		Expect(recoveredJVMLogs(lines, cursor, "BOOM")).To(BeFalse())
	})

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

	It("rejects a replacement pod during recovery", func() {
		gets := 0
		tc := &TestContext{Namespace: "test"}
		tc.KubeCli = requestClient(func(req *http.Request) (*http.Response, error) {
			if strings.HasSuffix(req.URL.Path, "/log") {
				return textResponse("2026-01-01T00:00:00Z BOOM\n"), nil
			}
			gets++
			pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "java", UID: "old"}}
			if gets > 1 {
				pod.UID = "new"
			}
			return jsonResponse(pod), nil
		})
		Expect(tc.thePodShouldStopPrintingLogLinesContaining(context.Background(), "java", "BOOM")).To(HaveOccurred())
	})

	It("shares log transport while retaining tail and latency semantics", func() {
		tc := &TestContext{Namespace: "test"}
		tc.KubeCli = requestClient(func(req *http.Request) (*http.Response, error) {
			Expect(req.URL.Query().Get("tailLines")).To(Equal("30"))
			if req.URL.Query().Get("timestamps") == "true" {
				return textResponse("2026-01-01T00:00:00Z 1. Hello World\n2026-01-01T00:00:01Z byteman Hello World\n2026-01-01T00:00:02Z 2. Hello World\n2026-01-01T00:00:04Z 3. Hello World\n"), nil
			}
			return textResponse("BOOM\ndiagnostic\n"), nil
		})
		lines, err := tc.podLogTail(context.Background(), "java", 30)
		Expect(err).NotTo(HaveOccurred())
		Expect(lines).To(Equal([]string{"BOOM", "diagnostic"}))
		timestamps, err := tc.podLogTimestamps(context.Background(), "java", 30)
		Expect(err).NotTo(HaveOccurred())
		Expect(timestamps).To(Equal([]time.Time{cursor, cursor.Add(2 * time.Second), cursor.Add(4 * time.Second)}))
		Expect(tc.thePodShouldPrintLogLinesWithLongIntervals(context.Background(), "java", 1000)).To(Succeed())
	})

	It("reports malformed timestamped logs", func() {
		tc := &TestContext{Namespace: "test", KubeCli: requestClient(func(*http.Request) (*http.Response, error) { return textResponse("invalid 1. Hello World\n"), nil })}
		_, err := tc.podLogTimestamps(context.Background(), "java", 3)
		Expect(err).To(HaveOccurred())
	})

	It("retains message detection and MySQL error response matching", func() {
		tc := &TestContext{Namespace: "test", KubeCli: requestClient(func(req *http.Request) (*http.Response, error) {
			if strings.HasSuffix(req.URL.Path, "/log") {
				return textResponse("BOOM\n"), nil
			}
			Expect(req.URL.Query().Get("sql")).To(Equal("select 1"))
			response := textResponse("BOOM mysql fault")
			response.StatusCode = http.StatusInternalServerError
			return response, nil
		})}
		Expect(tc.thePodShouldPrintLogLineContaining(context.Background(), "java", "BOOM")).To(Succeed())
		Expect(tc.theMySQLQueryAppReturns(context.Background(), "BOOM", "select 1")).To(Succeed())
	})

	It("propagates log API failures", func() {
		tc := &TestContext{Namespace: "test", KubeCli: requestClient(func(req *http.Request) (*http.Response, error) {
			if strings.HasSuffix(req.URL.Path, "/log") {
				return nil, fmt.Errorf("log unavailable")
			}
			return jsonResponse(&corev1.Pod{ObjectMeta: metav1.ObjectMeta{UID: "pod-1"}}), nil
		})}
		ctx, cancel := shortContext()
		defer cancel()
		Expect(tc.thePodShouldStopPrintingLogLinesContaining(ctx, "java", "BOOM")).To(HaveOccurred())
	})
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

	DescribeTable("retries a temporary read failure without replacing the recovery baseline", func(logRequest bool, failAt int) {
		podReads, logReads, baselines := 0, 0, 0
		tc := &TestContext{Namespace: "test", KubeCli: requestClient(func(req *http.Request) (*http.Response, error) {
			isLog := strings.HasSuffix(req.URL.Path, "/log")
			if isLog {
				logReads++
			} else {
				podReads++
			}
			if isLog == logRequest && ((isLog && logReads == failAt) || (!isLog && podReads == failAt)) {
				response := textResponse("temporary API failure")
				response.StatusCode = http.StatusServiceUnavailable
				return response, nil
			}
			if !isLog {
				return jsonResponse(pod), nil
			}
			if req.URL.Query().Get("tailLines") == "1" {
				baselines++
				Expect(baselines).To(Equal(1), "successful retries must retain the original cursor")
				return textResponse(baseline), nil
			}
			Expect(req.URL.Query().Get("sinceTime")).To(Equal("2026-01-01T00:00:00Z"))
			return textResponse(healthy), nil
		})}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		code, output := run(tc, ctx)
		Expect(code).To(BeZero(), output)
		Expect(baselines).To(Equal(1))
	},
		Entry("initial log query", true, 1),
		Entry("polling log query", true, 2),
		Entry("initial Pod query", false, 1),
		Entry("polling Pod query", false, 2))

	It("bounds every request with the same recovery deadline", func() {
		var deadlines []time.Time
		started := time.Now()
		tc := &TestContext{Namespace: "test", KubeCli: requestClient(func(req *http.Request) (*http.Response, error) {
			deadline, bounded := req.Context().Deadline()
			Expect(bounded).To(BeTrue(), "initial Pod and log reads need the recovery timeout too")
			deadlines = append(deadlines, deadline)
			if !strings.HasSuffix(req.URL.Path, "/log") {
				return jsonResponse(pod), nil
			}
			if req.URL.Query().Get("tailLines") == "1" {
				return textResponse(baseline), nil
			}
			return textResponse(healthy), nil
		})}
		code, output := run(tc, context.Background())
		Expect(code).To(BeZero(), output)
		Expect(deadlines).NotTo(BeEmpty())
		Expect(deadlines[0]).To(BeTemporally("~", started.Add(5*time.Minute), time.Second))
		for _, deadline := range deadlines {
			Expect(deadline).To(Equal(deadlines[0]))
		}
	})

	It("retains the original Pod UID after an initial log failure", func() {
		logReads := 0
		tc := &TestContext{Namespace: "test", KubeCli: requestClient(func(req *http.Request) (*http.Response, error) {
			if !strings.HasSuffix(req.URL.Path, "/log") {
				current := pod.DeepCopy()
				if logReads > 0 {
					current.UID = "replacement"
				}
				return jsonResponse(current), nil
			}
			logReads++
			if logReads == 1 {
				return nil, fmt.Errorf("log unavailable")
			}
			if req.URL.Query().Get("tailLines") == "1" {
				return textResponse(baseline), nil
			}
			return textResponse(healthy), nil
		})}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		code, output := run(tc, ctx)
		Expect(code).NotTo(BeZero(), output)
		Expect(output).To(ContainSubstring("was replaced during recovery"))
	})

	DescribeTable("never accepts malformed log evidence", func(initial bool) {
		tc := &TestContext{Namespace: "test", KubeCli: requestClient(func(req *http.Request) (*http.Response, error) {
			if !strings.HasSuffix(req.URL.Path, "/log") {
				return jsonResponse(pod), nil
			}
			if !initial && req.URL.Query().Get("tailLines") == "1" {
				return textResponse(baseline), nil
			}
			return textResponse("invalid 1. Hello World\ninvalid 2. Hello World\ninvalid 3. Hello World\n"), nil
		})}
		ctx, cancel := shortContext()
		defer cancel()
		code, output := run(tc, ctx)
		Expect(code).NotTo(BeZero(), output)
	}, Entry("initial log query", true), Entry("polling log query", false))

	It("rejects a Pod without a baseline UID", func() {
		tc := &TestContext{Namespace: "test", KubeCli: requestClient(func(req *http.Request) (*http.Response, error) {
			Expect(req.URL.Path).NotTo(HaveSuffix("/log"))
			return jsonResponse(&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "java"}}), nil
		})}
		code, output := run(tc, context.Background())
		Expect(code).NotTo(BeZero(), output)
		Expect(output).To(ContainSubstring("has no UID"))
	})

	DescribeTable("fails when read errors persist until the caller deadline", func(logRequest bool, failAt int) {
		podReads, logReads := 0, 0
		tc := &TestContext{Namespace: "test", KubeCli: requestClient(func(req *http.Request) (*http.Response, error) {
			isLog := strings.HasSuffix(req.URL.Path, "/log")
			if isLog {
				logReads++
			} else {
				podReads++
			}
			if isLog == logRequest && ((isLog && logReads >= failAt) || (!isLog && podReads >= failAt)) {
				return nil, fmt.Errorf("API unavailable")
			}
			if !isLog {
				return jsonResponse(pod), nil
			}
			return textResponse(baseline), nil
		})}
		ctx, cancel := shortContext()
		defer cancel()
		code, output := run(tc, ctx)
		Expect(code).NotTo(BeZero(), output)
		Expect(ctx.Err()).To(MatchError(context.DeadlineExceeded))
	},
		Entry("initial log query", true, 1),
		Entry("polling log query", true, 2),
		Entry("initial Pod query", false, 1),
		Entry("polling Pod query", false, 2))

	DescribeTable("does not accept recovery after cancellation during a read", func(cancelAt int) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		reads := 0
		tc := &TestContext{Namespace: "test", KubeCli: requestClient(func(req *http.Request) (*http.Response, error) {
			reads++
			if reads == cancelAt {
				cancel()
			}
			if !strings.HasSuffix(req.URL.Path, "/log") {
				return jsonResponse(pod), nil
			}
			if req.URL.Query().Get("tailLines") == "1" {
				return textResponse(baseline), nil
			}
			return textResponse(healthy), nil
		})}
		code, output := run(tc, ctx)
		Expect(code).NotTo(BeZero(), output)
		Expect(ctx.Err()).To(MatchError(context.Canceled))
	},
		Entry("initial Pod query", 1),
		Entry("initial log query", 2),
		Entry("polling Pod query", 3),
		Entry("polling log query", 4))
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
