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

package graph

import (
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGraph(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Graph Suite")
}

var _ = Describe("graph", func() {
	Context("NewGraph", func() {
		It("should return a non-nil graph with empty head map", func() {
			g := NewGraph()
			Expect(g).NotTo(BeNil())
			Expect(g.head).NotTo(BeNil())
			Expect(g.head).To(BeEmpty())
		})
	})

	Context("Insert", func() {
		It("should insert an edge with nil Next on first insert", func() {
			g := NewGraph()
			g.Insert(1, 2)

			edge := g.head[1]
			Expect(edge).NotTo(BeNil())
			Expect(edge.Source).To(Equal(uint32(1)))
			Expect(edge.Target).To(Equal(uint32(2)))
			Expect(edge.Next).To(BeNil())
		})

		It("should prepend and link Next on second insert to same source", func() {
			g := NewGraph()
			g.Insert(1, 2)
			g.Insert(1, 3)

			edge := g.head[1]
			Expect(edge.Target).To(Equal(uint32(3)))
			Expect(edge.Next).NotTo(BeNil())
			Expect(edge.Next.Target).To(Equal(uint32(2)))
			Expect(edge.Next.Next).To(BeNil())
		})

		It("should keep inserts to different sources independent", func() {
			g := NewGraph()
			g.Insert(1, 2)
			g.Insert(3, 4)

			Expect(g.head[1].Target).To(Equal(uint32(2)))
			Expect(g.head[3].Target).To(Equal(uint32(4)))
		})
	})

	Context("IterFrom", func() {
		It("should return the head edge for an existing source", func() {
			g := NewGraph()
			g.Insert(1, 2)

			edge := g.IterFrom(1)
			Expect(edge).NotTo(BeNil())
			Expect(edge.Target).To(Equal(uint32(2)))
		})

		It("should return nil for a non-existing source", func() {
			g := NewGraph()

			edge := g.IterFrom(99)
			Expect(edge).To(BeNil())
		})
	})

	Context("Flatten", func() {
		logger := logr.Discard()

		It("should return empty slice when source has no edges", func() {
			g := NewGraph()
			result := g.Flatten(1, logger)
			Expect(result).To(BeEmpty())
		})

		It("should return the single child when source has one edge", func() {
			g := NewGraph()
			g.Insert(1, 2)

			result := g.Flatten(1, logger)
			Expect(result).To(ConsistOf(uint32(2)))
		})

		It("should return all direct children when source has multiple edges", func() {
			g := NewGraph()
			g.Insert(1, 2)
			g.Insert(1, 3)

			result := g.Flatten(1, logger)
			Expect(result).To(ConsistOf(uint32(2), uint32(3)))
		})

		It("should include grandchildren in a multi-level tree", func() {
			g := NewGraph()
			g.Insert(1, 2)
			g.Insert(2, 4)
			g.Insert(2, 5)

			result := g.Flatten(1, logger)
			Expect(result).To(ConsistOf(uint32(2), uint32(4), uint32(5)))
		})

		It("should include all descendants in a deep chain", func() {
			g := NewGraph()
			g.Insert(1, 2)
			g.Insert(2, 3)
			g.Insert(3, 4)

			result := g.Flatten(1, logger)
			Expect(result).To(ConsistOf(uint32(2), uint32(3), uint32(4)))
		})
	})
})
