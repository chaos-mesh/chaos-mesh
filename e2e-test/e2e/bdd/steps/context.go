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

package steps

import (
	"context"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/pkg/client/versioned"
)

// ScenarioContext holds all runtime state shared across step definitions within
// a single Godog scenario. A fresh instance is created for every scenario so
// that steps are fully isolated from each other.
type ScenarioContext struct {
	// Kubernetes clients
	Namespace  string
	KubeCli    kubernetes.Interface
	Cli        client.Client
	ClientSet  *versioned.Clientset
	HTTPClient http.Client

	// State captured between steps for PodChaos scenarios
	InitialPodUIDs map[string]string // pod name -> UID snapshot

	// State captured between steps for NetworkChaos scenarios
	NetworkPeers []*corev1.Pod
	Ports        []uint16

	// cancelFuncs holds per-scenario context cancellation functions
	cancelFuncs []context.CancelFunc
}

// NewScenarioContext creates a fresh ScenarioContext with the supplied
// clients. The namespace is typically a unique per-scenario namespace
// provisioned by the test suite.
func NewScenarioContext(
	ns string,
	kubeCli kubernetes.Interface,
	cli client.Client,
	clientSet *versioned.Clientset,
	httpClient http.Client,
) *ScenarioContext {
	return &ScenarioContext{
		Namespace:      ns,
		KubeCli:        kubeCli,
		Cli:            cli,
		ClientSet:      clientSet,
		HTTPClient:     httpClient,
		InitialPodUIDs: make(map[string]string),
	}
}

// Background returns a context derived from context.Background. The
// cancel function is registered so that the suite can cancel all
// in-flight operations when a scenario ends.
func (s *ScenarioContext) Background() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFuncs = append(s.cancelFuncs, cancel)
	return ctx, cancel
}

// Cleanup cancels all registered contexts. Call this in an AfterScenario hook.
func (s *ScenarioContext) Cleanup() {
	for _, cancel := range s.cancelFuncs {
		cancel()
	}
}
