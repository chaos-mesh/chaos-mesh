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
	"time"
)

type MissedSchedule struct {
	MissedRun time.Time
}

func (m MissedSchedule) Type() string {
	return "Warning"
}

func (m MissedSchedule) Reason() string {
	return "MissSchedule"
}

func (m MissedSchedule) Message() string {
	return fmt.Sprintf("Missed scheduled time to start a job: %s", m.MissedRun.Format(time.RFC1123Z))
}

type ScheduleSpawn struct {
	Name string
}

func (s ScheduleSpawn) Type() string {
	return "Normal"
}

func (s ScheduleSpawn) Reason() string {
	return "Spawned"
}

func (s ScheduleSpawn) Message() string {
	return fmt.Sprintf("Create new object: %s", s.Name)
}

type ScheduleForbid struct {
	RunningName string
}

func (s ScheduleForbid) Type() string {
	return "Warning"
}

func (s ScheduleForbid) Reason() string {
	return "Forbid"
}

func (s ScheduleForbid) Message() string {
	return fmt.Sprintf("Forbid spawning new job because: %s is still running", s.RunningName)
}

type ScheduleSkipRemoveHistory struct {
	RunningName string
}

func (s ScheduleSkipRemoveHistory) Type() string {
	return "Warning"
}

func (s ScheduleSkipRemoveHistory) Reason() string {
	return "Skip"
}

func (s ScheduleSkipRemoveHistory) Message() string {
	return fmt.Sprintf("Skip removing history: %s is still running", s.RunningName)
}

func init() {
	register(MissedSchedule{}, ScheduleSpawn{}, ScheduleForbid{}, ScheduleSkipRemoveHistory{})
}
