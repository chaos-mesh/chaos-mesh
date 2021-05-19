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

// HTTPChaos is the Schema for the HTTPchaos API
type HTTPChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HTTPChaosSpec   `json:"spec,omitempty"`
	Status HTTPChaosStatus `json:"status,omitempty"`
}

type HTTPChaosSpec struct {
	PodSelector `json:",inline"`

	// Target is the object to be selected and injected, <Request|Response>.
	Target PodHttpChaosTarget `json:"target"`

	PodHttpChaosSelector `json:",inline"`

	PodHttpChaosActions `json:",inline"`

	// Duration represents the duration of the chaos action.
	// +optional
	Duration *string `json:"duration,omitempty"`
}

type HTTPChaosStatus struct {
	ChaosStatus `json:",inline"`

	// Instances always specifies podhttpchaos generation or empty
	// +optional
	Instances map[string]int64 `json:"instances,omitempty"`
}

func (obj *HTTPChaos) GetSelectorSpecs() map[string]interface{} {
	return map[string]interface{}{
		".": obj.Spec.PodSelector,
	}
}

func (obj *HTTPChaos) GetCustomStatus() interface{} {
	return &obj.Status.Instances
}
