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

package chaosdaemon

import (
	"testing"

	. "github.com/onsi/gomega"

	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

func Test_generateQdiscArgs(t *testing.T) {
	g := NewWithT(t)

	typ := "netem"

	t.Run("without parent and handle", func(t *testing.T) {

		args, err := generateQdiscArgs("add", &pb.Qdisc{Type: typ})

		g.Expect(err).To(BeNil())
		g.Expect(args).To(Equal([]string{"qdisc", "add", "dev", "eth0", "root", "handle", "1:0", typ}))
	})

	t.Run("with parent and handle", func(t *testing.T) {
		args, err := generateQdiscArgs("add", &pb.Qdisc{
			Type: typ,
			Parent: &pb.TcHandle{
				Major: 1,
				Minor: 1,
			},
			Handle: &pb.TcHandle{
				Major: 10,
				Minor: 0,
			},
		})

		g.Expect(err).To(BeNil())
		g.Expect(args).To(Equal([]string{"qdisc", "add", "dev", "eth0", "parent", "1:1", "handle", "10:0", typ}))
	})
}

func Test_convertNetemToArgs(t *testing.T) {
	g := NewWithT(t)

	t.Run("convert network delay", func(t *testing.T) {
		args := convertNetemToArgs(&pb.Netem{
			Time: 1000,
		})
		g.Expect(args).To(Equal("delay 1000"))

		args = convertNetemToArgs(&pb.Netem{
			Time:      1000,
			DelayCorr: 25,
		})
		g.Expect(args).To(Equal("delay 1000"))

		args = convertNetemToArgs(&pb.Netem{
			Time:      1000,
			Jitter:    10000,
			DelayCorr: 25,
		})
		g.Expect(args).To(Equal("delay 1000 10000 25.000000"))
	})

	t.Run("convert packet limit", func(t *testing.T) {
		args := convertNetemToArgs(&pb.Netem{
			Limit: 1000,
		})
		g.Expect(args).To(Equal("limit 1000"))
	})

	t.Run("convert packet loss", func(t *testing.T) {
		args := convertNetemToArgs(&pb.Netem{
			Loss: 100,
		})
		g.Expect(args).To(Equal("loss 100.000000"))

		args = convertNetemToArgs(&pb.Netem{
			Loss:     50,
			LossCorr: 12,
		})
		g.Expect(args).To(Equal("loss 50.000000 12.000000"))
	})

	t.Run("convert packet reorder", func(t *testing.T) {
		args := convertNetemToArgs(&pb.Netem{
			Reorder:     5,
			ReorderCorr: 10,
		})
		g.Expect(args).To(Equal(""))

		args = convertNetemToArgs(&pb.Netem{
			Time:        1000,
			Jitter:      10000,
			DelayCorr:   25,
			Reorder:     5,
			ReorderCorr: 10,
			Gap:         10,
		})
		g.Expect(args).To(Equal("delay 1000 10000 25.000000 reorder 5.000000 10.000000 gap 10"))

		args = convertNetemToArgs(&pb.Netem{
			Time:        1000,
			Jitter:      10000,
			DelayCorr:   25,
			Reorder:     5,
			ReorderCorr: 10,
			Gap:         10,
		})
		g.Expect(args).To(Equal("delay 1000 10000 25.000000 reorder 5.000000 10.000000 gap 10"))

		args = convertNetemToArgs(&pb.Netem{
			Time:      1000,
			Jitter:    10000,
			DelayCorr: 25,
			Reorder:   5,
			Gap:       10,
		})
		g.Expect(args).To(Equal("delay 1000 10000 25.000000 reorder 5.000000 gap 10"))
	})

	t.Run("convert packet duplication", func(t *testing.T) {
		args := convertNetemToArgs(&pb.Netem{
			Duplicate: 10,
		})
		g.Expect(args).To(Equal("duplicate 10.000000"))

		args = convertNetemToArgs(&pb.Netem{
			Duplicate:     10,
			DuplicateCorr: 50,
		})
		g.Expect(args).To(Equal("duplicate 10.000000 50.000000"))
	})

	t.Run("convert packet corrupt", func(t *testing.T) {
		args := convertNetemToArgs(&pb.Netem{
			Corrupt: 10,
		})
		g.Expect(args).To(Equal("corrupt 10.000000"))

		args = convertNetemToArgs(&pb.Netem{
			Corrupt:     10,
			CorruptCorr: 50,
		})
		g.Expect(args).To(Equal("corrupt 10.000000 50.000000"))
	})

	t.Run("complicate cases", func(t *testing.T) {
		args := convertNetemToArgs(&pb.Netem{
			Time:        1000,
			Jitter:      10000,
			Reorder:     5,
			Gap:         10,
			Corrupt:     10,
			CorruptCorr: 50,
		})
		g.Expect(args).To(Equal("delay 1000 10000 reorder 5.000000 gap 10 corrupt 10.000000 50.000000"))
	})

	t.Run("delay with rate", func(t *testing.T) {
		args := convertNetemToArgs(&pb.Netem{
			Time:   1000,
			Jitter: 10000,
			Rate:   "8000bit",
		})
		g.Expect(args).To(Equal("delay 1000 10000 rate 8000bit"))
	})
}
