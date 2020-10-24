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

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +chaos-mesh:base

// KernelChaos is the Schema for the kernelchaos API
type KernelChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a kernel chaos experiment
	Spec KernelChaosSpec `json:"spec"`

	// +optional
	// Most recently observed status of the kernel chaos experiment
	Status KernelChaosStatus `json:"status"`
}

// KernelChaosSpec defines the desired state of KernelChaos
type KernelChaosSpec struct {
	// Mode defines the mode to run chaos action.
	// Supported mode: one / all / fixed / fixed-percent / random-max-percent
	// +kubebuilder:validation:Enum=one;all;fixed;fixed-percent;random-max-percent
	Mode PodMode `json:"mode"`

	// Value is required when the mode is set to `FixedPodMode` / `FixedPercentPodMod` / `RandomMaxPercentPodMod`.
	// If `FixedPodMode`, provide an integer of pods to do chaos action.
	// If `FixedPercentPodMod`, provide a number from 0-100 to specify the percent of pods the server can do chaos action.
	// If `RandomMaxPercentPodMod`,  provide a number from 0-100 to specify the max percent of pods to do chaos action
	// +optional
	Value string `json:"value"`

	// Selector is used to select pods that are used to inject chaos action.
	Selector SelectorSpec `json:"selector"`

	// FailKernRequest defines the request of kernel injection
	FailKernRequest FailKernRequest `json:"failKernRequest"`

	// Duration represents the duration of the chaos action
	Duration *string `json:"duration,omitempty"`

	// Scheduler defines some schedule rules to control the running time of the chaos experiment about time.
	Scheduler *SchedulerSpec `json:"scheduler,omitempty"`
}

// GetSelector is a getter for Selector (for implementing SelectSpec)
func (in *KernelChaosSpec) GetSelector() SelectorSpec {
	return in.Selector
}

// GetMode is a getter for Mode (for implementing SelectSpec)
func (in *KernelChaosSpec) GetMode() PodMode {
	return in.Mode
}

// GetValue is a getter for Value (for implementing SelectSpec)
func (in *KernelChaosSpec) GetValue() string {
	return in.Value
}

// FailKernRequest defines the injection conditions
type FailKernRequest struct {
	// FailType indicates what to fail, can be set to '0' / '1' / '2'
	// If `0`, indicates slab to fail (should_failslab)
	// If `1`, indicates alloc_page to fail (should_fail_alloc_page)
	// If `2`, indicates bio to fail (should_fail_bio)
	// You can read:
	//   1. https://www.kernel.org/doc/html/latest/fault-injection/fault-injection.html
	//   2. http://github.com/iovisor/bcc/blob/master/tools/inject_example.txt
	// to learn more
	// +kubebuilder:validation:Maximum=2
	// +kubebuilder:validation:Minimum=0
	FailType int32 `json:"failtype"`

	// Headers indicates the appropriate kernel headers you need.
	// Eg: "linux/mmzone.h", "linux/blkdev.h" and so on
	Headers []string `json:"headers,omitempty"`

	// Callchain indicate a special call chain, such as:
	//     ext4_mount
	//       -> mount_subtree
	//          -> ...
	//             -> should_failslab
	// With an optional set of predicates and an optional set of
	// parameters, which used with predicates. You can read call chan
	// and predicate examples from https://github.com/chaos-mesh/bpfki/tree/develop/examples
	// to learn more.
	// If no special call chain, just keep Callchain empty, which means it will fail at any call chain
	// with slab alloc (eg: kmalloc).
	Callchain []Frame `json:"callchain,omitempty"`

	// Probability indicates the fails with probability.
	// If you want 1%, please set this field with 1.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	Probability uint32 `json:"probability,omitempty"`

	// Times indicates the max times of fails.
	// +kubebuilder:validation:Minimum=0
	Times uint32 `json:"times,omitempty"`
}

// Frame defines the function signature and predicate in function's body
type Frame struct {
	// Funcname can be find from kernel source or `/proc/kallsyms`, such as `ext4_mount`
	Funcname string `json:"funcname,omitempty"`

	// Parameters is used with predicate, for example, if you want to inject slab error
	// in `d_alloc_parallel(struct dentry *parent, const struct qstr *name)` with a special
	// name `bananas`, you need to set it to `struct dentry *parent, const struct qstr *name`
	// otherwise omit it.
	Parameters string `json:"parameters,omitempty"`

	// Predicate will access the arguments of this Frame, example with Parameters's, you can
	// set it to `STRNCMP(name->name, "bananas", 8)` to make inject only with it, or omit it
	// to inject for all d_alloc_parallel call chain.
	Predicate string `json:"predicate,omitempty"`
}

// KernelChaosStatus defines the observed state of KernelChaos
type KernelChaosStatus struct {
	ChaosStatus `json:",inline"`
}
