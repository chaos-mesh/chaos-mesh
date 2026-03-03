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

package utils

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestEncodeClkIds(t *testing.T) {
	g := NewGomegaWithT(t)

	mask, err := EncodeClkIds([]string{"CLOCK_REALTIME"})
	g.Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)
	g.Expect(mask).Should(Equal(uint64(1)))

	mask, err = EncodeClkIds([]string{"CLOCK_REALTIME", "CLOCK_MONOTONIC"})
	g.Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)
	g.Expect(mask).Should(Equal(uint64(3)))

	mask, err = EncodeClkIds([]string{"CLOCK_MONOTONIC"})
	g.Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)
	g.Expect(mask).Should(Equal(uint64(2)))
}
