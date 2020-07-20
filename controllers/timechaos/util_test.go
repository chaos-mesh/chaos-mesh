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

package timechaos

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

type SecAndNSecFromDurationTestCase struct {
	Duration time.Duration
	Sec      int64
	NSec     int64
}

func TestSecAndNSecFromDuration(t *testing.T) {
	g := NewGomegaWithT(t)
	cases := []SecAndNSecFromDurationTestCase{
		{time.Second * 100, 100, 0},
		{time.Second * -100, -100, 0},
		{time.Second*-100 + time.Microsecond*-20, -100, -20000},
		{time.Second*-100 + time.Microsecond*20, -99, -999980000},
		{time.Second*100 + time.Microsecond*20, 100, 20000},
		{time.Second*100 + time.Microsecond*-20, 99, 999980000},
	}

	for _, c := range cases {
		sec, nsec := secAndNSecFromDuration(c.Duration)
		g.Expect(sec).Should(Equal(c.Sec))
		g.Expect(nsec).Should(Equal(c.NSec))
	}
}
