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

package config

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"

	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/webhook/config/watcher"
)

func TestValidations(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Namespace scoped",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = Describe("Namespace-scoped Chaos", func() {
	Context("Namespace-scoped Chaos Configuration Validation", func() {
		It("Validation", func() {
			type TestCase struct {
				name        string
				config      config.ChaosControllerConfig
				expectValid bool
			}

			testCases := []TestCase{
				{
					name: "target namespace should not be empty with namespaced scope",
					config: config.ChaosControllerConfig{
						WatcherConfig: &watcher.Config{
							ClusterScoped:     false,
							TemplateNamespace: "",
							TargetNamespace:   "",
							TemplateLabels:    nil,
							ConfigLabels:      nil,
						},
						ClusterScoped:   false,
						TargetNamespace: "",
					},
					expectValid: false,
				},
				{
					name: "Watcher Config is always required",
					config: config.ChaosControllerConfig{
						WatcherConfig:   nil,
						ClusterScoped:   true,
						TargetNamespace: "",
					},
					expectValid: false,
				},
				{
					name: "clusterScope should keep constant",
					config: config.ChaosControllerConfig{
						WatcherConfig: &watcher.Config{
							ClusterScoped:     false,
							TemplateNamespace: "",
							TargetNamespace:   "",
							TemplateLabels:    nil,
							ConfigLabels:      nil,
						},
						ClusterScoped:   true,
						TargetNamespace: "",
					},
					expectValid: false,
				},
				{
					name: "ns should keep constant",
					config: config.ChaosControllerConfig{
						WatcherConfig: &watcher.Config{
							ClusterScoped:   false,
							TargetNamespace: "ns1",
						},
						ClusterScoped:   false,
						TargetNamespace: "ns2",
					},
					expectValid: false,
				},
			}

			for _, testCase := range testCases {
				By(testCase.name)
				err := validate(&testCase.config)
				if testCase.expectValid {
					Expect(err).NotTo(HaveOccurred())
				} else {
					Expect(err).To(HaveOccurred())
				}
			}
		})
	})
})
