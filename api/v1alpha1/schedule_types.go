// Copyright 2020 Chaos Mesh Authors.
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

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// Schedule is the cronly schedule object
type Schedule struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ScheduleSpec `json:"spec"`
	Status ScheduleStatus `json:"status"`
}

// ScheduleSpec is the specification of a schedule object
type ScheduleSpec struct {
	Schedule string `json:"schedule"`
}

// ScheduleStatus is the status of a schedule object
type ScheduleStatus struct {

}
