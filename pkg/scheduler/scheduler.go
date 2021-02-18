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

package scheduler

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

const (
	// Set the top bit if a star was included in the expression.
	starBit = 1 << 63
)

// LastTime returns the next time this schedule activated, less than or equal with the given time.
func LastTime(spec v1alpha1.SchedulerSpec, now time.Time) (*time.Time, error) {
	scheduler, err := cron.ParseStandard(spec.Cron)
	if err != nil {
		return nil, fmt.Errorf("fail to parse runner rule %s, %v", spec.Cron, err)
	}
	var next time.Time
	if cronSpec, ok := scheduler.(*cron.SpecSchedule); ok {
		scheduleLast := &cusSchedule{cronSpec}
		next = scheduleLast.Last(now)
	} else if cronSpec, ok := scheduler.(cron.ConstantDelaySchedule); ok {
		scheduleLast := &cusConstantDelaySchedule{cronSpec}
		next = scheduleLast.Last(now)
	} else {
		return nil, fmt.Errorf("assert cron spec failed")
	}
	return &next, nil
}

type cusConstantDelaySchedule struct {
	cron.ConstantDelaySchedule
}

// Last returns the last time this schedule activated, less than or equal with the given time.
// So it would always return now
func (s cusConstantDelaySchedule) Last(t time.Time) time.Time {
	return t
}

type cusSchedule struct {
	*cron.SpecSchedule
}

// Last returns the last time this schedule activated, less than or equal with the given time.
// If no time can be found to satisfy the schedule, return the zero time.
// Modified from the original `Next` function in robfig/cron at Dec 15, 2020
func (s *cusSchedule) Last(t time.Time) time.Time {
	// General approach:
	// For Month, Day, Hour, Minute, Second:
	// Check if the time value matches.  If yes, continue to the next field.
	// If the field doesn't match the schedule, then increment the field until it matches.
	// While incrementing the field, a wrap-around brings it back to the beginning
	// of the field list (since it is necessary to re-verify previous field
	// values)

	// Convert the given time into the schedule's timezone, if one is specified.
	// Save the original timezone so we can convert back after we find a time.
	// Note that schedules without a time zone specified (time.Local) are treated
	// as local to the time provided.
	origLocation := t.Location()
	loc := s.Location
	if loc == time.Local {
		loc = t.Location()
	}
	if s.Location != time.Local {
		t = t.In(s.Location)
	}

	// If no time is found within five years, return zero.
	yearLimit := t.Year() - 5

WRAP:
	if t.Year() < yearLimit {
		return time.Time{}
	}

	// Find the first applicable month.
	// If it's this month, then do nothing.
	for 1<<uint(t.Month())&s.Month == 0 {
		t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc).Add(-1 * time.Second)
		if t.Month() == time.December {
			goto WRAP
		}
	}

	// Now get a day in that month.
	for !dayMatches(s, t) {
		finalDay := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc).Add(-1 * time.Second).Day()
		t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc).Add(-1 * time.Second)
		if t.Day() == finalDay {
			goto WRAP
		}
	}

	for 1<<uint(t.Hour())&s.Hour == 0 {
		t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, loc).Add(-1 * time.Second)
		if t.Hour() == 23 {
			goto WRAP
		}
	}

	for 1<<uint(t.Minute())&s.Minute == 0 {
		t = t.Truncate(time.Minute).Add(-1 * time.Second)
		if t.Minute() == 59 {
			goto WRAP
		}
	}

	for 1<<uint(t.Second())&s.Second == 0 {
		t = t.Add(-1 * time.Second)
		if t.Second() == 59 {
			goto WRAP
		}
	}

	return t.In(origLocation)
}

// dayMatches returns true if the schedule's day-of-week and day-of-month
// restrictions are satisfied by the given time.
func dayMatches(s *cusSchedule, t time.Time) bool {
	var (
		domMatch bool = 1<<uint(t.Day())&s.Dom > 0
		dowMatch bool = 1<<uint(t.Weekday())&s.Dow > 0
	)
	if s.Dom&starBit > 0 || s.Dow&starBit > 0 {
		return domMatch && dowMatch
	}
	return domMatch || dowMatch
}
