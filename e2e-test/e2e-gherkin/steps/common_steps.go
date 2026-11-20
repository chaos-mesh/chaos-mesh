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
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type TestContext struct {
	KubeCli   kubernetes.Interface
	Client    client.Client
	Namespace string

	// State recorded within a scenario
	InitialPods       []corev1.Pod
	ContainerBaseline *recordedContainer
	OriginalImage     string
}

func (tc *TestContext) RegisterSteps(ctx *godog.ScenarioContext) {
	// Register common steps
	ctx.Step(`^a namespace is prepared$`, tc.aNamespaceIsPrepared)

	ctx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Minute)
		defer cancel()
		err := tc.cleanupNamespace(cleanupCtx)
		tc.Namespace = ""
		tc.InitialPods = nil
		tc.ContainerBaseline = nil
		tc.OriginalImage = ""
		return ctx, err
	})
}

func (tc *TestContext) cleanupNamespace(ctx context.Context) error {
	if tc.Namespace == "" {
		return nil
	}
	err := tc.KubeCli.CoreV1().Namespaces().Delete(ctx, tc.Namespace, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("delete namespace %s: %w", tc.Namespace, err)
	}
	err = wait.PollUntilContextCancel(ctx, 2*time.Second, true, func(ctx context.Context) (bool, error) {
		_, err := tc.KubeCli.CoreV1().Namespaces().Get(ctx, tc.Namespace, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, fmt.Errorf("get namespace %s: %w", tc.Namespace, err)
		}
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("wait for namespace %s deletion: %w", tc.Namespace, err)
	}
	return nil
}

func (tc *TestContext) aNamespaceIsPrepared(ctx context.Context) error {
	labels := map[string]string{
		"e2e-framework":                      "chaos-mesh",
		"pod-security.kubernetes.io/enforce": "privileged",
		"pod-security.kubernetes.io/warn":    "privileged",
		"pod-security.kubernetes.io/audit":   "privileged",
	}
	createdNs, err := framework.CreateTestingNS(ctx, "e2e-gherkin", tc.KubeCli, labels)
	if createdNs != nil {
		tc.Namespace = createdNs.Name
	}
	return err
}

func (tc *TestContext) parseChaosMode(mode string) (v1alpha1.SelectorMode, error) {
	switch strings.ToLower(mode) {
	case "one":
		return v1alpha1.OneMode, nil
	case "all":
		return v1alpha1.AllMode, nil
	case "fixed", "fixed-percent", "random-max-percent":
		return "", fmt.Errorf("chaos mode %q requires a value which is not yet supported in Gherkin steps", mode)
	default:
		return "", fmt.Errorf("unsupported chaos mode: %s", mode)
	}
}

func parseLabelSelector(selector string) (map[string]string, error) {
	parts := strings.Split(selector, "=")
	if len(parts) != 2 || parts[0] == "" {
		return nil, fmt.Errorf("invalid label selector format: %s", selector)
	}
	return map[string]string{parts[0]: parts[1]}, nil
}
