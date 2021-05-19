// Copyright 2021 Chaos Mesh Authors.
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

package recorder

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestParse(t *testing.T) {
	g := NewGomegaWithT(t)

	type casePair struct {
		message string
		ev      ChaosEvent
	}

	missedRun, _ := time.Parse(time.RFC1123Z, "Wed, 19 May 2021 18:36:06 +0000")
	testCases := []casePair{
		{"Successfully apply chaos for test", Applied{"test"}},
		{"Successfully recover chaos for test", Recovered{"test"}},

		{"Successfully update test of resource", Updated{"test"}},

		{"Experiment has been deleted", Deleted{}},
		{"Time up according to the duration", TimeUp{}},
		{"Experiment has been paused", Paused{}},
		{"Experiment has started", Started{}},

		{"Failed to test1: test2", Failed{"test1", "test2"}},

		{"Finalizer has been inited", FinalizerInited{}},
		{"Finalizer has been removed", FinalizerRemoved{}},

		{"Missed scheduled time to start a job: Wed, 19 May 2021 18:36:06 +0000", MissSchedule{MissedRun: missedRun}},
		{"Create new object: test", ScheduleSpawn{Name: "test"}},
		{"Forbid spawning new job because: test is still running", ScheduleForbid{RunningName: "test"}},
		{"Skip removing history: test is still running", ScheduleSkipRemoveHistory{RunningName: "test"}},
	}

	for _, c := range testCases {
		g.Expect(c.ev.Message()).To(Equal(c.message))

		g.Expect(Parse(c.message)).To(Equal(c.ev))
	}
}
