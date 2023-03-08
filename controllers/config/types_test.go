// Copyright 2021 Chaos Mesh Authors.
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

package config

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/chaos-mesh/chaos-mesh/pkg/config"
)

func TestValidations(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Namespace scoped")
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
						ClusterScoped:   false,
						TargetNamespace: "",
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
