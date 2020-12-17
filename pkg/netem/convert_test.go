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

package netem

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// These tests are written in BDD-style using Ginkgo framework. Refer to
// http://onsi.github.io/ginkgo to learn more.

var _ = Describe("NetworkChaos", func() {
	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context("convertUnitToBytes", func() {
		It("should convert number with unit successfully", func() {
			n, err := convertUnitToBytes("  10   mbPs  ")
			Expect(err).Should(Succeed())
			Expect(n).To(Equal(uint64(10 * 1024 * 1024)))
		})

		It("should return error with invalid unit", func() {
			n, err := convertUnitToBytes(" 10 cpbs")
			Expect(err).Should(HaveOccurred())
			Expect(n).To(Equal(uint64(0)))
		})
	})
})
