// Copyright 2021 Chaos Mesh Authors.
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

package v1alpha1

import (
	"fmt"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// Stress chaos is a chaos to generate plenty of stresses over a collection of pods.

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="duration",type=string,JSONPath=`.spec.duration`
// +chaos-mesh:experiment
// +genclient

// StressChaos is the Schema for the stresschaos API
type StressChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a time chaos experiment
	Spec StressChaosSpec `json:"spec"`

	// +optional
	// Most recently observed status of the time chaos experiment
	Status StressChaosStatus `json:"status,omitempty"`
}

var _ InnerObjectWithCustomStatus = (*StressChaos)(nil)
var _ InnerObjectWithSelector = (*StressChaos)(nil)
var _ InnerObject = (*StressChaos)(nil)

// StressChaosSpec defines the desired state of StressChaos
// +kubebuilder:validation:XValidation:rule="(has(self.stressngStressors) && size(self.stressngStressors) > 0) || has(self.stressors)",message="missing stressors"
type StressChaosSpec struct {
	ContainerSelector `json:",inline"`

	// Stressors defines plenty of stressors supported to stress system components out.
	// You can use one or more of them to make up various kinds of stresses. At least
	// one of the stressors should be specified.
	// +optional
	Stressors *Stressors `json:"stressors,omitempty"`

	// StressngStressors defines plenty of stressors just like `Stressors` except that it's an experimental
	// feature and more powerful. You can define stressors in `stress-ng` (see also `man stress-ng`) dialect,
	// however not all of the supported stressors are well tested. It maybe retired in later releases. You
	// should always use `Stressors` to define the stressors and use this only when you want more stressors
	// unsupported by `Stressors`. When both `StressngStressors` and `Stressors` are defined, `StressngStressors`
	// wins.
	// +optional
	StressngStressors string `json:"stressngStressors,omitempty"`

	// Duration represents the duration of the chaos action
	// +optional
	Duration *string `json:"duration,omitempty" webhook:"Duration"`

	// RemoteCluster represents the remote cluster where the chaos will be deployed
	// +optional
	RemoteCluster string `json:"remoteCluster,omitempty"`
}

// StressChaosStatus defines the observed state of StressChaos
type StressChaosStatus struct {
	ChaosStatus `json:",inline"`
	// Instances always specifies stressing instances
	// +optional
	Instances map[string]StressInstance `json:"instances,omitempty"`
}

// StressInstance is an instance generates stresses
type StressInstance struct {
	// UID is the stress-ng identifier
	// +optional
	UID string `json:"uid,omitempty"`
	// MemoryUID is the memStress identifier
	// +optional
	MemoryUID string `json:"memoryUid,omitempty"`
	// StartTime specifies when the stress-ng starts
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`
	// MemoryStartTime specifies when the memStress starts
	// +optional
	MemoryStartTime *metav1.Time `json:"memoryStartTime,omitempty"`
}

// Stressors defines plenty of stressors supported to stress system components out.
// You can use one or more of them to make up various kinds of stresses
// +kubebuilder:validation:XValidation:rule="has(self.memory) || has(self.cpu)",message="missing stressors"
type Stressors struct {
	// MemoryStressor stresses virtual memory out
	// +optional
	MemoryStressor *MemoryStressor `json:"memory,omitempty"`
	// CPUStressor stresses CPU out
	// +optional
	CPUStressor *CPUStressor `json:"cpu,omitempty"`
}

// Normalize the stressors to comply with stress-ng
func (in *Stressors) Normalize() (cpuStressors string, memoryStressors string, err error) {
	cpuStressors = ""
	memoryStressors = ""
	err = nil

	if memory := in.MemoryStressor; memory != nil && memory.Workers != 0 {
		if err := memory.validateSize(field.NewPath("stressors", "memory"), true); err != nil {
			return "", "", err
		}
		args := []string{"--workers", strconv.Itoa(memory.Workers)}

		if memory.Size != "" {
			args = append(args, "--size", memory.Size)
		}

		if memory.Time != "" {
			args = append(args, "--time", memory.Time)
		}

		for _, option := range memory.Options {
			args = append(args, strings.Fields(option)...)
		}
		memoryStressors = strings.Join(args, " ")
	}
	if in.CPUStressor != nil && in.CPUStressor.Workers != 0 {
		// Without these two args, we may not reach the resource limit of pod,
		// especially when we set cpu workers > 1
		// More details see: https://github.com/chaos-mesh/chaos-mesh/issues/3100
		cpuStressors += " --cpu-load-slice 10 --cpu-method sqrt"
		cpuStressors += fmt.Sprintf(" --cpu %d", in.CPUStressor.Workers)

		if in.CPUStressor.Load != nil {
			cpuStressors += fmt.Sprintf(" --cpu-load %d",
				*in.CPUStressor.Load)
		}

		if in.CPUStressor.Options != nil {
			for _, v := range in.CPUStressor.Options {
				cpuStressors += fmt.Sprintf(" %v ", v)
			}
		}
	}

	return
}

// Stressor defines common configurations of a stressor
type Stressor struct {
	// Workers specifies N workers to apply the stressor.
	// Maximum 8192 workers can run by stress-ng
	// +kubebuilder:validation:Maximum=8192
	Workers int `json:"workers"`
}

// MemoryStressor defines how to stress memory out
type MemoryStressor struct {
	Stressor `json:",inline"`

	// Size specifies N bytes consumed per vm worker, default is the total available memory.
	// One can specify the size as % of total available memory or in units of B, KB/KiB,
	// MB/MiB, GB/GiB, TB/TiB.
	// +optional
	Size string `json:"size,omitempty"`

	// Time specifies how long it takes for the workers' total memory usage to grow linearly to Size.
	// The workers then hold that memory until the experiment ends.
	// Duration controls how long the experiment runs.
	// By default, memory is allocated immediately.
	// +optional
	Time string `json:"time,omitempty" webhook:"Duration"`

	// OOMScoreAdj adjusts how likely the Linux out-of-memory (OOM) killer is to select the memStress process.
	// Higher values make memStress more likely to be killed, and lower values make it less likely.
	// The default value of 0 uses the inherited adjustment.
	// See https://docs.kernel.org/filesystems/proc.html#proc-pid-oom-adj-proc-pid-oom-score-adj-adjust-the-oom-killer-score for details.
	// +kubebuilder:validation:Minimum=-1000
	// +kubebuilder:validation:Maximum=1000
	// +kubebuilder:default=0
	// +optional
	OOMScoreAdj int `json:"oomScoreAdj,omitempty"`

	// Options passes extra command-line arguments to memStress.
	// When an option sets the same argument as a field, the option takes precedence.
	// +optional
	Options []string `json:"options,omitempty"`
}

// CPUStressor defines how to stress CPU out
type CPUStressor struct {
	Stressor `json:",inline"`
	// Load specifies P percent loading per CPU worker. 0 is effectively a sleep (no load) and 100
	// is full loading.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +optional
	Load *int `json:"load,omitempty"`

	// extend stress-ng options
	// +optional
	Options []string `json:"options,omitempty"`
}

func (obj *StressChaos) GetSelectorSpecs() map[string]any {
	return map[string]any{
		".": &obj.Spec.ContainerSelector,
	}
}

func (obj *StressChaos) GetCustomStatus() any {
	return &obj.Status.Instances
}
