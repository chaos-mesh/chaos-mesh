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
	InitialPods []corev1.Pod
}

func (tc *TestContext) RegisterSteps(ctx *godog.ScenarioContext) {
	// Register common steps
	ctx.Step(`^a namespace is prepared$`, tc.aNamespaceIsPrepared)

	// Hook to clean up namespace after scenario runs
	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if tc.Namespace != "" {
			_ = tc.KubeCli.CoreV1().Namespaces().Delete(ctx, tc.Namespace, metav1.DeleteOptions{})
			// Wait for namespace deletion without using Ginkgo-dependent framework function
			_ = wait.PollUntilContextTimeout(ctx, 2*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
				_, err := tc.KubeCli.CoreV1().Namespaces().Get(ctx, tc.Namespace, metav1.GetOptions{})
				if err != nil && apierrors.IsNotFound(err) {
					return true, nil
				}
				return false, nil
			})
		}
		return ctx, nil
	})
}

func (tc *TestContext) aNamespaceIsPrepared() error {
	labels := map[string]string{
		"e2e-framework":                      "chaos-mesh",
		"pod-security.kubernetes.io/enforce": "privileged",
		"pod-security.kubernetes.io/warn":    "privileged",
		"pod-security.kubernetes.io/audit":   "privileged",
	}
	createdNs, err := framework.CreateTestingNS(context.TODO(), "e2e-gherkin", tc.KubeCli, labels)
	if err != nil {
		return err
	}
	tc.Namespace = createdNs.Name
	return nil
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
