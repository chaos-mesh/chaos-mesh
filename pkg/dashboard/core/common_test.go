// Copyright 2025 Chaos Mesh Authors.
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

package core

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Core Suite")
}

var _ = Describe("Filter", func() {
	Describe("toMap", func() {
		Context("with empty filter", func() {
			It("should return a map with empty values", func() {
				filter := Filter{}
				expected := map[string]interface{}{
					"object_id": "",
					"start":     "",
					"end":       "",
					"namespace": "",
					"name":      "",
					"kind":      "",
					"limit":     "",
				}

				Expect(filter.toMap()).To(Equal(expected))
			})
		})

		Context("with filled filter", func() {
			It("should return a map with correct values", func() {
				filter := Filter{
					ObjectID:  "obj123",
					Start:     "2021-01-01 00:00:00",
					End:       "2021-01-02 00:00:00",
					Namespace: "default",
					Name:      "test",
					Kind:      "PodChaos",
					Limit:     "50",
				}
				expected := map[string]interface{}{
					"object_id": "obj123",
					"start":     "2021-01-01 00:00:00",
					"end":       "2021-01-02 00:00:00",
					"namespace": "default",
					"name":      "test",
					"kind":      "PodChaos",
					"limit":     "50",
				}

				Expect(filter.toMap()).To(Equal(expected))
			})
		})
	})

	Describe("ConstructQueryArgs", func() {
		DescribeTable("generates correct query and arguments",
			func(filter Filter, expectedQuery string, expectedArgs []interface{}) {
				query, args := filter.ConstructQueryArgs()
				Expect(query).To(Equal(expectedQuery))
				Expect(args).To(Equal(expectedArgs))
			},
			Entry("empty filter",
				Filter{},
				"",
				nil),
			Entry("filter with namespace only",
				Filter{Namespace: "default"},
				"namespace = ?",
				[]interface{}{"default"}),
			Entry("filter with namespace and name",
				Filter{Namespace: "default", Name: "test"},
				"name = ? AND namespace = ?",
				[]interface{}{"test", "default"}),
			Entry("filter with time range",
				Filter{
					Start:     "2021-01-01 00:00:00",
					End:       "2021-01-02 00:00:00",
					Namespace: "default",
				},
				"namespace = ? AND created_at BETWEEN ? AND ?",
				[]interface{}{"default", "2021-01-01 00:00:00", "2021-01-02 00:00:00"}),
			Entry("filter with start time only",
				Filter{Start: "2021-01-01 00:00:00"},
				"created_at >= ?",
				[]interface{}{"2021-01-01 00:00:00"}),
			Entry("filter with end time only",
				Filter{End: "2021-01-02 00:00:00"},
				"created_at <= ?",
				[]interface{}{"2021-01-02 00:00:00"}),
			Entry("complex filter",
				Filter{
					ObjectID:  "obj123",
					Start:     "2021-01-01 00:00:00",
					End:       "2021-01-02 00:00:00",
					Namespace: "default",
					Name:      "test",
					Kind:      "PodChaos",
					Limit:     "50", // Limit should not be in the query
				},
				"kind = ? AND name = ? AND namespace = ? AND object_id = ? AND created_at BETWEEN ? AND ?",
				[]interface{}{"PodChaos", "test", "default", "obj123", "2021-01-01 00:00:00", "2021-01-02 00:00:00"}),
		)
	})
})
