// Copyright 2026 Chaos Mesh Authors.
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

	"github.com/chaos-mesh/chaos-mesh/api/genericwebhook"
)

var _ = Describe("MemoryStressor time", func() {
	DescribeTable("passes time to memStress",
		func(stressor MemoryStressor, want string) {
			stressors := &Stressors{MemoryStressor: &stressor}
			_, memory, err := stressors.Normalize()
			Expect(err).NotTo(HaveOccurred())
			Expect(memory).To(Equal(want))
		},
		Entry("no time", MemoryStressor{Stressor: Stressor{Workers: 1}, Size: "1GiB"}, " --workers 1 --size 1GiB"),
		Entry("time", MemoryStressor{Stressor: Stressor{Workers: 2}, Size: "50%", Time: "5m"}, " --workers 2 --size 50% --time 5m"),
		Entry("time without size", MemoryStressor{Stressor: Stressor{Workers: 1}, Time: "30s"}, " --workers 1 --time 30s"),
	)

	DescribeTable("validates time as a duration",
		func(time string, invalid bool) {
			obj := &StressChaos{
				Spec: StressChaosSpec{
					ContainerSelector: ContainerSelector{PodSelector: PodSelector{Mode: AllMode}},
					Stressors: &Stressors{MemoryStressor: &MemoryStressor{
						Stressor: Stressor{Workers: 1},
						Time:     time,
					}},
				},
			}
			rejected := false
			for _, err := range genericwebhook.Validate(obj) {
				if strings.HasSuffix(err.Field, ".Time") {
					rejected = true
				}
			}
			Expect(rejected).To(Equal(invalid))
		},
		Entry("empty", "", false),
		Entry("seconds", "10s", false),
		Entry("compound", "1m30s", false),
		Entry("words", "ten minutes", true),
		Entry("no unit", "10", true),
	)
})
