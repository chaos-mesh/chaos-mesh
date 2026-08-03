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
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/pointer"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/pkg/fixture"
)

func (tc *TestContext) RegisterJVMChaosSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^a java pod named "([^"]*)" printing one log line per second is running$`, tc.aJavaPodPrintingLogsIsRunning)
	ctx.Step(`^a JVM latency chaos named "([^"]*)" with (\d+) ms latency is applied to class "([^"]*)" method "([^"]*)" for pods with label "([^"]*)"$`, tc.aJVMLatencyChaosIsApplied)
	ctx.Step(`^a JVM exception chaos named "([^"]*)" throwing exception "([^"]*)" with message "([^"]*)" is applied to class "([^"]*)" method "([^"]*)" for pods with label "([^"]*)"$`, tc.aJVMExceptionChaosIsApplied)
	ctx.Step(`^a JVM ruleData chaos named "([^"]*)" modifying method "([^"]*)" of class "([^"]*)" to return "([^"]*)" is applied to pods with label "([^"]*)"$`, tc.aJVMRuleDataChaosIsApplied)
	ctx.Step(`^the JVM chaos "([^"]*)" is deleted$`, tc.theJVMChaosIsDeleted)
	ctx.Step(`^the pod named "([^"]*)" should eventually print log lines with intervals longer than (\d+) milliseconds$`, tc.thePodShouldPrintLogLinesWithLongIntervals)
	ctx.Step(`^the pod named "([^"]*)" should eventually print a log line containing "([^"]*)"$`, tc.thePodShouldPrintLogLineContaining)
	ctx.Step(`^the pod named "([^"]*)" should eventually stop printing log lines containing "([^"]*)"$`, tc.thePodShouldStopPrintingLogLinesContaining)
}

func (tc *TestContext) aJavaPodPrintingLogsIsRunning(name string) error {
	pod := fixture.NewJavaHelloWorldPod(name, tc.Namespace)
	_, err := tc.KubeCli.CoreV1().Pods(tc.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	if err := tc.waitPodRunning(name); err != nil {
		return err
	}
	// Wait until the java process starts printing logs, so the java process
	// is ready before chaos is applied.
	return wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 5*time.Minute, true, func(ctx context.Context) (done bool, err error) {
		timestamps, err := tc.podLogTimestamps(ctx, name, 5)
		if err != nil {
			return false, nil
		}
		return len(timestamps) > 0, nil
	})
}

// newJVMChaos builds a JVMChaos object selecting all pods with the given
// label in the scenario namespace. The caller fills in action fields.
func (tc *TestContext) newJVMChaos(name, labelKeyVal string, param v1alpha1.JVMParameter, action v1alpha1.JVMChaosAction) (*v1alpha1.JVMChaos, error) {
	parts := strings.Split(labelKeyVal, "=")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid label selector format: %s", labelKeyVal)
	}

	return &v1alpha1.JVMChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: tc.Namespace,
		},
		Spec: v1alpha1.JVMChaosSpec{
			Action:       action,
			JVMParameter: param,
			ContainerSelector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Selector: v1alpha1.PodSelectorSpec{
						GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
							Namespaces: []string{tc.Namespace},
							LabelSelectors: map[string]string{
								parts[0]: parts[1],
							},
						},
					},
					Mode: v1alpha1.AllMode,
				},
			},
		},
	}, nil
}

func (tc *TestContext) aJVMLatencyChaosIsApplied(name string, latencyMs int, class, method, labelKeyVal string) error {
	param := v1alpha1.JVMParameter{
		JVMClassMethodSpec: v1alpha1.JVMClassMethodSpec{
			Class:  class,
			Method: method,
		},
		LatencyDuration: latencyMs,
	}
	jvmChaos, err := tc.newJVMChaos(name, labelKeyVal, param, v1alpha1.JVMLatencyAction)
	if err != nil {
		return err
	}
	return tc.Client.Create(context.TODO(), jvmChaos)
}

func (tc *TestContext) aJVMExceptionChaosIsApplied(name, exceptionClass, message, class, method, labelKeyVal string) error {
	param := v1alpha1.JVMParameter{
		JVMClassMethodSpec: v1alpha1.JVMClassMethodSpec{
			Class:  class,
			Method: method,
		},
		ThrowException: fmt.Sprintf("%s(%q)", exceptionClass, message),
	}
	jvmChaos, err := tc.newJVMChaos(name, labelKeyVal, param, v1alpha1.JVMExceptionAction)
	if err != nil {
		return err
	}
	return tc.Client.Create(context.TODO(), jvmChaos)
}

