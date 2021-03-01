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
	"strings"
	"testing"
	"time"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// Modified from the original `TestNext` function in robfig/cron at Dec 15, 2020
func TestLast(t *testing.T) {
	runs := []struct {
		time, spec string
		expected   string
	}{
		// `@every` cron
		{"Jul 9 14:45:00 2012", "@every 1s", "Jul 9 14:45:00 2012"},
		{"Jul 9 14:45:00 2012", "@every 1m", "Jul 9 14:45:00 2012"},
		{"Jul 9 14:45:00 2012", "@every 1h", "Jul 9 14:45:00 2012"},
		{"Jul 9 14:45:00 2012", "@every 1h1m1s", "Jul 9 14:45:00 2012"},

		// Simple cases
		{"Jul 9 14:45 2012", "0/15 * * * *", "Jul 9 14:45 2012"},
		{"Jul 9 14:59 2012", "0/15 * * * *", "Jul 9 14:45 2012"},
		{"Jul 9 14:59:59 2012", "0/15 * * * *", "Jul 9 14:45 2012"},

		// Wrap around hours
		{"Jul 9 16:15 2012", "20-35/15 * * * *", "Jul 9 15:35 2012"},

		// Wrap around days
		{"Jul 9 00:15 2012", "20-35/15 * * * *", "Jul 8 23:35 2012"},

		// Wrap around months
		{"Jul 9 23:35 2012", "0 0 10 Apr-Oct ?", "Jun 10 00:00 2012"},
		{"Jul 9 23:35 2012", "0 0 */5 Apr,Aug,Oct Mon", "Apr 30 00:00 2012"},
		{"Jul 9 23:35 2012", "0 0 */5 Oct Mon", "Oct 31 00:00 2011"},

		// Wrap around years
		{"Jan 9 23:35 2012", "0 0 * Feb Mon", "Feb 28 00:00 2011"},
		{"Jan 9 23:35 2012", "0 0 * Feb Mon/2", "Feb 28 00:00 2011"},

		// Leap year
		{"Jan 9 23:35 2012", "0 0 29 Feb ?", "Feb 29 00:00 2008"},

		// Daylight savings time 2am EST (-5) -> 3am EDT (-4)
		{"2012-03-11T03:59:00-0400", "TZ=America/New_York 30 1 11 Mar ?", "2012-03-11T02:30:00-0400"},

		// hourly job
		{"2012-03-11T00:59:00-0500", "TZ=America/New_York 0 * * * ?", "2012-03-11T00:00:00-0500"},
		{"2012-03-11T02:59:00-0400", "TZ=America/New_York 0 * * * ?", "2012-03-11T01:00:00-0500"},
		{"2012-03-11T03:59:00-0400", "TZ=America/New_York 0 * * * ?", "2012-03-11T03:00:00-0400"},
		{"2012-03-11T04:59:00-0400", "TZ=America/New_York 0 * * * ?", "2012-03-11T04:00:00-0400"},

		// hourly job using CRON_TZ
		{"2012-03-11T00:59:00-0500", "CRON_TZ=America/New_York 0 * * * ?", "2012-03-11T00:00:00-0500"},
		{"2012-03-11T02:59:00-0400", "CRON_TZ=America/New_York 0 * * * ?", "2012-03-11T01:00:00-0500"},
		{"2012-03-11T03:59:00-0400", "CRON_TZ=America/New_York 0 * * * ?", "2012-03-11T03:00:00-0400"},
		{"2012-03-11T04:59:00-0400", "CRON_TZ=America/New_York 0 * * * ?", "2012-03-11T04:00:00-0400"},

		// 1am nightly job
		{"2012-03-11T01:59:00-0500", "TZ=America/New_York 0 1 * * ?", "2012-03-11T01:00:00-0500"},
		{"2012-03-11T03:00:00-0400", "TZ=America/New_York 0 1 * * ?", "2012-03-11T01:00:00-0500"},

		// 2am nightly job (skipped)
		{"2012-03-12T01:00:00-0400", "TZ=America/New_York 0 2 * * ?", "2012-03-10T02:00:00-0500"},

		// Daylight savings time 2am EDT (-4) => 1am EST (-5)
		{"2012-11-04T02:30:00-0500", "TZ=America/New_York 0 0 04 Nov ?", "2012-11-04T00:00:00-0400"},
		{"2012-11-04T01:30:00-0500", "TZ=America/New_York 45 1 04 Nov ?", "2012-11-04T01:45:00-0400"},

		// hourly job
		{"2012-11-04T00:59:00-0400", "TZ=America/New_York 0 * * * ?", "2012-11-04T00:00:00-0400"},
		{"2012-11-04T01:00:00-0500", "TZ=America/New_York 30 * * * ?", "2012-11-04T01:30:00-0400"},
		{"2012-11-04T01:00:00-0500", "TZ=America/New_York 0 * * * ?", "2012-11-04T01:00:00-0500"},

		// 1am nightly job (runs twice)
		{"2012-11-04T01:59:00-0400", "TZ=America/New_York 0 1 * * ?", "2012-11-04T01:00:00-0400"},
		{"2012-11-04T02:00:00-0400", "TZ=America/New_York 0 1 * * ?", "2012-11-04T01:00:00-0500"},
		{"2012-11-04T01:00:00-0500", "TZ=America/New_York 0 1 * * ?", "2012-11-04T01:00:00-0500"},

		// 2am nightly job
		{"2012-11-04T01:59:00-0500", "TZ=America/New_York 0 2 * * ?", "2012-11-03T02:00:00-0400"},

		// Unsatisfiable
		{"Jul 9 23:35 2012", "0 0 30 Feb ?", ""},
		{"Jul 9 23:35 2012", "0 0 31 Apr ?", ""},
	}

	for _, c := range runs {
		ss := v1alpha1.SchedulerSpec{Cron: c.spec}
		actual, err := LastTime(ss, getTime(c.time))
		if err != nil {
			t.Error(err)
			continue
		}
		expected := getTime(c.expected)
		if !actual.Equal(expected) {
			t.Errorf("%s, \"%s\": (expected) %v != %v (actual)", c.time, c.spec, expected, actual)
		}
	}
}

func getTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}

	var location = time.Local
	if strings.HasPrefix(value, "TZ=") {
		parts := strings.Fields(value)
		loc, err := time.LoadLocation(parts[0][len("TZ="):])
		if err != nil {
			panic("could not parse location:" + err.Error())
		}
		location = loc
		value = parts[1]
	}

	var layouts = []string{
		"Jan 2 15:04 2006",
		"Jan 2 15:04:05 2006",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, value, location); err == nil {
			return t
		}
	}
	if t, err := time.ParseInLocation("2006-01-02T15:04:05-0700", value, location); err == nil {
		return t
	}
	panic("could not parse time value " + value)
}
