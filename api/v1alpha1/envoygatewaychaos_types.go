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
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="route-kind",type=string,JSONPath=`.spec.target.kind`
// +kubebuilder:printcolumn:name="route",type=string,JSONPath=`.spec.target.route`
// +kubebuilder:printcolumn:name="duration",type=string,JSONPath=`.spec.duration`
// +chaos-mesh:experiment
// +genclient

// EnvoyGatewayChaos injects faults through an Envoy Gateway BackendTrafficPolicy.
type EnvoyGatewayChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EnvoyGatewayChaosSpec   `json:"spec"`
	Status EnvoyGatewayChaosStatus `json:"status,omitempty"`
}

var _ InnerObjectWithSelector = (*EnvoyGatewayChaos)(nil)
var _ InnerObject = (*EnvoyGatewayChaos)(nil)

// EnvoyGatewayChaosSpec defines a route target and its fault configuration.
type EnvoyGatewayChaosSpec struct {
	// Duration represents the duration of the chaos action.
	// +optional
	Duration *string `json:"duration,omitempty" webhook:"Duration"`

	// Target identifies the Gateway API route affected by the experiment.
	Target EnvoyGatewayTarget `json:"target"`

	// Fault contains the delay and abort configuration.
	Fault EnvoyGatewayFault `json:"fault"`

	// RemoteCluster represents the remote cluster where the chaos will be deployed.
	// +optional
	RemoteCluster string `json:"remoteCluster,omitempty"`
}

// EnvoyGatewayRouteKind identifies a supported Gateway API route kind.
type EnvoyGatewayRouteKind string

const (
	// EnvoyGatewayHTTPRoute targets an HTTPRoute.
	EnvoyGatewayHTTPRoute EnvoyGatewayRouteKind = "HTTPRoute"
	// EnvoyGatewayGRPCRoute targets a GRPCRoute.
	EnvoyGatewayGRPCRoute EnvoyGatewayRouteKind = "GRPCRoute"
)

// EnvoyGatewayTarget identifies one HTTPRoute or GRPCRoute.
type EnvoyGatewayTarget struct {
	// Namespace contains the target route.
	// +kubebuilder:validation:MinLength=1
	Namespace string `json:"namespace"`

	// Kind is the Gateway API route kind.
	// +kubebuilder:validation:Enum=HTTPRoute;GRPCRoute
	Kind EnvoyGatewayRouteKind `json:"kind"`

	// Route is the target route name.
	// +kubebuilder:validation:MinLength=1
	Route string `json:"route"`
}

// Id implements selector.Target.
func (target *EnvoyGatewayTarget) Id() string {
	return fmt.Sprintf("%s/%s/%s", target.Namespace, target.Kind, target.Route)
}

// EnvoyGatewayFault contains delay and abort faults. At least one fault must be set.
type EnvoyGatewayFault struct {
	// Delay injects latency before forwarding a request.
	// +optional
	Delay *EnvoyGatewayDelay `json:"delay,omitempty"`

	// Abort terminates a request with an HTTP or gRPC status code.
	// +optional
	Abort *EnvoyGatewayAbort `json:"abort,omitempty"`
}

// EnvoyGatewayDelay defines a fixed-delay fault.
type EnvoyGatewayDelay struct {
	// FixedDelay is the latency added to selected requests.
	// +kubebuilder:validation:Pattern="^([0-9]{1,5}(h|m|s|ms)){1,4}$"
	FixedDelay string `json:"fixedDelay" webhook:"Duration"`

	// Percentage is the percentage of requests to delay.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	Percentage int32 `json:"percentage"`
}

// EnvoyGatewayAbort defines an HTTP or gRPC abort fault.
type EnvoyGatewayAbort struct {
	// HTTPStatus is returned for an HTTPRoute fault.
	// +optional
	// +kubebuilder:validation:Minimum=200
	// +kubebuilder:validation:Maximum=600
	HTTPStatus *int32 `json:"httpStatus,omitempty"`

	// GRPCStatus is returned for a GRPCRoute fault.
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=16
	GRPCStatus *int32 `json:"grpcStatus,omitempty"`

	// Percentage is the percentage of requests to abort.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	Percentage int32 `json:"percentage"`
}

// EnvoyGatewayChaosStatus represents the status of an EnvoyGatewayChaos.
type EnvoyGatewayChaosStatus struct {
	ChaosStatus `json:",inline"`
}

// GetSelectorSpecs returns the route target selected by this experiment.
func (obj *EnvoyGatewayChaos) GetSelectorSpecs() map[string]interface{} {
	return map[string]interface{}{
		".": &obj.Spec.Target,
	}
}
