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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// JVMChaosSpec defines the desired state of JVMChaos
type JVMChaosSpec struct {
	ContainerSelector `json:",inline"`

	// Duration represents the duration of the chaos action
	// +optional
	Duration *string `json:"duration,omitempty" webhook:"Duration"`

	// Action defines the specific jvm chaos action.
	// Supported action: latency;return;exception;stress;gc;ruleData
	// +kubebuilder:validation:Enum=latency;return;exception;stress;gc;ruleData;mysql
	Action JVMChaosAction `json:"action"`

	// JVMParameter represents the detail about jvm chaos action definition
	// +optional
	JVMParameter `json:",inline"`

	// RemoteCluster represents the remote cluster where the chaos will be deployed
	// +optional
	RemoteCluster string `json:"remoteCluster,omitempty"`
}

// JVMChaosAction represents the chaos action about jvm
type JVMChaosAction string

const (
	// JVMLatencyAction represents the JVM chaos action of invoke latency
	JVMLatencyAction JVMChaosAction = "latency"

	// JVMReturnAction represents the JVM chaos action of return value
	JVMReturnAction JVMChaosAction = "return"

	// JVMExceptionAction represents the JVM chaos action of throwing custom exceptions
	JVMExceptionAction JVMChaosAction = "exception"

	// JVMStressAction represents the JVM chaos action of stress like CPU and memory
	JVMStressAction JVMChaosAction = "stress"

	// JVMGCAction represents the JVM chaos action of trigger garbage collection
	JVMGCAction JVMChaosAction = "gc"

	// JVMRuleDataAction represents inject fault with byteman's rule
	// refer to https://downloads.jboss.org/byteman/4.0.14/byteman-programmers-guide.html#the-byteman-rule-language
	JVMRuleDataAction JVMChaosAction = "ruleData"

	// JVMMySQLAction represents the JVM chaos action of mysql java client fault injection
	JVMMySQLAction JVMChaosAction = "mysql"
)

// JVMParameter represents the detail about jvm chaos action definition
type JVMParameter struct {
	JVMCommonSpec `json:",inline"`

	JVMClassMethodSpec `json:",inline"`

	JVMStressCfgSpec `json:",inline"`

	JVMMySQLSpec `json:",inline"`

	// +optional
	// byteman rule name, should be unique, and will generate one if not set
	Name string `json:"name"`

	// +optional
	// the return value for action 'return'
	ReturnValue string `json:"value"`

	// +optional
	// the exception which needs to throw for action `exception`
	// or the exception message needs to throw in action `mysql`
	ThrowException string `json:"exception"`

	// +optional
	// the latency duration for action 'latency', unit ms
	// or the latency duration in action `mysql`
	LatencyDuration int `json:"latency"`

	// +optional
	// the byteman rule's data for action 'ruleData'
	RuleData string `json:"ruleData"`
}

// JVMCommonSpec is the common specification for JVMChaos
type JVMCommonSpec struct {
	// +optional
	// the port of agent server, default 9277
	Port int32 `json:"port,omitempty"`

	// the pid of Java process which needs to attach
	Pid int `json:"pid,omitempty"`
}

// JVMClassMethodSpec is the specification for class and method
type JVMClassMethodSpec struct {
	// +optional
	// Java class
	Class string `json:"class,omitempty"`

	// +optional
	// the method in Java class
	Method string `json:"method,omitempty"`
}

// JVMStressSpec is the specification for stress
type JVMStressCfgSpec struct {
	// +optional
	// the CPU core number needs to use, only set it when action is stress
	CPUCount int `json:"cpuCount,omitempty"`

	// +optional
	// the memory type needs to locate, only set it when action is stress, the value can be 'stack' or 'heap'
	MemoryType string `json:"memType,omitempty"`
}

// JVMMySQLSpec is the specification of MySQL fault injection in JVM
// only when SQL match the Database, Table and SQLType, JVMChaos mesh will inject fault
// for examle:
//
//	SQL is "select * from test.t1",
//	only when ((Database == "test" || Database == "") && (Table == "t1" || Table == "") && (SQLType == "select" || SQLType == "")) is true, JVMChaos will inject fault
type JVMMySQLSpec struct {
	// the version of mysql-connector-java, only support 5.X.X(set to "5") and 8.X.X(set to "8") now
	MySQLConnectorVersion string `json:"mysqlConnectorVersion,omitempty"`

	// the match database
	// default value is "", means match all database
	Database string `json:"database,omitempty"`

	// the match table
	// default value is "", means match all table
	Table string `json:"table,omitempty"`

	// the match sql type
	// default value is "", means match all SQL type.
	// The value can be 'select', 'insert', 'update', 'delete', 'replace'.
	SQLType string `json:"sqlType,omitempty"`
}

// JVMChaosStatus defines the observed state of JVMChaos
type JVMChaosStatus struct {
	ChaosStatus `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="action",type=string,JSONPath=`.spec.action`
// +kubebuilder:printcolumn:name="duration",type=string,JSONPath=`.spec.duration`
// +chaos-mesh:experiment

// JVMChaos is the Schema for the jvmchaos API
type JVMChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JVMChaosSpec   `json:"spec,omitempty"`
	Status JVMChaosStatus `json:"status,omitempty"`
}

var _ InnerObjectWithSelector = (*JVMChaos)(nil)
var _ InnerObject = (*JVMChaos)(nil)

func init() {
	SchemeBuilder.Register(&JVMChaos{}, &JVMChaosList{})
}

func (obj *JVMChaos) GetSelectorSpecs() map[string]interface{} {
	return map[string]interface{}{
		".": &obj.Spec.ContainerSelector,
	}
}
