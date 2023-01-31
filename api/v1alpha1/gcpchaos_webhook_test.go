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
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("gcpchaos_webhook", func() {
	Context("ChaosValidator of gcpchaos", func() {
		It("Validate", func() {

			type TestCase struct {
				name    string
				chaos   GCPChaos
				execute func(chaos *GCPChaos) error
				expect  string
			}
			tcs := []TestCase{
				{
					name: "simple ValidateCreate for DiskLoss",
					chaos: GCPChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo1",
						},
						Spec: GCPChaosSpec{
							Action: DiskLoss,
						},
					},
					execute: func(chaos *GCPChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "unknow action",
					chaos: GCPChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo6",
						},
					},
					execute: func(chaos *GCPChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
			}

			for _, tc := range tcs {
				err := tc.execute(&tc.chaos)
				if tc.expect == "error" {
					if err == nil {
						fmt.Println("Aha")
					}
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
			}
		})
	})
})
