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

package utils

import (
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/controller"
)

type ChaosStatusString string

const (
	Injecting ChaosStatusString = "injecting"
	Running   ChaosStatusString = "running"
	Finished  ChaosStatusString = "finished"
	Paused    ChaosStatusString = "paused"
)

type ScheduleStatusString string

const (
	ScheduleRunning ScheduleStatusString = "running"
	SchedulePaused  ScheduleStatusString = "paused"
)

func GetChaosState(obj v1alpha1.InnerObject) ChaosStatusString {
	selected := false
	allInjected := false
	for _, c := range obj.GetChaos().Status.Conditions {
		if c.Status == corev1.ConditionTrue {
			switch c.Type {
			case v1alpha1.ConditionPaused:
				return Paused
			case v1alpha1.ConditionSelected:
				selected = true
			case v1alpha1.ConditionAllInjected:
				allInjected = true
			}
		}
	}
	if controller.IsChaosFinished(obj, time.Now()) {
		return Finished
	}
	if selected && allInjected {
		return Running
	}
	return Injecting
}

func GetScheduleState(sch v1alpha1.Schedule) ScheduleStatusString {
	if sch.IsPaused() {
		return SchedulePaused
	}
	return ScheduleRunning
}