func (tc *TestContext) aJVMRuleDataChaosIsApplied(name, method, class, returnValue, labelKeyVal string) error {
	ruleData := fmt.Sprintf(
		"RULE modify return value\nCLASS %s\nMETHOD %s\nAT ENTRY\nIF true\nDO\n    return %s\nENDRULE",
		class, method, returnValue,
	)
	param := v1alpha1.JVMParameter{
		RuleData: ruleData,
	}
	jvmChaos, err := tc.newJVMChaos(name, labelKeyVal, param, v1alpha1.JVMRuleDataAction)
	if err != nil {
		return err
	}
	return tc.Client.Create(context.TODO(), jvmChaos)
}

func (tc *TestContext) theJVMChaosIsDeleted(name string) error {
	jvmChaos := &v1alpha1.JVMChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: tc.Namespace,
		},
	}
	return tc.Client.Delete(context.TODO(), jvmChaos)
}

func (tc *TestContext) thePodShouldPrintLogLinesWithLongIntervals(name string, intervalMs int) error {
	expected := time.Duration(intervalMs) * time.Millisecond
	return wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 5*time.Minute, true, func(ctx context.Context) (done bool, err error) {
		// Read more lines than needed: when the byteman agent runs in
		// verbose mode, its own output is mixed into stdout, and
		// podLogTimestamps only keeps the application log lines.
		timestamps, err := tc.podLogTimestamps(ctx, name, 30)
		if err != nil {
			return false, nil
		}
		if len(timestamps) < 3 {
			return false, nil
		}
		timestamps = timestamps[len(timestamps)-3:]
		// The chaos takes effect when consecutive log lines are separated by
		// more than the injected latency.
		for i := 1; i < len(timestamps); i++ {
			if timestamps[i].Sub(timestamps[i-1]) <= expected {
				return false, nil
			}
		}
		return true, nil
	})
}

func (tc *TestContext) thePodShouldPrintLogLineContaining(name, message string) error {
	return wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 5*time.Minute, true, func(ctx context.Context) (done bool, err error) {
		lines, err := tc.podLogTail(ctx, name, 10)
		if err != nil {
			return false, nil
		}
		for _, line := range lines {
			if strings.Contains(line, message) {
				return true, nil
			}
		}
		return false, nil
	})
}

func (tc *TestContext) thePodShouldStopPrintingLogLinesContaining(name, message string) error {
	return wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 5*time.Minute, true, func(ctx context.Context) (done bool, err error) {
		// Look at the last few lines only: old lines from before the
		// recovery still contain the message.
		lines, err := tc.podLogTail(ctx, name, 3)
		if err != nil {
			return false, nil
		}
		if len(lines) == 0 {
			return false, nil
		}
		for _, line := range lines {
			if strings.Contains(line, message) {
				return false, nil
			}
		}
		return true, nil
	})
}

// podLogTail reads the last lines of the pod log.
func (tc *TestContext) podLogTail(ctx context.Context, name string, tailLines int64) ([]string, error) {
	req := tc.KubeCli.CoreV1().Pods(tc.Namespace).GetLogs(name, &corev1.PodLogOptions{
		TailLines: pointer.Int64(tailLines),
	})
	data, err := req.DoRaw(ctx)
	if err != nil {
		return nil, err
	}

	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

// podLogTimestamps reads the last lines of the pod log and returns the
// timestamps of the "Hello World" lines printed by the java test app.
func (tc *TestContext) podLogTimestamps(ctx context.Context, name string, tailLines int64) ([]time.Time, error) {
	req := tc.KubeCli.CoreV1().Pods(tc.Namespace).GetLogs(name, &corev1.PodLogOptions{
		Timestamps: true,
		TailLines:  pointer.Int64(tailLines),
	})
	data, err := req.DoRaw(ctx)
	if err != nil {
		return nil, err
	}

	var timestamps []time.Time
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "Hello World") {
			continue
		}
		fields := strings.SplitN(line, " ", 2)
		if len(fields) != 2 {
			continue
		}
		ts, err := time.Parse(time.RFC3339Nano, fields[0])
		if err != nil {
			continue
		}
		timestamps = append(timestamps, ts)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return timestamps, nil
}
