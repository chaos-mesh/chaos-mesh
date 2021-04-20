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

// PodHttpChaosSpec defines the desired state of IoChaos
type PodHttpChaosSpec struct {
	// ProxyPorts represents the target ports to be proxy of.
	ProxyPorts []int32 `json:"proxyPorts"`

	// TODO: support multiple different container to inject in one pod
	// +optional
	Container *string `json:"container,omitempty"`

	// Pid represents a running toda process id
	// +optional
	Pid int64 `json:"pid,omitempty"`

	// StartTime represents the start time of a toda process
	// +optional
	StartTime int64 `json:"startTime,omitempty"`

	// Rules are a list of injection rule for http request
	// +optional
	Rules []*PodHttpChaosRule `json:"requestRules,omitempty"`
}

// PodHttpChaosRule defines the injection rule for http request
type PodHttpChaosRule struct {
	Target   PodHttpChaosTarget   `json:"target"`
	Selector PodHttpChaosSelector `json:"selector"`
	Actions  PodHttpChaosActions  `json:"actions"`
}

type PodHttpChaosSelector struct {
	// +optional
	Port *int32 `json:"port,omitempty"`

	// +optional
	Path *string `json:"path,omitempty"`

	// +optional
	Method *string `json:"method,omitempty"`

	// +optional
	Code *int32 `json:"code,omitempty"`

	// +optional
	RequestHeaders map[string]string `json:"request_headers,omitempty"`

	// +optional
	ResponseHeaders map[string]string `json:"response_headers,omitempty"`
}

// HttpChaosAction defines possible actions of HttpChaos
type PodHttpChaosActions struct {
	// +optional
	Abort *bool `json:"abort,omitempty"`

	// Delay represents the delay of the target request/response.
	// A duration string is a possibly signed sequence of
	// decimal numbers, each with optional fraction and a unit suffix,
	// such as "300ms", "-1.5h" or "2h45m".
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	// +optional
	Delay *string `json:"delay,omitempty"`

	// +optional
	Append *PodHttpChaosAppendActions `json:"append,omitempty"`

	// +optional
	Replace *PodHttpChaosReplaceActions `json:"replace,omitempty"`
}

// PodHttpChaosAppendActions defines possible append-actions of HttpChaos
type PodHttpChaosAppendActions struct {
	// +optional
	Queries [][]string `json:"queries,omitempty"`

	// +optional
	Headers [][]string `json:"headers,omitempty"`
}

// PodHttpChaosReplaceActions defines possible replace-actions of HttpChaos
type PodHttpChaosReplaceActions struct {
	// +optional
	Path *string `json:"path,omitempty"`

	// +optional
	Method *string `json:"method,omitempty"`

	// +optional
	Code *int32 `json:"code,omitempty"`

	// +optional
	Body []byte `json:"body,omitempty"`

	// +optional
	Queries map[string]string `json:"queries,omitempty"`

	// +optional
	Headers map[string]string `json:"headers,omitempty"`
}

// PodHttpChaosTarget represents the type of an HttpChaos Action
type PodHttpChaosTarget string

const (
	// PodHttpRequest represents injecting chaos for http request
	PodHttpRequest PodHttpChaosTarget = "Request"

	// PodHttpResponse represents injecting chaos for http response
	PodHttpResponse PodHttpChaosTarget = "Response"
)

const KindPodHttpChaos = "PodHttpChaos"

// +kubebuilder:object:root=true

// PodHttpChaos is the Schema for the podiochaos API
type PodHttpChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PodHttpChaosSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// PodHttpChaosList contains a list of PodHttpChaos
type PodHttpChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PodHttpChaos `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PodHttpChaos{}, &PodHttpChaosList{})
}
