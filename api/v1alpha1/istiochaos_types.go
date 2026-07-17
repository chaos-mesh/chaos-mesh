// Copyright 2026 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="virtual-service",type=string,JSONPath=`.spec.target.virtualService`
// +kubebuilder:printcolumn:name="route",type=string,JSONPath=`.spec.target.httpRoute`
// +kubebuilder:printcolumn:name="duration",type=string,JSONPath=`.spec.duration`
// +chaos-mesh:experiment
// +genclient

// IstioChaos is the Schema for the istiochaos API.
type IstioChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IstioChaosSpec   `json:"spec"`
	Status IstioChaosStatus `json:"status,omitempty"`
}

var _ InnerObjectWithSelector = (*IstioChaos)(nil)
var _ InnerObject = (*IstioChaos)(nil)

// IstioChaosSpec defines an Istio fault injected through a VirtualService.
type IstioChaosSpec struct {
	// Duration represents the duration of the chaos action.
	// +optional
	Duration *string `json:"duration,omitempty" webhook:"Duration"`

	// Target identifies the VirtualService HTTP route to clone and fault.
	Target IstioTarget `json:"target"`

	// Fault contains the Istio fault configuration.
	Fault IstioFault `json:"fault"`

	// RemoteCluster represents the remote cluster where the chaos will be deployed.
	// +optional
	RemoteCluster string `json:"remoteCluster,omitempty"`
}

// IstioTarget identifies one named HTTP route in a VirtualService.
type IstioTarget struct {
	// Namespace is the namespace containing the VirtualService.
	// +kubebuilder:validation:MinLength=1
	Namespace string `json:"namespace"`

	// VirtualService is the name of the target VirtualService.
	// +kubebuilder:validation:MinLength=1
	VirtualService string `json:"virtualService"`

	// HTTPRoute is the name of the target entry in spec.http.
	// +kubebuilder:validation:MinLength=1
	HTTPRoute string `json:"httpRoute"`
}

// Id implements selector.Target.
func (target *IstioTarget) Id() string {
	return types.NamespacedName{
		Namespace: target.Namespace,
		Name:      target.VirtualService,
	}.String()
}

// IstioFault contains delay and abort faults. At least one fault must be set.
type IstioFault struct {
	// Delay injects latency before forwarding a request.
	// +optional
	Delay *IstioDelay `json:"delay,omitempty"`

	// Abort terminates a request with an HTTP status code.
	// +optional
	Abort *IstioAbort `json:"abort,omitempty"`
}

// IstioDelay defines an Istio fixed-delay fault.
type IstioDelay struct {
	// FixedDelay is the latency added to selected requests.
	FixedDelay string `json:"fixedDelay" webhook:"Duration"`

	// Percentage is the percentage of requests to delay.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	Percentage int32 `json:"percentage"`
}

// IstioAbort defines an Istio HTTP-abort fault.
type IstioAbort struct {
	// HTTPStatus is the status code returned to selected requests.
	// +kubebuilder:validation:Minimum=100
	// +kubebuilder:validation:Maximum=599
	HTTPStatus int32 `json:"httpStatus"`

	// Percentage is the percentage of requests to abort.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	Percentage int32 `json:"percentage"`
}

// IstioChaosStatus represents the status of an IstioChaos.
type IstioChaosStatus struct {
	ChaosStatus `json:",inline"`
}

// GetSelectorSpecs returns the VirtualService target selected by this experiment.
func (obj *IstioChaos) GetSelectorSpecs() map[string]interface{} {
	return map[string]interface{}{
		".": &obj.Spec.Target,
	}
}
