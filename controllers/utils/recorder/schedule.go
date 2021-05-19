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
	"strings"
	"time"
)

type MissSchedule struct {
	MissedRun time.Time
}

func (m MissSchedule) Type() string {
	return "Warning"
}

func (m MissSchedule) Reason() string {
	return "MissSchedule"
}

func (m MissSchedule) Message() string {
	return fmt.Sprintf("Missed scheduled time to start a job: %s", m.MissedRun.Format(time.RFC1123Z))
}

func (m MissSchedule) Parse(message string) ChaosEvent {
	prefix := "Missed scheduled time to start a job: "
	if strings.HasPrefix(message, prefix) {
		missedRun, err := time.Parse(time.RFC1123Z, strings.TrimPrefix(message, prefix))
		if err == nil {
			return MissSchedule{
				MissedRun: missedRun,
			}
		}
	}

	return nil
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

func (m ScheduleSpawn) Parse(message string) ChaosEvent {
	prefix := "Create new object: "
	if strings.HasPrefix(message, prefix) {
		return ScheduleSpawn{
			Name: strings.TrimPrefix(message, prefix),
		}
	}

	return nil
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

func (m ScheduleForbid) Parse(message string) ChaosEvent {
	prefix := "Forbid spawning new job because: "
	suffix := " is still running"
	if strings.HasPrefix(message, prefix) && strings.HasSuffix(message, suffix) {
		return ScheduleForbid{
			RunningName: strings.TrimSuffix(strings.TrimPrefix(message, prefix), suffix),
		}
	}

	return nil
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

func (m ScheduleSkipRemoveHistory) Parse(message string) ChaosEvent {
	prefix := "Skip removing history: "
	suffix := " is still running"
	if strings.HasPrefix(message, prefix) && strings.HasSuffix(message, suffix) {
		return ScheduleSkipRemoveHistory{
			RunningName: strings.TrimSuffix(strings.TrimPrefix(message, prefix), suffix),
		}
	}

	return nil
}

func init() {
	register(MissSchedule{}, ScheduleSpawn{}, ScheduleForbid{}, ScheduleSkipRemoveHistory{})
}
