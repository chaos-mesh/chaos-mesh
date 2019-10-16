// Copyright 2019 PingCAP, Inc.
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

package util

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestSleep(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name      string
		sleepTime time.Duration
		isStop    bool
	}

	tcs := []TestCase{
		{
			name:      "sleep 3s",
			sleepTime: 3 * time.Second,
			isStop:    false,
		},
		{
			name:      "stop sleep",
			sleepTime: 10 * time.Second,
			isStop:    true,
		},
	}

	for _, tc := range tcs {
		stopC := make(chan struct{})

		if tc.isStop {
			close(stopC)
		}

		start := time.Now()
		Sleep(stopC, tc.sleepTime)

		elapsed := time.Since(start)

		if tc.isStop {
			g.Expect(elapsed).Should(BeNumerically("<", tc.sleepTime), tc.name)
		} else {
			g.Expect(elapsed).Should(BeNumerically(">=", tc.sleepTime), tc.name)
		}
	}
}
