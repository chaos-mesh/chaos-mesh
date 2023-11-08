// Copyright 2023 Chaos Mesh Authors.
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
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="duration",type=string,JSONPath=`.spec.duration`
// +chaos-mesh:experiment

// K8SChaos is the Schema for the Kubernetes Chaos API
type K8SChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a pod chaos experiment
	Spec K8SChaosSpec `json:"spec"`

	// +optional
	// Most recently observed status of the chaos experiment about pods
	Status K8SChaosStatus `json:"status,omitempty"`
}

var (
	_ InnerObjectWithSelector = (*K8SChaos)(nil)
	_ InnerObject             = (*K8SChaos)(nil)
)

// K8SChaosSpec defines the desired state of K8SChaos
type K8SChaosSpec struct {
	// Duration represents the duration of the chaos action
	// +optional
	Duration *string `json:"duration,omitempty" webhook:"Duration"`

	// +kubebuilder:validation:Required
	APIObjects *K8SChaosAPIObjects `json:"apiObjects"`

	// AllowPatching specifies that the chaos should patch & restore the modified object,
	// rather than create & delete.
	// +optional
	AllowPatching bool `json:"allowPatching,omitempty"`

	// RemoteCluster represents the remote cluster where the chaos will be deployed
	// +optional
	RemoteCluster string `json:"remoteCluster,omitempty"`
}

func (spec *K8SChaosSpec) Validate() error {
	if spec.APIObjects == nil || spec.APIObjects.Value == "" {
		return errors.New("APIObjects.Value must be set")
	}
	return nil
}

// K8SChaosAPIObjects defines ways to specify Kubernetes API resources to be applied by K8SChaos
type K8SChaosAPIObjects struct {
	// Literal string containing Kubernetes API resources
	Value string `json:"value"`
}

// K8SChaosStatus defines the observed state of K8SChaos
type K8SChaosStatus struct {
	ChaosStatus `json:",inline"`
}

func (obj *K8SChaos) GetSelectorSpecs() map[string]interface{} {
	return map[string]interface{}{
		".": obj.Spec.APIObjects,
	}
}
