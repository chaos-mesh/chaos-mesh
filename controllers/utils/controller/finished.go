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

package controller

import (
	"time"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func IsChaosFinished(obj v1alpha1.InnerObject, now time.Time) bool {
	status := obj.GetStatus()

	finished := true

	// If one of the record has not been recovered, it's not finished
	for _, record := range status.Experiment.Records {
		if record.Phase != v1alpha1.NotInjected {
			finished = false
		}
	}

	// If the duration hasn't exceeded, it's not finished
	duration, err := obj.GetDuration()
	if err != nil {
		return finished
	}
	if obj.GetObjectMeta().CreationTimestamp.Add(*duration).After(now) {
		finished = false
	}

	return finished
}
