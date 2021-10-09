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

package v1alpha1

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("physicalmachinechaos_webhook", func() {
	Context("webhook.Defaultor of physicalmachinechaos", func() {
		It("Default", func() {
			physicalMachineChaos := &PhysicalMachineChaos{
				Spec: PhysicalMachineChaosSpec{
					Action: "stress-cpu",
					PhysicalMachineSelector: PhysicalMachineSelector{
						Address: []string{
							"123.123.123.123:123",
							"234.234.234.234:234",
						},
					},
					ExpInfo: ExpInfo{
						UID: "",
					},
				},
			}
			physicalMachineChaos.Default()
			Expect(physicalMachineChaos.Spec.UID).ToNot(Equal(""))
			Expect(physicalMachineChaos.Spec.Address).To(BeEquivalentTo([]string{
				"http://123.123.123.123:123",
				"http://234.234.234.234:234",
			}))
		})
	})
	Context("webhook.Validator of physicalmachinechaos", func() {
		It("Validate", func() {
			testCases := []struct {
				chaos PhysicalMachineChaos
				err   string
			}{
				{
					PhysicalMachineChaos{
						Spec: PhysicalMachineChaosSpec{
							Action: "stress-cpu",
							PhysicalMachineSelector: PhysicalMachineSelector{
								Address: []string{""},
							},
							ExpInfo: ExpInfo{},
						},
					},
					"the address is empty",
				},
				{
					PhysicalMachineChaos{
						Spec: PhysicalMachineChaosSpec{
							Action: "stress-cpu",
							PhysicalMachineSelector: PhysicalMachineSelector{Address: []string{
								"123.123.123.123:123",
								"234.234.234.234:234",
							}},
							ExpInfo: ExpInfo{},
						},
					},
					"the configuration corresponding to action is empty",
				},
			}

			for _, testCase := range testCases {
				err := testCase.chaos.ValidateCreate()
				Expect(strings.Contains(err.Error(), testCase.err)).To(BeTrue())
			}
		})
	})
})
