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

package v1alpha1

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("physicalmachine_webhook", func() {
	Context("webhook.Defaultor of physicalmachine", func() {
		It("Default", func() {
			physicalMachine := &PhysicalMachine{
				Spec: PhysicalMachineSpec{
					Address: "123.123.123.123:123",
				},
			}
			physicalMachine.Default()
			Expect(physicalMachine.Spec.Address).To(BeEquivalentTo("http://123.123.123.123:123"))
		})
	})
	Context("webhook.Validator of physicalmachine", func() {
		It("Validate", func() {
			testCases := []struct {
				physicalMachine PhysicalMachine
				err             string
			}{
				{
					PhysicalMachine{
						Spec: PhysicalMachineSpec{
							Address: "",
						},
					},
					"the address is required",
				}, {
					PhysicalMachine{
						Spec: PhysicalMachineSpec{
							Address: "123",
						},
					},
					"the address is invalid",
				}, {
					PhysicalMachine{
						Spec: PhysicalMachineSpec{
							Address: "http://123.123.123.123:123",
						},
					},
					"",
				},
			}

			for _, testCase := range testCases {
				_, err := testCase.physicalMachine.ValidateCreate()
				if len(testCase.err) != 0 {
					Expect(err).To(HaveOccurred())
					Expect(strings.Contains(err.Error(), testCase.err)).To(BeTrue())
				} else {
					Expect(err).ToNot(HaveOccurred())
				}
			}
		})
	})
})
