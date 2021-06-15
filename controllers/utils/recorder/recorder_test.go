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
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestGenerateAnnotations(t *testing.T) {
	g := NewGomegaWithT(t)

	type casePair struct {
		annotations map[string]string
		ev          ChaosEvent
	}

	missedRun, _ := time.Parse(time.RFC3339Nano, "2021-05-19T18:36:06Z")
	testCases := []casePair{
		{map[string]string{"chaos-mesh.org/id": "", "chaos-mesh.org/type": "applied"}, Applied{}},
		{map[string]string{"chaos-mesh.org/id": "test", "chaos-mesh.org/type": "applied"}, Applied{"test"}},
		{map[string]string{"chaos-mesh.org/id": "test", "chaos-mesh.org/type": "recovered"}, Recovered{"test"}},

		{map[string]string{"chaos-mesh.org/field": "test", "chaos-mesh.org/type": "updated"}, Updated{"test"}},

		{map[string]string{"chaos-mesh.org/type": "deleted"}, Deleted{}},
		{map[string]string{"chaos-mesh.org/type": "time-up"}, TimeUp{}},
		{map[string]string{"chaos-mesh.org/type": "paused"}, Paused{}},
		{map[string]string{"chaos-mesh.org/type": "started"}, Started{}},

		{map[string]string{"chaos-mesh.org/activity": "test1", "chaos-mesh.org/err": "test2", "chaos-mesh.org/type": "failed"}, Failed{"test1", "test2"}},
		{map[string]string{"chaos-mesh.org/type": "not-supported", "chaos-mesh.org/activity": "pausing a workflow schedule"}, NotSupported{Activity: "pausing a workflow schedule"}},

		{map[string]string{"chaos-mesh.org/type": "finalizer-inited"}, FinalizerInited{}},
		{map[string]string{"chaos-mesh.org/type": "finalizer-removed"}, FinalizerRemoved{}},

		{map[string]string{"chaos-mesh.org/missed-run": "2021-05-19T18:36:06Z", "chaos-mesh.org/type": "missed-schedule"}, MissedSchedule{MissedRun: missedRun}},
		{map[string]string{"chaos-mesh.org/name": "test", "chaos-mesh.org/type": "schedule-spawn"}, ScheduleSpawn{Name: "test"}},
		{map[string]string{"chaos-mesh.org/running-name": "test", "chaos-mesh.org/type": "schedule-forbid"}, ScheduleForbid{RunningName: "test"}},
		{map[string]string{"chaos-mesh.org/running-name": "test", "chaos-mesh.org/type": "schedule-skip-remove-history"}, ScheduleSkipRemoveHistory{RunningName: "test"}},
		{map[string]string{"chaos-mesh.org/type": "nodes-created", "chaos-mesh.org/child-nodes": "[\"node-a\",\"node-b\"]"}, NodesCreated{ChildNodes: []string{"node-a", "node-b"}}},
	}

	for _, c := range testCases {
		g.Expect(generateAnnotations(c.ev)).To(Equal(c.annotations))

		ev, err := FromAnnotations(c.annotations)
		if err != nil {
			fmt.Printf("fail")
		}
		g.Expect(ev).To(Equal(c.ev))
	}
}

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

		{"Missed scheduled time to start a job: Wed, 19 May 2021 18:36:06 +0000", MissedSchedule{MissedRun: missedRun}},
		{"Create new object: test", ScheduleSpawn{Name: "test"}},
		{"Forbid spawning new job because: test is still running", ScheduleForbid{RunningName: "test"}},
		{"Skip removing history: test is still running", ScheduleSkipRemoveHistory{RunningName: "test"}},
	}

	for _, c := range testCases {
		g.Expect(c.ev.Message()).To(Equal(c.message))
	}
}
