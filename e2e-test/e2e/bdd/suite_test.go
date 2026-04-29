// Copyright 2024 Chaos Mesh Authors.
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

// Package bdd provides a Godog-based BDD test layer for Chaos Mesh E2E tests.
//
// # Running
//
//	cd e2e-test
//	go test ./e2e/bdd/... -v \
//	    --kubeconfig=$HOME/.kube/config \
//	    --namespace=chaos-testing
//
// Feature files live under e2e/bdd/features/:
//
//	features/podchaos/    – PodChaos scenarios
//	features/networkchaos/ – NetworkChaos scenarios
//
// Step definitions live under e2e/bdd/steps/.
//
// # Architecture
//
// Each Godog scenario gets a fresh [steps.ScenarioContext] that holds
// Kubernetes clients, per-scenario state (e.g. initial pod UIDs), and a list
// of context.CancelFunc values cleaned up in the AfterScenario hook.
//
// The BDD layer intentionally reuses fixture helpers from
// e2e-test/pkg/fixture and util helpers from e2e-test/e2e/util so that
// test infrastructure is not duplicated.
package bdd

import (
	"context"
	"flag"
	"net/http"
	"os"
	"testing"

	"github.com/cucumber/godog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/bdd/steps"
	chaosmeshclient "github.com/chaos-mesh/chaos-mesh/pkg/client/versioned"
)

var (
	kubeconfig = flag.String("kubeconfig", os.Getenv("KUBECONFIG"), "path to kubeconfig")
	namespace  = flag.String("namespace", "chaos-testing", "namespace for test resources")
)

func TestBDD(t *testing.T) {
	flag.Parse()

	restCfg, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		t.Fatalf("build kubeconfig: %v", err)
	}

	kubeCli, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		t.Fatalf("build kube client: %v", err)
	}

	scheme := chaosmeshv1alpha1.SchemeBuilder.GroupVersion.Group
	_ = scheme // referenced to force import

	ctrlCli, err := client.New(restCfg, client.Options{})
	if err != nil {
		t.Fatalf("build controller-runtime client: %v", err)
	}

	clientSet, err := chaosmeshclient.NewForConfig(restCfg)
	if err != nil {
		t.Fatalf("build chaos-mesh clientset: %v", err)
	}

	// Ensure the test namespace exists.
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: *namespace}}
	_, _ = kubeCli.CoreV1().Namespaces().Create(t.Context(), ns, metav1.CreateOptions{})

	suite := godog.TestSuite{
		Name: "chaos-mesh-bdd",
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			ctx := steps.NewScenarioContext(
				*namespace,
				kubeCli,
				ctrlCli,
				clientSet,
				http.Client{},
			)

			steps.RegisterPodChaosSteps(sc, ctx)
			steps.RegisterNetworkChaosSteps(sc, ctx)

			sc.After(func(_ context.Context, _ *godog.Scenario, err error) (context.Context, error) {
				ctx.Cleanup()
				return nil, err
			})
		},
		Options: &godog.Options{
			Format:    "pretty",
			Paths:     []string{"features"},
			Randomize: 0,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("BDD scenarios failed")
	}
}
