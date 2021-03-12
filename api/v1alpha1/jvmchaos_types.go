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

// JVMChaosSpec defines the desired state of JVMChaos
type JVMChaosSpec struct {
	// Mode defines the mode to run chaos action.
	// Supported mode: one / all / fixed / fixed-percent / random-max-percent
	Mode PodMode `json:"mode"`

	// Value is required when the mode is set to `FixedPodMode` / `FixedPercentPodMod` / `RandomMaxPercentPodMod`.
	// If `FixedPodMode`, provide an integer of pods to do chaos action.
	// If `FixedPercentPodMod`, provide a number from 0-100 to specify the max % of pods the server can do chaos action.
	// If `RandomMaxPercentPodMod`,  provide a number from 0-100 to specify the % of pods to do chaos action
	// +optional
	Value string `json:"value"`

	// Selector is used to select pods that are used to inject chaos action.
	Selector SelectorSpec `json:"selector"`

	// Duration represents the duration of the chaos action
	// +optional
	Duration *string `json:"duration,omitempty"`

	// Scheduler defines some schedule rules to control the running time of the chaos experiment about time.
	// +optional
	Scheduler *SchedulerSpec `json:"scheduler,omitempty"`

	// Action defines the specific jvm chaos action.
	// Supported action: delay, return, script, cfl, oom, ccf, tce, delay4servlet, tce4servlet
	// +kubebuilder:validation:Enum=delay;return;script;cfl;oom;ccf;tce;delay4servlet;tce4servlet
	Action JVMChaosAction `json:"action"`

	// JVMParameter represents the detail about jvm chaos action definition
	// +optional
	JVMParameter `json:",inline"`
}

// GetSelector is a getter for Selector (for implementing SelectSpec)
func (in *JVMChaosSpec) GetSelector() SelectorSpec {
	return in.Selector
}

// GetMode is a getter for Mode (for implementing SelectSpec)
func (in *JVMChaosSpec) GetMode() PodMode {
	return in.Mode
}

// GetValue is a getter for Value (for implementing SelectSpec)
func (in *JVMChaosSpec) GetValue() string {
	return in.Value
}

// JVMChaosAction represents the chaos action about jvm
type JVMChaosAction string

const (
	// JVMDelayAction represents the JVM chaos action of invoke delay
	JVMDelayAction JVMChaosAction = "delay"

	// JVMReturnAction represents the JVM chaos action of return value
	JVMReturnAction JVMChaosAction = "return"

	// JVMReturnAction represents the JVM chaos action for complex failure scenarios.
	// Write Java or Groovy scripts, such as tampering with parameters, modifying return values,
	// throwing custom exceptions, and so on
	JVMScriptAction JVMChaosAction = "script"

	// JVMCpuFullloadAction represents the JVM chaos action of CPU is full
	JVMCpuFullloadAction JVMChaosAction = "cfl"

	// JVMOOMAction represents the JVM chaos action of OOM exception
	JVMOOMAction JVMChaosAction = "oom"

	// JVMCodeCacheFillingAction represents the JVM chaos action of code cache filling
	JVMCodeCacheFillingAction JVMChaosAction = "ccf"

	// JVMExceptionAction represents the JVM chaos action of throwing custom exceptions
	JVMExceptionAction JVMChaosAction = "tce"

	// ServletDelayAction represents the JVM chaos action of Servlet response delay
	ServletDelayAction JVMChaosAction = "delay4servlet"

	// ServletExceptionAction represents the JVM chaos action of Servlet throwing custom exceptions
	ServletExceptionAction JVMChaosAction = "tce4servlet"
)

// JVMParameter represents the detail about jvm chaos action definition
type JVMParameter struct {
	// EffectCount represents the number of affect
	// +optional
	EffectCount int `json:"effectcount"`

	// EffectPercent represents the percent of affect
	// +optional
	EffectPercent int `json:"effectpercent"`

	// Delay represents the detail about JVM chaos action of invoke delay
	// +optional
	Delay *JVMDelaySpec `json:"delay,omitempty"`

	// Return represents the detail about JVM chaos action of return value
	// +optional
	Return *JVMReturnSpec `json:"return,omitempty"`

	// Script represents the detail about JVM chaos action of Java or Groovy scripts
	// +optional
	Script *JVMScriptSpec `json:"script,omitempty"`

	// CpuFullload represents the detail about JVM chaos action of CPU is full
	// +optional
	CpuFullload *JVMCpufullloadSpec `json:"cfl,omitempty"`

	// OOM represents the detail about JVM chaos action of OOM exception
	// +optional
	OOM *JVMOOMSpec `json:"oom,omitempty"`

	// Exception represents the detail about JVM chaos action of throwing custom exceptions
	// +optional
	Exception *JVMExceptionSpec `json:"tce,omitempty"`

	// Delay4Servlet represents the detail about JVM chaos action of Servlet response delay
	// +optional
	Delay4Servlet *ServletDelaySpec `json:"delay4servlet,omitempty"`

	// Exception4Servlet represents the detail about JVM chaos action of Servlet throwing custom exceptions
	// +optional
	Exception4Servlet *ServletExceptionSpec `json:"tce4servlet,omitempty"`
}

// JVMExceptionSpec represents the detail about JVM chaos action of throwing custom exceptions
type JVMExceptionSpec struct {
	// JVMCommonParameter represents the common jvm chaos parameter
	JVMCommonParameter `json:",inline"`

	// Exception represents the Exception class, with the full package name, must inherit from java.lang.Exception or Java.lang.Exception itself
	Exception string `json:"exception"`

	// Message represents specifies the exception class information.
	// +optional
	Message string `json:"message"`
}

