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
	ContainerSelector `json:",inline"`

	// Duration represents the duration of the chaos action
	// +optional
	Duration *string `json:"duration,omitempty"`

	// Action defines the specific jvm chaos action.
	// Supported action: delay;return;script;cfl;oom;ccf;tce;cpf;tde;tpf
	// +kubebuilder:validation:Enum=delay;return;script;cfl;oom;ccf;tce;cpf;tde;tpf
	Action JVMChaosAction `json:"action"`

	// JVMParameter represents the detail about jvm chaos action definition
	// +optional
	JVMParameter `json:",inline"`

	// Target defines the specific jvm chaos target.
	// Supported target: servlet;psql;jvm;jedis;http;dubbo;rocketmq;tars;mysql;druid;redisson;rabbitmq;mongodb
	// +kubebuilder:validation:Enum=servlet;psql;jvm;jedis;http;dubbo;rocketmq;tars;mysql;druid;redisson;rabbitmq;mongodb
	Target JVMChaosTarget `json:"target"`
}

type JVMChaosTarget string

const (
	// SERVLET represents servlet as a target of chaos
	SERVLET JVMChaosTarget = "servlet"

	// PSQL represents Postgresql JDBC as a target of chaos
	PSQL JVMChaosTarget = "psql"

	// JVM represents JVM as a target of chaos
	JVM JVMChaosTarget = "jvm"

	// JEDIS represents jedis (a java redis client) as a target of chaos
	JEDIS JVMChaosTarget = "jedis"

	// HTTP represents http client as a target of chaos
	HTTP JVMChaosTarget = "http"

	// DUBBO represents a Dubbo services as a target of chaos
	DUBBO JVMChaosTarget = "dubbo"

	// ROCKETMQ represents Rocketmq java client as a target of chaos
	ROCKETMQ JVMChaosTarget = "rocketmq"

	// MYSQL represents Mysql JDBC as a target of chaos
	MYSQL JVMChaosTarget = "mysql"

	// DRUID represents the Druid database connection pool as a target of chaos
	DRUID JVMChaosTarget = "druid"

	// TARS represents the Tars service as a target of chaos
	TARS JVMChaosTarget = "tars"

	// REDISSON represents Redisson (a java redis client) as a target of chaos
	REDISSON JVMChaosTarget = "redisson"

	// RABBITMQ represents the Rabbitmq java client as a target of chaos
	RABBITMQ JVMChaosTarget = "rabbitmq"

	// MONGODB represents the Mongodb java client as a target of chaos
	MONGODB JVMChaosTarget = "mongodb"
)

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

	// JVMConnectionPoolFullAction represents the JVM chaos action of Connection Pool Full
	JVMConnectionPoolFullAction JVMChaosAction = "cpf"

	// JVMThrowDeclaredExceptionAction represents the JVM chaos action of throwing declared exception
	JVMThrowDeclaredExceptionAction JVMChaosAction = "tde"

	// JVMThreadPoolFullAction represents the JVM chaos action of thread pool full
	JVMThreadPoolFullAction JVMChaosAction = "tpf"
)

// JVMParameter represents the detail about jvm chaos action definition
type JVMParameter struct {

	// Flags represents the flags of action
	// +optional
	Flags map[string]string `json:"flags,omitempty"`

	// Matchers represents the matching rules for the target
	// +optional
	Matchers map[string]string `json:"matchers,omitempty"`
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

func (obj *JVMChaos) GetSelectorSpecs() map[string]interface{} {
	return map[string]interface{}{
		".": &obj.Spec.PodSelector,
	}
}
