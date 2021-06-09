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

package cron

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func TestGetRecentUnmetScheduleTime(t *testing.T) {
	g := NewGomegaWithT(t)

	type testCase struct {
		now               string
		lastScheduleTime  string
		creationTimeStamp string
		schedule          string
		missedRun         *string
		nextRun           *string
		err               error
	}

	zeroTime := "0001-01-01T00:00:00.000Z"
	testCases := []testCase{
		{
			now:               "2021-04-28T05:59:43.5Z",
			lastScheduleTime:  "2021-04-28T05:59:38.0Z",
			creationTimeStamp: zeroTime,
			schedule:          "@every 5s",
			missedRun:         pointer.StringPtr("2021-04-28T05:59:43.0Z"),
			nextRun:           pointer.StringPtr("2021-04-28T05:59:48.0Z"),
			err:               nil,
		},
		{
			now:               "2021-04-28T06:49:35.079Z",
			lastScheduleTime:  "2021-04-28T06:49:35.000Z",
			creationTimeStamp: zeroTime,
			schedule:          "@every 5s",
			missedRun:         nil,
			nextRun:           pointer.StringPtr("2021-04-28T06:49:40.000Z"),
			err:               nil,
		},
		{
			now:               "2021-04-28T06:49:35.079Z",
			lastScheduleTime:  zeroTime,
			creationTimeStamp: "2021-04-28T06:49:35.000Z",
			schedule:          "@every 5s",
			missedRun:         nil,
			nextRun:           pointer.StringPtr("2021-04-28T06:49:40.000Z"),
			err:               nil,
		},
		{
			now:               "2021-04-28T06:49:38.079Z",
			lastScheduleTime:  zeroTime,
			creationTimeStamp: "2021-04-28T06:49:35.000Z",
			schedule:          "@every 5s",
			missedRun:         nil,
			nextRun:           pointer.StringPtr("2021-04-28T06:49:40.000Z"),
			err:               nil,
		},
		{
			now:               "2021-04-28T06:49:40.079Z",
			lastScheduleTime:  zeroTime,
			creationTimeStamp: "2021-04-28T06:49:35.000Z",
			schedule:          "@every 5s",
			missedRun:         pointer.StringPtr("2021-04-28T06:49:40.000Z"),
			nextRun:           pointer.StringPtr("2021-04-28T06:49:45.000Z"),
			err:               nil,
		},
	}
	for _, t := range testCases {
		now, err := time.Parse(time.RFC3339, t.now)
		g.Expect(err).To(BeNil())
		lastScheduleTime, err := time.Parse(time.RFC3339, t.lastScheduleTime)
		g.Expect(err).To(BeNil())
		createTimeStamp, err := time.Parse(time.RFC3339, t.creationTimeStamp)
		g.Expect(err).To(BeNil())

		schedule := v1alpha1.Schedule{
			ObjectMeta: metav1.ObjectMeta{
				CreationTimestamp: metav1.Time{
					Time: createTimeStamp,
				},
			},
			Spec: v1alpha1.ScheduleSpec{
				Schedule: t.schedule,
			},
			Status: v1alpha1.ScheduleStatus{
				LastScheduleTime: metav1.Time{
					Time: lastScheduleTime,
				},
			},
		}
		missedRun, nextRun, err := getRecentUnmetScheduleTime(&schedule, now)

		expectedMissedRun := BeNil()
		expectedNextRun := BeNil()
		expectedErr := BeNil()
		if t.missedRun != nil {
			missedRun, err := time.Parse(time.RFC3339, *t.missedRun)
			g.Expect(err).To(BeNil())
			expectedMissedRun = Equal(&missedRun)
		}
		if t.nextRun != nil {
			nextRun, err := time.Parse(time.RFC3339, *t.nextRun)
			g.Expect(err).To(BeNil())
			expectedNextRun = Equal(&nextRun)
		}
		if t.err != nil {
			expectedErr = Equal(t.err)
		}

		g.Expect(err).To(expectedErr)
		g.Expect(missedRun).To(expectedMissedRun)
		g.Expect(nextRun).To(expectedNextRun)
	}
}
