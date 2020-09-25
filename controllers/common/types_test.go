// Copyright 2020 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/webhook/config/watcher"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Namespace scoped",
		[]Reporter{envtest.NewlineReporter{}})
}

var _ = Describe("Namespace-scoped Chaos", func() {
	Context("Namespace-scoped Chaos Configuration Validation", func() {
		It("target namespace should not be empty with namespaced scope", func() {
			c := config.ChaosControllerConfig{
				WatcherConfig: &watcher.Config{
					ClusterScoped:     false,
					TemplateNamespace: "",
					TargetNamespace:   "",
					TemplateLabels:    nil,
					ConfigLabels:      nil,
				},
				ClusterScoped:   false,
				TargetNamespace: "",
			}
			By("validating")
			Expect(validate(&c)).ShouldNot(Succeed())
		})

		It("Watcher Config is always required", func() {
			c := config.ChaosControllerConfig{
				WatcherConfig:   nil,
				ClusterScoped:   true,
				TargetNamespace: "",
			}
			By("validating")
			Expect(validate(&c)).ShouldNot(Succeed())
		})
		It("clusterScope should keep constant", func() {
			c := config.ChaosControllerConfig{
				WatcherConfig: &watcher.Config{
					ClusterScoped:     false,
					TemplateNamespace: "",
					TargetNamespace:   "",
					TemplateLabels:    nil,
					ConfigLabels:      nil,
				},
				ClusterScoped:   true,
				TargetNamespace: "",
			}
			By("validating")
			Expect(validate(&c)).ShouldNot(Succeed())
		})
		It("ns should keep constant", func() {
			c := config.ChaosControllerConfig{
				WatcherConfig: &watcher.Config{
					ClusterScoped:   false,
					TargetNamespace: "ns1",
				},
				ClusterScoped:   false,
				TargetNamespace: "ns2",
			}
			By("validating")
			Expect(validate(&c)).ShouldNot(Succeed())
		})
	})
})
