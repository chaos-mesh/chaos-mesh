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

	"github.com/cucumber/godog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
			_ = tc.KubeCli.CoreV1().Namespaces().Delete(context.Background(), tc.Namespace, metav1.DeleteOptions{})
		}
		return ctx, nil
	})
}

func (tc *TestContext) aNamespaceIsPrepared() error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "e2e-gherkin-",
			Labels: map[string]string{
				"e2e-framework":                      "chaos-mesh",
				"pod-security.kubernetes.io/enforce": "privileged",
				"pod-security.kubernetes.io/warn":    "privileged",
				"pod-security.kubernetes.io/audit":   "privileged",
			},
		},
	}
	createdNs, err := tc.KubeCli.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	tc.Namespace = createdNs.Name
	return nil
}
