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
	"fmt"
	"time"

	"github.com/robfig/cron"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// Get this function from Kubernetes

// getRecentUnmetScheduleTime gets the most recent time that have passed when a Job should have started but did not.
//
// If there are too many (>100) unstarted times, just give up and return a nil.
func getRecentUnmetScheduleTime(schedule *v1alpha1.Schedule, now time.Time) (*time.Time, *time.Time, error) {
	sched, err := cron.ParseStandard(schedule.Spec.Schedule)
	if err != nil {
		return nil, nil, fmt.Errorf("unparseable schedule: %s : %s", schedule.Spec.Schedule, err)
	}

	var earliestTime time.Time
	if !schedule.Status.LastScheduleTime.UTC().IsZero() {
		earliestTime = schedule.Status.LastScheduleTime.Time
	} else {
		earliestTime = schedule.ObjectMeta.CreationTimestamp.Time
	}
	if schedule.Spec.StartingDeadlineSeconds != nil {
		schedulingDeadline := now.Add(-time.Second * time.Duration(*schedule.Spec.StartingDeadlineSeconds))

		if schedulingDeadline.After(earliestTime) {
			earliestTime = schedulingDeadline
		}
	}
	if earliestTime.After(now) {
		return nil, nil, fmt.Errorf("earliestTime is later than now: earliestTime: %v, now: %v", earliestTime, now)
	}

	iterateTime := 0
	var missedRun *time.Time
	nextRun := sched.Next(earliestTime)
	for t := sched.Next(earliestTime); !t.After(now); t = sched.Next(t) {
		t := t

		missedRun = &t
		nextRun = sched.Next(*missedRun)

		iterateTime++
		if iterateTime > 100 {
			// We can't get the most recent times so just return an empty slice
			return nil, nil, fmt.Errorf("too many missed start time (> 100). Set or decrease .spec.startingDeadlineSeconds or check clock skew")
		}
	}

	return missedRun, &nextRun, nil
}