// JVMOOMSpec represents the detail about JVM chaos action of OOM exception
type JVMOOMSpec struct {
	// JVMCommonParameter represents the common jvm chaos parameter
	JVMCommonParameter `json:",inline"`

	// Area represents JVM memory area, currently supported [HEAP, NOHEAP, OFFHEAP], required.
	// Eden+Old is denoted by HEAP
	// Metaspace is denoted by NOHEAP
	// off-heap memory is denoted by OFFHEAP
	// +kubebuilder:validation:Enum=HEAP;NOHEAP;OFFHEAP
	Area string `json:"area"`

	// Block represents specifies the size of the object that supports only the HEAP and OFFHEAP areas in MB
	// +optional
	Block string `json:"block"`

	// Interval represents unit MS, default interval between 500 OOM exceptions, only in non-violent mode, can slow down the frequency of GC without worrying about the process being unresponsive
	// +optional
	Interval int `json:"interval"`

	// WildMode represents default false, whether to turn on wild mode or not.
	// If it is wild mode, the memory created before will not be released after OOM occurrence, which may cause the application process to be unresponsive
	// +optional
	WildMode bool `json:"wildmode"`
}

// JVMCpufullloadSpec represents the detail about JVM chaos action of CPU is full
type JVMCpufullloadSpec struct {
	// JVMCommonParameter represents the common jvm chaos parameter
	JVMCommonParameter `json:",inline"`

	// CpuCount represents the number of CPU cores to bind to, that is, specify how many cores are full
	CpuCount int `json:"cpucount"`
}

// JVMScriptSpec represents the detail about JVM chaos action of Java or Groovy scripts
type JVMScriptSpec struct {
	// JVMCommonParameter represents the common jvm chaos parameter
	JVMCommonParameter `json:",inline"`

	// Content represents the script content is Base64 encoded content.
	// Note that it cannot be used with file
	// +optional
	Content string `json:"content"`

	// File represents script file, absolute path to file
	// Note that it cannot be used with content
	// +optional
	File string `json:"file"`

	// Name represents script name, use for logging
	// +optional
	Name string `json:"name"`

	// Type represents script type, java or groovy, default to java
	// +kubebuilder:validation:Enum=java;groovy;
	// +optional
	Type string `json:"type"`
}

// JVMReturnSpec represents the detail about JVM chaos action of return value
type JVMReturnSpec struct {
	// JVMCommonParameter represents the common jvm chaos parameter
	JVMCommonParameter `json:",inline"`

	// Value represents specifies the return value of a class method, supporting only primitive, null, and String return values. required
	Value string `json:"value"`
}

// JVMDelaySpec represents the detail about JVM chaos action of invoke delay
type JVMDelaySpec struct {
	// JVMCommonParameter represents the common jvm chaos parameter
	JVMCommonParameter `json:",inline"`

	// Time represents delay time, in milliseconds, required
	Time int `json:"time"`

	// Offset represents delay fluctuation time
	// +optional
	Offset int `json:"offset"`
}

// JVMCommonParameter represents the common jvm chaos parameter
type JVMCommonParameter struct {
	// Classname represents specify the class name, which must be an implementation class with a full package name, such as com.xxx.xxx.XController. required
	Classname string `json:"classname"`

	// Methodname represents specify the method name. Note that methods with the same method name will be injected with the same fault. required
	Methodname string `json:"methodname"`

	// After represents method execution is completed before the injection failure is returned.
	// +optional
	After bool `json:"after"`
}

// ServletExceptionSpec represents the detail about JVM chaos action of Servlet throwing custom exceptions
type ServletExceptionSpec struct {
	// ServletCommonParameter represents the common servlet chaos parameter
	ServletCommonParameter `json:",inline"`

	// Exception represents the Exception class, with the full package name, must inherit from java.lang.Exception or Java.lang.Exception itself
	Exception string `json:"exception"`

	// Message represents specifies the exception class information.
	// +optional
	Message string `json:"message"`
}

// ServletDelaySpec represents the detail about JVM chaos action of Servlet response delay
type ServletDelaySpec struct {
	// ServletCommonParameter represents the common servlet chaos parameter
	ServletCommonParameter `json:",inline"`

	// Time represents delay time, in milliseconds, required
	Time int `json:"time"`

	// Offset represents delay fluctuation time
	// +optional
	Offset int `json:"offset"`
}

// ServletCommonParameter represents the common servlet chaos parameter
type ServletCommonParameter struct {
	// Method represents HTTP request method, such as GET, POST, DELETE or PUT. Default is GET
	// +kubebuilder:validation:Enum=GET;POST;PUT;DELETE
	// +optional
	Method string `json:"method"`

	// QueryString represents HTTP request query string
	// +optional
	QueryString string `json:"querystring"`

	// RequestPath represents HTTP request path. The path should start with /
	RequestPath string `json:"requestpath"`
}

// JVMChaosStatus defines the observed state of JVMChaos
type JVMChaosStatus struct {
	ChaosStatus `json:",inline"`
}

// +kubebuilder:object:root=true
// +chaos-mesh:base

// JVMChaos is the Schema for the jvmchaos API
type JVMChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JVMChaosSpec   `json:"spec,omitempty"`
	Status JVMChaosStatus `json:"status,omitempty"`
}

func init() {
	SchemeBuilder.Register(&JVMChaos{}, &JVMChaosList{})
}
