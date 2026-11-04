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
	"math"
	"testing"
	"time"

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

	// mustConvertNetemToArgs asserts the conversion succeeds and returns the args.
	mustConvertNetemToArgs := func(netem *pb.Netem) string {
		args, err := convertNetemToArgs(netem)
		g.Expect(err).To(BeNil())
		return args
	}

	t.Run("convert network delay", func(t *testing.T) {
		args := mustConvertNetemToArgs(&pb.Netem{
			Time: "1000ms",
		})
		g.Expect(args).To(Equal("delay 1s"))

		args = mustConvertNetemToArgs(&pb.Netem{
			Time:      "1000ms",
			DelayCorr: 25,
		})
		g.Expect(args).To(Equal("delay 1s"))

		args = mustConvertNetemToArgs(&pb.Netem{
			Time:      "1000us",
			Jitter:    "10000ns",
			DelayCorr: 25,
		})
		g.Expect(args).To(Equal("delay 1ms 10us 25.000000"))
	})

	t.Run("convert network delay with units unknown to tc", func(t *testing.T) {
		// tc only understands "s", "ms" and "us", while these are all valid
		// values for the API, so they must be normalized instead of being
		// forwarded as is.
		args := mustConvertNetemToArgs(&pb.Netem{
			Time: "1m",
		})
		g.Expect(args).To(Equal("delay 60s"))

		args = mustConvertNetemToArgs(&pb.Netem{
			Time: "0m30s",
		})
		g.Expect(args).To(Equal("delay 30s"))

		args = mustConvertNetemToArgs(&pb.Netem{
			Time: "1m30s",
		})
		g.Expect(args).To(Equal("delay 90s"))

		args = mustConvertNetemToArgs(&pb.Netem{
			Time: "1500ns",
		})
		g.Expect(args).To(Equal("delay 2us"))
	})

	t.Run("convert network delay below one millisecond", func(t *testing.T) {
		// "0.5s" is lexicographically smaller than "0ms", it used to be
		// silently dropped, leaving the experiment without any delay.
		args := mustConvertNetemToArgs(&pb.Netem{
			Time: "0.5s",
		})
		g.Expect(args).To(Equal("delay 500ms"))

		args = mustConvertNetemToArgs(&pb.Netem{
			Time:      "0.25s",
			Jitter:    "0.1s",
			DelayCorr: 25,
		})
		g.Expect(args).To(Equal("delay 250ms 100ms 25.000000"))
	})

	t.Run("convert zero network delay", func(t *testing.T) {
		// "0s" and "0us" are lexicographically greater than "0ms" while being
		// zero, they used to generate a pointless `delay 0s`.
		for _, zero := range []string{"", "0ms", "0s", "0us", "0ns"} {
			args := mustConvertNetemToArgs(&pb.Netem{
				Time: zero,
			})
			g.Expect(args).To(Equal(""), "delay %q should be dropped", zero)

			args = mustConvertNetemToArgs(&pb.Netem{
				Time:      "1s",
				Jitter:    zero,
				DelayCorr: 25,
			})
			g.Expect(args).To(Equal("delay 1s"), "jitter %q should be dropped", zero)
		}
	})

	t.Run("reject invalid durations", func(t *testing.T) {
		_, err := convertNetemToArgs(&pb.Netem{
			Time: "1000",
		})
		g.Expect(err).To(HaveOccurred())

		_, err = convertNetemToArgs(&pb.Netem{
			Time:   "1s",
			Jitter: "not a duration",
		})
		g.Expect(err).To(HaveOccurred())
	})

	t.Run("reject negative durations", func(t *testing.T) {
		// A negative duration would compare as "not greater than zero" and be
		// indistinguishable from an unset field, so it must not be accepted.
		_, err := convertNetemToArgs(&pb.Netem{
			Time: "-1s",
		})
		g.Expect(err).To(HaveOccurred())

		_, err = convertNetemToArgs(&pb.Netem{
			Time:   "1s",
			Jitter: "-1ms",
		})
		g.Expect(err).To(HaveOccurred())
	})

	t.Run("convert packet limit", func(t *testing.T) {
		args := mustConvertNetemToArgs(&pb.Netem{
			Limit: 1000,
		})
		g.Expect(args).To(Equal("limit 1000"))
	})

	t.Run("convert packet loss", func(t *testing.T) {
		args := mustConvertNetemToArgs(&pb.Netem{
			Loss: 100,
		})
		g.Expect(args).To(Equal("loss 100.000000"))

		args = mustConvertNetemToArgs(&pb.Netem{
			Loss:     50,
			LossCorr: 12,
		})
		g.Expect(args).To(Equal("loss 50.000000 12.000000"))
	})

	t.Run("convert packet reorder", func(t *testing.T) {
		args := mustConvertNetemToArgs(&pb.Netem{
			Reorder:     5,
			ReorderCorr: 10,
		})
		g.Expect(args).To(Equal(""))

		args = mustConvertNetemToArgs(&pb.Netem{
			Time:        "1000ms",
			Jitter:      "10000ms",
			DelayCorr:   25,
			Reorder:     5,
			ReorderCorr: 10,
			Gap:         10,
		})
		g.Expect(args).To(Equal("delay 1s 10s 25.000000 reorder 5.000000 10.000000 gap 10"))

		args = mustConvertNetemToArgs(&pb.Netem{
			Time:        "1000ms",
			Jitter:      "10000ms",
			DelayCorr:   25,
			Reorder:     5,
			ReorderCorr: 10,
			Gap:         10,
		})
		g.Expect(args).To(Equal("delay 1s 10s 25.000000 reorder 5.000000 10.000000 gap 10"))

		args = mustConvertNetemToArgs(&pb.Netem{
			Time:      "1000ms",
			Jitter:    "10000ms",
			DelayCorr: 25,
			Reorder:   5,
			Gap:       10,
		})
		g.Expect(args).To(Equal("delay 1s 10s 25.000000 reorder 5.000000 gap 10"))
	})

	t.Run("convert packet duplication", func(t *testing.T) {
		args := mustConvertNetemToArgs(&pb.Netem{
			Duplicate: 10,
		})
		g.Expect(args).To(Equal("duplicate 10.000000"))

		args = mustConvertNetemToArgs(&pb.Netem{
			Duplicate:     10,
			DuplicateCorr: 50,
		})
		g.Expect(args).To(Equal("duplicate 10.000000 50.000000"))
	})

	t.Run("convert packet corrupt", func(t *testing.T) {
		args := mustConvertNetemToArgs(&pb.Netem{
			Corrupt: 10,
		})
		g.Expect(args).To(Equal("corrupt 10.000000"))

		args = mustConvertNetemToArgs(&pb.Netem{
			Corrupt:     10,
			CorruptCorr: 50,
		})
		g.Expect(args).To(Equal("corrupt 10.000000 50.000000"))
	})

	t.Run("complicate cases", func(t *testing.T) {
		args := mustConvertNetemToArgs(&pb.Netem{
			Time:        "1000ms",
			Jitter:      "10000ms",
			Reorder:     5,
			Gap:         10,
			Corrupt:     10,
			CorruptCorr: 50,
		})
		g.Expect(args).To(Equal("delay 1s 10s reorder 5.000000 gap 10 corrupt 10.000000 50.000000"))
	})

	t.Run("delay with rate", func(t *testing.T) {
		args := mustConvertNetemToArgs(&pb.Netem{
			Time:   "1000ms",
			Jitter: "10000ms",
			Rate:   "8000bit",
		})
		g.Expect(args).To(Equal("delay 1s 10s rate 8000bit"))
	})
}

func Test_formatTcDuration(t *testing.T) {
	g := NewWithT(t)

	// tc only understands the "s", "ms" and "us" suffixes.
	for _, tc := range []struct {
		duration time.Duration
		expected string
	}{
		{0, "0s"},
		{time.Second, "1s"},
		{90 * time.Second, "90s"},
		{time.Minute, "60s"},
		{time.Hour, "3600s"},
		{100 * time.Millisecond, "100ms"},
		{1500 * time.Millisecond, "1500ms"},
		{time.Microsecond, "1us"},
		{1500 * time.Microsecond, "1500us"},
		// Sub-microsecond values are rounded up, tc cannot express them and
		// rounding down would turn the delay into a no-op.
		{time.Nanosecond, "1us"},
		{1500 * time.Nanosecond, "2us"},
		{time.Millisecond + time.Nanosecond, "1001us"},
		// The largest duration time.ParseDuration can return must not overflow.
		{time.Duration(math.MaxInt64), "9223372036854776us"},
	} {
		g.Expect(formatTcDuration(tc.duration)).To(Equal(tc.expected), "%v", tc.duration)
	}
}
