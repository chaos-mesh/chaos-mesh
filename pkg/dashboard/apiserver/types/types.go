// Copyright 2022 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package types

import (
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
	"github.com/chaos-mesh/chaos-mesh/pkg/status"
)

// Archive defines the basic information of an archive.
type Archive = core.ObjectBase

/*
ArchiveDetail represents an archive instance.

It inherits `Archive` and adds complete definition of an experiment.
*/
type ArchiveDetail struct {
	Archive
	KubeObject core.KubeObjectDesc `json:"kube_object"`
}

// Experiment defines the basic information of an experiment.
type Experiment struct {
	core.ObjectBase
	Status        status.ChaosStatus `json:"status"`
	FailedMessage string             `json:"failed_message,omitempty"`
}

/*
ExperimentDetail represents an experiment instance.

It inherits `Experiment` and adds complete definition of an experiment.
*/
type ExperimentDetail struct {
	Experiment
	KubeObject core.KubeObjectDesc `json:"kube_object"`
}

// PhysicalMachine defines the basic information of a physical machine.
type PhysicalMachine struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Address   string `json:"address"`
}

// Pod defines the basic information of a pod.
type Pod struct {
	IP        string `json:"ip"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	State     string `json:"state"`
}

// Schedule defines the basic information of a schedule.
type Schedule struct {
	core.ObjectBase
	Status status.ScheduleStatus `json:"status"`
}

/*
ScheduleDetail represents an archive instance.

It inherits `Schedule` and adds complete definition of a schedule.
*/
type ScheduleDetail struct {
	Schedule
	ExperimentUIDs []string            `json:"experiment_uids"`
	KubeObject     core.KubeObjectDesc `json:"kube_object"`
}

type StatusCheckTemplateBase struct {
	Namespace   string `json:"namespace"`
	Name        string `json:"name"`
	UID         string `json:"uid"`
	Description string `json:"description,omitempty"`
	Created     string `json:"created_at"`
}

/*
StatusCheckTemplateDetail represents an archive instance.

It inherits `StatusCheckTemplateBase` and adds complete definition of a Status Check template.
*/
type StatusCheckTemplateDetail struct {
	StatusCheckTemplateBase `json:",inline,omitempty"`
	Spec                    v1alpha1.StatusCheckTemplate `json:"spec"`
}

type StatusCheckTemplate struct {
	Namespace   string                       `json:"namespace"`
	Name        string                       `json:"name"`
	Description string                       `json:"description,omitempty"`
	Spec        v1alpha1.StatusCheckTemplate `json:"spec"`
}
